package cloudrun

import (
	"strings"
	"time"

	pb "cloud.google.com/go/run/apiv2/runpb"
)

type ServiceState struct {
	isReconciling      bool
	observedGeneration int64
	latestRevision     string
	isReady            bool
}

func NewServiceState() ServiceState {
	return ServiceState{
		isReconciling:      false,
		observedGeneration: -1,
		latestRevision:     "",
		isReady:            false,
	}
}

func ErrorServiceState() ServiceState {
	return ServiceState{
		isReconciling:      false,
		observedGeneration: -1,
		latestRevision:     "",
		isReady:            false,
	}
}

func GetServiceState(service *pb.Service) ServiceState {
	return ServiceState{
		isReconciling:      service.Reconciling,
		observedGeneration: service.ObservedGeneration,
		latestRevision:     service.LatestReadyRevision,
		isReady:            service.TerminalCondition.Type == "Ready",
	}
}

func (s *ServiceState) IsReconciling() bool {
	return s.isReconciling
}

func (s *ServiceState) IsReady() bool {
	return s.isReady
}

func (s *ServiceState) String() string {
	result := "STOPPED"
	if s.IsReady() {
		result = "READY"
	}

	if s.IsReconciling() {
		result += "(*)"
	}

	return result
}

type RevisionState struct {
	isReconciling      bool
	observedGeneration int64
	isDeleted          bool
}

func GetRevisionState(revision *pb.Revision) RevisionState {
	return RevisionState{
		isReconciling:      revision.Reconciling,
		observedGeneration: revision.ObservedGeneration,
		isDeleted:          revision.DeleteTime != nil,
	}
}

func NewRevisionState() RevisionState {
	return RevisionState{
		isReconciling:      false,
		observedGeneration: -1,
		isDeleted:          false,
	}
}

func (s *RevisionState) IsReconciling() bool {
	return s.isReconciling
}

func (s *RevisionState) IsRunning() bool {
	return !s.isDeleted
}

func (s *RevisionState) String() string {
	result := "DELETED"
	if s.IsRunning() {
		result = "RUNNING"
	}

	if s.IsReconciling() {
		result += "(*)"
	}

	return result
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
		Name:           ParseRevisionName(revision.Name),
		MinScale:       uint32(revision.Scaling.MinInstanceCount),
		MaxScale:       uint32(revision.Scaling.MaxInstanceCount),
		MaxConcurrency: uint32(revision.MaxInstanceRequestConcurrency),
		Timeout:        revision.Timeout.AsDuration(),
		Containers:     containers,
	}
}

func ParseServiceName(resource string) string {
	return strings.SplitN(resource, "/", 6)[5]
}

func ParseRevisionName(resource string) string {
	return strings.SplitN(resource, "/", 8)[7]
}
