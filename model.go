package magnet

import "context"

// IaaS is an abstraction for a particular IaaS.
type IaaS interface {
	State(ctx context.Context) (*State, error)
	Converge(ctx context.Context, state *State) error
}

// State represents the resources in a Cloud Foundry deployment.
type State struct {
	Hosts []*Host
	VMs   []*VM
	Rules []*Rule
}

// VM is a virtual machine in a Cloud Foundry depoyment.
type VM struct {
	Name      string
	ID        string
	HostUUID  string
	HostName  string
	Job       string
	Reference string
}

// Host is a host in a Cloud Foundry deployment.
type Host struct {
	Name string
	ID   string
}

type Rule struct {
	Name      string
	ID        string
	Enabled   bool
	Mandatory bool
	VMs       []*VM
}
