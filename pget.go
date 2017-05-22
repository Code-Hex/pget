package pget

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
)

type pget struct {
	Options
	URLs       []string
	GoRoutines int
}

func New() *pget {
	return &pget{
		GoRoutines: runtime.NumCPU(),
	}
}

func (p *pget) Run() int {
	if e := p.run(); e != nil {
		exitCode, err := UnwrapErrors(e)
		// for ignoreError
		if err == nil {
			return exitCode
		}

		if p.StackTrace {
			fmt.Fprintf(os.Stderr, "Error:\n  %+v\n", e)
		} else {
			fmt.Fprintf(os.Stderr, "Error:\n  %v\n", err)
		}
		return exitCode
	}
	return 0
}

func (p *pget) run() error {
	if err := p.prepare(); err != nil {
		return err
	}
	return nil
}

func (p *pget) prepare() error {
	if procs := os.Getenv("GOMAXPROCS"); procs == "" {
		runtime.GOMAXPROCS(p.GoRoutines)
	}

	args, err := parseOptions(&p.Options, os.Args[1:])
	if err != nil {
		return errors.Wrap(err, "Failed to parse command line args")
	}

	urls, err := parseURLs(args)
	if err != nil {
		return errors.Wrap(err, "Failed to parse url from arguments or stdin")
	}
	p.URLs = urls

	if p.Options.TargetDir != "" {
		dir := p.Options.TargetDir
		info, err := os.Stat(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				return makeResourceError(errors.Wrap(err, "Invalid directory"))
			}
			if err := os.MkdirAll(dir, os.ModeDir); err != nil {
				return makeResourceError(errors.Wrapf(err, `Failed to create diretory "%s"`, dir))
			}
		} else if !info.IsDir() {
			return makeResourceError(errors.New("Invalid directory"))
		}
		p.Options.TargetDir = strings.TrimSuffix(dir, "/")
	}
	return nil
}

func parseURLs(urls []string) ([]string, error) {
	var URLs []string
	// Looking for url from args
	for _, url := range urls {
		if govalidator.IsURL(url) {
			URLs = append(URLs, url)
		}
	}

	// Read from stdin if did not pass url arguments
	if len(URLs) < 1 {
		fmt.Fprintf(os.Stdout, "Please input url separate with space or newline\n")
		fmt.Fprintf(os.Stdout, "Start download with ^D\n")

		// scanning url from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			scan := scanner.Text()
			urls := strings.Split(scan, " ")
			for _, url := range urls {
				if govalidator.IsURL(url) {
					URLs = append(URLs, url)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, makeResourceError(errors.Wrap(err, "Failed to parse url from stdin"))
		}
		if len(URLs) < 1 {
			return nil, makeArgumentsError(errors.New("Download URL is must needed"))
		}
	}

	return URLs, nil
}

func parseOptions(opts *Options, argv []string) ([]string, error) {
	if len(argv) == 0 {
		os.Stdout.Write(opts.usage())
		return nil, makeIgnoreErr()
	}

	o, err := opts.parse(argv)
	if err != nil {
		return nil, makeArgumentsError(errors.Wrap(err, "Failed to parse command line options"))
	}
	if opts.Help {
		os.Stdout.Write(opts.usage())
		return nil, makeIgnoreErr()
	}
	if opts.Version {
		os.Stdout.Write([]byte(msg))
		return nil, makeIgnoreErr()
	}
	if opts.Update {
		result, err := opts.isupdate()
		if err != nil {
			return nil, makeArgumentsError(errors.Wrap(err, "Failed to parse command line options"))
		}
		os.Stdout.Write(result)
		return nil, makeIgnoreErr()
	}

	return o, nil
}
