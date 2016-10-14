package magnet

import "context"

// IaaS is an abstraction for a particular IaaS.
type IaaS interface {
	State(ctx context.Context) (*State, error)
	Converge(ctx context.Context, state *State, rec *RuleRecommendation) error
}

// State represents the resources in a Cloud Foundry deployment.
type State struct {
	RuleContainer string
	VMContainer   string
	Hosts         []*Host
	VMs           []*VM
	Rules         []*Rule
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

// Rule can be used to achieve anti-affinity
type Rule struct {
	Name      string
	ID        string
	Enabled   bool
	Mandatory bool
	VMs       []*VM
}

// RuleRecommendation is a reccomendation for how to achieve anti-affinity
// based on the current state of the system.
type RuleRecommendation struct {
	Valid   []Rule // already exist and should be left unchanged
	Stale   []Rule // outdated and should be removed
	Missing []Rule // don't yet exist and need to be created
}
