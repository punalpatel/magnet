package magnet

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"time"
)

// Daemon wraps up the logic for periodically
// checking and rebalancing a deployment.
type Daemon struct {
	IaaS    IaaS
	Period  int
	running int32
}

// Run runs the main daemon loop.  It blocks until
// one of the following conditions are met:
//  - the context is cancelled
//  - the process receives a SIGINT
//
// If the first check fails, Run terminates and returns
// the error.  If subsequent checks fail, Run will
// continue to poll the IaaS and will not return.
func (d *Daemon) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := d.Poll(ctx)
	if err != nil {
		return err
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			if err == context.Canceled {
				return nil
			}
			return err
		case <-time.After(time.Duration(d.Period) * time.Minute):
			ctx2, cancel2 := context.WithTimeout(ctx, 60*time.Second)
			defer cancel2()
			d.Poll(ctx2)
		case <-c:
			cancel()
		}
	}
}

func (d *Daemon) startRunning() bool {
	return atomic.CompareAndSwapInt32(&d.running, 0, 1)
}

func (d *Daemon) stopRunning() {
	atomic.StoreInt32(&d.running, 0)
}

// Poll is a wrapper for Check that ensures that multiple
// invocations of Check won't run concurrently.
func (d *Daemon) Poll(ctx context.Context) error {
	if !d.startRunning() {
		return nil
	}

	defer func() {
		d.stopRunning()
	}()
	err := Check(ctx, d.IaaS)
	if err != nil {
		// Need to log this
		return err
	}

	return nil
}
