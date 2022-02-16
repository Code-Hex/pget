package pget

/**
 * @website http://albulescu.ro
 * @author Cosmin Albulescu <cosmin@albulescu.ro>
 */

import (
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
		if err := DownloadFile(link, dest); err != nil {
			return err
		}
	}
	return nil
}
func DownloadFile(url string, dest string) error {
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
	req.Header.Add("user-agent", "curl/7.81.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	size, err := strconv.Atoi(resp.Header.Get("content-length"))
	if err != nil {
		return err
	}
	bar := pb.Start64(int64(size)).SetWriter(stdout).Set(pb.Bytes, true)
	defer bar.Finish()
	rd := bar.NewProxyReader(resp.Body)

	_, err = io.Copy(out, rd)

	if err != nil {
		return err
	}

	return nil
}
