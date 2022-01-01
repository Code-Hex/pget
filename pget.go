package pget

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
)

// Pget structs
type Pget struct {
	Trace     bool
	Filename  string
	TargetDir string
	Procs     int
	URLs      []string

	args      []string
	timeout   int
	useragent string
	referer   string
}

// New for pget package
func New() *Pget {
	return &Pget{
		Trace:   false,
		Procs:   runtime.NumCPU(), // default
		timeout: 10,
	}
}

// Run execute methods in pget package
func (pget *Pget) Run(ctx context.Context, version string, args []string) error {
	if err := pget.Ready(version, args); err != nil {
		return errTop(err)
	}

	target, err := Check(ctx, &CheckConfig{
		URLs:    pget.URLs,
		Timeout: time.Duration(pget.timeout) * time.Second,
	})
	if err != nil {
		return err
	}

	filename := target.Filename
	if pget.Filename != "" {
		filename = pget.Filename
	}

	opts := []DownloadOption{
		WithUserAgent(pget.useragent),
		WithReferer(pget.referer),
	}

	return Download(ctx, &DownloadConfig{
		Filename:      filename,
		Dirname:       GetDirname(pget.TargetDir, filename, pget.Procs),
		ContentLength: target.ContentLength,
		Procs:         pget.Procs,
		URLs:          target.URLs,
	}, opts...)
}

// Ready method define the variables required to Download.
func (pget *Pget) Ready(version string, args []string) error {
	if procs := os.Getenv("GOMAXPROCS"); procs == "" {
		runtime.GOMAXPROCS(pget.Procs)
	}

	opts, err := pget.parseOptions(args, version)
	if err != nil {
		return errors.Wrap(errTop(err), "failed to parse command line args")
	}

	if opts.Trace {
		pget.Trace = opts.Trace
	}

	if opts.Procs > 2 {
		pget.Procs = opts.Procs
	}

	if opts.Timeout > 0 {
		pget.timeout = opts.Timeout
	}

	if err := pget.parseURLs(); err != nil {
		return errors.Wrap(err, "failed to parse of url")
	}

	if opts.Output != "" {
		pget.Filename = opts.Output
	}

	if opts.UserAgent != "" {
		pget.useragent = opts.UserAgent
	}

	if opts.Referer != "" {
		pget.referer = opts.Referer
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
		opts.TargetDir = strings.TrimSuffix(opts.TargetDir, string(filepath.Separator))
	}
	pget.TargetDir = opts.TargetDir

	return nil
}

func (pget *Pget) parseOptions(argv []string, version string) (*Options, error) {
	var opts Options
	if len(argv) == 0 {
		os.Stdout.Write(opts.usage(version))
		return nil, makeIgnoreErr()
	}

	o, err := opts.parse(argv, version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse command line options")
	}

	if opts.Help {
		os.Stdout.Write(opts.usage(version))
		return nil, makeIgnoreErr()
	}

	if opts.Update {
		result, err := opts.isupdate(version)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse command line options")
		}

		os.Stdout.Write(result)
		return nil, makeIgnoreErr()
	}

	pget.args = o

	return &opts, nil
}

func (pget *Pget) parseURLs() error {

	// find url in args
	for _, argv := range pget.args {
		if govalidator.IsURL(argv) {
			pget.URLs = append(pget.URLs, argv)
		}
	}

	if len(pget.URLs) < 1 {
		fmt.Fprintf(os.Stdout, "Please input url separate with space or newline\n")
		fmt.Fprintf(os.Stdout, "Start download at ^D\n")

		// scanning url from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			scan := scanner.Text()
			urls := strings.Split(scan, " ")
			for _, url := range urls {
				if govalidator.IsURL(url) {
					pget.URLs = append(pget.URLs, url)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return errors.Wrap(err, "failed to parse url from stdin")
		}

		if len(pget.URLs) < 1 {
			return errors.New("urls not found in the arguments passed")
		}
	}

	return nil
}
