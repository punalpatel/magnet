package vsphere

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/kelseyhightower/envconfig"
	"github.com/pivotalservices/magnet"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
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

func (c *vsphereconfig) hostAndPort() string {
	if c.Scheme == "http" && c.Port != "80" {
		return fmt.Sprintf("%s:%s", c.Hostname, c.Port)
	}
	if c.Scheme == "https" && c.Port != "443" {
		return fmt.Sprintf("%s:%s", c.Hostname, c.Port)
	}
	return c.Hostname
}

// IaaS is the vSphere implementation of IaaS.
type IaaS struct {
	URL    *url.URL
	config *vsphereconfig
	client *govmomi.Client
}

func (i *IaaS) Converge(ctx context.Context, state *magnet.State) error {
	return nil
}

func (i *IaaS) State(ctx context.Context) (*magnet.State, error) {
	c, err := govmomi.NewClient(ctx, i.URL, i.config.Insecure)
	if err != nil {
		return nil, err
	}
	i.client = c
	if !i.client.IsVC() {
		return nil, fmt.Errorf("%s is not a vCenter", i.config.hostAndPort())
	}
	fmt.Println("Connected to", i.config.hostAndPort())
	return i.state(ctx, c)
}

func (z *collector) hydrate(ctx context.Context, c *govmomi.Client) {
	if len(z.dcRefs) > 0 {
		c.PropertyCollector().Retrieve(ctx, z.dcRefs, nil, &z.dcs)
		for _, dc := range z.dcs {
			fmt.Println("dc:", dc.Name)
		}
	}
	if len(z.hostRefs) > 0 {
		c.PropertyCollector().Retrieve(ctx, z.hostRefs, nil, &z.hosts)
		for _, host := range z.hosts {
			fmt.Println("host:", host.Name)
		}
	}
	if len(z.vmRefs) > 0 {
		c.PropertyCollector().Retrieve(ctx, z.vmRefs, nil, &z.vms)
		for _, vm := range z.vms {
			fmt.Println("vm:", vm.Name)
		}
	}
	if len(z.clusterRefs) > 0 {
		c.PropertyCollector().Retrieve(ctx, z.clusterRefs, nil, &z.clusters)
		for _, cluster := range z.clusters {
			fmt.Println("cluster:", cluster.Name)
		}
	}
	if len(z.rpRefs) > 0 {
		var rps []mo.ResourcePool
		c.PropertyCollector().Retrieve(ctx, z.rpRefs, nil, &rps)
		var retrieve func(rps []types.ManagedObjectReference)
		retrieve = func(r []types.ManagedObjectReference) {
			if len(r) == 0 {
				return
			}
			var childrps []mo.ResourcePool
			c.PropertyCollector().Retrieve(ctx, r, nil, &childrps)
			rps = append(rps, childrps...)
		}
		for _, rp := range rps {
			if len(rp.ResourcePool) == 0 {
				continue
			}
			retrieve(rp.ResourcePool)
		}
		z.rps = append(z.rps, rps...)
		for _, rp := range rps {
			fmt.Println("rp:", rp.Name)
		}
	}

	z.dcRefs = nil
	z.hostRefs = nil
	z.vmRefs = nil
	z.clusterRefs = nil
	z.rpRefs = nil
}

func (z *collector) enumerate(ctx context.Context, c *govmomi.Client, objs []object.Reference) {
	for i := range objs {
		ref := objs[i]
		switch ref.Reference().Type {
		case "Folder":
			z.folderRefs = append(z.folderRefs, ref.Reference())
			f := object.NewFolder(c.Client, ref.Reference())
			children, err := f.Children(ctx)
			if err == nil {
				z.enumerate(ctx, c, children)
			}
		case "VirtualMachine":
			z.vmRefs = append(z.vmRefs, ref.Reference())
		case "HostSystem":
			z.hostRefs = append(z.hostRefs, ref.Reference())
		case "ClusterComputeResource":
			z.clusterRefs = append(z.clusterRefs, ref.Reference())
		case "ResourcePool":
			z.rpRefs = append(z.rpRefs, ref.Reference())
		case "Datacenter":
			z.dcRefs = append(z.dcRefs, ref.Reference())
		}
	}
}

