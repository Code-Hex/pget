package pget

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const version = "test_version"

func TestParts_of_ready(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		wantProcs int
		wantURLs  int
	}{
		{
			name: "one URL",
			args: []string{
				"pget",
				"-p",
				"2",
				"http://example.com/filename.tar.gz",
				"--trace",
				"--output",
				"filename.tar.gz",
			},
			wantProcs: 2,
			wantURLs:  1,
		},
		{
			name: "two URLs",
			args: []string{
				"pget",
				"-p",
				"2",
				"http://example.com/filename.tar.gz",
				"http://example2.com/filename.tar.gz",
				"--trace",
				"--output",
				"filename.tar.gz",
			},
			wantProcs: 4,
			wantURLs:  2,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := New()

			if err := p.Ready(version, tc.args); err != nil {
				t.Errorf("failed to parse command line args: %s", err)
			}

			assert.Equal(t, true, p.Trace, "failed to parse arguments of trace")
			assert.Equal(t, tc.wantProcs, p.Procs, "failed to parse arguments of procs")
			assert.Equal(t, "filename.tar.gz", p.Output, "failed to parse output")

			assert.Len(t, p.URLs, tc.wantURLs)
		})
	}
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
