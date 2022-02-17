package mdns

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() { plugin.Register("mdns", setup) }

func setup(c *caddy.Controller) error {
	m := MDNS{}
	c.Next()
	if c.NextArg() {
		return plugin.Error("mdns", c.ArgErr())
	}

	c.OnStartup(func() error {
		return m.ConnectDBus()
	})

	c.OnShutdown(func() error {
		return m.Disconnect()
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		m.Next = next
		return m
	})

	return nil
}
