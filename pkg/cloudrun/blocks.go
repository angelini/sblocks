package cloudrun

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type RevisionInstance struct {
	definition *Revision
	state      RevisionState
}

type ServiceInstance struct {
	name      string
	state     ServiceState
	uri       string
	revisions map[string]*RevisionInstance
}

func (si *ServiceInstance) orderedRevisions() []*RevisionInstance {
	keys := make([]string, 0, len(si.revisions))
	for key := range si.revisions {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	instances := make([]*RevisionInstance, 0, len(keys))
	for _, key := range keys {
		instances = append(instances, si.revisions[key])
	}

	return instances
}

type ServiceBlock struct {
	name     string
	public   bool
	labels   map[string]string
	services map[string]*ServiceInstance
}

func CreateServiceBlock(ctx context.Context, client *CloudRunClient, rootName string, public bool, size int, labels map[string]string, revision *Revision) (*ServiceBlock, error) {
	services := make(map[string]*ServiceInstance, size)
	name := fmt.Sprintf("%s-%s", rootName, randomString(6))

	{
		group, ctx := errgroup.WithContext(ctx)

		for i := 0; i < size; i++ {
			serviceName := fmt.Sprintf("%s-%d", name, i)
			group.Go(func() error {
				service, err := client.Create(ctx, serviceName, labels, revision)
				if err != nil {
					return err
				}

				if public {
					err = client.AllowPublicAccess(ctx, serviceName)
					if err != nil {
						return err
					}
				}

				services[serviceName] = &ServiceInstance{
					name:  serviceName,
					uri:   service.Uri,
					state: GetServiceState(service),
				}
				return nil
			})
		}

		err := group.Wait()
		if err != nil {
			return nil, err
		}
	}

	sb := ServiceBlock{
		name,
		public,
		labels,
		services,
	}

	err := sb.loadRevisions(ctx, client)
	if err != nil {
		return nil, err
	}

	return &sb, nil
}

func LoadServiceBlock(ctx context.Context, client *CloudRunClient, name string) (*ServiceBlock, error) {
	services, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

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
			name:  ParseServiceName(service.Name),
			state: GetServiceState(service),
			uri:   service.Uri,
		}
	}

	sb := ServiceBlock{
		name:     name,
		labels:   labels,
		services: instances,
	}

	err = sb.loadRevisions(ctx, client)
	if err != nil {
		return nil, err
	}

	return &sb, nil
}

func (sb *ServiceBlock) loadRevisions(ctx context.Context, client *CloudRunClient) error {
	group, ctx := errgroup.WithContext(ctx)

	for _, service := range sb.services {
		service := service
		group.Go(func() error {
			revisions, err := client.ListRevisions(ctx, service.name)
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

			service.revisions = revisionInstances
			return nil
		})
	}

	return group.Wait()
}

func (sb *ServiceBlock) CreateRevision(ctx context.Context, client *CloudRunClient, revision *Revision) error {
	group, ctx := errgroup.WithContext(ctx)

	for _, service := range sb.services {
		service := service
		group.Go(func() error {
			err := client.Update(ctx, service.name, revision)
			if err != nil {
				return err
			}

			sb.services[service.name].revisions[revision.Name] = &RevisionInstance{
				definition: revision,
				state:      NewRevisionState(),
			}
			return nil
		})
	}

	return group.Wait()
}

func (sb *ServiceBlock) Display() []string {
	labels := make([]string, 0, len(sb.labels))
	for key, value := range sb.labels {
		labels = append(labels, fmt.Sprintf("%s=%s", key, value))
	}

	results := []string{
		fmt.Sprintf("%s [%s]:", sb.name, strings.Join(labels, ", ")),
	}

	for _, service := range sb.orderedServices() {
		results = append(results, fmt.Sprintf("  > %s: %s", service.name, service.state.String()))
		results = append(results, fmt.Sprintf("      %s", service.uri))
		for _, revision := range service.orderedRevisions() {
			results = append(results, fmt.Sprintf("    - %s: %s", strings.TrimPrefix(revision.definition.Name, service.name+"-"), revision.state.String()))
		}
	}

	return results
}

func (sb *ServiceBlock) orderedServices() []*ServiceInstance {
	keys := make([]string, 0, len(sb.services))
	for key := range sb.services {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	instances := make([]*ServiceInstance, 0, len(keys))
	for _, key := range keys {
		instances = append(instances, sb.services[key])
	}

	return instances
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
