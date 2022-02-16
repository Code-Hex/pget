package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Code-Hex/pget"
)

var version string

func main() {
	cli := pget.New()
	if err := cli.Run(context.Background(), version, os.Args[1:]); err != nil {
		if pget.ISNotSupportRequestRange(err) {
			if err = pget.DownloadFiles(cli.URLs, "./"); err != nil {
				fmt.Fprintf(os.Stderr, "Error:\n%+v\n", err)
				os.Exit(1)
			}
		}
		if cli.Trace {
			fmt.Fprintf(os.Stderr, "Error:\n%+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error:\n  %v\n", err)
		}
		os.Exit(1)
	}
}
