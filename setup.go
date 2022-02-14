package referral

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {
	caddy.RegisterPlugin("referral", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {

	a := New()

	handler, err := initForward(c)

	if err != nil {
		return plugin.Error("referral", err)
	}

	a.handlers = append(a.handlers, handler)

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		a.Next = next
		return a
	})

	c.OnShutdown(func() error {
		for _, handler := range a.handlers {
			if err := handler.OnShutdown(); err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}
