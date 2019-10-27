package pget

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Code-Hex/updater"
	"github.com/asaskevich/govalidator"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

// Options struct for parse command line arguments
type Options struct {
	Help      bool          `short:"h" long:"help"`
	Version   bool          `short:"v" long:"version"`
	Procs     int           `short:"p" long:"procs" default:"2"`
	Output    string        `short:"o" long:"output"`
	TargetDir string        `short:"d" long:"target-dir"`
	Timeout   time.Duration `short:"t" long:"timeout" default:"10s"`
	UserAgent string        `short:"u" long:"user-agent"`
	Referer   string        `short:"r" long:"referer"`
	Update    bool          `long:"check-update"`
	Trace     bool          `long:"trace"`
}

func (o *Object) parseOptions(argv []string) error {
	if len(argv) == 0 {
		os.Stdout.Write(o.usage())
		return o.makeIgnoreErr()
	}
	if err := o.parse(argv); err != nil {
		return errors.Wrap(err, "failed to parse command line options")
	}
	if err := o.arrangementOptions(); err != nil {
		return errors.Wrap(err, "failed to arrangement options")
	}
	return o.shouldIgnore()
}

func (o *Object) parse(argv []string) (err error) {
	p := flags.NewParser(o.Options, flags.PrintErrors)
	o.Args, err = p.ParseArgs(argv)
	if err != nil {
		os.Stderr.Write(o.usage())
		return errors.Wrap(err, "invalid command line options")
	}
	return nil
}

func (o *Object) shouldIgnore() error {
	if o.Help {
		os.Stdout.Write(o.usage())
		return o.makeIgnoreErr()
	}
	if o.Version {
		os.Stdout.Write([]byte(msg))
		return o.makeIgnoreErr()
	}
	if o.Update {
		result, err := o.isUpdate()
		if err != nil {
			return errors.Wrap(err, "failed to parse command line options")
		}
		os.Stdout.Write(result)
		return o.makeIgnoreErr()
	}
	return nil
}

func (o *Object) arrangementOptions() error {
	if err := o.parseURLs(); err != nil {
		return errors.Wrap(err, "failed to parse of url")
	}
	if o.Output != "" {
		o.filename = o.Output
	}
	return o.prepareTargetDir()
}

func (o *Object) prepareTargetDir() error {
	if o.TargetDir != "" {
		o.TargetDir = path.Clean(o.TargetDir)
		info, err := os.Stat(o.TargetDir)
		if err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "target dir is invalid")
			}
			if err := os.MkdirAll(o.TargetDir, 0755); err != nil {
				return errors.Wrapf(err, "failed to create diretory at %s", o.TargetDir)
			}
		} else if !info.IsDir() {
			return errors.New("target dir is not a valid directory")
		}
	}
	return nil
}

func (o *Object) parseURLs() error {
	// find url in args
	for _, arg := range o.Args {
		if govalidator.IsURL(arg) {
			o.URLs = append(o.URLs, arg)
		}
	}

	if len(o.URLs) < 1 {
		fmt.Println("Please input url separate with space or newline")
		fmt.Println("Start download at ^D")

		// scanning url from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			scan := scanner.Text()
			urls := strings.Split(scan, " ")
			for _, url := range urls {
				if govalidator.IsURL(url) {
					o.URLs = append(o.URLs, url)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return errors.Wrap(err, "failed to parse url from stdin")
		}

		if len(o.URLs) < 1 {
			return errors.New("urls not found in the arguments passed")
		}
	}
	return nil
}

func (opts *Options) usage() []byte {
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

func (opts *Options) isUpdate() ([]byte, error) {
	buf := bytes.Buffer{}
	result, err := updater.CheckWithTag("Code-Hex", "pget", version)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(&buf, result+"\n")

	return buf.Bytes(), nil
}
