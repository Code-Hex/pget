package pget

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

// Options struct for parse command line arguments
type Options struct {
	Help    bool   `short:"h" long:"help" description:"print usage and exit"`
	Version bool   `short:"v" long:"version" description:"display the version of pget and exit"`
	Procs   int    `short:"p" long:"procs" description:"split ratio to download file"`
	Output  string `short:"o" long:"output" description:"output file to FILENAME"`
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

	fmt.Fprintf(&buf, "Pget "+version+", a parallel file download client\n"+
		`Usage: pget [options] URL

Options:
`)

	var description string
	t := reflect.TypeOf(opts)

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag

		if sh := tag.Get("short"); sh != "" {
			description = fmt.Sprintf("-%s,  --%s", sh, tag.Get("long"))
		} else {
			description = fmt.Sprintf("--%s", tag.Get("long"))
		}

		fmt.Fprintf(&buf, "  %-20s %s\n", description, tag.Get("description"))
	}

	return buf.Bytes()
}
