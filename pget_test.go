package pget

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPget(t *testing.T) {
	// listening file server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/moo", http.StatusFound)
	})

	mux.HandleFunc("/moo", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/mooo", http.StatusFound)
	})

	mux.HandleFunc("/mooo", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/test.tar.gz", http.StatusFound)
	})

	mux.HandleFunc("/test.tar.gz", func(w http.ResponseWriter, r *http.Request) {
		fp := "_testdata/test.tar.gz"
		data, err := ioutil.ReadFile(fp)
		if err != nil {
			t.Errorf("failed to readfile: %s", err)
		}
		http.ServeContent(w, r, fp, time.Now(), bytes.NewReader(data))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// begin tests
	url := ts.URL
	TChecking(t, url)
	TDownload(t, url)
	TBindwithFiles(t)
}

func TChecking(t *testing.T, url string) {
	fmt.Fprintf(os.Stdout, "Testing checking_test\n")

	p := New()
	p.URLs = append(p.URLs, url)

	if err := p.Checking(); err != nil {
		t.Errorf("failed to check header: %s", err)
	}

	// could redirect?
	assert.NotEqual(t, p.TargetURLs[0], url, "failed to get of the last url in the redirect")
	fmt.Fprintf(os.Stdout, "checking_test Done\n\n")
}

func TDownload(t *testing.T, url string) {
	fmt.Fprintf(os.Stdout, "Testing download_test\n")

	dir := fmt.Sprintf("_test.tar.gz.%d", runtime.NumCPU())
	p := New()

	p.URLs = append(p.URLs, url)
	p.Utils = &Data{
		filename: "test.tar.gz",
		dirname:  "_test.tar.gz",
	}

	if err := p.Checking(); err != nil {
		t.Errorf("failed to check header: %s", err)
	}

	assert.Equal(t, p.FileName(), "test.tar.gz", "expected 'test.tar.gz' got %s", p.FileName())
	assert.Equal(t, p.DirName(), dir, "expected '"+dir+"' got %s", p.DirName())
	assert.Equal(t, p.FileSize(), uint(1719652), "expected '1719652' got %d", p.DirName())

	p.Procs = runtime.NumCPU()

	if err := p.Download(); err != nil {
		t.Errorf("failed to download: %s", err)
	}

	// check of the file to exists
	for i := 0; i < p.Procs; i++ {
		_, err := os.Stat(fmt.Sprintf("_test.tar.gz.4/test.tar.gz.2.%d", i))
		if err == nil {
			t.Errorf("file not exist: %s", err)
		}
	}

	fmt.Fprintf(os.Stdout, "download_test Done\n\n")
}

func TBindwithFiles(t *testing.T) {
	fmt.Fprintf(os.Stdout, "Testing bind_test\n")

	dir := fmt.Sprintf("_test.tar.gz.%d", runtime.NumCPU())

	p := New()
	p.Procs = runtime.NumCPU()

	p.Utils = &Data{
		filename:     "test.tar.gz",
		filesize:     uint(1719652),
		dirname:      dir,
		fullfilename: "test.tar.gz",
	}

	fp := "_testdata/test.tar.gz"
	original, err := get2md5(fp)

	if err != nil {
		t.Errorf("failed to md5sum of original file: %s", err)
	}

	if err := p.BindwithFiles(p.Procs); err != nil {
		t.Errorf("failed to BindwithFiles: %s", err)
	}

	resultfp, err := get2md5(p.FileName())
	if err != nil {
		t.Errorf("failed to md5sum of result file: %s", err)
	}

	assert.Equal(t, original, resultfp, "expected %s got %s", original, resultfp)

	if err := os.Remove(p.FileName()); err != nil {
		t.Errorf("failed to remove of result file: %s", err)
	}

	fmt.Fprintf(os.Stdout, "bind_test Done\n\n")
}

func get2md5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer f.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}

	// get the 16 bytes hash
	bytes := hash.Sum(nil)[:16]

	return hex.EncodeToString(bytes), nil
}
