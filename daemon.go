package magnet

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"time"
)

type Daemon struct {
	IaaS    IaaS
	running int32
}

func (d *Daemon) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := Check(ctx, d.IaaS)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(10 * time.Second)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			err = ctx.Err()
			if err == context.Canceled {
				return nil
			}
			return err
		case <-ticker.C:
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
