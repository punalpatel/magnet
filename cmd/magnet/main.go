package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pivotalservices/magnet"
	"github.com/pivotalservices/magnet/vsphere"
)

func main() {
	v, err := vsphere.New()
	if err != nil {
		exit(err)
	}
	d := &magnet.Daemon{IaaS: v}
	err = d.Run(context.Background())
	if err != nil {
		exit(err)
	}
	os.Exit(0)
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}
