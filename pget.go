package pget

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"runtime"

	"github.com/pkg/errors"
)

const version = "0.0.1"

// New for pget package
func New() *Pget {
	return &Pget{
		ARGV:  os.Args,
		Trace: false,
		Utils: &Data{},
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

	if err := pget.BindwithFiles(pget.procs); err != nil {
		return err
	}

	return nil
}

func (pget *Pget) ready() error {
	if procs := os.Getenv("GOMAXPROCS"); procs == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	var opts Options
	if err := pget.parseOptions(&opts, &pget.args, pget.ARGV); err != nil {
		return errors.Wrap(err, "failed to parse command line args")
	}

	if opts.Trace {
		pget.Trace = opts.Trace
	}

	if opts.Procs <= 0 {
		pget.procs = 2
	} else {
		pget.procs = opts.Procs
	}

	if err := pget.parseURLs(); err != nil {
		return errors.Wrap(err, "failed to parse of url")
	}

	if opts.Output != "" {
		pget.SetFileName(opts.Output)
	} else {
		pget.URLFileName(pget.url)
	}

	fmt.Fprintf(os.Stdout, "Checking now %s\n", pget.url)
	if err := pget.Checking(); err != nil {
		return errors.Wrap(err, "faild to check header")
	}

	return nil
}

func (pget Pget) makeIgnoreErr() ignore {
	return ignore{
		err: errors.New("this is ignore message"),
	}
}

// Error for version, usage
func (i ignore) Error() string {
	return i.err.Error()
}

func (i ignore) Cause() error {
	return i.err
}

func (pget *Pget) parseOptions(opts *Options, args *[]string, argv []string) error {

	if len(argv) == 1 {
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
		os.Stdout.Write([]byte("Pget " + version + ", a parallel file download client\n"))
		return pget.makeIgnoreErr()
	}

	*args = o

	return nil
}

func (pget *Pget) parseURLs() error {

	r := regexp.MustCompile(`^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z0-9]{1,4}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)$`)

	// find url in args
	for _, argv := range pget.args {
		if r.MatchString(argv) {
			pget.url = argv
			break
		}
	}

	if pget.url == "" {
		return errors.New("url has not been set in argument")
	}

	u, err := url.Parse(pget.url)
	if err != nil {
		return errors.Wrap(err, "faild to url parse")
	}
	pget.url = u.String()

	return nil
}
