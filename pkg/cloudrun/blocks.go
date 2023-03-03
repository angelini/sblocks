package cloudrun

import (
	"context"
	"fmt"
	"math/rand"
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
			state:     ServiceMissing,
			revisions: make(map[string]*RevisionInstance),
		}
	}

	return &ServiceBlock{
		name,
		labels,
		services,
	}
}

func (sb *ServiceBlock) Create(ctx context.Context, client *CloudRunClient, revision *Revision) error {
	group, ctx := errgroup.WithContext(ctx)

	for _, service := range sb.services {
		service.revisions[revision.Name] = &RevisionInstance{
			definition: revision,
			state:      RevisionMissing,
		}
	}

	for _, service := range sb.services {
		service := service
		group.Go(func() error {
			service.state = Creating
			service.revisions[revision.Name].state = Starting

			err := client.Create(ctx, service.name, sb.labels, revision)
			if err != nil {
				service.state = ServiceMissing
				service.revisions[revision.Name].state = RevisionMissing
				return err
			}

			service.state = Created
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
