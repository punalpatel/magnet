package magnet

import "context"

type IaaS interface {
	State(ctx context.Context) (*State, error)
	Converge(ctx context.Context, state *State) error
}

type State struct {
	Clusters []*Cluster
	Hosts    []*Host
	VMs      []*VM
	Rules    []*Rule
}

type Cluster struct {
	ID            string
	Name          string
	ResourcePools []*ResourcePool
	Rules         []*Rule
	VMs           []*VM
}

type ResourcePool struct {
	ID       string
	Name     string
	Implicit bool
}

type VM struct {
	ID           string
	Name         string
	Host         string
	Cluster      string
	ResourcePool string
	Job          string
}

type Host struct {
	ID      string
	Name    string
	Cluster *Cluster
	VMs     []*VM
}

type Rule struct {
	ID     string
	Name   string
	Parent string
	VMs    []*VM
}

type Job struct {
	ID   string
	Name string
	VMs  []*VM
}
