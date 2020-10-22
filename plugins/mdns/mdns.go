package mdns

import (
	"strings"
	"sync"
	"time"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/grandcat/zeroconf"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

var log = clog.NewWithPlugin("mdns")

type MDNS struct {
	Next      plugin.Handler
	mutex     *sync.RWMutex
	mdnsHosts *map[string]*zeroconf.ServiceEntry
}

func (m MDNS) AddARecord(msg *dns.Msg, state *request.Request, hosts map[string]*zeroconf.ServiceEntry, name string) bool {
	// Add A and AAAA record for name (if it exists) to msg.
	// A records need to be returned in A queries, this function
	// provides common code for doing so.
	answerEntry, present := hosts[name]
	if present {
		if answerEntry.AddrIPv4 != nil && state.QType() == dns.TypeA {
			aheader := dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}
			// TODO: Support multiple addresses
			msg.Answer = append(msg.Answer, &dns.A{Hdr: aheader, A: answerEntry.AddrIPv4[0]})
		}
		if answerEntry.AddrIPv6 != nil && state.QType() == dns.TypeAAAA {
			aaaaheader := dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 60}
			msg.Answer = append(msg.Answer, &dns.AAAA{Hdr: aaaaheader, AAAA: answerEntry.AddrIPv6[0]})
		}
		return true
	}
	return false
}

func (m MDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	msg := new(dns.Msg)
	state := request.Request{W: w, Req: r}
	mdnsHosts := *m.mdnsHosts
	hostName := strings.ToLower(state.QName())

	// Prepare message
	msg.SetReply(r)
	msg.Authoritative = true
	msg.RecursionAvailable = true

	// Check requirements
	if !strings.HasSuffix(state.QName(), ".local.") {
		log.Debugf("Skipping due to query '%s' not '.local'", state.QName())
		return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
	}

	if state.QType() != dns.TypeA && state.QType() != dns.TypeAAAA {
		log.Debugf("Skipping due to unrecognized query type %v", state.QType())
		return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
	}

	msg.Answer = []dns.RR{}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.AddARecord(msg, &state, mdnsHosts, hostName) {
		log.Debug(msg)
		w.WriteMsg(msg)
		return dns.RcodeSuccess, nil
	}

	log.Debugf("No records found for '%s', forwarding to next plugin.", state.QName())
	return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
}

func (m *MDNS) BrowseMDNS() {
	entriesSrv := make(chan *zeroconf.ServiceEntry)
	mdnsHosts := make(map[string]*zeroconf.ServiceEntry)
	discovery := []string{}

	// Retrieve Services
	go func(results <-chan *zeroconf.ServiceEntry) {
		log.Debug("Retrieving mDNS services")
		for entry := range results {
			serviceName := strings.TrimSuffix(entry.Instance, ".local")
			log.Debugf("Service: %s\n", serviceName)
			discovery = append(discovery, serviceName)
		}
	}(entriesSrv)

	// Get all available services
	queryService("_services._dns-sd._udp", entriesSrv, 10)

	// Discover hosts
	for _, serviceName := range discovery {
		processService(serviceName, mdnsHosts)

		// Update fast the list to get soon a answer
		for k, v := range mdnsHosts {
			if _, found := (*m.mdnsHosts)[k]; !found {
				(*m.mdnsHosts)[k] = v
			}
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	// Clear maps so we don't have stale entries
	for k := range *m.mdnsHosts {
		delete(*m.mdnsHosts, k)
	}
	// Copy values into the shared maps only after we've collected all of them.
	// This prevents us from having to lock during the entire mdns browse time.
	for k, v := range mdnsHosts {
		(*m.mdnsHosts)[k] = v
	}
}

func processService(service string, mdnsHosts map[string]*zeroconf.ServiceEntry) {
	entriesHost := make(chan *zeroconf.ServiceEntry)

	// Retrieve Hosts
	go func(results <-chan *zeroconf.ServiceEntry) {
		log.Debug("Retrieving mDNS entries")
		for entry := range results {
			// Make a copy of the entry so zeroconf can't later overwrite our changes
			localEntry := *entry
			if localEntry.HostName != "" {
				log.Debugf("Instance: %s, HostName: %s, AddrIPv4: %s, AddrIPv6: %s\n", localEntry.Instance, localEntry.HostName, localEntry.AddrIPv4, localEntry.AddrIPv6)
				mdnsHosts[strings.ToLower(localEntry.HostName)] = entry
			} else {
				log.Debugf("Ignore Instance: %v", localEntry)
			}
		}
	}(entriesHost)

	queryService(service, entriesHost, 8)
}

func queryService(service string, channel chan *zeroconf.ServiceEntry, timeout int) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Errorf("Failed to initialize %s resolver: %s", service, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout) * time.Second)
	defer cancel()
	err = resolver.Browse(ctx, service, "local.", channel)
	if err != nil {
		log.Errorf("Failed to browse %s records: %s", service, err.Error())
		return
	}
	<-ctx.Done()
}

func (m MDNS) Name() string { return "mdns" }
