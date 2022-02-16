package pget

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/cheggaaa/pb/v3"
)

func DownloadFiles(urls []string, dest string) error {
	for _, link := range urls {
		if err := DownloadFile(link, dest, context.Background()); err != nil {
			return err
		}
	}
	return nil
}
func DownloadFile(url string, dest string, ctx context.Context) error {
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
	req.Header.Add("user-agent", "curl/7.81.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bar := new(pb.ProgressBar)
	size, err := strconv.Atoi(resp.Header.Get("content-length"))
	// when server return content-length
	if err == nil {
		bar.SetTotal(int64(size))
	}
	bar.SetWriter(stdout).Set(pb.Bytes, true).Start()
	bar.Start()
	defer bar.Finish()
	rd := bar.NewProxyReader(resp.Body)
	_, err = io.Copy(out, rd)
	return err
}
