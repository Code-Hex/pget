package pget

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type AssignmentConfig struct {
	Procs         int
	TaskSize      int64 // download filesize per task
	ContentLength int64 // full download filesize
	URLs          []string
	PartialDir    string
	Filename      string
}

type Task struct {
	ID         int
	Procs      int
	URL        string
	Range      Range
	PartialDir string
	Filename   string
}

func (t *Task) destPath() string {
	return filepath.Join(
		t.PartialDir,
		fmt.Sprintf("%s.%d.%d", t.Filename, t.Procs, t.ID),
	)
}

func (t *Task) String() string {
	return fmt.Sprintf("task[%d]: %q", t.ID, t.destPath())
}

type makeRequestOption struct {
	useragent string
	referer   string
}

func (t *Task) makeRequest(ctx context.Context, opt *makeRequestOption) (*http.Request, error) {
	req, err := http.NewRequest("GET", t.URL, nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to make a new request: %d", t.ID))
	}
	req = req.WithContext(ctx)

	// set download ranges
	req.Header.Set("Range", t.Range.BytesRange())

	// set useragent
	if opt.useragent != "" {
		req.Header.Set("User-Agent", opt.useragent)
	}

	// set referer
	if opt.referer != "" {
		req.Header.Set("Referer", opt.referer)
	}

	return req, nil
}

// Assignment method that to each goroutine gives the task
func Assignment(c *AssignmentConfig) []*Task {
	tasks := make([]*Task, 0, c.Procs)

	var totalActiveProcs int
	for i := 0; i < c.Procs; i++ {

		r := makeRange(i, c.Procs, c.TaskSize, c.ContentLength)

		partName := filepath.Join(
			c.PartialDir,
			fmt.Sprintf("%s.%d.%d", c.Filename, c.Procs, i),
		)

		if info, err := os.Stat(partName); err == nil {
			infosize := info.Size()
			// check if the part is fully downloaded
			if i == c.Procs-1 {
				if infosize == r.high-r.low {
					continue
				}
			} else if infosize == c.TaskSize {
				// skip as the part is already downloaded
				continue
			}

			// make low range from this next byte
			r.low += infosize
		}

		tasks = append(tasks, &Task{
			ID:         i,
			Procs:      c.Procs,
			URL:        c.URLs[totalActiveProcs%len(c.URLs)],
			Range:      r,
			PartialDir: c.PartialDir,
			Filename:   c.Filename,
		})

		totalActiveProcs++
	}

	return tasks
}

type DownloadConfig struct {
	Filename      string
	Dirname       string
	ContentLength int64
	Procs         int
	URLs          []string

	*makeRequestOption
}

type DownloadOption func(c *DownloadConfig)

func WithUserAgent(ua string) DownloadOption {
	return func(c *DownloadConfig) {
		c.makeRequestOption.useragent = ua
	}
}

func WithReferer(referer string) DownloadOption {
	return func(c *DownloadConfig) {
		c.makeRequestOption.referer = referer
	}
}

func Download(ctx context.Context, c *DownloadConfig, opts ...DownloadOption) error {
	partialDir := getPartialDirname(c.Dirname, c.Filename, c.Procs)
	// create download location
	if err := os.MkdirAll(partialDir, 0755); err != nil {
		return errors.Wrap(err, "failed to mkdir for download location")
	}

	c.makeRequestOption = &makeRequestOption{}

	for _, opt := range opts {
		opt(c)
	}

	tasks := Assignment(&AssignmentConfig{
		Procs:         c.Procs,
		TaskSize:      c.ContentLength / int64(c.Procs),
		ContentLength: c.ContentLength,
		URLs:          c.URLs,
		PartialDir:    partialDir,
		Filename:      c.Filename,
	})

	if err := parallelDownload(ctx, c, tasks); err != nil {
		return err
	}

	return bindFiles(c, partialDir)
}

func parallelDownload(ctx context.Context, c *DownloadConfig, tasks []*Task) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return ProgressBar(ctx, c.ContentLength, c.Dirname)
	})

	for _, task := range tasks {
		task := task
		eg.Go(func() error {
			req, err := task.makeRequest(ctx, c.makeRequestOption)
			if err != nil {
				return err
			}
			return task.download(req)
		})
	}

	return eg.Wait()
}

func (t *Task) download(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to get response: %q", t.String())
	}
	defer resp.Body.Close()

	output, err := os.OpenFile(t.destPath(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return errors.Wrapf(err, "failed to create: %q", t.String())
	}
	defer output.Close()

	if _, err := io.Copy(output, resp.Body); err != nil {
		return errors.Wrapf(err, "failed to write response body: %q", t.String())
	}

	return nil
}

func bindFiles(c *DownloadConfig, partialDir string) error {
	fmt.Println("\nbinding with files...")

	destPath := filepath.Join(c.Dirname, c.Filename)
	f, err := os.Create(destPath)
	if err != nil {
		return errors.Wrap(err, "failed to create a file in download location")
	}
	defer f.Close()

	bar := pb.Start64(c.ContentLength)

	copyFn := func(name string) error {
		subfp, err := os.Open(name)
		if err != nil {
			return errors.Wrapf(err, "failed to open %q in download location", name)
		}

		defer subfp.Close()

		proxy := bar.NewProxyReader(subfp)
		if _, err := io.Copy(f, proxy); err != nil {
			return errors.Wrapf(err, "failed to copy %q", name)
		}

		// remove a file in download location for join
		if err := os.Remove(name); err != nil {
			return errors.Wrapf(err, "failed to remove %q in download location", name)
		}
		return nil
	}

	for i := 0; i < c.Procs; i++ {
		name := fmt.Sprintf("%s/%s.%d.%d", partialDir, c.Filename, c.Procs, i)
		if err := copyFn(name); err != nil {
			return err
		}
	}

	bar.Finish()

	// remove download location
	// RemoveAll reason: will create .DS_Store in download location if execute on mac
	if err := os.RemoveAll(partialDir); err != nil {
		return errors.Wrap(err, "failed to remove download location")
	}

	return nil
}
