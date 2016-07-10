package pget

import "golang.org/x/net/context"

type Ch struct {
	Err  chan error
	Done chan bool
}

func (ch *Ch) Close() {
	close(ch.Err)
	close(ch.Done)
}

func (ch *Ch) Listen(ctx context.Context, cancelAll context.CancelFunc, totalActiveProcs int) (e error) {

	for i := 0; i < totalActiveProcs; i++ {
		select {
		case <-ctx.Done():
			i -= 1
		case err := <-ch.Err:
			cancelAll()
			e = err
		case <-ch.Done:
		}
	}

	return
}
