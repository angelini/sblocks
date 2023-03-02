package cloudrun

import (
	"context"
	"fmt"

	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
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

func asPbContainers(containers []Container) []*runpb.Container {
	result := make([]*runpb.Container, len(containers))
	for i, container := range containers {
		result[i] = &runpb.Container{
			Name:  container.Name,
			Image: container.Image,
		}
	}
	return result
}

func (c *CloudRunClient) Create(ctx context.Context, service Service, name, revision string) error {
	req := &runpb.CreateServiceRequest{
		Parent:    c.parent,
		ServiceId: name,
		Service: &runpb.Service{
			Description: "Managed by sblocks",
			Labels:      service.Labels,
			Ingress:     runpb.IngressTraffic_INGRESS_TRAFFIC_ALL,
			Template: &runpb.RevisionTemplate{
				Revision:   fmt.Sprintf("%s-%s", name, revision),
				Labels:     service.Labels,
				Containers: asPbContainers(service.Containers),
			},
		},
	}

	op, err := c.services.CreateService(ctx, req)
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

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
