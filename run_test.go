package pget

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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

	os.Args = []string{
		"pget",
		"-p",
		"3",
		fmt.Sprintf("%s/%s", url, "file.name"),
		"--timeout",
		"5",
	}

	if err := copy("_testdata/resume.tar.gz", "resume.tar.gz"); err != nil {
		t.Errorf("failed to copy: %s", err)
	}

	if err := archiver.NewTarGz().Unarchive("resume.tar.gz", "."); err != nil {
		t.Errorf("failed to untargz: %s", err)
	}

	p := New()
	if err := p.Run(); err != nil {
		t.Errorf("failed to Run: %s", err)
	}

	if err := os.Remove("resume.tar.gz"); err != nil {
		t.Errorf("failed to remove of test file: %s", err)
	}

	if err := os.Remove(p.FileName()); err != nil {
		t.Errorf("failed to remove of result file: %s", err)
	}
	fmt.Fprintf(os.Stdout, "Done\n")
	fmt.Fprintf(os.Stdout, "Second\n")

	os.Args = []string{
		"pget",
		fmt.Sprintf("%s/%s", url, "file.name"),
		"-d",
		"./target_dir",
		"--trace",
	}

	if err := p.Run(); err != nil {
		t.Errorf("failed to Run: %s", err)
	}

	// check exist file
	if _, err := os.Stat("./target_dir"); os.IsNotExist(err) {
		t.Errorf("failed to output to destination")
	}

	if err := os.RemoveAll("./target_dir"); err != nil {
		t.Errorf("failed to remove of result file: %s", err)
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
