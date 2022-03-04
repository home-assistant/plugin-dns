package mdns

import (
	"net"
	"strings"
	"syscall"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/mdns/resolve1"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

var log = clog.NewWithPlugin("mdns")

type MDNS struct {
	Next     plugin.Handler
	Resolver *resolve1.Manager
	Ifc      int32
}

func insertAnAnswer(answers []dns.RR, answer dns.RR, index int) []dns.RR {
	return append(answers[:index], append([]dns.RR{answer}, answers[index:]...)...)
}

func (m MDNS) AddARecord(msg *dns.Msg, state *request.Request, name string, addresses []struct {
	V0 int32
	V1 int32
	V2 []byte
}) bool {
	// Add A and AAAA record for name (if it exists) to msg.
	// A records need to be returned in A queries, this function
	// provides common code for doing so.
	// Success is always returned if any answers found, even if they don't match question type
	// A noerror on A and nxdomain on AAAA (or vice versa) breaks some clients (musl)
	if len(addresses) == 0 {
		return false
	}

	ifc_index := 0
	for i := 0; i < len(addresses); i++ {
		addr := addresses[i]
		var ip net.IP = addr.V2
		if addr.V1 == syscall.AF_INET && state.QType() == dns.TypeA {
			aheader := dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}

			if ip.Mask(net.CIDRMask(23, 32)).Equal(net.IPv4(172, 30, 32, 0)) {
				// Prefer an address within hassio network if one is returned by inserting at front
				msg.Answer = insertAnAnswer(msg.Answer, &dns.A{Hdr: aheader, A: ip}, 0)
				ifc_index = 1

			} else if addr.V0 == m.Ifc {
				// Primary interface is next most preferred if we found it
				msg.Answer = insertAnAnswer(msg.Answer, &dns.A{Hdr: aheader, A: ip}, ifc_index)

			} else {
				msg.Answer = append(msg.Answer, &dns.A{Hdr: aheader, A: ip})
			}

		} else if addr.V1 == syscall.AF_INET6 && state.QType() == dns.TypeAAAA {
			aaaaheader := dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 60}
			msg.Answer = append(msg.Answer, &dns.AAAA{Hdr: aaaaheader, AAAA: ip})
		}
	}
	return true
}

func (m MDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	if m.Resolver == nil {
		return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
	}

	msg := new(dns.Msg)
	state := request.Request{W: w, Req: r}
	hostName := strings.ToLower(state.QName())

	// Prepare message
	msg.SetReply(r)
	msg.Authoritative = true
	msg.RecursionAvailable = true

	// Check requirements
	if !(strings.HasSuffix(state.QName(), ".local.") || len(strings.Split(state.QName(), ".")) == 2) {
		return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
	}

	if state.QType() != dns.TypeA && state.QType() != dns.TypeAAAA {
		return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
	}

	msg.Answer = []dns.RR{}
	addresses, _, _, err := m.Resolver.ResolveHostname(ctx, 0, hostName, syscall.AF_UNSPEC, 0)

	if err != nil {
		// Usually the error will say that it couldn't find a host with that name
		// There may be uncommon errors though so not swallowing it while debugging
		log.Debug(err)

	} else if m.AddARecord(msg, &state, hostName, addresses) {
		log.Debug(msg)
		w.WriteMsg(msg)
		return dns.RcodeSuccess, nil
	}

	log.Debugf("No records found for '%s', forwarding to next plugin.", state.QName())
	return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
}

func GetPrimaryInterface(ctx context.Context, resolver *resolve1.Manager) int32 {
	names, _, err := resolver.ResolveAddress(ctx, 0, syscall.AF_INET, []byte{8, 8, 8, 8}, 0)

	if err != nil {
		log.Error("could not locate primary interface due to: ", err)
		return 0
	}
	if len(names) == 0 {
		log.Error("could not locate primary interface, possible network issue")
		return 0
	}
	return names[0].V0
}

func (m MDNS) Name() string { return "mdns" }
