package referral

import (
	"context"
	"github.com/coredns/coredns/plugin/forward"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"math/rand"
	"time"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"
	"github.com/miekg/dns"
)

const name = "referral"

var log = clog.NewWithPlugin("referral")

type Referral struct {
	Next     plugin.Handler
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

	if len(r.Question) == 0 {
		log.Debugf("the query had no questions, refusing to serve it...")
		return dns.RcodeRefused, nil
	}

	log.Debugf("resolving query for question %s", r.Question[0].String())

	nw := nonwriter.New(w)
	rcode, err := plugin.NextOrFailure(rf.Name(), rf.Next, ctx, nw, r)

	if nw.Msg != nil && isReferral(nw.Msg) {
		log.Debugf("found extras in the referral response, %d", len(nw.Msg.Extra))
		for _, e := range nw.Msg.Extra {
			log.Debugf("extra: %s", e.String())
		}
		var (
			rcode = dns.RcodeServerFailure
			err   error
		)
		extras := shuffleExtra(nw.Msg.Extra)
		for _, rec := range extras {
			if a, ok := rec.(*dns.A); ok {
				host := a.A.String()
				log.Debugf("the referral ip address, %s", a.A.String())
				rnw := nonwriter.New(w)
				if f, ok := rf.handlers[host]; ok {
					rcode, err = f.ServeDNS(ctx, rnw, r)
				} else {
					f := forward.New()
					p := forward.NewProxy(host+":53", transport.DNS)
					p.SetExpire(defaultExpire)
					f.SetProxy(p)
					rf.handlers[host] = f
					rcode, err = f.ServeDNS(ctx, rnw, r)
				}
				if rnw.Msg != nil && rcode == dns.RcodeSuccess {
					w.WriteMsg(rnw.Msg)
					return rcode, err
				}
			}
		}
		return rcode, err
	}

	if nw.Msg != nil {
		log.Debugf("upstream response code is: %d", nw.Msg.Rcode)
		for _, a := range nw.Msg.Answer {
			log.Debugf("with answer: %s", a.String())
		}
		w.WriteMsg(nw.Msg)
	}

	return rcode, err
}

func isReferral(msg *dns.Msg) bool {
	return len(msg.Answer) == 0 && len(msg.Ns) > 0 && len(msg.Extra) > 1
}

func shuffleExtra(es []dns.RR) []dns.RR {
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(es), func(i, j int) {
		es[i], es[j] = es[j], es[i]
	})
	return es
}
