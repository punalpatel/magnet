package magnet_test

import (
	"context"
	"errors"

	"github.com/pivotalservices/magnet"
	"github.com/pivotalservices/magnet/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("State", func() {
	var (
		i *mock.IaaS
	)
	BeforeEach(func() {
		i = &mock.IaaS{}
	})

	Context("when Check()ing a mock IaaS", func() {
		It("gets the state of the IaaS", func() {
			gotState := false
			i.StateFn = func(ctx context.Context) (*magnet.State, error) {
				gotState = true
				return nil, nil
			}
			magnet.Check(context.Background(), i)
			Ω(gotState).Should(BeTrue())
		})

		It("returns an error if it can't connect to the IaaS", func() {
			var noState = errors.New("couldn't get state")
			i.StateFn = func(ctx context.Context) (*magnet.State, error) {
				return nil, noState
			}
			err := magnet.Check(context.Background(), i)
			Ω(err).Should(Equal(noState))
		})
	})

	Context("IsBalanced", func() {
		It("reports a single-host state as balanced", func() {
			state := &magnet.State{
				Hosts: []*magnet.Host{
					&magnet.Host{},
				},
			}
			Ω(magnet.IsBalanced(state)).Should(BeTrue())
		})

		It("detects an unbalanced state when two jobs are on the same host", func() {
			host1 := &magnet.Host{ID: "host1"}
			host2 := &magnet.Host{ID: "host2"}
			routerVM1 := &magnet.VM{Job: "router", Host: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", Host: host1.ID}

			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2},
				VMs:   []*magnet.VM{routerVM1, routerVM2},
			}
			Ω(magnet.IsBalanced(state)).Should(BeFalse())
		})

		It("detects a balanced state when two jobs are on separate hosts", func() {
			host1 := &magnet.Host{ID: "host1"}
			host2 := &magnet.Host{ID: "host2"}
			routerVM1 := &magnet.VM{Job: "router", Host: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", Host: host2.ID}

			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2},
				VMs:   []*magnet.VM{routerVM1, routerVM2},
			}
			Ω(magnet.IsBalanced(state)).Should(BeTrue())
		})

		It("detects a balanced state when 4 jobs are split on separate hosts", func() {
			host1 := &magnet.Host{ID: "host1"}
			host2 := &magnet.Host{ID: "host2"}
			routerVM1 := &magnet.VM{Job: "router", Host: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", Host: host2.ID}
			cellVM1 := &magnet.VM{Job: "diego_cell", Host: host1.ID}
			cellVM2 := &magnet.VM{Job: "diego_cell", Host: host2.ID}

			// throw some extra jobs in the mix
			clockGlobalVM := &magnet.VM{Job: "clock_global", Host: host2.ID}
			cloudControllerVM := &magnet.VM{Job: "cloud_controller", Host: host2.ID}
			haproxyVM := &magnet.VM{Job: "ha_proxy", Host: host1.ID}

			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2},
				VMs:   []*magnet.VM{routerVM1, routerVM2, cellVM1, cellVM2, clockGlobalVM, cloudControllerVM, haproxyVM},
			}
			Ω(magnet.IsBalanced(state)).Should(BeTrue())
		})

		It("detects an unbalanced state when 4 jobs are incorrectly split on separate hosts", func() {
			host1 := &magnet.Host{ID: "host1"}
			host2 := &magnet.Host{ID: "host2"}
			// both routers on the same host, and both cells on the same host
			routerVM1 := &magnet.VM{Job: "router", Host: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", Host: host1.ID}
			cellVM1 := &magnet.VM{Job: "diego_cell", Host: host2.ID}
			cellVM2 := &magnet.VM{Job: "diego_cell", Host: host2.ID}

			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2},
				VMs:   []*magnet.VM{routerVM1, routerVM2, cellVM1, cellVM2},
			}
			Ω(magnet.IsBalanced(state)).Should(BeFalse())
		})

		It("detects a balanced state when some duplication is required", func() {
			host1 := &magnet.Host{ID: "host1"}
			host2 := &magnet.Host{ID: "host2"}
			host3 := &magnet.Host{ID: "host3"}

			routerVM1 := &magnet.VM{Job: "router", Host: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", Host: host2.ID}
			routerVM3 := &magnet.VM{Job: "router", Host: host3.ID}
			routerVM4 := &magnet.VM{Job: "router", Host: host1.ID}
			routerVM5 := &magnet.VM{Job: "router", Host: host2.ID}

			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2, host3},
				VMs:   []*magnet.VM{routerVM1, routerVM2, routerVM3, routerVM4, routerVM5},
			}
			Ω(magnet.IsBalanced(state)).Should(BeTrue())
		})

		It("detects an unbalanced state when some duplication is required", func() {
			host1 := &magnet.Host{ID: "host1"}
			host2 := &magnet.Host{ID: "host2"}
			host3 := &magnet.Host{ID: "host3"}

			routerVM1 := &magnet.VM{Job: "router", Host: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", Host: host2.ID}

			// put 3 routers on a single host, when the best-case
			// scenario is a max of 2 routers per host
			routerVM3 := &magnet.VM{Job: "router", Host: host3.ID}
			routerVM4 := &magnet.VM{Job: "router", Host: host3.ID}
			routerVM5 := &magnet.VM{Job: "router", Host: host3.ID}

			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2, host3},
				VMs:   []*magnet.VM{routerVM1, routerVM2, routerVM3, routerVM4, routerVM5},
			}
			Ω(magnet.IsBalanced(state)).Should(BeFalse())
		})

		It("detects a balanced state when some hosts are not utilized", func() {
			host1 := &magnet.Host{ID: "host1"}
			host2 := &magnet.Host{ID: "host2"}
			host3 := &magnet.Host{ID: "host3"}
			host4 := &magnet.Host{ID: "host4"}
			host5 := &magnet.Host{ID: "host5"}

			routerVM1 := &magnet.VM{Job: "router", Host: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", Host: host2.ID}
			routerVM3 := &magnet.VM{Job: "router", Host: host3.ID}

			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2, host3, host4, host5},
				VMs:   []*magnet.VM{routerVM1, routerVM2, routerVM3},
			}
			Ω(magnet.IsBalanced(state)).Should(BeTrue())
		})
	})
})
