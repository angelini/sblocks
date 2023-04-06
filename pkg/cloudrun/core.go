package cloudrun

import (
	"fmt"
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

func (s *ServiceState) String() string {
	result := "STOPPED"
	if s.isReady {
		result = "READY"
	}

	if s.isReconciling {
		result += "(*)"
	}

	return result
}

type RevisionState struct {
	isReconciling      bool
	observedGeneration int64
	isDeleted          bool
	traffic            int
}

func GetRevisionState(revision *pb.Revision, traffic int) RevisionState {
	return RevisionState{
		isReconciling:      revision.Reconciling,
		observedGeneration: revision.ObservedGeneration,
		isDeleted:          revision.DeleteTime != nil,
		traffic:            traffic,
	}
}

func NewRevisionState() RevisionState {
	return RevisionState{
		isReconciling:      false,
		observedGeneration: -1,
		isDeleted:          false,
	}
}

func (r *RevisionState) String() string {
	result := "INACTIVE"

	if r.traffic > 0 {
		result = fmt.Sprintf("ACTIVE (%d)", r.traffic)
	}

	if r.isDeleted {
		result = "DELETED"
	}

	if r.isReconciling {
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

type TrafficStatus struct {
	latest    bool
	revisions map[string]int32
}

func NewTrafficStatus(statuses []*pb.TrafficTargetStatus) *TrafficStatus {
	if len(statuses) == 1 && statuses[0].Type == pb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST {
		return &TrafficStatus{latest: true}
	}

	revisions := make(map[string]int32, len(statuses))
	for _, status := range statuses {
		revisions[status.Revision] = status.Percent
	}

	return &TrafficStatus{
		latest:    false,
		revisions: revisions,
	}
}
