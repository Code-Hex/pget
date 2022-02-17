package pget

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_DownloadFiles(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test.tar.gz", func(w http.ResponseWriter, r *http.Request) {
		fp := "_testdata/test.tar.gz"
		data, err := ioutil.ReadFile(fp)
		if err != nil {
			t.Errorf("failed to readfile: %s", err)
		}
		w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
		http.ServeContent(w, r, fp, time.Now(), bytes.NewReader(data))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// begin tests
	url := ts.URL
	cli := New()
	tmpdir := t.TempDir()
	t.Run("download_without_range", func(t *testing.T) {

		err := cli.downloadFiles(context.Background(), []string{url + "/test.tar.gz"}, tmpdir)
		if err != nil {
			t.Fatal(err)
		}
		// check of the file to exists
		filename := filepath.Join(tmpdir, "test.tar.gz")
		_, err = os.Stat(filename)
		if err != nil {
			t.Errorf("%q does not exist: %v", filename, err)
		}

		cmpFileChecksum(t, "_testdata/test.tar.gz", filename)
	})
}
