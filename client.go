package pget

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/rs/dnscache"
)

func httpClient() *http.Client {
	client := http.DefaultClient
	resolver := new(dnscache.Resolver)
	client.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
			separator := strings.LastIndex(addr, ":")
			ips, err := resolver.LookupHost(ctx, addr[:separator])
			if err != nil {
				return nil, err
			}
			for _, ip := range ips {
				conn, err = net.Dial(network, ip+addr[separator:])
				if err == nil {
					break
				}
			}
			return
		},
	}
	return client
}
