package vsphere

import "fmt"

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
