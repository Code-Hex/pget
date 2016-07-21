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
	low    uint
	high   uint
	worker uint
}

func isNotLastURL(url, purl string) bool {
	return url != purl && url != ""
}

func isLastProc(i, procs uint) bool {
	return i == procs-1
}

// Checking is check to can request
func (p *Pget) Checking() error {

	ctx, cancelAll := context.WithTimeout(context.Background(), time.Duration(p.timeout)*time.Second)

	ch := MakeCh()
	defer ch.Close()

	for _, url := range p.URLs {
		fmt.Fprintf(os.Stdout, "Checking now %s\n", url)
		go p.CheckMirrors(ctx, url, ch)
	}

	// listen for error or size channel
	size, err := ch.CheckingListen(ctx, cancelAll, len(p.URLs))
	if err != nil {
		return err
	}

	filename := p.Utils.URLFileName(p.TargetDir, p.TargetURLs[0])
	p.SetFileName(filename)
	p.SetFullFileName(p.TargetDir, filename)
	p.Utils.SetDirName(p.TargetDir, filename, p.Procs)

	p.SetFileSize(size)

	return nil
}

// CheckMirrors method check be able to range access. also get redirected url.
func (p *Pget) CheckMirrors(ctx context.Context, url string, ch *Ch) {

	res, err := ctxhttp.Head(ctx, http.DefaultClient, url)
	if res != nil {
		res.Body.Close()
	}
	if err != nil {
		ch.Err <- errors.Wrap(err, "failed to head request: "+url)
		return
	}

	if res.Header.Get("Accept-Ranges") != "bytes" {
		ch.Err <- errors.Errorf("not supported range access: %s", url)
		return
	}

	// To perform with the correct "range access"
	// get the last url in the redirect
	_url := res.Request.URL.String()
	if isNotLastURL(_url, url) {
		p.TargetURLs = append(p.TargetURLs, _url)
	} else {
		p.TargetURLs = append(p.TargetURLs, url)
	}

	// get of ContentLength
	size := uint(res.ContentLength)
	if size <= 0 {
		ch.Err <- errors.New("invalid content length")
	} else {
		ch.Size <- size
	}
}

// Download method distributes the task to each goroutine for each URL
func (p *Pget) Download() error {

	procs := uint(p.Procs)

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

	ch := MakeCh()
	defer ch.Close()

	totalActiveProcs := 1 // 1 is progressbar

	// on an assignment for request
	p.Assignment(&totalActiveProcs, ctx, procs, split, ch)

	go p.Utils.ProgressBar(ctx, ch)

	// listen for error or done channel
	if err := ch.DownloadListen(ctx, cancelAll, totalActiveProcs); err != nil {
		return err
	}

	return nil
}

// Assignment method that to each goroutine gives the task
func (p Pget) Assignment(totalActiveProcs *int, ctx context.Context, procs, split uint, ch *Ch) {
	filename := p.FileName()
	dirname := p.DirName()

	assignment := uint(p.Procs / len(p.TargetURLs))

	var lasturl string
	for i := uint(0); i < procs; i++ {
		partName := fmt.Sprintf("%s/%s.%d.%d", dirname, filename, procs, i)

		// make range
		r := p.Utils.MakeRange(i, split, procs)

		if info, err := os.Stat(partName); err == nil {
			infosize := uint(info.Size())
			// check if the part is fully downloaded
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

		url := p.TargetURLs[0]

		// give efficiency and equality work
		if uint(*totalActiveProcs-1)%assignment == 0 {
			// Like shift method
			if len(p.TargetURLs) > 1 {
				p.TargetURLs = p.TargetURLs[1:]
			}

			// check whether to output the message
			if lasturl != url {
				fmt.Fprintf(os.Stdout, "Download start from %s\n", url)
				lasturl = url
			}
		}

		// execute get request
		go func(r Range, url string) {
			if err := p.Requests(ctx, r, filename, dirname, url); err != nil {
				ch.Err <- err
			} else {
				ch.Done <- true
			}
		}(r, url)
	}
}

// Requests method will download the file
func (p Pget) Requests(ctx context.Context, r Range, filename, dirname, url string) error {

	res, err := p.MakeResponse(ctx, r.low, r.high, r.worker, url) // ctxhttp
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to split get requests: %d", r.worker))
	}

	defer res.Body.Close()

	partName := fmt.Sprintf("%s/%s.%d.%d", dirname, filename, p.Procs, r.worker)
	output, err := os.OpenFile(partName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create %s in %s", filename, dirname))
	}

	defer output.Close()

	io.Copy(output, res.Body)

	return nil
}

// MakeResponse return *http.Response include context and range header
func (p *Pget) MakeResponse(ctx context.Context, low, high, worker uint, url string) (*http.Response, error) {
	// create get request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to split NewRequest for get: %d", worker))
	}

	// set download ranges
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", low, high))

	// set user agent
	if p.useragent != "" {
		req.Header.Set("User-Agent", p.useragent)
	}

	return ctxhttp.Do(ctx, http.DefaultClient, req)
}
