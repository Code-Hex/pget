package pget

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
)

func getPartialDirname(targetDir, filename string, procs int) string {
	if targetDir == "" {
		return fmt.Sprintf("_%s.%d", filename, procs)
	}
	return filepath.Join(targetDir, fmt.Sprintf("_%s.%d", filename, procs))
}

// checkProgress In order to confirm the degree of progress
func checkProgress(dirname string) (int64, error) {
	return subDirsize(dirname)
}

func subDirsize(dirname string) (int64, error) {
	var size int64
	err := filepath.Walk(dirname, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})

	return size, err
}

func makeRange(i, procs int, rangeSize, contentLength int64) Range {
	low := rangeSize * int64(i)
	if i == procs-1 {
		return Range{
			low:  low,
			high: contentLength,
		}
	}
	return Range{
		low:  low,
		high: low + rangeSize - 1,
	}
}

func (r Range) BytesRange() string {
	return fmt.Sprintf("bytes=%d-%d", r.low, r.high)
}

func progressBar(ctx context.Context, contentLength int64, dirname string) error {
	bar := pb.Start64(contentLength).SetWriter(stdout).Set(pb.Bytes, true)
	defer bar.Finish()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(100 * time.Millisecond): // To save cpu resource
			size, err := checkProgress(dirname)
			if err != nil {
				return errors.Wrap(err, "failed to get directory size")
			}

			if size < contentLength {
				bar.SetCurrent(size)
			} else {
				bar.SetCurrent(contentLength)
				return nil
			}
		}
	}
}
