package pget

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParts_of_ready(t *testing.T) {

	// begin test
	fmt.Fprintf(os.Stdout, "Testing parse_test\n")
	url := "http://example.com/filename.tar.gz"

	args := []string{
		"pget",
		"-p",
		"2",
		url,
		"--trace",
		"-o",
		"Hello.tar.gz",
	}

	p := New()
	var opts Options
	if err := p.parseOptions(&opts, args); err != nil {
		t.Errorf("failed to parse command line args: %s", err)
	}

	assert.Equal(t, true, opts.Trace, "failed to parse arguments of trace")
	assert.Equal(t, opts.Procs, 2, "failed to parse arguments of procs")

	if err := p.parseURLs(); err != nil {
		t.Errorf("failed to parse of url: %s", err)
	}

	p.SetFileName(opts.Output)
	assert.Equal(t, p.FileName(), "Hello.tar.gz", "failed to parse arguments of output")

	p.URLFileName(url)
	assert.Equal(t, p.FileName(), "filename.tar.gz", "failed to get of filename from url")

	fmt.Fprintf(os.Stdout, "parse_test Done\n\n")
}

func TestShowhelp(t *testing.T) {
	// begin test
	fmt.Fprintf(os.Stdout, "Testing showhelp_test\n")

	args := []string{
		"pget",
		"-h",
	}

	p := New()
	opts := Options{}
	err := p.parseOptions(&opts, args)
	assert.NotNil(t, err)

	args = []string{
		"pget",
		"--help",
	}

	p = New()
	opts = Options{}
	err = p.parseOptions(&opts, args)
	assert.NotNil(t, err)

	fmt.Fprintf(os.Stdout, "showhelp_test Done\n\n")
}

func TestShowversion(t *testing.T) {
	// begin test
	fmt.Fprintf(os.Stdout, "Testing showversion_test\n")

	args := []string{
		"pget",
		"-v",
	}

	p := New()
	opts := Options{}
	err := p.parseOptions(&opts, args)
	assert.NotNil(t, err)

	args = []string{
		"pget",
		"--version",
	}

	p = New()
	opts = Options{}
	err = p.parseOptions(&opts, args)
	assert.NotNil(t, err)

	fmt.Fprintf(os.Stdout, "showversion_test Done\n\n")
}
