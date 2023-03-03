package cloudrun

import (
	"context"
	"fmt"

	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/angelini/sblocks/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/iterator"
)

type CloudRunClient struct {
	parent   string
	services *run.ServicesClient
}

func NewCloudRunClient(ctx context.Context, project, location string) (*CloudRunClient, error) {
	services, err := run.NewServicesClient(ctx)
	if err != nil {
		return nil, err
	}

	return &CloudRunClient{
		parent:   fmt.Sprintf("projects/%s/locations/%s", project, location),
		services: services,
	}, nil
}

func (c *CloudRunClient) Close() error {
	return c.services.Close()
}

func asPbContainers(containers map[string]*Container) []*runpb.Container {
	result := make([]*runpb.Container, 0, len(containers))
	for _, container := range containers {
		result = append(result, &runpb.Container{
			Name:  container.Name,
			Image: container.Image,
		})
	}
	return result
}

func (c *CloudRunClient) Create(ctx context.Context, name string, labels map[string]string, revision *Revision) error {
	req := &runpb.CreateServiceRequest{
		Parent:    c.parent,
		ServiceId: name,
		Service: &runpb.Service{
			Description: "Managed by sblocks",
			Labels:      labels,
			Ingress:     runpb.IngressTraffic_INGRESS_TRAFFIC_ALL,
			Template: &runpb.RevisionTemplate{
				Revision:   fmt.Sprintf("%s-%s", name, revision.Name),
				Labels:     labels,
				Containers: asPbContainers(revision.Containers),
			},
		},
	}

	log.Info(ctx, "start create service", zap.String("name", name))
	op, err := c.services.CreateService(ctx, req)
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

	log.Info(ctx, "finished create service", zap.String("name", name))
	return nil
}

func (c *CloudRunClient) List(ctx context.Context) ([]string, error) {
	req := &runpb.ListServicesRequest{
		Parent: c.parent,
	}

	results := []string{}
	it := c.services.ListServices(ctx, req)

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		results = append(results, resp.Name)

	}

	return results, nil
}

func (c *CloudRunClient) Update(ctx context.Context) error {
	req := &runpb.UpdateServiceRequest{}

	c.services.UpdateService(ctx, req)

	return nil
}

func (c *CloudRunClient) DeleteAll(ctx context.Context) error {
	names, err := c.List(ctx)
	if err != nil {
		return err
	}

	group, ctx := errgroup.WithContext(ctx)

	for _, name := range names {
		name := name

		group.Go(func() error {
			req := &runpb.DeleteServiceRequest{
				Name: name,
			}

			log.Info(ctx, "start delete service", zap.String("name", name))
			op, err := c.services.DeleteService(ctx, req)
			if err != nil {
				return err
			}

			_, err = op.Wait(ctx)
			if err != nil {
				return err
			}

			return nil
		})
	}

	return group.Wait()
}
