package pget

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
)

const (
	version = "0.0.3"
	msg     = "Pget v" + version + ", parallel file download client\n"
)

// Pget structs
type Pget struct {
	Trace bool
	Utils
	procs   int
	args    []string
	url     string
	timeout int
}

type ignore struct {
	err error
}

type cause interface {
	Cause() error
}

// New for pget package
func New() *Pget {
	return &Pget{
		Trace:   false,
		Utils:   &Data{},
		procs:   runtime.NumCPU(), // default
		timeout: 10,
	}
}

// ErrTop get important message from wrapped error message
func (pget Pget) ErrTop(err error) error {
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

// Run execute methods in pget package
func (pget *Pget) Run() error {
	if err := pget.ready(); err != nil {
		return pget.ErrTop(err)
	}

	if err := pget.download(); err != nil {
		return err
	}

	if err := pget.Utils.BindwithFiles(pget.procs); err != nil {
		return err
	}

	return nil
}

func (pget *Pget) ready() error {
	if procs := os.Getenv("GOMAXPROCS"); procs == "" {
		runtime.GOMAXPROCS(pget.procs)
	}

	var opts Options
	if err := pget.parseOptions(&opts, os.Args[1:]); err != nil {
		return errors.Wrap(err, "failed to parse command line args")
	}

	if opts.Trace {
		pget.Trace = opts.Trace
	}

	if opts.Procs > 2 {
		pget.procs = opts.Procs
	}

	if opts.Timeout > 0 {
		pget.timeout = opts.Timeout
	}

	if err := pget.parseURLs(); err != nil {
		return errors.Wrap(err, "failed to parse of url")
	}

	if opts.Output != "" {
		abs, err := filepath.Abs(opts.Output)
		if err != nil {
			return errors.Wrap(err, "failed to parse of output")
		}

		file, path, err := pget.Utils.SplitFilePath(abs)
		if err != nil {
			return errors.Wrap(err, "failed to parse of output")
		}
		pget.Utils.SetFileName(file)

		// directory name use to parallel download
		pget.Utils.SetDirName(path, file, pget.procs)
	} else {
		pget.Utils.URLFileName(pget.url)

		// directory name use to parallel download
		pget.Utils.SetDirName("", pget.Utils.FileName(), pget.procs)
	}

	fmt.Fprintf(os.Stdout, "Checking now %s\n", pget.url)
	if err := pget.Checking(); err != nil {
		return errors.Wrap(err, "failed to check header")
	}

	return nil
}

func (pget Pget) makeIgnoreErr() ignore {
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

func (pget *Pget) parseOptions(opts *Options, argv []string) error {

	if len(argv) == 0 {
		os.Stdout.Write(opts.usage())
		return pget.makeIgnoreErr()
	}

	o, err := opts.parse(argv)
	if err != nil {
		return errors.Wrap(err, "failed to parse command line options")
	}

	if opts.Help {
		os.Stdout.Write(opts.usage())
		return pget.makeIgnoreErr()
	}

	if opts.Version {
		os.Stdout.Write([]byte(msg))
		return pget.makeIgnoreErr()
	}

	if opts.Update {
		result, err := opts.isupdate()
		if err != nil {
			return errors.Wrap(err, "failed to parse command line options")
		}

		os.Stdout.Write(result)
		return pget.makeIgnoreErr()
	}

	pget.args = o

	return nil
}

func (pget *Pget) parseURLs() error {

	// find url in args
	for _, argv := range pget.args {
		if govalidator.IsURL(argv) {
			pget.url = argv
			break
		}
	}

	if pget.url == "" {
		return errors.New("url has not been set in argument")
	}

	return nil
}
