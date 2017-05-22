package pget

import "github.com/pkg/errors"

type ignore struct {
	err error
}

type cause interface {
	Cause() error
}

// UnwrapError get important message from wrapped error message
func UnwrapError(err error) error {
	for e := err; e != nil; {
		switch e.(type) {
		case ignore:
			return nil
		case cause:
			e = e.(cause).Cause()
		default:
			return e
		}
	}
	return nil
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
