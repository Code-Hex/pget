package pget

import (
	"bytes"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

// Options struct for parse command line arguments
type Options struct {
	Help    bool   `short:"h" long:"help" description:"print usage and exit"`
	Version bool   `short:"v" long:"version" description:"display the version of pget and exit"`
	Procs   int    `short:"p" long:"procs" description:"split ratio to download file"`
	Output  string `short:"o" long:"output" description:"output file to FILENAME"`
	Timeout int    `long:"timeout" description:"timeout of checking request in seconds"`
	Trace   bool   `long:"trace" description:"display detail error messages"`
	// File    string `long:"file" description:"urls has same hash in a file to download"`
}

func (opts *Options) parse(argv []string) ([]string, error) {
	p := flags.NewParser(opts, flags.PrintErrors)
	args, err := p.ParseArgs(argv)

	if err != nil {
		os.Stderr.Write(opts.usage())
		return nil, errors.Wrap(err, "invalid command line options")
	}

	return args, nil
}

func (opts Options) usage() []byte {
	buf := bytes.Buffer{}

	fmt.Fprintf(&buf, msg+
		`Usage: pget [options] URL

Options:
  -h,  --help                   print usage and exit
  -v,  --version                display the version of pget and exit
  -p,  --procs <num>            split ratio to download file
  -o,  --output <filename>      output file to FILENAME
  --timeout <seconds>           timeout of checking request in seconds
  --trace                       display detail error messages
`)

	return buf.Bytes()
}
