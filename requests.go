package pget

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// Range struct for range access
type Range struct {
	low    uint64
	high   uint64
	worker uint64
}

func (p Pget) isNotLastURL(url string) bool {
	return url != p.url && url != ""
}

// Checking is check to can request
func (p *Pget) Checking() error {

	url := p.url

	// checking
	client := http.Client{
		Timeout: time.Duration(p.timeout) * time.Second,
	}
	res, err := client.Get(url)
	if err != nil {
		return errors.Wrap(err, "failed to head request: "+url)
	}
	defer res.Body.Close()

	if res.Header.Get("Accept-Ranges") != "bytes" {
		return errors.Errorf("not supported range access: %s", url)
	}

	// To perform with the correct "range access"
	// get the last url in the redirect
	_url := res.Request.URL.String()
	if p.isNotLastURL(_url) {
		p.url = _url
	}

	// get of ContentLength
	size := res.ContentLength
	if size <= 0 {
		return errors.New("invalid content length")
	}

	p.SetFileSize(uint64(size))

	return nil
}

func (p *Pget) download() error {

	fmt.Fprintf(os.Stdout, "Download start %s\n", p.url)

	procs := uint64(p.procs)

	filesize := p.FileSize()
	filename := p.FileName()
	dirname := p.DirName()

	// create download location
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return errors.Wrap(err, "failed to mkdir for download location")
	}

	// calculate split file size
	split := filesize / procs

	if err := p.Utils.IsFree(split); err != nil {
		return err
	}

	ctx, cancelAll := context.WithCancel(context.Background())

	chErr := make(chan error)
	chDone := make(chan bool)

	totalActiveProcs := 0
	for i := uint64(0); i < procs; i++ {
		partName := fmt.Sprintf("%s/%s.%d.%d", dirname, filename, procs, i)
		info, err := os.Stat(partName)
		r := p.Utils.MakeRange(i, split, procs)
		if err == nil {
			//check if the part is fully downloaded
			if uint64(info.Size()) == split {
				// skip as the part is already downloaded
				continue
			}

			// make low range from this next byte
			r.low += uint64(info.Size()) + 1
		}
		totalActiveProcs += 1
		go func(r Range) {
			if err := p.requests(ctx, r, filename, dirname); err != nil {
				chErr <- err
			}
			chDone <- true
		}(r)
	}

	go p.Utils.ProgressBar(ctx, chErr, chDone)

	// listen for error or done channel
	for ch := 0; ch < totalActiveProcs; ch++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-chErr:
			cancelAll()
			return err
		case <-chDone:
		}
	}

	close(chErr)
	close(chDone)

	return nil
}

func (p Pget) requests(ctx context.Context, r Range, filename, dirname string) error {

	res, err := p.MakeResponse(ctx, r.low, r.high, r.worker) // ctxhttp
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to split get requests: %d", r.worker))
	}

	defer res.Body.Close()

	partName := fmt.Sprintf("%s/%s.%d.%d", dirname, filename, p.procs, r.worker)
	output, err := os.OpenFile(partName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create %s in %s", filename, dirname))
	}

	defer output.Close()

	io.Copy(output, res.Body)

	return nil
}

// MakeResponse return *http.Response include context and range header
func (p Pget) MakeResponse(ctx context.Context, low, high, worker uint64) (*http.Response, error) {
	// create get request
	req, err := http.NewRequest("GET", p.url, nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to split NewRequest for get: %d", worker))
	}

	// set download ranges
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", low, high))
	client := new(http.Client)

	return ctxhttp.Do(ctx, client, req)
}
