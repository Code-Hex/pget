package pget

import (
	"errors"

	"golang.org/x/net/context"
)

// Ch struct for request
type Ch struct {
	Err  chan error
	Size chan uint
	Done chan bool
}

// MakeCh instead of init
func MakeCh() *Ch {
	return &Ch{
		Err:  make(chan error),
		Size: make(chan uint),
		Done: make(chan bool),
	}
}

// Close method will close channel in Ch struct
func (ch *Ch) Close() {
	close(ch.Err)
	close(ch.Size)
	close(ch.Done)
}

// CheckingListen method wait all channels for Checking method in requests
func (ch *Ch) CheckingListen(ctx context.Context, cancelAll context.CancelFunc, totalActiveProcs int) (size uint, e error) {
	ContentLens := uint(0)
	for i := 0; i < totalActiveProcs; i++ {
		select {
		case <-ctx.Done():
			i--
		case err := <-ch.Err:
			if e != nil {
				cancelAll()
			}
			e = err
		case size = <-ch.Size:
			if ContentLens == 0 {
				ContentLens = size
			} else {
				if ContentLens != size {
					if e != nil {
						cancelAll()
					}
					e = errors.New("Not match the file on each mirrors")
				}
			}
		}
	}

	return
}

// DownloadListen method wait all channels for Download method in requests
func (ch *Ch) DownloadListen(ctx context.Context, cancelAll context.CancelFunc, totalActiveProcs int) (e error) {

	for i := 0; i < totalActiveProcs; i++ {
		select {
		case <-ctx.Done():
			i--
		case err := <-ch.Err:
			if e != nil {
				cancelAll()
			}
			e = err
		case <-ch.Done:
		}
	}

	return
}
