package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/disk"
)

// IsFree is check your disk space for size needed to download
func IsFree(filesize, procs uint) error {
	free, err := freeSpace()
	if err != nil {
		return err
	}
	// calculate split file size
	split := filesize / procs
	want := filesize + split
	if free < uint64(want) {
		return errors.New("there is not sufficient free space in a disk")
	}
	return nil
}

func freeSpace() (uint64, error) {
	path := func() string {
		if runtime.GOOS == "windows" {
			return "C:\\"
		}
		return "/"
	}()
	usage, err := disk.Usage(path)
	if err != nil {
		return 0, err
	}
	return usage.Free, nil
}

// FileNameFromURL gets from url
func FileNameFromURL(url string) (filepath string) {
	filename := path.Base(url)
	filepath = filename
	// create unique filename
	for i := 1; ; i++ {
		if _, err := os.Stat(filepath); err == nil {
			filepath = fmt.Sprintf("%s-%d", filename, i)
		} else {
			break
		}
	}
	return
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
