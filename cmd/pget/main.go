package main

import (
	"fmt"
	"os"

	"github.com/Code-Hex/pget"
)

func main() {

	cli := pget.New()
	if err := cli.Run(); err != nil {
		if cli.Trace {
			fmt.Fprintf(os.Stderr, "Error:\n%+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error:\n  %v\n", cli.ErrTop(err))
		}
		os.Exit(1)
	}

	os.Exit(0)
}
