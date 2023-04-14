package runtime

import (
	"context"

	cr "github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/angelini/sblocks/pkg/core"
	"github.com/google/uuid"
)

type Operation interface {
	Runtime() string
	Execute(context.Context, *cr.Client) error
}

type CreateFreeServiceOp struct {
	runtimeName string
	public      bool
	labels      map[string]string
	revision    core.Revision
}

func (c *CreateFreeServiceOp) Runtime() string {
	return c.runtimeName
}

func (c *CreateFreeServiceOp) Execute(ctx context.Context, client *cr.Client) error {
	serviceName := uuid.New().String()

	_, err := client.Create(ctx, serviceName, c.labels, c.revision)
	if err != nil {
		return err
	}

	if c.public {
		err = client.AllowPublicAccess(ctx, serviceName)
		if err != nil {
			return err
		}
	}

	return nil
}

type DeleteFreeServiceOp struct {
	runtimeName string
	serviceName string
}

func (c *DeleteFreeServiceOp) Runtime() string {
	return c.runtimeName
}

func (c *DeleteFreeServiceOp) Execute(ctx context.Context, client *cr.Client) error {
	return client.Delete(ctx, c.serviceName)
}
