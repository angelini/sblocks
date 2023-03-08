package cloudrun

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type RevisionInstance struct {
	definition *Revision
	state      RevisionState
}

func (ri *RevisionInstance) ShortName() string {
	return strings.SplitN(ri.definition.Name, "/", 8)[7]
}

type ServiceInstance struct {
	name      string
	state     ServiceState
	revisions map[string]*RevisionInstance
}

func (si *ServiceInstance) ShortName() string {
	return strings.SplitN(si.name, "/", 6)[5]
}

type ServiceBlock struct {
	name     string
	labels   map[string]string
	services map[string]*ServiceInstance
}

func CreateServiceBlock(ctx context.Context, client *CloudRunClient, rootName string, size int, labels map[string]string, revision *Revision) (*ServiceBlock, error) {
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

				services[serviceName] = &ServiceInstance{
					name:  fmt.Sprintf("%s/services/%s", client.Parent, serviceName),
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
			name:  service.Name,
			state: GetServiceState(service),
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

	for _, instance := range sb.services {
		instance := instance
		group.Go(func() error {
			revisions, err := client.ListRevisions(ctx, instance.ShortName())
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

			instance.revisions = revisionInstances
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

	for _, service := range sb.services {
		results = append(results, fmt.Sprintf("  > %s: %s", service.ShortName(), service.state.String()))
		for _, revision := range service.revisions {
			results = append(results, fmt.Sprintf("    - %s: %s", strings.TrimPrefix(revision.ShortName(), service.ShortName()+"-"), revision.state.String()))
		}
	}

	return results
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
