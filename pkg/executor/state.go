package executor

import (
	"context"
	"fmt"

	"cloud.google.com/go/run/apiv2/runpb"
	cr "github.com/angelini/sblocks/pkg/cloudrun"
	rt "github.com/angelini/sblocks/pkg/runtime"
	"golang.org/x/sync/errgroup"
)

type State struct {
	runtimes map[string]*rt.RuntimeInstance
}

func NewState() *State {
	return &State{
		runtimes: make(map[string]*rt.RuntimeInstance),
	}
}

func (s *State) AddRuntime(runtime rt.Runtime) error {
	if _, found := s.runtimes[runtime.Name]; found {
		return fmt.Errorf("runtime with name '%s' already exists in state", runtime.Name)
	}

	s.runtimes[runtime.Name] = rt.NewRuntimeInstance(runtime)
	return nil
}

func (s *State) Refresh(ctx context.Context, client *cr.Client) error {
	services, revisions, err := fetchAll(ctx, client)
	if err != nil {
		return err
	}

	group, ctx := errgroup.WithContext(ctx)

	for _, runtime := range s.runtimes {
		runtime := runtime
		group.Go(func() error {
			return runtime.Refresh(ctx, client, services, revisions)
		})
	}

	return group.Wait()
}

func (s *State) Converge(ctx context.Context, client *cr.Client) error {
	var operations []rt.Operation

	for _, runtime := range s.runtimes {
		operations = append(operations, runtime.Converge()...)
	}

	group, ctx := errgroup.WithContext(ctx)

	for _, operation := range operations {
		operation := operation
		group.Go(func() error {
			return operation.Execute(ctx, client)
		})
	}

	return group.Wait()
}

func fetchAll(ctx context.Context, client *cr.Client) ([]*runpb.Service, []*runpb.Revision, error) {
	var services []*runpb.Service
	var revisions []*runpb.Revision

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		var err error
		services, err = client.List(ctx)
		return err
	})

	group.Go(func() error {
		var err error
		revisions, err = client.ListAllRevisions(ctx)
		return err
	})

	err := group.Wait()
	if err != nil {
		return nil, nil, err
	}

	return services, revisions, nil
}
