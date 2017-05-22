package pget

import (
	"bytes"
	"fmt"
	"os"

	"github.com/Code-Hex/updater"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

const (
	version = "1.0.0"
	msg     = "Pget v" + version + ", The fastest file download client\n"
)

// Options struct for parse command line arguments
type Options struct {
	Help       bool   `short:"h" long:"help"`
	Version    bool   `short:"v" long:"version"`
	Procs      int    `short:"p" long:"procs"`
	Output     string `short:"o" long:"output"`
	TargetDir  string `short:"d" long:"target-dir"`
	Timeout    int    `short:"t" long:"timeout"`
	UserAgent  string `short:"u" long:"user-agent"`
	Referer    string `short:"r" long:"referer"`
	Update     bool   `long:"check-update"`
	StackTrace bool   `long:"trace"`
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
  -o,  --output <filename>      output file to <filename>
  -d,  --target-dir <path>    	path to the directory to save the downloaded file, filename will be taken from url
  -t,  --timeout <seconds>      timeout of checking request in seconds
  -u,  --user-agent <agent>     identify as <agent>
  -r,  --referer <referer>      identify as <referer>
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
