package mdns

import (
	"sync"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/grandcat/zeroconf"
)

func init() { plugin.Register("mdns", setup) }

func setup(c *caddy.Controller) error {
	mdnsHosts := make(map[string]*zeroconf.ServiceEntry)
	mutex := sync.RWMutex{}
	m := MDNS{mutex: &mutex, mdnsHosts: &mdnsHosts}

	for c.Next() {
		if c.NextArg() {
			return plugin.Error("mdns", c.ArgErr())
		}
	}

	c.OnStartup(func() error {
		go browseLoop(&m)
		return nil
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		m.Next = next
		return m
	})

	return nil
}

func browseLoop(m *MDNS) {
	for {
		m.BrowseMDNS()
		time.Sleep(120 * time.Second)
	}
}
