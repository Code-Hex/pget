package pget

import (
	"os"
	"runtime"
	"strings"

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
	procs      int
	args       []string
	timeout    int
	urls       []string
	targetURLs []string
	TargetDir  string
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
	if err := pget.Ready(); err != nil {
		return pget.ErrTop(err)
	}

	if err := pget.Checking(); err != nil {
		return errors.Wrap(err, "failed to check header")
	}

	if err := pget.Download(); err != nil {
		return err
	}

	if err := pget.Utils.BindwithFiles(pget.procs); err != nil {
		return err
	}

	return nil
}

// Ready method define the variables required to Download.
func (pget *Pget) Ready() error {
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

	if opts.TargetDir != "" {
		info, err := os.Stat(opts.TargetDir)
		if err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "target dir is invalid")
			}

			if err := os.MkdirAll(opts.TargetDir, 0755); err != nil {
				return errors.Wrapf(err, "failed to create diretory at %s", opts.TargetDir)
			}

		} else if !info.IsDir() {
			return errors.New("target dir is not a valid directory")
		}
	}
	opts.TargetDir = strings.TrimSuffix(opts.TargetDir, "/")
	pget.TargetDir = opts.TargetDir

	filename := pget.Utils.URLFileName(pget.TargetDir, pget.urls[0])
	pget.SetFileName(filename)
	pget.SetFullFileName(pget.TargetDir, filename)
	pget.Utils.SetDirName(pget.TargetDir, filename, pget.procs)

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
			pget.urls = append(pget.urls, argv)
		}
	}

	if len(pget.urls) < 1 {
		return errors.New("urls not found in the arguments passed")
	}

	return nil
}
