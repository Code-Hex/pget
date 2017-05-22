package pget

type causer interface {
	Cause() error
}

type exit interface {
	ExitCode() int
}

type baseError struct {
	err  error
	code int
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
type argumentsError struct{ baseError }

func makeArgumentsError(err error) argumentsError {
	return argumentsError{baseError{err: err, code: 65}}
}

func (a argumentsError) Error() string { return a.err.Error() }
func (a argumentsError) ExitCode() int { return a.code }

// Error for resources
type resourceError struct{ baseError }

func makeResourceError(err error) resourceError {
	return resourceError{baseError{err: err, code: 72}}
}

func (r resourceError) Error() string { return r.err.Error() }
func (r resourceError) ExitCode() int { return r.code }
