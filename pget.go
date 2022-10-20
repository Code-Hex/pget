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

	"github.com/Songmu/prompter"
	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
)

// Pget structs
type Pget struct {
	Trace  bool
	Output string
	Procs  int
	URLs   []string

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

	// TODO(codehex): calc maxIdleConnsPerHost
	client := newDownloadClient(16)

	target, err := Check(ctx, &CheckConfig{
		URLs:    pget.URLs,
		Timeout: time.Duration(pget.timeout) * time.Second,
		Client:  client,
	})
	if err != nil {
		return err
	}

	filename := target.Filename

	var dir string
	if pget.Output != "" {
		fi, err := os.Stat(pget.Output)
		if err == nil && fi.IsDir() {
			dir = pget.Output
		} else {
			dir, filename = filepath.Split(pget.Output)
			if dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return errors.Wrapf(err, "failed to create diretory at %s", dir)
				}
			}
		}
	}

	opts := []DownloadOption{
		WithUserAgent(pget.useragent, version),
		WithReferer(pget.referer),
	}

	return Download(ctx, &DownloadConfig{
		Filename:      filename,
		Dirname:       dir,
		ContentLength: target.ContentLength,
		Procs:         pget.Procs,
		URLs:          target.URLs,
		Client:        client,
	}, opts...)
}

const (
	warningNumConnection = 4
	warningMessage       = "[WARNING] Using a large number of connections to 1 URL can lead to DOS attacks.\n" +
		"In most cases, `4` or less is enough. In addition, the case is increasing that if you use multiple connections to 1 URL does not increase the download speed with the spread of CDNs.\n" +
		"See: https://github.com/Code-Hex/pget#disclaimer\n" +
		"\n" +
		"Would you execute knowing these?\n"
)

// Ready method define the variables required to Download.
func (pget *Pget) Ready(version string, args []string) error {
	opts, err := pget.parseOptions(args, version)
	if err != nil {
		return errors.Wrap(errTop(err), "failed to parse command line args")
	}

	if opts.Trace {
		pget.Trace = opts.Trace
	}

	if opts.Timeout > 0 {
		pget.timeout = opts.Timeout
	}

	if err := pget.parseURLs(); err != nil {
		return errors.Wrap(err, "failed to parse of url")
	}

	if opts.NumConnection > warningNumConnection && !prompter.YN(warningMessage, false) {
		return makeIgnoreErr()
	}

	pget.Procs = opts.NumConnection * len(pget.URLs)

	if opts.Output != "" {
		pget.Output = opts.Output
	}

	if opts.UserAgent != "" {
		pget.useragent = opts.UserAgent
	}

	if opts.Referer != "" {
		pget.referer = opts.Referer
	}

	return nil
}

func (pget *Pget) parseOptions(argv []string, version string) (*Options, error) {
	var opts Options
	if len(argv) == 0 {
		stdout.Write(opts.usage(version))
		return nil, makeIgnoreErr()
	}

	o, err := opts.parse(argv, version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse command line options")
	}

	if opts.Help {
		stdout.Write(opts.usage(version))
		return nil, makeIgnoreErr()
	}

	if opts.Update {
		result, err := opts.isupdate(version)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse command line options")
		}

		stdout.Write(result)
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
		fmt.Fprintf(stdout, "Please input url separate with space or newline\n")
		fmt.Fprintf(stdout, "Start download with ^D\n")

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
