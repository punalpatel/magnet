package magnet_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotalservices/magnet"
	"github.com/pivotalservices/magnet/mock"
)

var _ = Describe("daemon tests", func() {

	Context("when using a mock IaaS", func() {
		var (
			i magnet.IaaS
			d *magnet.Daemon
		)
		BeforeEach(func() {
			i = mock.New()
			d = &magnet.Daemon{IaaS: i}
		})

		It("can be cancelled", func() {
			ctx, cancel := context.WithCancel(context.Background())
			time.AfterFunc(100*time.Millisecond, cancel)

			// we explicitly cancelled, so this is not considered an error
			Ω(d.Run(ctx)).Should(Succeed())
		})

		It("errors when configured to time out", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			err := d.Run(ctx)

			// a timeout is considered an error
			Ω(err).Should(HaveOccurred())
		})
	})
})
