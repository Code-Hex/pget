package pget

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
)

func (pget *Pget) DownloadFiles(ctx context.Context, urls []string, dest string) error {
	defer ctx.Done()
	for _, link := range urls {
		if err := pget.downloadFile(context.Background(), link, dest); err != nil {
			return err
		}
	}
	return nil
}
func (pget *Pget) downloadFile(ctx context.Context, url string, dest string) error {
	file := path.Base(url)
	var path = filepath.Join(dest, file)
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	if pget.useragent != "" {
		req.Header.Add("user-agent", pget.useragent)
	}
	if pget.referer != "" {
		req.Header.Add("Referer", pget.referer)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bar := new(pb.ProgressBar)
	if resp.ContentLength > 0 {
		bar.SetTotal(int64(resp.ContentLength))
	}
	bar.SetWriter(stdout).Set(pb.Bytes, true).Start()
	bar.Start()
	defer bar.Finish()
	rd := bar.NewProxyReader(resp.Body)
	_, err = io.Copy(out, rd)
	return err
}
