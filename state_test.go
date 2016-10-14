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
				return &magnet.State{}, nil
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

		It("attempts to converge an unbalanced state to a balanced one", func() {
			i.StateFn = func(ctx context.Context) (*magnet.State, error) {
				host1 := &magnet.Host{ID: "host1"}
				host2 := &magnet.Host{ID: "host2"}

				routerVM1 := &magnet.VM{Job: "router", HostUUID: host1.ID}
				routerVM2 := &magnet.VM{Job: "router", HostUUID: host1.ID}

				cellVM1 := &magnet.VM{Job: "diego_cell", HostUUID: host2.ID}
				cellVM2 := &magnet.VM{Job: "diego_cell", HostUUID: host2.ID}

				state := &magnet.State{
					Hosts: []*magnet.Host{host1, host2},
					VMs:   []*magnet.VM{routerVM1, routerVM2, cellVM1, cellVM2},
				}
				return state, nil
			}

			calledConverge := false
			i.ConvergeFn = func(ctx context.Context, state *magnet.State) error {
				calledConverge = true
				return nil
			}

			Ω(magnet.Check(context.Background(), i)).Should(Succeed())
			Ω(calledConverge).Should(BeTrue())
		})

		It("returns an error if it can't converge to the new state", func() {
			i.StateFn = func(ctx context.Context) (*magnet.State, error) {
				host1 := &magnet.Host{ID: "host1"}
				host2 := &magnet.Host{ID: "host2"}

				routerVM1 := &magnet.VM{Job: "router", HostUUID: host1.ID}
				routerVM2 := &magnet.VM{Job: "router", HostUUID: host1.ID}

				cellVM1 := &magnet.VM{Job: "diego_cell", HostUUID: host2.ID}
				cellVM2 := &magnet.VM{Job: "diego_cell", HostUUID: host2.ID}

				state := &magnet.State{
					Hosts: []*magnet.Host{host1, host2},
					VMs:   []*magnet.VM{routerVM1, routerVM2, cellVM1, cellVM2},
				}
				return state, nil
			}

			i.ConvergeFn = func(ctx context.Context, state *magnet.State) error {
				return errors.New("couldn't converge")
			}

			Ω(magnet.Check(context.Background(), i)).ShouldNot(Succeed())
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
			routerVM1 := &magnet.VM{Job: "router", HostUUID: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", HostUUID: host1.ID}

			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2},
				VMs:   []*magnet.VM{routerVM1, routerVM2},
			}
			Ω(magnet.IsBalanced(state)).Should(BeFalse())
		})

		It("detects a balanced state when two jobs are on separate hosts", func() {
			host1 := &magnet.Host{ID: "host1"}
			host2 := &magnet.Host{ID: "host2"}
			routerVM1 := &magnet.VM{Job: "router", HostUUID: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", HostUUID: host2.ID}

			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2},
				VMs:   []*magnet.VM{routerVM1, routerVM2},
			}
			Ω(magnet.IsBalanced(state)).Should(BeTrue())
		})

		It("detects a balanced state when 4 jobs are split on separate hosts", func() {
			host1 := &magnet.Host{ID: "host1"}
			host2 := &magnet.Host{ID: "host2"}
			routerVM1 := &magnet.VM{Job: "router", HostUUID: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", HostUUID: host2.ID}
			cellVM1 := &magnet.VM{Job: "diego_cell", HostUUID: host1.ID}
			cellVM2 := &magnet.VM{Job: "diego_cell", HostUUID: host2.ID}

			// throw some extra jobs in the mix
			clockGlobalVM := &magnet.VM{Job: "clock_global", HostUUID: host2.ID}
			cloudControllerVM := &magnet.VM{Job: "cloud_controller", HostUUID: host2.ID}
			haproxyVM := &magnet.VM{Job: "ha_proxy", HostUUID: host1.ID}

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
			routerVM1 := &magnet.VM{Job: "router", HostUUID: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", HostUUID: host1.ID}
			cellVM1 := &magnet.VM{Job: "diego_cell", HostUUID: host2.ID}
			cellVM2 := &magnet.VM{Job: "diego_cell", HostUUID: host2.ID}

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

			routerVM1 := &magnet.VM{Job: "router", HostUUID: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", HostUUID: host2.ID}
			routerVM3 := &magnet.VM{Job: "router", HostUUID: host3.ID}
			routerVM4 := &magnet.VM{Job: "router", HostUUID: host1.ID}
			routerVM5 := &magnet.VM{Job: "router", HostUUID: host2.ID}

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

			routerVM1 := &magnet.VM{Job: "router", HostUUID: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", HostUUID: host2.ID}

			// put 3 routers on a single host, when the best-case
			// scenario is a max of 2 routers per host
			routerVM3 := &magnet.VM{Job: "router", HostUUID: host3.ID}
			routerVM4 := &magnet.VM{Job: "router", HostUUID: host3.ID}
			routerVM5 := &magnet.VM{Job: "router", HostUUID: host3.ID}

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

			routerVM1 := &magnet.VM{Job: "router", HostUUID: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", HostUUID: host2.ID}
			routerVM3 := &magnet.VM{Job: "router", HostUUID: host3.ID}

			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2, host3, host4, host5},
				VMs:   []*magnet.VM{routerVM1, routerVM2, routerVM3},
			}
			Ω(magnet.IsBalanced(state)).Should(BeTrue())
		})
	})

	Context("RuleRecommendations (2 hosts)", func() {
		var (
			recommendations                   *magnet.RuleRecommendation
			validRule, missingRule, bogusRule *magnet.Rule
		)
		BeforeEach(func() {
			host1 := &magnet.Host{ID: "host1"}
			host2 := &magnet.Host{ID: "host2"}

			routerVM1 := &magnet.VM{Job: "router", HostUUID: host1.ID}
			routerVM2 := &magnet.VM{Job: "router", HostUUID: host1.ID}

			cellVM1 := &magnet.VM{Job: "diego_cell", HostUUID: host2.ID}
			cellVM2 := &magnet.VM{Job: "diego_cell", HostUUID: host2.ID}

			clockGlobalVM := &magnet.VM{Job: "clock_global", HostUUID: host1.ID}

			validRule = &magnet.Rule{Name: "diego_cell", Enabled: true, Mandatory: true, VMs: []*magnet.VM{cellVM1, cellVM2}}
			missingRule = &magnet.Rule{Name: "router", Enabled: true, Mandatory: true, VMs: []*magnet.VM{routerVM1, routerVM2}}
			bogusRule = &magnet.Rule{Name: "bogus", Enabled: true, Mandatory: true, VMs: []*magnet.VM{routerVM1, clockGlobalVM}}

			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2},
				VMs:   []*magnet.VM{routerVM1, routerVM2, cellVM1, cellVM2, clockGlobalVM},
				Rules: []*magnet.Rule{validRule, bogusRule},
			}
			recommendations = magnet.RuleRecommendations(state)
		})

		It("Identifies missing rules", func() {
			Ω(recommendations.Missing).Should(ConsistOf(*missingRule))
		})

		It("Identifies valid rules", func() {
			Ω(recommendations.Valid).Should(ConsistOf(*validRule))
		})

		It("Identifies stale rules", func() {
			Ω(recommendations.Stale).Should(ConsistOf(*bogusRule))
		})
	})

	Context("RuleRecommendations (incorrect/stale rule)", func() {
		var (
			recommendations        *magnet.RuleRecommendation
			staleRule1, staleRule2 *magnet.Rule
			routerVM1, routerVM2   *magnet.VM
			cellVM1, cellVM2       *magnet.VM
		)
		BeforeEach(func() {
			host1 := &magnet.Host{ID: "host1"}
			host2 := &magnet.Host{ID: "host2"}

			routerVM1 = &magnet.VM{Job: "router", HostUUID: host1.ID}
			routerVM2 = &magnet.VM{Job: "router", HostUUID: host1.ID}

			cellVM1 = &magnet.VM{Job: "diego_cell", HostUUID: host2.ID}
			cellVM2 = &magnet.VM{Job: "diego_cell", HostUUID: host2.ID}

			staleRule1 = &magnet.Rule{
				Name:      "router",
				Enabled:   true,
				Mandatory: true,
				VMs:       []*magnet.VM{routerVM1, cellVM1},
			}
			staleRule2 = &magnet.Rule{
				Name:      "diego_cell",
				Enabled:   true,
				Mandatory: true,
				VMs:       []*magnet.VM{routerVM2, cellVM2},
			}

			// create a state that has 2 rules, but for the wrong VMs
			state := &magnet.State{
				Hosts: []*magnet.Host{host1, host2},
				VMs:   []*magnet.VM{routerVM1, routerVM2, cellVM1, cellVM2},
				Rules: []*magnet.Rule{staleRule1, staleRule2},
			}
			recommendations = magnet.RuleRecommendations(state)
		})

		It("Identifies valid rules", func() {
			Ω(recommendations.Valid).Should(BeEmpty())
		})

		It("Identifies missing rules", func() {
			router := magnet.Rule{
				Name:      "router",
				Enabled:   true,
				Mandatory: true,
				VMs:       []*magnet.VM{routerVM1, routerVM2},
			}
			diegoCell := magnet.Rule{
				Name:      "diego_cell",
				Enabled:   true,
				Mandatory: true,
				VMs:       []*magnet.VM{cellVM1, cellVM2},
			}
			Ω(recommendations.Missing).Should(ConsistOf(router, diegoCell))
		})

		It("Identifies stale rules", func() {
			Ω(recommendations.Stale).Should(ConsistOf(*staleRule1, *staleRule2))
		})
	})
})
