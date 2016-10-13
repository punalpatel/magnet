package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/pivotalservices/magnet"
	"github.com/pivotalservices/magnet/vsphere"
)

var Version = "dev"

var (
	ver  = flag.Bool("v", false, "print the version")
	poll = flag.Int("p", 5, "polling period (minutes)")
)

func main() {
	flag.Parse()
	if *ver {
		printVersion()
		return
	}

	v, err := vsphere.New()
	if err != nil {
		exit(err)
	}
	d := &magnet.Daemon{IaaS: v, Period: *poll}
	err = d.Run(context.Background())
	if err != nil {
		exit(err)
	}
	os.Exit(0)
}

func printVersion() {
	fmt.Println(Version)
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}
