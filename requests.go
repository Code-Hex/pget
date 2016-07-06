package pget

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

func (p Pget) isNotLastURL(url string) bool {
	return url != p.url && url != ""
}

// Checking is check to can request
func (p *Pget) Checking() error {

	url := p.url

	// checking
	res, err := http.Get(url)
	defer res.Body.Close()

	if err != nil {
		return errors.Wrap(err, "failed to head request: "+url)
	}

	if res.Header.Get("Accept-Ranges") != "bytes" {
		return errors.Errorf("not supported range access: %s", url)
	}

	// To perform to the correct "range access"
	// get of the last url in the redirect
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

	// directory name use to parallel download
	p.SetDirName(filename)

	dirname := p.DirName()

	// create download location
	if err := os.Mkdir(dirname, 0755); err != nil {
		return errors.Wrap(err, "faild to mkdir for download location")
	}

	// calculate split file size
	split := filesize / procs

	if err := p.Utils.IsFree(split); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(int(procs))
	for i := uint64(0); i < procs; i++ {
		go func(i uint64) {
			defer wg.Done()
			r := p.Utils.MakeRange(i, split, procs)
			if err := p.requests(ctx, r, filename); err != nil {
				context.Canceled = err
				cancel()
			}
		}(i)
	}

	if err := p.Utils.ProgressBar(ctx); err != nil {
		context.Canceled = err
		cancel()
	}

	wg.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

func (p Pget) requests(ctx context.Context, r Range, filename string) error {

	dirname := p.DirName()

	res, err := p.MakeResponse(ctx, r.low, r.high, r.worker) // ctxhttp
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("faild to split get requests: %d", r.worker))
	}

	defer res.Body.Close()

	output, err := os.Create(fmt.Sprintf("%s/%s.%d", dirname, filename, r.worker))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("faild to create %s in %s", filename, dirname))
	}

	defer output.Close()

	io.Copy(output, res.Body)

	return nil
}

func (p Pget) MakeResponse(ctx context.Context, low, high, worker uint64) (*http.Response, error) {
	// create get request
	req, err := http.NewRequest("GET", p.url, nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("faild to split NewRequest for get: %d", worker))
	}

	// set download ranges
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", low, high))
	client := new(http.Client)

	return ctxhttp.Do(ctx, client, req)
}
