package pget

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

func (o *Object) Head(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new HEAD request")
	}
	return o.client.Do(req.WithContext(ctx))
}

func (o *Object) Get(ctx context.Context, r *ranges, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new GET request")
	}
	// set download ranges
	req.Header.Set("Range", r.String())
	// set useragent
	if o.UserAgent != "" {
		req.Header.Set("User-Agent", o.UserAgent)
	}
	// set referer
	if o.Referer != "" {
		req.Header.Set("Referer", o.Referer)
	}
	return o.client.Do(req.WithContext(ctx))
}

// checkIsCorrectMirrors method checks be able to range access. also get redirected url.
func (o *Object) checkIsCorrectMirrors(ctx context.Context, url string) func() error {
	return func() error {
		res, err := o.Head(ctx, url)
		if err != nil {
			return errors.Wrap(err, "failed to send HEAD reuqest")
		}
		if res.Header.Get("Accept-Ranges") != "bytes" {
			return errors.Errorf("not supported range access: %s", url)
		}

		// To perform with the correct "range access"
		// get the last url in the redirect
		maybeRedirected := res.Request.URL.String()
		if isRedirectedURL(maybeRedirected, url) {
			o.TargetURLs = append(o.TargetURLs, maybeRedirected)
		} else {
			o.TargetURLs = append(o.TargetURLs, url)
		}
		// get of ContentLength
		if res.ContentLength <= 0 {
			return errors.New("invalid content length")
		}
		o.chans.size <- uint(res.ContentLength)
		return nil
	}
}

func isRedirectedURL(maybeRedirected, url string) bool {
	return url != "" && maybeRedirected != url
}

// Requests method will download the file
func (o *Object) Requests(ctx context.Context, r *ranges, url string) func() error {
	return func() error {
		res, err := o.Get(ctx, r, url)
		if err != nil {
			return errors.Wrapf(err, "failed to split get requests: %d", r.worker)
		}

		partName := fmt.Sprintf(
			"%s/%s.%d.%d",
			o.tmpDirName,
			o.filename,
			o.Procs,
			r.worker,
		)

		output, err := os.OpenFile(partName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return errors.Wrapf(err, "failed to create %s in %s", o.filename, o.tmpDirName)
		}
		defer output.Close()

		if _, err := io.Copy(output, res.Body); err != nil {
			return errors.Wrapf(err, "failed to write response body to %s", partName)
		}

		return nil
	}
}
