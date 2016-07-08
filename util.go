package pget

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/cheggaaa/pb.v1"

	"github.com/pkg/errors"
	"github.com/ricochet2200/go-disk-usage/du"
	"golang.org/x/net/context"
)

// Data struct has file of relational data
type Data struct {
	filename string
	filesize uint64
	dirname  string
}

// Utils interface indicate function
type Utils interface {
	ProgressBar(context.Context, chan error, chan bool)
	BindwithFiles(int) error
	IsFree(uint64) error
	MakeRange(uint64, uint64, uint64) Range

	// like setter
	SetFileName(string)
	URLFileName(string)
	SetDirName(string)
	SetFileSize(uint64)

	// like getter
	FileName() string
	FileSize() uint64
	DirName() string
}

// FileName get from Data structs member
func (d Data) FileName() string {
	return d.filename
}

// FileSize get from Data structs member
func (d Data) FileSize() uint64 {
	return d.filesize
}

// DirName get from Data structs member
func (d Data) DirName() string {
	return d.dirname
}

// SetFileSize set to Data structs member
func (d *Data) SetFileSize(size uint64) {
	d.filesize = size
}

// SetFileName set to Data structs member
func (d *Data) SetFileName(output string) {
	d.filename = output
}

// URLFileName set to Data structs member using url
func (d *Data) URLFileName(url string) {
	token := strings.Split(url, "/")

	// get of filename from url
	var original string
	for i := 1; original == ""; i++ {
		original = token[len(token)-i]
	}

	filename := original

	// create uniq filename
	for i := 1; true; i++ {
		if _, err := os.Stat(filename); err == nil {
			filename = fmt.Sprintf("%s-%d", original, i)
		} else {
			break
		}
	}
	d.filename = filename
}

// SetDirName set to Data structs member
func (d *Data) SetDirName(filename string) {
	original := "_" + filename
	dirname := original

	// create uniq directory name
	for i := 1; true; i++ {
		if _, err := os.Stat(dirname); err == nil {
			dirname = fmt.Sprintf("%s-%d", original, i)
		} else {
			break
		}
	}
	d.dirname = dirname
}

func (d Data) isDos() bool {
	return runtime.GOOS == "windows"
}

func (d Data) freeSpace() (freespace uint64) {

	if d.isDos() {
		freespace = du.NewDiskUsage("C:\\").Free()
	} else {
		freespace = du.NewDiskUsage("/").Free()
	}

	return
}

// IsFree is check your disk space for size needed to download
func (d Data) IsFree(split uint64) error {
	want := d.filesize + split
	if d.freeSpace() < want {
		return errors.Errorf("there is not sufficient free space in a disk")
	}

	return nil
}

func (d Data) subDirsize(dirname string) (int64, error) {
	var size int64
	err := filepath.Walk(dirname, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})

	return size, err
}

// MakeRange will return Range struct to download function
func (d *Data) MakeRange(i, split, procs uint64) Range {
	low := split * i
	high := low + split - 1
	if i == procs-1 {
		high = d.FileSize()
	}

	return Range{
		low:    low,
		high:   high,
		worker: i,
	}
}

// ProgressBar is to show progressbar
func (d Data) ProgressBar(ctx context.Context, chErr chan error, chDone chan bool) {
	filesize := int64(d.filesize)
	dirname := d.dirname

	bar := pb.New64(filesize)
	bar.Start()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			size, err := d.subDirsize(dirname)
			if err != nil {
				chErr <- errors.Wrap(err, "failed to get directory size")
				return
			}

			if size < filesize {
				bar.Set64(size)
			} else {
				bar.Set64(filesize)
				bar.Finish()
				chDone <- true
				return
			}

			// To save cpu resource
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// BindwithFiles function for file binding after split download
func (d *Data) BindwithFiles(procs int) error {

	fmt.Println("\nbinding with files...")

	filesize := d.filesize
	filename := d.filename
	dirname := d.dirname

	fh, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "failed to create a file in download location")
	}

	defer fh.Close()

	bar := pb.New64(int64(filesize))
	bar.Start()

	for i := 0; i < procs; i++ {
		f := fmt.Sprintf("%s/%s.%d", dirname, filename, i)
		subfp, err := os.Open(f)
		if err != nil {
			return errors.Wrap(err, "failed to open "+f+" in download location")
		}
		defer subfp.Close()

		proxy := bar.NewProxyReader(subfp)
		io.Copy(fh, proxy)

		// remove a file in download location for join
		if err := os.Remove(f); err != nil {
			return errors.Wrap(err, "failed to remove a file in download location")
		}
	}

	bar.Finish()

	// remove download location
	if err := os.Remove(dirname); err != nil {
		return errors.Wrap(err, "failed to remove download location")
	}

	fmt.Println("Complete")

	return nil
}
