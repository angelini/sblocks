package cloudrun

import (
	"context"
	"fmt"

	iampb "cloud.google.com/go/iam/apiv1/iampb"
	run "cloud.google.com/go/run/apiv2"
	pb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/angelini/sblocks/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/iterator"
)

type CloudRunClient struct {
	Parent    string
	services  *run.ServicesClient
	revisions *run.RevisionsClient
}

func NewCloudRunClient(ctx context.Context, project, location string) (*CloudRunClient, error) {
	services, err := run.NewServicesClient(ctx)
	if err != nil {
		return nil, err
	}

	revisions, err := run.NewRevisionsClient(ctx)
	if err != nil {
		return nil, err
	}

	return &CloudRunClient{
		Parent:    fmt.Sprintf("projects/%s/locations/%s", project, location),
		services:  services,
		revisions: revisions,
	}, nil
}

func (c *CloudRunClient) Close() error {
	return c.services.Close()
}

func asPbContainers(containers map[string]Container) []*pb.Container {
	result := make([]*pb.Container, 0, len(containers))
	for _, container := range containers {
		result = append(result, &pb.Container{
			Name:  container.Name,
			Image: container.Image,
		})
	}
	return result
}

func (c *CloudRunClient) Create(ctx context.Context, name string, labels map[string]string, revision *Revision) (*pb.Service, error) {
	req := &pb.CreateServiceRequest{
		Parent:    c.Parent,
		ServiceId: name,
		Service: &pb.Service{
			Description: "Managed by sblocks",
			Labels:      labels,
			Ingress:     pb.IngressTraffic_INGRESS_TRAFFIC_ALL,
			Template: &pb.RevisionTemplate{
				Revision:   fmt.Sprintf("%s-%s", name, revision.Name),
				Labels:     labels,
				Containers: asPbContainers(revision.Containers),
			},
		},
	}

	log.Info(ctx, "start create service", zap.String("name", name))
	op, err := c.services.CreateService(ctx, req)
	if err != nil {
		return nil, err
	}

	service, err := op.Wait(ctx)
	if err != nil {
		return nil, err
	}

	log.Info(ctx, "finished create service", zap.String("name", name))
	return service, nil
}

func (c *CloudRunClient) List(ctx context.Context) ([]*pb.Service, error) {
	req := &pb.ListServicesRequest{
		Parent: c.Parent,
	}

	var services []*pb.Service
	it := c.services.ListServices(ctx, req)

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		services = append(services, resp)

	}

	return services, nil
}

func (c *CloudRunClient) ListRevisions(ctx context.Context, serviceName string) ([]*pb.Revision, error) {
	req := &pb.ListRevisionsRequest{
		Parent: fmt.Sprintf("%s/services/%s", c.Parent, serviceName),
	}

	var revisions []*pb.Revision
	it := c.revisions.ListRevisions(ctx, req)

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		revisions = append(revisions, resp)

	}

	return revisions, nil
}

func (c *CloudRunClient) GetIAM(ctx context.Context, serviceName string) error {
	req := &iampb.GetIamPolicyRequest{
		Resource: fmt.Sprintf("%s/services/%s", c.Parent, serviceName),
	}

	resp, err := c.services.GetIamPolicy(ctx, req)
	if err != nil {
		return err
	}

	log.Info(ctx, "iam", zap.String("policy", resp.String()))
	for _, binding := range resp.Bindings {
		log.Info(ctx, "iam binding", zap.String("binding", binding.String()))
	}

	return nil
}

func (c *CloudRunClient) Update(ctx context.Context) error {
	req := &pb.UpdateServiceRequest{}

	c.services.UpdateService(ctx, req)

	return nil
}

func (c *CloudRunClient) DeleteAll(ctx context.Context) error {
	services, err := c.List(ctx)
	if err != nil {
		return err
	}

	group, ctx := errgroup.WithContext(ctx)

	for _, service := range services {
		name := service.Name

		group.Go(func() error {
			req := &pb.DeleteServiceRequest{
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
