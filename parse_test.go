package pget

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const version = "test_version"

func TestParts_of_ready(t *testing.T) {
	// begin test
	url := "http://example.com/filename.tar.gz"

	args := []string{
		"pget",
		"-p",
		"2",
		url,
		"--trace",
	}

	p := New()
	opts, err := p.parseOptions(args, version)
	if err != nil {
		t.Errorf("failed to parse command line args: %s", err)
	}

	assert.Equal(t, true, opts.Trace, "failed to parse arguments of trace")
	assert.Equal(t, opts.Procs, 2, "failed to parse arguments of procs")

	if err := p.parseURLs(); err != nil {
		t.Errorf("failed to parse of url: %s", err)
	}

	assert.Equal(t, "filename.tar.gz", p.Filename, "failed to get of filename from url")
}

func TestShowhelp(t *testing.T) {
	args := []string{
		"pget",
		"-h",
	}

	p := New()
	_, err := p.parseOptions(args, version)
	assert.NotNil(t, err)

	args = []string{
		"pget",
		"--help",
	}

	p = New()
	_, err = p.parseOptions(args, version)
	assert.NotNil(t, err)
}

func TestShowisupdate(t *testing.T) {
	args := []string{
		"pget",
		"--check-update",
	}

	p := New()
	_, err := p.parseOptions(args, version)
	assert.NotNil(t, err)
}
