package mdns

import (
	"net"
	"strings"
	"syscall"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/mdns/dbusgen"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/godbus/dbus/v5"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

var log = clog.NewWithPlugin("mdns")

type MDNS struct {
	conn     *dbus.Conn
	Next     plugin.Handler
	Resolver *dbusgen.Org_Freedesktop_Resolve1_Manager
}

func (m MDNS) AddARecord(ctx context.Context, msg *dns.Msg, state *request.Request, name string) bool {
	// Add A and AAAA record for name (if it exists) to msg.
	// A records need to be returned in A queries, this function
	// provides common code for doing so.
	addresses, _, _, err := m.Resolver.ResolveHostname(ctx, 0, name, syscall.AF_UNSPEC, 0)

	if err != nil {
		log.Error("Failed to reach systemd resolver: ", err)
		return false
	}

	resolved := false
	for i := 0; i < len(addresses); i++ {
		addr := addresses[i]
		if addr.V1 == syscall.AF_INET && state.QType() == dns.TypeA {
			ip := net.IPv4(addr.V2[0], addr.V2[1], addr.V2[2], addr.V2[3])
			aheader := dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}
			msg.Answer = append(msg.Answer, &dns.A{Hdr: aheader, A: ip})
			resolved = true

		} else if addr.V1 == syscall.AF_INET6 && state.QType() == dns.TypeAAAA {
			var ip net.IP = addr.V2
			aaaaheader := dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 60}
			msg.Answer = append(msg.Answer, &dns.AAAA{Hdr: aaaaheader, AAAA: ip})
			resolved = true
		}
	}
	return resolved
}

func (m MDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	msg := new(dns.Msg)
	state := request.Request{W: w, Req: r}
	hostName := strings.ToLower(state.QName())

	// Prepare message
	msg.SetReply(r)
	msg.Authoritative = true
	msg.RecursionAvailable = true

	// Check requirements
	if !strings.HasSuffix(state.QName(), ".local.") {
		return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
	}

	if state.QType() != dns.TypeA && state.QType() != dns.TypeAAAA {
		return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
	}

	msg.Answer = []dns.RR{}

	if m.AddARecord(ctx, msg, &state, hostName) {
		log.Debug(msg)
		w.WriteMsg(msg)
		return dns.RcodeSuccess, nil
	}

	log.Debugf("No records found for '%s', forwarding to next plugin.", state.QName())
	return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
}

func (m MDNS) ConnectDBus() error {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return err
	}
	m.conn = conn
	m.Resolver = dbusgen.NewOrg_Freedesktop_Resolve1_Manager(conn.Object("org.freedesktop.resolve1", "/org/freedesktop/resolve1"))
	return nil
}

func (m MDNS) Disconnect() error {
	return m.conn.Close()
}

func (m MDNS) Name() string { return "mdns" }
