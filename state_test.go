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
})
