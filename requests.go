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

func isNotLastURL(url, purl string) bool {
	return url != purl && url != ""
}

func isLastProc(i, procs uint64) bool {
	return i == procs-1
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
	if isNotLastURL(_url, p.url) {
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

	ch := &Ch{
		Err:  make(chan error),
		Done: make(chan bool),
	}
	defer ch.Close()

	totalActiveProcs := 1 // 1 is progressbar

	// on an assignment for request
	p.assignment(&totalActiveProcs, ctx, procs, split, ch)

	go p.Utils.ProgressBar(ctx, ch)

	// listen for error or done channel
	if err := ch.Listen(ctx, cancelAll, totalActiveProcs); err != nil {
		return err
	}

	return nil
}

func (p Pget) assignment(totalActiveProcs *int, ctx context.Context, procs, split uint64, ch *Ch) {
	filename := p.FileName()
	dirname := p.DirName()

	for i := uint64(0); i < procs; i++ {
		partName := fmt.Sprintf("%s/%s.%d.%d", dirname, filename, procs, i)
		r := p.Utils.MakeRange(i, split, procs)

		if info, err := os.Stat(partName); err == nil {
			infosize := uint64(info.Size())
			//check if the part is fully downloaded
			if isLastProc(i, procs) {
				if infosize == r.high-r.low {
					continue
				}
			} else if infosize == split {
				// skip as the part is already downloaded
				continue
			}

			// make low range from this next byte
			r.low += infosize
		}
		*totalActiveProcs++
		go func(r Range) {
			if err := p.requests(ctx, r, filename, dirname); err != nil {
				ch.Err <- err
			} else {
				ch.Done <- true
			}
		}(r)
	}
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
