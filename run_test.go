package pget

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/mholt/archiver"
)

func TestRun(t *testing.T) {
	// listening file server
	mux := http.NewServeMux()

	mux.HandleFunc("/file.name", func(w http.ResponseWriter, r *http.Request) {
		fp := "_testdata/test.tar.gz"
		data, err := ioutil.ReadFile(fp)
		if err != nil {
			t.Errorf("failed to readfile: %s", err)
		}
		http.ServeContent(w, r, fp, time.Now(), bytes.NewReader(data))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	url := ts.URL

	if err := copy("_testdata/resume.tar.gz", "resume.tar.gz"); err != nil {
		t.Errorf("failed to copy: %s", err)
	}

	if err := archiver.NewTarGz().Unarchive("resume.tar.gz", "."); err != nil {
		t.Errorf("failed to untargz: %s", err)
	}

	p := New()
	if err := p.Run(context.Background(), version, []string{
		"pget",
		"-p",
		"3",
		fmt.Sprintf("%s/%s", url, "file.name"),
		"--timeout",
		"5",
	}); err != nil {
		t.Errorf("failed to Run: %s", err)
	}

	if err := os.Remove("resume.tar.gz"); err != nil {
		t.Errorf("failed to remove of test file: %s", err)
	}

	tmpDir := t.TempDir()
	if err := p.Run(context.Background(), version, []string{
		"pget",
		path.Join(url, "file.name"),
		"-d",
		tmpDir,
		"--trace",
	}); err != nil {
		t.Errorf("failed to Run: %s", err)
	}

	// check exist file
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("failed to output to destination")
	}
}

func copy(src, dest string) error {
	srcp, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcp.Close()

	dst, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, srcp); err != nil {
		return err
	}

	return nil
}
