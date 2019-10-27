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

	// To save cpu resource
	ticker := time.NewTicker(100 * time.Millisecond)
	return func() error {
		bar.Start()
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				size, err := progress(dirname)
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
			}
		}
	}
}

// Progress In order to confirm the degree of progress
var progress = utils.SubDirsize
