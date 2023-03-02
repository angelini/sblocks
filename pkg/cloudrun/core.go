package cloudrun

type ServiceState int

const (
	ServiceMissing ServiceState = iota
	Creating
	Created
	Deleting
)

type RevisionState int

const (
	RevisionMissing RevisionState = iota
	Starting
	Started
	Running
	Stopping
	Stopped
)

type Container struct {
	Name  string
	Image string
}

type Service struct {
	RootName   string
	Labels     map[string]string
	Containers []Container
}
