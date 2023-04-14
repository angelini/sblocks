package cloudrun

import (
	"context"
	"fmt"
	"strings"

	iampb "cloud.google.com/go/iam/apiv1/iampb"
	run "cloud.google.com/go/run/apiv2"
	pb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/angelini/sblocks/internal/log"
	"github.com/angelini/sblocks/pkg/core"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/iterator"
)

type Client struct {
	Parent    string
	services  *run.ServicesClient
	revisions *run.RevisionsClient
}

func NewClient(ctx context.Context, project, location string) (*Client, error) {
	services, err := run.NewServicesClient(ctx)
	if err != nil {
		return nil, err
	}

	revisions, err := run.NewRevisionsClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		Parent:    fmt.Sprintf("projects/%s/locations/%s", project, location),
		services:  services,
		revisions: revisions,
	}, nil
}

func (c *Client) Close() error {
	return c.services.Close()
}

func asPbContainers(containers map[string]core.Container) []*pb.Container {
	result := make([]*pb.Container, 0, len(containers))
	for _, container := range containers {
		result = append(result, &pb.Container{
			Name:  container.Name,
			Image: container.Image,
		})
	}
	return result
}

func (c *Client) Create(ctx context.Context, name string, labels map[string]string, revision core.Revision) (*pb.Service, error) {
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

func (c *Client) List(ctx context.Context) ([]*pb.Service, error) {
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

func (c *Client) AllowPublicAccess(ctx context.Context, serviceName string) error {
	req := &iampb.SetIamPolicyRequest{
		Resource: fmt.Sprintf("%s/services/%s", c.Parent, serviceName),
		Policy: &iampb.Policy{
			Bindings: []*iampb.Binding{
				{Role: "roles/run.invoker", Members: []string{"allUsers"}},
			},
		},
	}

	_, err := c.services.SetIamPolicy(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ListAllRevisions(ctx context.Context) ([]*pb.Revision, error) {
	req := &pb.ListRevisionsRequest{
		Parent: "-",
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

func (c *Client) ListRevisions(ctx context.Context, serviceName string) ([]*pb.Revision, error) {
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

func (c *Client) Update(ctx context.Context, serviceName string, labels map[string]string, revision *core.Revision) error {
	req := &pb.UpdateServiceRequest{
		Service: &pb.Service{
			Name:   fmt.Sprintf("%s/services/%s", c.Parent, serviceName),
			Labels: labels,
			Template: &pb.RevisionTemplate{
				Revision:   fmt.Sprintf("%s-%s", serviceName, revision.Name),
				Labels:     labels,
				Containers: asPbContainers(revision.Containers),
			},
		},
	}

	log.Info(ctx, "start update service", zap.String("name", serviceName))
	op, err := c.services.UpdateService(ctx, req)
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

	log.Info(ctx, "finished update service", zap.String("name", serviceName))
	return nil
}

func (c *Client) Delete(ctx context.Context, name string) error {
	req := &pb.DeleteServiceRequest{
		Name: fmt.Sprintf("%s/services/%s", c.Parent, name),
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

	log.Info(ctx, "finished delete service", zap.String("name", name))
	return nil
}

func (c *Client) DeleteAll(ctx context.Context) error {
	services, err := c.List(ctx)
	if err != nil {
		return err
	}

	group, ctx := errgroup.WithContext(ctx)

	for _, service := range services {
		name := service.Name
		shortName := strings.SplitN(name, "/", 6)[5]

		group.Go(func() error {
			req := &pb.DeleteServiceRequest{
				Name: name,
			}

			log.Info(ctx, "start delete service", zap.String("name", shortName))
			op, err := c.services.DeleteService(ctx, req)
			if err != nil {
				return err
			}

			_, err = op.Wait(ctx)
			if err != nil {
				return err
			}

			log.Info(ctx, "finished delete service", zap.String("name", shortName))
			return nil
		})
	}

	return group.Wait()
}
