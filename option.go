package pget

import (
	"bytes"
	"fmt"

	"github.com/Code-Hex/updater"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

// Options struct for parse command line arguments
type Options struct {
	Help      bool   `short:"h" long:"help"`
	Procs     int    `short:"p" long:"procs"`
	Output    string `short:"o" long:"output"`
	Timeout   int    `short:"t" long:"timeout"`
	UserAgent string `short:"u" long:"user-agent"`
	Referer   string `short:"r" long:"referer"`
	Update    bool   `long:"check-update"`
	Trace     bool   `long:"trace"`
}

func (opts *Options) parse(argv []string, version string) ([]string, error) {
	p := flags.NewParser(opts, flags.PrintErrors)
	args, err := p.ParseArgs(argv)

	if err != nil {
		stdout.Write(opts.usage(version))
		return nil, errors.Wrap(err, "invalid command line options")
	}

	return args, nil
}

func (opts Options) usage(version string) []byte {
	buf := bytes.Buffer{}

	msg := "Pget %s, The fastest file download client\n"
	fmt.Fprintf(&buf, msg+
		`Usage: pget [options] URL
  Options:
  -h,  --help                   print usage and exit
  -p,  --procs <num>            split ratio to download file
  -o,  --output <filename>      output file to <filename>
  -t,  --timeout <seconds>      timeout of checking request in seconds
  -u,  --user-agent <agent>     identify as <agent>
  -r,  --referer <referer>      identify as <referer>
  --check-update                check if there is update available
  --trace                       display detail error messages
`, version)
	return buf.Bytes()
}

func (opts Options) isupdate(version string) ([]byte, error) {
	buf := bytes.Buffer{}
	result, err := updater.CheckWithTag("Code-Hex", "pget", version)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(&buf, result+"\n")

	return buf.Bytes(), nil
}
