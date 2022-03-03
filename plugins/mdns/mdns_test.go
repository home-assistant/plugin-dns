package mdns

import (
	"net"
	"strings"
	"syscall"
	"testing"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

func makeAddress(ifc int32, ip net.IP) (address struct {
	V0 int32
	V1 int32
	V2 []byte
}) {
	var family int32
	if ip.To4() == nil {
		family = syscall.AF_INET6
	} else {
		family = syscall.AF_INET
	}
	address.V0 = ifc
	address.V1 = family
	address.V2 = ip
	return
}

func makeAddressList(addr ...struct {
	V0 int32
	V1 int32
	V2 []byte
}) []struct {
	V0 int32
	V1 int32
	V2 []byte
} {
	return addr
}

var (
	ipv4 = makeAddressList(makeAddress(1, net.ParseIP("10.1.1.1")))
	ipv6 = makeAddressList(makeAddress(1, net.ParseIP("2001::1")))
	ha   = makeAddressList(
		makeAddress(1, net.ParseIP("10.1.1.1")),
		makeAddress(2, net.ParseIP("172.17.0.1")),
		makeAddress(3, net.ParseIP("172.30.32.1")),
	)
	multiple = makeAddressList(
		makeAddress(2, net.ParseIP("192.168.1.1")),
		makeAddress(1, net.ParseIP("10.1.1.1")),
	)
)

func answerToString(answer []dns.RR) string {
	answers := []string{}
	for _, ans := range answer {
		answers = append(answers, ans.String())
	}
	return strings.Join(answers, "\n")
}

func TestAddARecord(t *testing.T) {
	testCases := []struct {
		tcase          string
		name           string
		qtype          uint16
		responseWriter dns.ResponseWriter
		ifc            int32
		addresses      []struct {
			V0 int32
			V1 int32
			V2 []byte
		}
		expected    []string
		expectedRet bool
	}{
		{"valid local ipv4", "mymachine.local", dns.TypeA, nilResponseWriter{}, 1, ipv4, []string{"mymachine.local	60	IN	A	10.1.1.1"}, true},
		{"valid local ipv6", "mymachine.local", dns.TypeAAAA, nilResponseWriter{}, 1, ipv6, []string{"mymachine.local	60	IN	AAAA	2001::1"}, true},
		{"priority test - docker, ip, other", "homeassistant.local", dns.TypeA, nilResponseWriter{}, 1, ha, []string{
			"homeassistant.local	60	IN	A	172.30.32.1",
			"homeassistant.local	60	IN	A	10.1.1.1",
			"homeassistant.local	60	IN	A	172.17.0.1",
		}, true},
		{"interface match first", "myservice.local", dns.TypeA, nilResponseWriter{}, 1, multiple, []string{
			"myservice.local	60	IN	A	10.1.1.1",
			"myservice.local	60	IN	A	192.168.1.1",
		}, true},
		{"no interface, keep order", "myservice.local", dns.TypeA, nilResponseWriter{}, 0, multiple, []string{
			"myservice.local	60	IN	A	192.168.1.1",
			"myservice.local	60	IN	A	10.1.1.1",
		}, true},
		{"success on AAAA if only A found", "mymachine.local", dns.TypeAAAA, nilResponseWriter{}, 1, ipv4, []string{}, true},
		{"success on A if only AAAA found", "mymachine.local", dns.TypeA, nilResponseWriter{}, 1, ipv6, []string{}, true},
	}
	for _, tc := range testCases {
		m := MDNS{nil, nil, tc.ifc}
		msg := new(dns.Msg)
		req := new(dns.Msg)
		req.SetQuestion("", tc.qtype)
		state := request.Request{W: tc.responseWriter, Req: req}
		success := m.AddARecord(msg, &state, tc.name, tc.addresses)
		if success != tc.expectedRet {
			t.Errorf("case[%v]: Failed", tc.tcase)
		}
		if success {
			passed := len(msg.Answer) == len(tc.expected)
			if passed {
				for i, addr := range msg.Answer {
					passed = passed && addr.String() == tc.expected[i]
				}
			}
			if !passed {
				t.Errorf("case[%v]: expected %v, got %v", tc.tcase, strings.Join(tc.expected, "\n"), answerToString(msg.Answer))
			}
		}
	}
}

type nilResponseWriter struct {
}

func (nilResponseWriter) LocalAddr() net.Addr {
	return nil
}

func (nilResponseWriter) RemoteAddr() net.Addr {
	return nil
}

func (nilResponseWriter) Close() error {
	return nil
}

func (nilResponseWriter) Hijack() {
}

func (nilResponseWriter) TsigStatus() error {
	return nil
}

func (nilResponseWriter) WriteMsg(msg *dns.Msg) error {
	return nil
}

func (nilResponseWriter) Write(b []byte) (int, error) {
	return 0, nil
}

func (nilResponseWriter) TsigTimersOnly(b bool) {
}
