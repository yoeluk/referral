package referral

import (
	"context"
	"github.com/coredns/coredns/plugin/forward"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"time"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"
	"github.com/miekg/dns"
)

const name = "referral"

var log = clog.NewWithPlugin("referral")

type Referral struct {
	Next  plugin.Handler
	handlers map[string]HandlerWithCallbacks
}

type HandlerWithCallbacks interface {
	plugin.Handler
	OnStartup() error
	OnShutdown() error
}

func New() (rf *Referral) {
	return &Referral{handlers: make(map[string]HandlerWithCallbacks)}
}

func (rf *Referral) Name() string {
	return name
}

const defaultExpire = 10 * time.Second

func (rf *Referral) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	log.Debugf("resolving query for question %s", r.Question[0].String())

	nw := nonwriter.New(w)
	rcode, err := plugin.NextOrFailure(rf.Name(), rf.Next, ctx, nw, r)

	if nw.Msg != nil && isReferral(nw.Msg) {
		log.Debugf("found extras in the referral response, %d", len(nw.Msg.Extra))
		first := nw.Msg.Extra[0]
		if a, ok := first.(*dns.A); ok {
			host := a.A.String()
			log.Debugf("the ip address is, %s", a.A.String())
			if f, kk := rf.handlers[host]; kk {
				return f.ServeDNS(ctx, w, r)
			}
			f := forward.New()
			p := forward.NewProxy(host+":53", transport.DNS)
			p.SetExpire(defaultExpire)
			f.SetProxy(p)
			rf.handlers[host] = f
			return f.ServeDNS(ctx, nw, r)
		}
	}

	if nw.Msg != nil {
		w.WriteMsg(nw.Msg)
	}

	return rcode, err
}

func isReferral(msg *dns.Msg) bool {
	return len(msg.Answer) == 0 && len(msg.Ns) > 0 && len(msg.Extra) > 0
}