package cloudrun

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/angelini/sblocks/internal/maps"
	"golang.org/x/sync/errgroup"
)

type RevisionInstance struct {
	definition *Revision
	labels     map[string]string
	state      RevisionState
}

type ServiceInstance struct {
	name      string
	state     ServiceState
	uri       string
	traffic   *TrafficStatus
	revisions map[string]*RevisionInstance
}

type ServiceBlock struct {
	name     string
	public   bool
	labels   map[string]string
	services map[string]*ServiceInstance
}

func CreateServiceBlock(ctx context.Context, client *Client, public bool, size int, labels map[string]string, revision *Revision) (*ServiceBlock, error) {
	services := make(map[string]*ServiceInstance, size)
	name := randomString(6)

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
					name:    serviceName,
					uri:     service.Uri,
					state:   GetServiceState(service),
					traffic: NewTrafficStatus(service.TrafficStatuses),
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

func LoadServiceBlocks(ctx context.Context, client *Client, environment string) (map[string]*ServiceBlock, error) {
	services, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	blocks := make(map[string]*ServiceBlock)

	for _, service := range services {
		envLabel, found := service.Labels["sb_environment"]
		if !found || envLabel != environment {
			continue
		}

		serviceName := ParseServiceName(service.Name)
		blockName := strings.Split(serviceName, "-")[0]

		block, found := blocks[blockName]
		if !found {
			block = &ServiceBlock{
				name:     blockName,
				labels:   service.Labels,
				services: make(map[string]*ServiceInstance),
			}
			blocks[blockName] = block
		}

		block.services[serviceName] = &ServiceInstance{
			name:    serviceName,
			state:   GetServiceState(service),
			uri:     service.Uri,
			traffic: NewTrafficStatus(service.TrafficStatuses),
		}
	}

	for _, block := range blocks {
		err = block.loadRevisions(ctx, client)
		if err != nil {
			return nil, err
		}
	}

	return blocks, nil
}

func (sb *ServiceBlock) loadRevisions(ctx context.Context, client *Client) error {
	group, ctx := errgroup.WithContext(ctx)

	for _, service := range sb.services {
		service := service
		group.Go(func() error {
			revisions, err := client.ListRevisions(ctx, service.name)
			if err != nil {
				return err
			}

			revisionInstances := make(map[string]*RevisionInstance, len(revisions))
			for idx, revision := range revisions {
				percentage := 0
				if idx == 0 && service.traffic.latest {
					percentage = 100
				} else {
					percentage = int(service.traffic.revisions[revision.Name])
				}

				definition := RevisionDefinition(revision)
				revisionInstances[revision.Name] = &RevisionInstance{
					definition: &definition,
					labels:     revision.Labels,
					state:      GetRevisionState(revision, percentage),
				}
			}

			service.revisions = revisionInstances
			return nil
		})
	}

	return group.Wait()
}

func (sb *ServiceBlock) CreateRevision(ctx context.Context, client *Client, revision *Revision) error {
	group, ctx := errgroup.WithContext(ctx)

	for _, service := range sb.services {
		service := service
		group.Go(func() error {
			err := client.Update(ctx, service.name, sb.labels, revision)
			if err != nil {
				return err
			}

			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		return err
	}

	err = sb.loadRevisions(ctx, client)
	if err != nil {
		return err
	}

	return nil
}

func (sb *ServiceBlock) Display() []string {
	results := []string{
		fmt.Sprintf("%s [%s]:", sb.name, formatLabels(sb.labels)),
	}

	for _, service := range maps.SortedValues(sb.services) {
		results = append(results, fmt.Sprintf("  > %s: %s", service.name, service.state.String()))
		results = append(results, fmt.Sprintf("    uri: %s", service.uri))
		for _, revision := range maps.SortedValues(service.revisions) {
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
