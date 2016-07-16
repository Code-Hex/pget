package pget

import (
	"bytes"
	"fmt"
	"os"

	"github.com/Code-Hex/updater"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

// Options struct for parse command line arguments
type Options struct {
	Help      bool   `short:"h" long:"help" description:"print usage and exit"`
	Version   bool   `short:"v" long:"version" description:"display the version of pget and exit"`
	Procs     int    `short:"p" long:"procs" description:"split ratio to download file"`
	TargetDir string `short:"d" long:"target-dir" description:"path to directory to store the downloaded file"`
	Timeout   int    `short:"t" long:"timeout" description:"timeout of checking request in seconds"`
	Update    bool   `long:"check-update" description:"check if there is update available"`
	Trace     bool   `long:"trace" description:"display detail error messages"`
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
  -d,  --target-dir <PATH>    	path to the directory to save the downloaded file, filename will be taken from url
  -t,  --timeout <seconds>      timeout of checking request in seconds
  --check-update                check if there is update available
  --trace                       display detail error messages
`)
	return buf.Bytes()
}

func (opts Options) isupdate() ([]byte, error) {
	buf := bytes.Buffer{}
	result, err := updater.CheckWithTag("Code-Hex", "pget", version)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(&buf, result+"\n")

	return buf.Bytes(), nil
}
