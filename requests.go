package pget

import (
	"context"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var ErrNotSupportRequestRange = errors.New("does not support range request")

func ISNotSupportRequestRange(e error) bool {
	return strings.Contains(e.Error(), ErrNotSupportRequestRange.Error())
}

// Range struct for range access
type Range struct {
	low  int64
	high int64
}

func isNotLastURL(url, purl string) bool {
	return url != purl && url != ""
}

// CheckConfig is a configuration to check download target.
type CheckConfig struct {
	URLs    []string
	Timeout time.Duration
	Client  *http.Client
}

// Target represensts download target.
type Target struct {
	Filename      string
	ContentLength int64
	URLs          []string
}

// Check checks be able to download from targets
func Check(ctx context.Context, c *CheckConfig) (*Target, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	if len(c.URLs) == 0 {
		return nil, errors.New("URL is required at least one")
	}

	client := newClient(c.Client)

	infos, err := getMirrorInfos(ctx, client, c.URLs)
	if err != nil {
		return nil, err
	}

	if err := checkEachContent(infos); err != nil {
		return nil, err
	}

	urls := make([]string, len(infos))
	for i, info := range infos {
		urls[i] = info.RetrievedURL
	}

	return &Target{
		Filename:      path.Base(infos[0].RetrievedURL),
		ContentLength: infos[0].ContentLength,
		URLs:          urls,
	}, nil
}

func getMirrorInfos(ctx context.Context, client *http.Client, urls []string) ([]*mirrorInfo, error) {
	var mu sync.Mutex
	eg, ctx := errgroup.WithContext(ctx)

	infos := make([]*mirrorInfo, 0, len(urls))

	for _, url := range urls {
		url := url
		eg.Go(func() error {
			info, err := getMirrorInfo(ctx, client, url)
			if err != nil {
				return errors.Wrap(err, url)
			}

			mu.Lock()
			infos = append(infos, info)
			mu.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return infos, nil
}

type mirrorInfo struct {
	RetrievedURL  string
	ContentLength int64
}

func getMirrorInfo(ctx context.Context, client *http.Client, url string) (*mirrorInfo, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make head request")
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to head request")
	}

	if resp.Header.Get("Accept-Ranges") != "bytes" {
		return nil, ErrNotSupportRequestRange
	}

	if resp.ContentLength <= 0 {
		return nil, errors.New("invalid content length")
	}

	// To perform with the correct "range access"
	// get the last url in the redirect
	_url := resp.Request.URL.String()
	if isNotLastURL(_url, url) {
		return &mirrorInfo{
			RetrievedURL:  _url,
			ContentLength: resp.ContentLength,
		}, nil
	}

	return &mirrorInfo{
		RetrievedURL:  url,
		ContentLength: resp.ContentLength,
	}, nil
}

// check contents are the same on each mirrors
func checkEachContent(infos []*mirrorInfo) error {
	var contentLength int64
	for _, info := range infos {
		if contentLength == 0 {
			contentLength = info.ContentLength
			continue
		}
		if contentLength != info.ContentLength {
			return errors.New("does not match content length on each mirrors")
		}
	}
	return nil
}
