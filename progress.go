package pget

import (
	"context"
	"time"

	"github.com/Code-Hex/pget/internal/utils"

	"github.com/pkg/errors"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// progressBar is to show progressbar
func (o *Object) progressBar(ctx context.Context) func() error {
	dirname := o.tmpDirName
	filesize := int64(o.filesize)
	bar := pb.New64(filesize)
	return func() error {
		bar.Start()
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				size, err := o.Progress(dirname)
				if err != nil {
					return errors.Wrap(err, "failed to get directory size")
				}
				if size < filesize {
					bar.Set64(size)
				} else {
					bar.Set64(filesize)
					bar.Finish()
					return nil
				}

				// To save cpu resource
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// Progress In order to confirm the degree of progress
func (o *Object) Progress(dirname string) (int64, error) {
	return utils.SubDirsize(dirname)
}
