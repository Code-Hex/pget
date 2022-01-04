package pget

import (
	"context"
	"net"
	"net/http"
	"runtime"
	"time"
)

func newDownloadClient(maxIdleConnsPerHost int) *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	dialer := newDialRateLimiter(&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	})
	tr.DialContext = dialer.DialContext
	tr.MaxIdleConns = 0 // no limit
	tr.MaxIdleConnsPerHost = maxIdleConnsPerHost
	tr.DisableCompression = true
	tr.ForceAttemptHTTP2 = false
	return &http.Client{
		Transport: tr,
	}
}

func newClient(client *http.Client) *http.Client {
	if client == nil {
		return http.DefaultClient
	}
	return client
}

// Prevents too many dials happening at once, because we've observed that that increases the thread
// count in the app, to several times more than is actually necessary - presumably due to a blocking OS
// call somewhere. It's tidier to avoid creating those excess OS threads.
// Even our change from Dial (deprecated) to DialContext did not replicate the effect of dialRateLimiter.
//
// see: https://github.com/Azure/azure-storage-azcopy/blob/058bd5bc5b970074520e4ee088b15328d888c483/ste/mgr-JobPartMgr.go#L117-L124
type dialRateLimiter struct {
	dialer *net.Dialer
	sem    chan struct{}
}

func newDialRateLimiter(dialer *net.Dialer) *dialRateLimiter {
	// exact value doesn't matter too much, but too low will be too slow,
	// and too high will reduce the beneficial effect on thread count
	const concurrentDialsPerCpu = 10

	return &dialRateLimiter{
		dialer: dialer,
		sem:    make(chan struct{}, concurrentDialsPerCpu*runtime.NumCPU()),
	}
}

func (d *dialRateLimiter) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d.sem <- struct{}{}
	defer func() { <-d.sem }()
	return d.dialer.DialContext(ctx, network, address)
}
