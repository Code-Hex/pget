package pget

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mholt/archiver"
)

func TestRunResume(t *testing.T) {
	// listening file server
	mux := http.NewServeMux()

	mux.HandleFunc("/file.name", func(w http.ResponseWriter, r *http.Request) {
		fp := filepath.Join("_testdata", "test.tar.gz")
		data, err := os.ReadFile(fp)
		if err != nil {
			t.Errorf("failed to readfile: %s", err)
		}
		http.ServeContent(w, r, fp, time.Now(), bytes.NewReader(data))
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	url := ts.URL
	targetURL := fmt.Sprintf("%s/%s", url, "file.name")
	tmpDir := t.TempDir()

	// resume.tar.gz is included resumable file structures.
	// _file.name.3
	// ├── file.name.3.0
	// ├── file.name.3.1
	// └── file.name.3.2
	resumeFilePath := filepath.Join(tmpDir, "resume.tar.gz")
	if err := copy(
		filepath.Join("_testdata", "resume.tar.gz"),
		resumeFilePath,
	); err != nil {
		t.Fatalf("failed to copy: %s", err)
	}

	if err := archiver.NewTarGz().Unarchive(resumeFilePath, tmpDir); err != nil {
		t.Fatalf("failed to untargz: %s", err)
	}

	p := New()
	if err := p.Run(context.Background(), version, []string{
		"pget",
		"-p",
		"3",
		targetURL,
		"--timeout",
		"5",
		"--output",
		tmpDir,
	}); err != nil {
		t.Errorf("failed to Run: %s", err)
	}

	cmpFileChecksum(t,
		filepath.Join("_testdata", "test.tar.gz"),
		filepath.Join(tmpDir, "file.name"),
	)
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
