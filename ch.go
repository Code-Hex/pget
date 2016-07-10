package pget

import "golang.org/x/net/context"

// Ch struct for request
type Ch struct {
	Err  chan error
	Done chan bool
}

// Close method will close channel in Ch struct
func (ch *Ch) Close() {
	close(ch.Err)
	close(ch.Done)
}

// Listen method wait all channels
func (ch *Ch) Listen(ctx context.Context, cancelAll context.CancelFunc, totalActiveProcs int) (e error) {

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
