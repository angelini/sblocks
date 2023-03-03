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
	Name    string
	Image   string
	Command string
	Args    []string
}

type Revision struct {
	Name       string
	Containers map[string]*Container
}
