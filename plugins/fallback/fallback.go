// Package fallback implements a fallback plugin for CoreDNS
package fallback

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"

	"github.com/miekg/dns"
)

// Fallback plugin allows an fallback set of upstreams be specified which will be used
// if the plugin chain returns specific error messages.
type Fallback struct {
	Next     plugin.Handler
	rules    map[int]rule
	original bool // At least one rule has "original" flag
	handlers []HandlerWithCallbacks
}

type rule struct {
	original bool
	handler  HandlerWithCallbacks
}

// HandlerWithCallbacks interface is made for handling the requests
type HandlerWithCallbacks interface {
	plugin.Handler
	OnStartup() error
	OnShutdown() error
}

// New initializes Fallback plugin
func New() (f *Fallback) {
	return &Fallback{rules: make(map[int]rule)}
}

// ServeDNS implements the plugin.Handler interface.
func (f Fallback) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// If fallback has original option set for any code then copy original request to use it instead of changed
	var originalRequest *dns.Msg
	if f.original {
		originalRequest = r.Copy()
	}
	nw := nonwriter.New(w)
	rcode, err := plugin.NextOrFailure(f.Name(), f.Next, ctx, nw, r)

	//By default the rulesIndex is equal rcode, so in such way we handle the case
	//when rcode is SERVFAIL and nw.Msg is nil, otherwise we use nw.Msg.Rcode
	//because, for example, for the following cases like NXDOMAIN, REFUSED the rcode is 0 (returned by forward)
	//A forward doesn't return 0 only in case SERVFAIL
	rulesIndex := rcode
	if nw.Msg != nil {
		rulesIndex = nw.Msg.Rcode
	}

	if u, ok := f.rules[rulesIndex]; ok {
		if u.original && originalRequest != nil {
			return u.handler.ServeDNS(ctx, w, originalRequest)
		}
		return u.handler.ServeDNS(ctx, w, r)
	}
	if nw.Msg != nil {
		w.WriteMsg(nw.Msg)
	}
	return rcode, err
}

// Name implements the Handler interface.
func (f Fallback) Name() string { return "fallback" }
