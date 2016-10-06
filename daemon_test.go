package magnet_test

import (
	"context"
	"errors"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotalservices/magnet"
	"github.com/pivotalservices/magnet/mock"
)

var _ = Describe("daemon tests", func() {
	var (
		i *mock.IaaS
		d *magnet.Daemon
	)
	BeforeEach(func() {
		i = &mock.IaaS{}
		d = &magnet.Daemon{IaaS: i}
	})

	Context("when Run()ing a daemon", func() {
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

		It("returns immediately if already running", func() {
			count := 0
			ctx, cancel := context.WithCancel(context.Background())
			i.StateFn = func(ctx context.Context) (*magnet.State, error) {
				count = count + 1
				<-ctx.Done()
				return nil, errors.New("no")
			}
			var finished, started sync.WaitGroup
			started.Add(2)
			finished.Add(2)
			for j := 0; j < 2; j++ {
				go func() {
					started.Done()
					d.Poll(ctx)
					finished.Done()
				}()
			}
			started.Wait()
			cancel()
			finished.Wait()
			Ω(count).Should(Equal(1))
		})
	})
})
