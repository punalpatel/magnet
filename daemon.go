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
	ticker := time.NewTicker(5 * time.Second)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	cancelled := make(chan bool)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				cancelled <- true
			case <-ticker.C:
				Check(d.IaaS)
			case <-c:
				cancel()
			}
		}
	}()
	<-cancelled

	return nil
}

func Check(iaas IaaS) {
	fmt.Println("checking IaaS")
	iaas.Connect()
}
