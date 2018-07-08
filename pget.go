package pget

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"golang.org/x/sync/errgroup"
	pb "gopkg.in/cheggaaa/pb.v1"

	"github.com/Code-Hex/pget/internal/utils"
	"github.com/pkg/errors"
)

var (
	version string
	msg     = "Pget v" + version + ", parallel file download client\n"
)

type Object struct {
	*Options
	Args       []string
	URLs       []string
	TargetURLs []string

	*info
	*chans

	client     *http.Client
	grp        errgroup.Group
	goProgress chan struct{}
}

type info struct {
	filename   string
	filesize   uint
	tmpDirName string
}

type chans struct {
	size    chan uint
	setSize chan struct{}
}

func New() *Object {
	return &Object{
		Options: new(Options),
		client:  httpClient(),
		info:    new(info),
		chans: &chans{
			size:    make(chan uint),
			setSize: make(chan struct{}),
		},
	}
}

func (o *Object) Run() int {
	if err := o.run(); err != nil {
		if o.Trace {
			fmt.Fprintf(os.Stderr, "Error:\n%+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error:\n  %v\n", o.getRootErr(err))
		}
		return 1
	}
	return 0
}

func (o *Object) run() error {
	if err := o.prepare(); err != nil {
		return errors.Wrap(err, "failed to prepare pget")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go o.checkSizeDifference(cancel)
	if err := o.check(ctx); err != nil {
		return errors.Wrap(err, "failed to check header")
	}
	if err := o.download(ctx); err != nil {
		return err
	}
	return o.bindwithFiles()
}

// prepare method defines the variables required to Download.
func (o *Object) prepare() error {
	if err := o.parseOptions(os.Args[1:]); err != nil {
		return errors.Wrap(err, "failed to parse command line args")
	}
	if procs := os.Getenv("GOMAXPROCS"); procs == "" {
		runtime.GOMAXPROCS(o.Procs)
	}
	return nil
}

// check method checks it can request
func (o *Object) check(ctx context.Context) error {
	tctx, cancel := context.WithTimeout(ctx, o.Timeout)
	defer cancel()
	eg, ectx := errgroup.WithContext(tctx)
	for _, url := range o.URLs {
		fmt.Printf("Checking now %s\n", url)
		eg.Go(o.checkIsCorrectMirrors(ectx, url))
	}
	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "failed to check mirrors")
	}
	return o.setup()
}

func (o *Object) setup() error {
	o.setFileSize()
	// did already get filename from -o option
	if o.filename == "" {
		o.filename = utils.FileNameFromURL(o.TargetURLs[0])
	}
	o.tmpDirName = fmt.Sprintf("_%s.%d", o.filename, o.Procs)

	if err := utils.IsFree(o.filesize, uint(o.Procs)); err != nil {
		return errors.Wrap(err, "failed to check is disk free")
	}
	// create download location
	if err := os.MkdirAll(o.tmpDirName, 0755); err != nil {
		return errors.Wrap(err, "failed to mkdir for download location")
	}
	return nil
}

func (o *Object) download(ctx context.Context) error {
	eg, ectx := errgroup.WithContext(ctx)
	// on an assignment for request
	o.assignment(ectx, eg)
	eg.Go(o.progressBar(ctx))
	return eg.Wait()
}

// assignment method that to each goroutine gives the task
func (o *Object) assignment(ctx context.Context, eg *errgroup.Group) {
	filename := o.filename
	dirname := o.tmpDirName

	procs := uint(o.Procs)
	split := o.filesize / procs
	assignment := procs / uint(len(o.URLs))

	var lasturl string
	totalActiveProcs := uint(0)
	for i := uint(0); i < procs; i++ {
		partName := fmt.Sprintf("%s/%s.%d.%d", dirname, filename, procs, i)

		// make range
		r := o.makeRanges(i, split)
		if info, err := os.Stat(partName); err == nil {
			infosize := uint(info.Size())
			// check if the part is fully downloaded
			if isLastProc(i, procs) {
				if infosize == r.filesize() {
					continue
				}
			} else if infosize == split {
				// skip as the part is already downloaded
				continue
			}
			// make low range from this next byte
			r.low += infosize
		}

		totalActiveProcs++

		url := o.TargetURLs[0]

		// give efficiency and equality work
		if totalActiveProcs%assignment == 0 {
			// Like shift method
			if len(o.TargetURLs) > 1 {
				o.TargetURLs = o.TargetURLs[1:]
			}

			// check whether to output the message
			if lasturl != url {
				fmt.Printf("Download start from %s\n", url)
				lasturl = url
			}
		}

		// execute get request
		eg.Go(o.Requests(ctx, r, url))
	}
}

func isLastProc(i, procs uint) bool {
	return i == procs-1
}

// bindwithFiles function for file binding after split download
func (o *Object) bindwithFiles() error {

	fmt.Println("\nbinding with files...")

	filesize := o.filesize
	filename := o.filename
	dirname := o.tmpDirName
	fh, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "failed to create a file in download location")
	}
	defer fh.Close()

	bar := pb.New64(int64(filesize))
	bar.Start()

	var f string
	for i := 0; i < o.Procs; i++ {
		f = fmt.Sprintf("%s/%s.%d.%d", dirname, filename, o.Procs, i)
		subfp, err := os.Open(f)
		if err != nil {
			return errors.Wrap(err, "failed to open "+f+" in download location")
		}

		proxy := bar.NewProxyReader(subfp)
		io.Copy(fh, proxy)

		// Not use defer
		subfp.Close()

		// remove a file in download location for join
		if err := os.Remove(f); err != nil {
			return errors.Wrap(err, "failed to remove a file in download location")
		}
	}

	bar.Finish()

	// remove download location
	// RemoveAll reason: will create .DS_Store in download location if execute on mac
	if err := os.RemoveAll(dirname); err != nil {
		return errors.Wrap(err, "failed to remove download location")
	}

	fmt.Println("Complete")

	return nil
}
