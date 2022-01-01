package pget

import "github.com/pkg/errors"

type causer interface {
	Cause() error
}

type ignore struct {
	err error
}

func makeIgnoreErr() ignore {
	return ignore{
		err: errors.New("this is ignore message"),
	}
}

// Error for options: version, usage
func (i ignore) Error() string {
	return i.err.Error()
}

func (i ignore) Cause() error {
	return i.err
}

// errTop get important message from wrapped error message
func errTop(err error) error {
	for e := err; e != nil; {
		switch e.(type) {
		case ignore:
			return nil
		case causer:
			e = e.(causer).Cause()
		default:
			return e
		}
	}

	return nil
}
