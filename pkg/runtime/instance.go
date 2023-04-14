package runtime

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"cloud.google.com/go/run/apiv2/runpb"
	"github.com/angelini/sblocks/internal/maps"
	cr "github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/angelini/sblocks/pkg/core"
)

type RevisionInstance struct {
	definition *core.Revision
	labels     map[string]string
	state      core.RevisionState
}

type ServiceInstance struct {
	name      string
	state     core.ServiceState
	uri       string
	traffic   *core.TrafficStatus
	revisions []*RevisionInstance
}

type RuntimeInstance struct {
	definition Runtime

	mutex            sync.RWMutex
	freeServices     []*ServiceInstance
	assignedServices map[string]*ServiceInstance
}

func NewRuntimeInstance(definition Runtime) *RuntimeInstance {
	return &RuntimeInstance{
		definition:       definition,
		freeServices:     nil,
		assignedServices: make(map[string]*ServiceInstance),
	}
}

func (r *RuntimeInstance) Refresh(ctx context.Context, client *cr.Client, allServices []*runpb.Service, allRevisions []*runpb.Revision) error {
	var free []*ServiceInstance
	assigned := make(map[string]*ServiceInstance)
	revisions := make(map[string][]*RevisionInstance)

	for _, revision := range allRevisions {
		runtimeLabel, found := revision.Labels["sb_runtime"]
		if !found || runtimeLabel != r.definition.Name {
			continue
		}

		serviceName := core.ParseServiceName(revision.Name)

		percentage := 0
		// FIXME
		// if idx == 0 && service.traffic.latest {
		// 	percentage = 100
		// } else {
		// 	percentage = int(service.traffic.revisions[revision.Name])
		// }

		definition := core.RevisionDefinition(revision)
		revisions[serviceName] = append(revisions[serviceName], &RevisionInstance{
			definition: &definition,
			labels:     revision.Labels,
			state:      core.GetRevisionState(revision, percentage),
		})
	}

	for _, service := range allServices {
		runtimeLabel, found := service.Labels["sb_runtime"]
		if !found || runtimeLabel != r.definition.Name {
			continue
		}

		serviceName := core.ParseServiceName(service.Name)
		instance := &ServiceInstance{
			name:      serviceName,
			state:     core.GetServiceState(service),
			uri:       service.Uri,
			traffic:   core.NewTrafficStatus(service.TrafficStatuses),
			revisions: revisions[serviceName],
		}

		assignmentLabel, found := service.Labels["sb_assigment"]
		if found {
			free = append(free, instance)
		} else {
			assigned[assignmentLabel] = instance
		}
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.freeServices = free
	r.assignedServices = assigned

	return nil
}

func (r *RuntimeInstance) Converge() []Operation {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var operations []Operation

	for i := 0; i < (r.definition.FreeSize - len(r.freeServices)); i++ {
		operations = append(operations, &CreateFreeServiceOp{
			runtimeName: r.definition.Name,
			public:      r.definition.Public,
			labels:      r.definition.Labels,
			revision:    r.definition.Revision,
		})
	}

	for i := 0; i < (len(r.freeServices) - r.definition.FreeSize); i++ {
		operations = append(operations, &DeleteFreeServiceOp{
			runtimeName: r.definition.Name,
			serviceName: r.freeServices[len(r.freeServices)-(i+1)].name,
		})
	}

	return operations
}

func (r *RuntimeInstance) Display() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	results := []string{
		fmt.Sprintf("%s [%s]:", r.definition.Name, formatLabels(r.definition.Labels)),
	}

	for _, service := range r.freeServices {
		results = append(results, fmt.Sprintf("  > %s: %s", service.name, service.state.String()))
		results = append(results, fmt.Sprintf("    uri: %s", service.uri))
		for _, revision := range service.revisions {
			results = append(results, fmt.Sprintf(
				"    - %s[%s]: %s",
				strings.TrimPrefix(revision.definition.Name, service.name+"-"),
				formatLabels(revision.labels),
				revision.state.String(),
			))
		}
	}

	return results
}

func formatLabels(labels map[string]string) string {
	entries := make([]string, 0, len(labels))
	for _, key := range maps.SortedKeys(labels) {
		entries = append(entries, fmt.Sprintf("%s=%s", key, labels[key]))
	}

	return strings.Join(entries, ", ")
}
