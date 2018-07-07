package pget

import "context"

func (o *Object) checkSizeDifference(cancel context.CancelFunc) {
	var (
		changed  bool
		baseSize uint
	)
	for size := range o.chans.size {
		if !changed {
			changed = true
			baseSize = size
			continue
		}
		if baseSize != size {
			cancel()
			return
		}
	}
	o.filesize = baseSize
	o.chans.setSize <- struct{}{}
	close(o.chans.setSize)
}

func (o *Object) setFileSize() {
	close(o.chans.size)
	<-o.chans.setSize
}
