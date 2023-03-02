package cloudrun

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"
)

type ServiceInstance struct {
	name      string
	state     ServiceState
	revisions map[string]RevisionState
}

type ServiceBlock struct {
	name       string
	definition Service
	instances  map[string]*ServiceInstance
}

func NewServiceBlock(service Service, size int) *ServiceBlock {
	instances := make(map[string]*ServiceInstance, size)
	name := fmt.Sprintf("%s-%s", service.RootName, randomString(6))

	for i := 0; i < size; i++ {
		name := fmt.Sprintf("%s-%d", name, i)
		instances[name] = &ServiceInstance{
			name:      name,
			state:     ServiceMissing,
			revisions: make(map[string]RevisionState),
		}
	}

	return &ServiceBlock{
		name,
		service,
		instances,
	}
}

func (sb *ServiceBlock) Create(ctx context.Context, client *CloudRunClient, revision string) error {
	group, ctx := errgroup.WithContext(ctx)

	for _, instance := range sb.instances {
		instance.revisions[revision] = RevisionMissing
	}

	for _, instance := range sb.instances {
		instance := instance
		group.Go(func() error {
			instance.state = Creating
			instance.revisions[revision] = Starting

			err := client.Create(ctx, sb.definition, instance.name, revision)
			if err != nil {
				instance.state = ServiceMissing
				instance.revisions[revision] = RevisionMissing
				return err
			}

			instance.state = Created
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
