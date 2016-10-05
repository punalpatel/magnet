package mock

import "github.com/pivotalservices/magnet"

type IaaS struct {
}

func New() magnet.IaaS {
	return &IaaS{}
}

func (m *IaaS) Connect() error {
	return nil
}
func (m *IaaS) IsConnected() bool {
	return true
}

func (m *IaaS) Jobs() ([]magnet.Job, error) {
	return nil, nil
}
func (m *IaaS) VMs() ([]magnet.VM, error) {
	return nil, nil
}
func (m *IaaS) Rules() ([]magnet.Rule, error) {
	return nil, nil
}
