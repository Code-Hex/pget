package pget

type ignore struct{}

func (o *Object) makeIgnoreErr() error {
	return &ignore{}
}

// Error for options: version, usage
func (i *ignore) Error() string {
	return "This messeage is not reach"
}

// getRootErr gets important message from wrapped error message
func (o *Object) getRootErr(err error) error {
	type causer interface {
		Cause() error
	}
	for e := err; e != nil; {
		switch e.(type) {
		case *ignore:
			return nil
		case causer:
			e = e.(causer).Cause()
		default:
			return e
		}
	}
	return nil
}
