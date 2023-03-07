package cloudrun

import (
	"time"

	pb "cloud.google.com/go/run/apiv2/runpb"
)

type ServiceState struct {
	isReconciling      bool
	observedGeneration int64
	latestRevision     string
	isTerminal         bool
}

func NewServiceState() ServiceState {
	return ServiceState{
		isReconciling:      false,
		observedGeneration: -1,
		latestRevision:     "",
		isTerminal:         false,
	}
}

func ErrorServiceState() ServiceState {
	return ServiceState{
		isReconciling:      false,
		observedGeneration: -1,
		latestRevision:     "",
		isTerminal:         true,
	}
}

func GetServiceState(service *pb.Service) ServiceState {
	return ServiceState{
		isReconciling:      service.Reconciling,
		observedGeneration: service.ObservedGeneration,
		latestRevision:     service.LatestReadyRevision,
		isTerminal:         service.TerminalCondition != nil,
	}
}

func (s *ServiceState) IsReconciling() bool {
	return s.isReconciling
}

func (s *ServiceState) IsRunning() bool {
	return !s.isTerminal && s.latestRevision != ""
}

type RevisionState struct {
	isReconciling      bool
	observedGeneration int64
}

func GetRevisionState(revision *pb.Revision) RevisionState {
	return RevisionState{
		isReconciling:      revision.Reconciling,
		observedGeneration: revision.ObservedGeneration,
	}
}

func NewRevisionState() ServiceState {
	return ServiceState{
		isReconciling:      false,
		observedGeneration: -1,
	}
}

func (s *RevisionState) IsReconciling() bool {
	return s.isReconciling
}

type Container struct {
	Name    string
	Image   string
	Command string
	Args    []string
}

func ContainerDefinition(container *pb.Container) Container {
	command := ""
	if len(container.Command) > 0 {
		command = container.Command[0]
	}

	return Container{
		Name:    container.Name,
		Image:   container.Image,
		Command: command,
		Args:    container.Args,
	}
}

type Revision struct {
	Name           string
	MinScale       uint32
	MaxScale       uint32
	MaxConcurrency uint32
	Timeout        time.Duration
	Containers     map[string]Container
}

func RevisionDefinition(revision *pb.Revision) Revision {
	containers := make(map[string]Container, len(revision.Containers))
	for _, container := range revision.Containers {
		containers[container.Name] = ContainerDefinition(container)
	}

	return Revision{
		Name:           revision.Name,
		MinScale:       uint32(revision.Scaling.MinInstanceCount),
		MaxScale:       uint32(revision.Scaling.MaxInstanceCount),
		MaxConcurrency: uint32(revision.MaxInstanceRequestConcurrency),
		Timeout:        revision.Timeout.AsDuration(),
		Containers:     containers,
	}
}
