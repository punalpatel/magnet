package magnet

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

type Daemon struct {
	IaaS IaaS
}

func (d *Daemon) Run(ctx context.Context) error {
	err := d.IaaS.Connect()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ticker := time.NewTicker(2 * time.Second)
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
			Check(d.IaaS)
		case <-c:
			cancel()
		}
	}
}

func Check(iaas IaaS) {
	fmt.Println("checking IaaS")
	iaas.Connect()
}
