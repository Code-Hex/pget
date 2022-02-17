package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/Code-Hex/pget"
)

var version string

func main() {
	cli := pget.New()
	if err := cli.Run(context.Background(), version, os.Args[1:]); err != nil {
		if errors.Is(err, pget.ErrNotSupportRequestRange) {
			if e := cli.DownloadFiles(context.Background(), cli.URLs, "./"); e != nil {
				fmt.Fprintf(os.Stderr, "Error:\n%+v\n", e)
				os.Exit(1)
			}
			os.Exit(0)
		}
		if cli.Trace {
			fmt.Fprintf(os.Stderr, "Error:\n%+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error:\n  %v\n", err)
		}
		os.Exit(1)
	}
}
