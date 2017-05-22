package pget

import "github.com/pkg/errors"

type causer interface {
	Cause() error
}

type exit interface {
	ExitCode() int
}

// UnwrapErrors get important message from wrapped error message
func UnwrapErrors(err error) (int, error) {
	for e := err; e != nil; {
		switch e.(type) {
		case exit:
			return e.(exit).ExitCode(), e
		case ignoreError:
			return 0, nil
		case causer:
			e = e.(causer).Cause()
		default:
			return 1, e // default error
		}
	}
	return 0, nil
}

// Error for options: version, usage
type ignoreError struct{}

func makeIgnoreErr() ignoreError  { return ignoreError{} }
func (ignoreError) Error() string { return "" }

// Error for arguments
type argumentsError struct {
	err  error
	code int
}

func makeArgumentsError(err error, message string) argumentsError {
	return argumentsError{
		err:  errors.Wrap(err, message),
		code: 65,
	}
}

func (a argumentsError) Cause() error  { return a.err }
func (a argumentsError) Error() string { return a.err.Error() }
func (a argumentsError) ExitCode() int { return a.code }
