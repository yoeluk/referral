package referral

import (
	"context"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

const name = "referral"

var log = clog.NewWithPlugin("referral")

type Referral struct {
	Next  plugin.Handler

	handlers []HandlerWithCallbacks
}

// HandlerWithCallbacks interface is made for handling the requests
type HandlerWithCallbacks interface {
	plugin.Handler
	OnStartup() error
	OnShutdown() error
}

func (p *Referral) Name() string {
	return name
}

// New initializes Alternate plugin
func New() (f *Referral) {
	return &Referral{}
}

func (p *Referral) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{Req: r, W: w}
	qName := state.Name()
	qType := state.QType()

	log.Debugf("query name, %s", qName)
	log.Debugf("query type %d", qType)

	return dns.RcodeSuccess, nil
}