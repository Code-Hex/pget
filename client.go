package pget

import (
	"net/http"
	"time"

	"go.mercari.io/go-dnscache"
	"go.uber.org/zap"
)

func httpClient() *http.Client {
	client := http.DefaultClient
	resolver, _ := dnscache.New(3*time.Second, 5*time.Second, zap.NewNop())
	client.Transport = &http.Transport{
		DialContext: dnscache.DialFunc(
			resolver,
			http.DefaultTransport.(*http.Transport).DialContext,
		),
	}
	return client
}
