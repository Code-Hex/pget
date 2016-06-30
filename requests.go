package pget

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/pkg/errors"
)

func (p Pget) isNotLastURL(url string) bool {
	return url != p.url && url != ""
}

// Checking is check to can request
func (p *Pget) Checking() error {

	url := p.url

	// checking
	res, err := http.Head(url)
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

	// make directory for paralell download
	p.SetDirName(filename)

	dirname := p.DirName()

	// create download location
	if err := os.Mkdir(dirname, 0755); err != nil {
		return errors.Wrap(err, "faild to mkdir for download location")
	}

	cerr := make(chan error, procs)

	// calculate split file size
	split := filesize / procs

	if err := p.IsFree(split); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(int(procs))
	for i := uint64(0); i < procs; i++ {
		go func(i uint64) {
			defer wg.Done()
			low := split * i
			high := low + split - 1
			if i == procs-1 {
				high = filesize
			}
			if err := p.requests(i, low, high, filename); err != nil {
				cerr <- err
			}
		}(i)
	}

	// listen to requests error
	go func() {
		for err := range cerr {
			panic(err.Error())
		}
	}()

	if err := p.ProgressBar(); err != nil {
		return err
	}

	wg.Wait()
	close(cerr)

	return nil
}

func (p Pget) requests(i, low, high uint64, filename string) error {

	url := p.url
	dirname := p.DirName()

	// create get request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("faild to split NewRequest for get: %d", i))
	}

	// set download ranges
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", low, high))
	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("faild to split get requests: %d", i))
	}

	defer res.Body.Close()

	output, err := os.Create(fmt.Sprintf("%s/%s.%d", dirname, filename, i))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("faild to create %s in %s", filename, dirname))
	}

	defer output.Close()

	io.Copy(output, res.Body)

	return nil
}
