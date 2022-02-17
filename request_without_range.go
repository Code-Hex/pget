package pget

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/cheggaaa/pb/v3"
)

func (pget *Pget) downloadFiles(ctx context.Context, urls []string, dest string) error {
	defer ctx.Done()
	for _, link := range urls {
		if err := pget.downloadFile(context.Background(), link, dest); err != nil {
			return err
		}
	}
	return nil
}
func (pget *Pget) downloadFile(ctx context.Context, url string, dest string) error {

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

	file := getFilename(resp)
	var path = filepath.Join(dest, file)
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rd)
	return err
}

func getFilename(resp *http.Response) string {
	var fileName = path.Base(resp.Request.URL.String())
	var p = regexp.MustCompile(`.+filename="(.+?)".*`)
	var contentDisposition = resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		var m = p.FindAllStringSubmatch(contentDisposition, -1)
		if len(m) > 0 && len(m[0]) > 0 {
			fileName = m[0][0]
		}
	}
	return fileName
}
