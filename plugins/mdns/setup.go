package mdns

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/mdns/resolve1"

	"github.com/godbus/dbus/v5"
)

func init() { plugin.Register("mdns", setup) }

func setup(c *caddy.Controller) error {
	for c.Next() {
		if c.NextArg() {
			return plugin.Error("mdns", c.ArgErr())
		}
	}

	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return plugin.Error("mdns", err)
	}
	bus_object := conn.Object("org.freedesktop.resolve1", "/org/freedesktop/resolve1")
	resolver := resolve1.NewManager(bus_object)

	m := MDNS{
		Resolver: resolver,
		Ifc:      GetPrimaryInterface(conn.Context(), resolver),
	}

	c.OnShutdown(func() error {
		return conn.Close()
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		m.Next = next
		return m
	})

	return nil
}
