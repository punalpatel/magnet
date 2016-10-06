package vsphere

import (
	"context"
	"fmt"
	"net/url"

	"github.com/kelseyhightower/envconfig"
	"github.com/pivotalservices/magnet"
	"github.com/vmware/govmomi"
)

type vsphereconfig struct {
	Scheme       string `default:"https"`
	Hostname     string `required:"true"`
	Port         string `default:"443"`
	Username     string `required:"true"`
	Password     string `required:"true"`
	Insecure     bool   `default:"false"`
	Cluster      string `required:"true"`
	ResourcePool string `default:""`
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

type IaaS struct {
	URL    *url.URL
	config *vsphereconfig
}

func (v *IaaS) Connect() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fmt.Printf("Connected to %s\n", v.URL.String())
	c, err := govmomi.NewClient(ctx, v.URL, v.config.Insecure)
	if err != nil {
		return err
	}

	if !c.IsVC() {
		return fmt.Errorf("%s is not a vCenter", v.config.HostAndPort())
	}
	fmt.Println("Connected to", v.config.HostAndPort())
	return nil
}

func (v *IaaS) IsConnected() bool {
	return true
}

func (v *IaaS) Jobs() ([]magnet.Job, error) {
	return nil, nil
}
func (v *IaaS) VMs() ([]magnet.VM, error) {
	return nil, nil
}
func (v *IaaS) Rules() ([]magnet.Rule, error) {
	return nil, nil
}

// New creates an IaaS that connects to the vCenter API.
// It is configured with the following environment variables:
//   - VSPHERE_SCHEME        (default https)
//   - VSPHERE_HOSTNAME      (required)
//   - VSPHERE_PORT          (default 443)
//   - VSPHERE_USERNAME      (required)
//   - VSPHERE_PASSWORD      (required)
//   - VSPHERE_INSECURE      (default false)
//   - VSPHERE_CLUSTER       (required)
//   - VSPHERE_RESOURCE_POOL (default "")
func New() (magnet.IaaS, error) {
	var config vsphereconfig
	err := envconfig.Process("vsphere", &config)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s:%s@%s/sdk", config.Scheme, url.QueryEscape(config.Username), url.QueryEscape(config.Password), config.HostAndPort())
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	v := &IaaS{URL: parsed, config: &config}
	return v, nil
}

func (v *IaaS) Datacenters() (*[]Datacenter, error) {
	return nil, nil
}

type Datacenter struct {
	ID   string
	Name string
}

func (dc *Datacenter) Clusters() (*[]Cluster, error) {
	return nil, nil
}

type Cluster struct {
	ID            string
	ResourcePools *[]ResourcePool
	Rules         *[]magnet.Rule
	VMs           *[]magnet.VM
}

type ResourcePool struct {
	ID   string
	Name string
}

func (dc *Cluster) Clusters() (*[]ResourcePool, error) {
	return nil, nil
}
