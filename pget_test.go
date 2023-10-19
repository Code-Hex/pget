package pget

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	stdout = io.Discard
	os.Exit(m.Run())
}

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
		data, err := os.ReadFile(fp)
		if err != nil {
			t.Errorf("failed to readfile: %s", err)
		}
		http.ServeContent(w, r, fp, time.Now(), bytes.NewReader(data))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// begin tests
	url := ts.URL

	tmpdir := t.TempDir()

	cfg := &DownloadConfig{
		Filename:      "test.tar.gz",
		ContentLength: 1719652,
		Dirname:       tmpdir,
		Procs:         4,
		URLs:          []string{ts.URL},
		Client:        newDownloadClient(1),
	}

	t.Run("check", func(t *testing.T) {
		target, err := Check(context.Background(), &CheckConfig{
			URLs:    []string{url},
			Timeout: 10 * time.Second,
		})

		if err != nil {
			t.Fatalf("failed to check header: %s", err)
		}

		if len(target.URLs) == 0 {
			t.Fatalf("invalid URL length %d", len(target.URLs))
		}

		// could redirect?
		assert.NotEqual(t, target.URLs[0], url, "failed to get of the last url in the redirect")
	})

	t.Run("download", func(t *testing.T) {
		err := Download(context.Background(), cfg)
		if err != nil {
			t.Fatal(err)
		}
		// check of the file to exists
		for i := 0; i < cfg.Procs; i++ {
			filename := filepath.Join(tmpdir, "_test.tar.gz.4", fmt.Sprintf("test.tar.gz.2.%d", i))
			_, err := os.Stat(filename)
			if err == nil {
				t.Errorf("%q does not exist: %v", filename, err)
			}
		}

		cmpFileChecksum(t, "_testdata/test.tar.gz", filepath.Join(tmpdir, cfg.Filename))
	})
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

func cmpFileChecksum(t *testing.T, wantPath, gotPath string) {
	t.Helper()
	want, err := get2md5(wantPath)

	if err != nil {
		t.Fatalf("failed to md5sum of original file: %s", err)
	}

	resultfp, err := get2md5(gotPath)
	if err != nil {
		t.Fatalf("failed to md5sum of result file: %s", err)
	}

	if want != resultfp {
		t.Errorf("expected %s got %s", want, resultfp)
	}
}
