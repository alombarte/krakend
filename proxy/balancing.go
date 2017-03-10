package proxy

import (
	"context"
	"net/url"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/sd"
	"github.com/devopsfaith/krakend/sd/dnssrv"
)

// NewRoundRobinLoadBalancedMiddleware creates proxy middleware adding a round robin balancer
func NewRoundRobinLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return newLoadBalancedMiddleware(sd.NewRoundRobinLB(newSubscriber(remote)))
}

// NewRandomLoadBalancedMiddleware creates proxy middleware adding a random balancer
func NewRandomLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return newLoadBalancedMiddleware(sd.NewRandomLB(newSubscriber(remote), time.Now().UnixNano()))
}

func newSubscriber(remote *config.Backend) sd.Subscriber {
	if remote.DNSSVR {
		return dnssrv.New(remote.Host[0])
	}
	return sd.FixedSubscriber(remote.Host)
}

func newLoadBalancedMiddleware(lb sd.Balancer) Middleware {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, request *Request) (*Response, error) {
			host, err := lb.Host()
			if err != nil {
				return nil, err
			}
			r := request.Clone()

			rawURL := []byte{}
			rawURL = append(rawURL, host...)
			rawURL = append(rawURL, r.Path...)
			r.URL, err = url.Parse(string(rawURL))
			if err != nil {
				return nil, err
			}
			r.URL.RawQuery = r.Query.Encode()

			return next[0](ctx, &r)
		}
	}
}