func debug(i interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "   ")
	enc.Encode(i)
}

type collector struct {
	dcs         []mo.Datacenter
	dcRefs      []types.ManagedObjectReference
	hosts       []mo.HostSystem
	hostRefs    []types.ManagedObjectReference
	vms         []mo.VirtualMachine
	vmRefs      []types.ManagedObjectReference
	clusters    []mo.ClusterComputeResource
	clusterRefs []types.ManagedObjectReference
	rps         []mo.ResourcePool
	rpRefs      []types.ManagedObjectReference
	folders     []mo.Folder
	folderRefs  []types.ManagedObjectReference
}

func jobForVM(vm *mo.VirtualMachine) string {
	if vm == nil || len(vm.Value) == 0 {
		return ""
	}

	fieldKey := int32(-1)
	for _, field := range vm.AvailableField {
		if field.Name == "job" {
			fieldKey = field.Key
		}
	}

	for _, v := range vm.Value {
		if v.GetCustomFieldValue().Key == fieldKey {
			if cv, ok := v.(*types.CustomFieldStringValue); ok {
				return cv.Value
			}
		}
	}
	return ""
}

func (c *collector) toState(ctx context.Context, client *govmomi.Client) (*magnet.State, error) {
	state := &magnet.State{}
	for _, vm := range c.vms {
		v := &magnet.VM{
			ID:      vm.Reference().String(),
			Name:    vm.Name,
			Host:    "",
			Cluster: "",

			Job: jobForVM(&vm),
		}
		if vm.ResourcePool != nil {
			v.ResourcePool = vm.ResourcePool.String()
		}
		debug(v)
		state.VMs = append(state.VMs, v)
	}
	for _, host := range c.hosts {
		state.Hosts = append(state.Hosts, &magnet.Host{
			ID:   host.Reference().String(),
			Name: host.Name,
		})
	}
	for _, cluster := range c.clusters {
		state.Clusters = append(state.Clusters, &magnet.Cluster{
			ID:   cluster.Reference().String(),
			Name: cluster.Name,
		})
	}

	return nil, nil
}

func (i *IaaS) state(ctx context.Context, c *govmomi.Client) (*magnet.State, error) {
	f := find.NewFinder(c.Client, true)
	collector := &collector{}
	objects, err := f.DatacenterList(ctx, "*")
	if err != nil {
		return nil, err
	}
	var refs []object.Reference
	for _, dc := range objects {
		refs = append(refs, object.NewReference(c.Client, dc.Reference()))
	}

	collector.enumerate(ctx, c, refs)
	collector.hydrate(ctx, c)

	for _, dc := range collector.dcs {
		var dcrefs []object.Reference
		d := object.NewDatacenter(c.Client, dc.Reference())
		f.SetDatacenter(d)
		dcFolders, err := d.Folders(ctx)
		if err != nil {
			return nil, err
		}

		hosts, _ := f.HostSystemList(ctx, path.Join(dcFolders.HostFolder.InventoryPath, "*"))
		for _, host := range hosts {
			dcrefs = append(dcrefs, host.Reference())
		}
		rps, _ := f.ResourcePoolList(ctx, "*")
		for _, rp := range rps {
			dcrefs = append(dcrefs, rp.Reference())
		}
		children, err := dcFolders.VmFolder.Children(ctx)
		if err != nil {
			return nil, err
		}
		dcrefs = append(dcrefs, children...)
		children, err = dcFolders.HostFolder.Children(ctx)
		if err != nil {
			return nil, err
		}
		dcrefs = append(dcrefs, children...)
		collector.enumerate(ctx, c, dcrefs)
	}

	collector.hydrate(ctx, c)
	return collector.toState(ctx, c)
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

	uri := fmt.Sprintf("%s://%s:%s@%s/sdk", config.Scheme, url.QueryEscape(config.Username), url.QueryEscape(config.Password), config.hostAndPort())
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	i := &IaaS{URL: parsed, config: &config}
	return i, nil
}
