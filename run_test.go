package pget

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	// listening file server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/file", http.StatusFound)
	})

	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		fp := "_testdata/test.tar.gz"
		data, err := ioutil.ReadFile(fp)
		if err != nil {
			t.Errorf("failed to readfile: %s", err)
		}
		http.ServeContent(w, r, fp, time.Now(), bytes.NewReader(data))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// begin test
	fmt.Fprintf(os.Stdout, "Testing pget_test\n")
	url := ts.URL

	os.Args = []string{
		"pget",
		"-p",
		"2",
		url,
	}

	p := New()
	if err := p.Run(); err != nil {
		t.Errorf("failed to Run: %s", err)
	}

	if err := os.Remove(p.FileName()); err != nil {
		t.Errorf("failed to remove of result file: %s", err)
	}

	fmt.Fprintf(os.Stdout, "pget_test Done\n\n")
}
