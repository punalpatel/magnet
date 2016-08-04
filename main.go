package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"golang.org/x/net/context"

	"github.com/kelseyhightower/envconfig"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type vsphereconfig struct {
	Scheme   string `default:"https"`
	Hostname string `required:"true"`
	Port     string `default:"443"`
	Username string `required:"true"`
	Password string `required:"true"`
	Insecure bool   `default:"false"`
}

func (c *vsphereconfig) HostAndPort() string {
	if c.Scheme == "http" && c.Port != "80" {
		return fmt.Sprintf("%s:%s", c.Hostname, c.Port)
	}
	if c.Scheme == "https" && c.Port != "443" {
		return fmt.Sprintf("%s:%s", c.Hostname, c.Port)
	}
	return c.Hostname
}

func main() {
	var config vsphereconfig
	err := envconfig.Process("vsphere", &config)
	if err != nil {
		log.Fatal(err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	urlString := fmt.Sprintf("%s://%s:%s@%s/sdk", config.Scheme, url.QueryEscape(config.Username), url.QueryEscape(config.Password), config.HostAndPort())

	url, err := url.Parse(urlString)
	if err != nil {
		panic(fmt.Sprintf("Error: %s\n", err))
	}

	c, err := govmomi.NewClient(ctx, url, config.Insecure)

	if err != nil {
		panic(fmt.Sprintf("Failed to initialize vmware.govmomi Client: %v", err))
	}

	fmt.Println(fmt.Sprintf("Connected to a vCenter? %v", c.IsVC()))

	f := find.NewFinder(c.Client, true)

	// Find one and only datacenter
	dc, err := f.DefaultDatacenter(ctx)
	if err != nil {
		exit(err)
	}

	// Make future calls local to this datacenter
	f.SetDatacenter(dc)
	clusters, err := f.ClusterComputeResourceList(ctx, "*")
	fmt.Println("Fetching Clusters...")
	for _, cluster := range clusters {
		fmt.Println(fmt.Sprintf("Cluster: %s", cluster.Name()))
		hosts, _ := cluster.Hosts(ctx)
		for _, host := range hosts {
			fmt.Println(fmt.Sprintf("Host: %s", host.Name()))
		}
	}

	pc := property.DefaultCollector(c.Client)

	// Convert clusters into list of references
	var refs []types.ManagedObjectReference
	for _, cluster := range clusters {
		refs = append(refs, cluster.Reference())
	}

	// Retrieve summary property for all clusters
	var mclusters []mo.ClusterComputeResource
	err = pc.Retrieve(ctx, refs, nil, &mclusters)
	if err != nil {
		exit(err)
	}
	for _, cluster := range mclusters {
		fmt.Println(cluster.Name)
		for _, rule := range cluster.Configuration.Rule {
			ruleinfo := rule.GetClusterRuleInfo()
			fmt.Println(fmt.Sprintf("Rule: %s", ruleinfo.Name))
		}
	}
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}
