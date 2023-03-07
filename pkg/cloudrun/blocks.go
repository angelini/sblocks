package cloudrun

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/angelini/sblocks/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type RevisionInstance struct {
	definition *Revision
	state      RevisionState
}

type ServiceInstance struct {
	name      string
	state     ServiceState
	revisions map[string]*RevisionInstance
}

type ServiceBlock struct {
	name     string
	labels   map[string]string
	services map[string]*ServiceInstance
}

func NewServiceBlock(rootName string, size int, labels map[string]string) *ServiceBlock {
	services := make(map[string]*ServiceInstance, size)
	name := fmt.Sprintf("%s-%s", rootName, randomString(6))

	for i := 0; i < size; i++ {
		name := fmt.Sprintf("%s-%d", name, i)
		services[name] = &ServiceInstance{
			name:      name,
			state:     NewServiceState(),
			revisions: make(map[string]*RevisionInstance),
		}
	}

	return &ServiceBlock{
		name,
		labels,
		services,
	}
}

func LoadServiceBlock(ctx context.Context, client *CloudRunClient, name string) (*ServiceBlock, error) {
	services, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	group, ctx := errgroup.WithContext(ctx)

	var labels map[string]string
	instances := make(map[string]*ServiceInstance)

	for _, service := range services {
		if !strings.HasPrefix(service.Name, fmt.Sprintf("%s/services/%s-", client.Parent, name)) {
			continue
		}

		if labels == nil {
			labels = service.Labels
		}

		instances[service.Name] = &ServiceInstance{
			name:  service.Name,
			state: GetServiceState(service),
		}

		log.Info(ctx, "instances", zap.Any("i", instances))
	}

	for _, instance := range instances {
		instance := instance
		group.Go(func() error {
			name := strings.TrimPrefix(instance.name, fmt.Sprintf("%s/services/", client.Parent))
			client.GetIAM(ctx, name)

			revisions, err := client.ListRevisions(ctx, name)
			if err != nil {
				return err
			}

			revisionInstances := make(map[string]*RevisionInstance, len(revisions))
			for _, revision := range revisions {
				definition := RevisionDefinition(revision)
				revisionInstances[revision.Name] = &RevisionInstance{
					definition: &definition,
					state:      GetRevisionState(revision),
				}
			}

			log.Info(ctx, "revisions", zap.Any("r", revisionInstances))

			instance.revisions = revisionInstances
			return nil
		})
	}

	err = group.Wait()
	if err != nil {
		return nil, err
	}

	return &ServiceBlock{
		name:     name,
		labels:   labels,
		services: instances,
	}, nil
}

func (sb *ServiceBlock) Create(ctx context.Context, client *CloudRunClient, revision *Revision) error {
	group, ctx := errgroup.WithContext(ctx)

	// for _, service := range sb.services {
	// 	service.revisions[revision.Name] = &RevisionInstance{
	// 		definition: revision,
	// 		state:      RevisionMissing,
	// 	}
	// }

	for _, instance := range sb.services {
		instance := instance
		group.Go(func() error {
			// instance.revisions[revision.Name].state = Starting

			service, err := client.Create(ctx, instance.name, sb.labels, revision)
			if err != nil {
				instance.state = ErrorServiceState()
				// instance.revisions[revision.Name].state = RevisionMissing
				return err
			}

			instance.state = GetServiceState(service)
			return nil
		})
	}

	return group.Wait()
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
