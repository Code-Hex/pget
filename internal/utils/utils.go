package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
	"github.com/ricochet2200/go-disk-usage/du"
)

// IsFree is check your disk space for size needed to download
func IsFree(filesize, procs uint) error {
	// calculate split file size
	split := filesize / procs
	want := filesize + split
	if freeSpace() < want {
		return errors.New("there is not sufficient free space in a disk")
	}
	return nil
}

func freeSpace() uint {
	if runtime.GOOS == "windows" {
		return uint(du.NewDiskUsage("C:\\").Free())
	}
	return uint(du.NewDiskUsage("/").Free())
}

// FileNameFromURL gets from url
func FileNameFromURL(url string) string {
	filename := path.Base(url)
	filepath := filename
	// create unique filename
	for i := 1; ; i++ {
		if _, err := os.Stat(filepath); err == nil {
			filepath = fmt.Sprintf("%s-%d", filename, i)
		} else {
			break
		}
	}
	return filepath
}

// SubDirsize calcs direcory size
func SubDirsize(dirname string) (int64, error) {
	var size int64
	err := filepath.Walk(dirname, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
