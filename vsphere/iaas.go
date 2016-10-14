package vsphere

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/pivotalservices/magnet"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// IaaS is the vSphere implementation of IaaS.
type IaaS struct {
	URL    *url.URL
	config *vsphereconfig
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
//   - VSPHERE_RESOURCEPOOL  (default "")
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

// Converge applies the specified reccomendations in order to achieve anti-affinity.
func (i *IaaS) Converge(ctx context.Context, state *magnet.State, rec *magnet.RuleRecommendation) error {
	c, err := govmomi.NewClient(ctx, i.URL, i.config.Insecure)
	if err != nil {
		return err
	}
	if !c.IsVC() {
		return fmt.Errorf("%s is not a vCenter", i.config.hostAndPort())
	}

	// TODO:
	// - lookup cluster (state.RuleContainer)
	// - remove any rules in rec.Stale
	// - add any rules in rec.Missing

	return nil
}

// State gets the current state of the deployment on vSphere.
func (i *IaaS) State(ctx context.Context) (*magnet.State, error) {
	c, err := govmomi.NewClient(ctx, i.URL, i.config.Insecure)
	if err != nil {
		return nil, err
	}
	if !c.IsVC() {
		return nil, fmt.Errorf("%s is not a vCenter", i.config.hostAndPort())
	}
	fmt.Println("Connected to", i.config.hostAndPort())
	return i.state(ctx, c)
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

func (i *IaaS) state(ctx context.Context, client *govmomi.Client) (*magnet.State, error) {
	f := find.NewFinder(client.Client, true)
	collector := &collector{}
	objects, err := f.DatacenterList(ctx, "*")
	if err != nil {
		return nil, err
	}
	var refs []object.Reference
	for _, dc := range objects {
		refs = append(refs, object.NewReference(client.Client, dc.Reference()))
	}

	collector.enumerate(ctx, client, refs)
	collector.hydrate(ctx, client)

	for _, dc := range collector.dcs {
		var dcrefs []object.Reference
		d := object.NewDatacenter(client.Client, dc.Reference())
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
		collector.enumerate(ctx, client, dcrefs)
	}

	collector.hydrate(ctx, client)
	collector.filter(i.config.Cluster, i.config.ResourcePool)
	return collector.toState(ctx, client)
}

type collector struct {
	dcs          []mo.Datacenter
	dcRefs       []types.ManagedObjectReference
	hosts        []mo.HostSystem
	hostRefs     []types.ManagedObjectReference
	vms          []mo.VirtualMachine
	vmRefs       []types.ManagedObjectReference
	vmToHosts    map[string]string // vm reference to host UUID
	hostnames    map[string]string // host UUID to hostname
	clusters     []mo.ClusterComputeResource
	clusterRefs  []types.ManagedObjectReference
	rps          []mo.ResourcePool
	rpRefs       []types.ManagedObjectReference
	folders      []mo.Folder
	folderRefs   []types.ManagedObjectReference
	cluster      *mo.ClusterComputeResource
	resourcepool *mo.ResourcePool
}

func (c *collector) toState(ctx context.Context, client *govmomi.Client) (*magnet.State, error) {
	state := &magnet.State{}
	state.VMContainer = c.cluster.Reference().String()
	state.RuleContainer = c.resourcepool.Reference().String()
	for _, host := range c.hosts {
		state.Hosts = append(state.Hosts, &magnet.Host{
			ID:   host.Reference().String(),
			Name: host.Name,
		})
	}

	vmLookup := make(map[string]*magnet.VM)
	for i := range c.vms {
		job := jobForVM(&c.vms[i])
		if job == "" {
			continue
		}
		uuid := c.vmToHosts[c.vms[i].Reference().Value]
		v := &magnet.VM{
			ID:        c.vms[i].Config.Uuid,
			Reference: c.vms[i].Self.Value,
			Name:      c.vms[i].Name,
			HostUUID:  uuid,
			HostName:  c.hostnames[uuid],
			Job:       job,
		}
		vmLookup[c.vms[i].Self.Value] = v
		state.VMs = append(state.VMs, v)
	}

	for _, cluster := range c.clusters {
		for _, rule := range cluster.Configuration.Rule {
			aa, ok := rule.(*types.ClusterAntiAffinityRuleSpec)
			if !ok {
				continue
			}

			ptrToBool := func(b *bool) bool {
				if b == nil {
					return false
				}
				return *b
			}
			rule := &magnet.Rule{
				Name:      aa.Name,
				ID:        aa.RuleUuid,
				Enabled:   ptrToBool(aa.Enabled),
				Mandatory: ptrToBool(aa.Mandatory),
				VMs:       []*magnet.VM{},
			}

			for _, vm := range aa.Vm {
				rule.VMs = append(rule.VMs, vmLookup[vm.Value])
			}

			state.Rules = append(state.Rules, rule)
		}
	}

	return state, nil
}

func (c *collector) filter(cluster string, resourcepool string) {
	for _, cl := range c.clusters {
		if strings.EqualFold(cl.Name, cluster) {
			c.cluster = &cl
			break
		}
	}

	if c.cluster == nil {
		// TODO: This is invalid; but may result from the renaming of a cluster
		panic("Cannot find cluster")
	}

	for _, r := range c.rps {
		if strings.EqualFold(r.Reference().String(), c.cluster.ResourcePool.Reference().String()) {
			c.resourcepool = &r
			break
		}
	}

	if strings.TrimSpace(resourcepool) != "" {
		var filtered []mo.ResourcePool
		var recurse func(rpRefs []types.ManagedObjectReference)
		recurse = func(rpRefs []types.ManagedObjectReference) {
			for _, r := range rpRefs {
				for _, res := range c.rps {
					if strings.EqualFold(r.String(), res.Reference().String()) {
						filtered = append(filtered, res)
						if res.ResourcePool != nil {
							recurse(res.ResourcePool)
						}
					}
				}
			}
		}
		recurse(c.resourcepool.ResourcePool)
		for _, r := range filtered {
			if strings.EqualFold(r.Name, resourcepool) {
				c.resourcepool = &r
				break
			}
		}
	}

	if c.resourcepool == nil {
		panic("Cannot find resource pool")
	}

	var vms []mo.VirtualMachine
	for _, vm := range c.vms {
		if vm.ResourcePool != nil &&
			strings.EqualFold(vm.ResourcePool.String(), c.resourcepool.Reference().String()) &&
			!strings.HasPrefix(vm.Name, "sc") &&
			!strings.HasPrefix(vm.Name, "tpl") {
			vms = append(vms, vm)
		}
	}

	var hosts []mo.HostSystem
	for _, host := range c.cluster.Host {
		for _, h := range c.hosts {
			if strings.EqualFold(host.String(), h.Reference().String()) {
				hosts = append(hosts, h)
			}
		}
	}

	c.vms = vms
	c.hosts = hosts
}

// properties to retrieve from vCenter
var (
	// https://pubs.vmware.com/vsphere-60/index.jsp#com.vmware.wssdk.apiref.doc/vim.ResourcePool.html
	rpProps = []string{"name", "value", "resourcePool"}

	// https://pubs.vmware.com/vsphere-60/index.jsp#com.vmware.wssdk.apiref.doc/vim.Datacenter.html
	dcProps = []string{"name", "hostFolder", "vmFolder"}

	// https://pubs.vmware.com/vsphere-60/index.jsp#com.vmware.wssdk.apiref.doc/vim.HostSystem.html
	hostProps = []string{"name", "vm", "hardware"}

	// https://pubs.vmware.com/vsphere-60/index.jsp#com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	vmProps = []string{"name", "value", "resourcePool", "availableField", "customValue", "config"}

	// https://pubs.vmware.com/vsphere-60/index.jsp#com.vmware.wssdk.apiref.doc/vim.ClusterComputeResource.html
	clusterProps = []string{"name", "host", "resourcePool", "configuration"}
)

func (c *collector) hydrate(ctx context.Context, client *govmomi.Client) {
	if len(c.rpRefs) > 0 {
		var rps []mo.ResourcePool
		client.PropertyCollector().Retrieve(ctx, c.rpRefs, rpProps, &rps)
		var retrieve func(rps []types.ManagedObjectReference)
		retrieve = func(r []types.ManagedObjectReference) {
			if len(r) == 0 {
				return
			}
			var childrps []mo.ResourcePool
			client.PropertyCollector().Retrieve(ctx, r, rpProps, &childrps)
			rps = append(rps, childrps...)
		}
		for _, rp := range rps {
			if len(rp.ResourcePool) == 0 {
				continue
			}
			retrieve(rp.ResourcePool)
		}
		c.rps = append(c.rps, rps...)
	}

	if len(c.dcRefs) > 0 {
		client.PropertyCollector().Retrieve(ctx, c.dcRefs, dcProps, &c.dcs)
	}
	if len(c.hostRefs) > 0 {
		client.PropertyCollector().Retrieve(ctx, c.hostRefs, hostProps, &c.hosts)
		c.vmToHosts = make(map[string]string)
		c.hostnames = make(map[string]string)
		for _, host := range c.hosts {
			for _, vm := range host.Vm {
				uuid := host.Hardware.SystemInfo.Uuid
				c.vmToHosts[vm.Reference().Value] = uuid
				c.hostnames[uuid] = host.Name
			}
		}
	}
	if len(c.vmRefs) > 0 {
		client.PropertyCollector().Retrieve(ctx, c.vmRefs, vmProps, &c.vms)
	}
	if len(c.clusterRefs) > 0 {
		client.PropertyCollector().Retrieve(ctx, c.clusterRefs, clusterProps, &c.clusters)
	}

	c.dcRefs = nil
	c.hostRefs = nil
	c.vmRefs = nil
	c.clusterRefs = nil
	c.rpRefs = nil
}

func (c *collector) enumerate(ctx context.Context, client *govmomi.Client, objs []object.Reference) {
	for i := range objs {
		ref := objs[i]
		switch ref.Reference().Type {
		case "Folder":
			c.folderRefs = append(c.folderRefs, ref.Reference())
			f := object.NewFolder(client.Client, ref.Reference())
			children, err := f.Children(ctx)
			if err == nil {
				c.enumerate(ctx, client, children)
			}
		case "VirtualMachine":
			c.vmRefs = append(c.vmRefs, ref.Reference())
		case "HostSystem":
			c.hostRefs = append(c.hostRefs, ref.Reference())
		case "ClusterComputeResource":
			c.clusterRefs = append(c.clusterRefs, ref.Reference())
		case "ResourcePool":
			c.rpRefs = append(c.rpRefs, ref.Reference())
		case "Datacenter":
			c.dcRefs = append(c.dcRefs, ref.Reference())
		}
	}
}

func debug(i interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "   ")
	enc.Encode(i)
}
