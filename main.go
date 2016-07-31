package main

import (
	"fmt"

	"github.com/vmware/govmomi"
)

func main() {
	ctx := nil //http://godoc.org/golang.org/x/net/context#Context
	url := nil //http://godoc.org/net/url#Userinfo
	insecure := false

	client, err := govmomi.NewClient(ctx, url, insecure)

	if err != nil {
		panic(fmt.Sprintf("Failed to initialize vmware.govmomi Client: %v", err))
	}

	fmt.Printf("Connected to a vCenter? %v", client.IsVC())
}
