package pget

import (
	"fmt"
	"os"
	"testing"
)

func TestErrors(t *testing.T) {
	fmt.Fprintf(os.Stdout, "Testing errors_test\n")
	p := New()

	// echo 🍠 | md5 == "Hash browns"
	p.url = "http://b721d4258a46e85d64807a3c407d01ac.com/filename.dat"
	p.procs = 2
	p.Utils = &Data{
		filename: "filename.dat",
		dirname:  "_filename.dat",
	}

	err := p.download()
	if err == nil {
		t.Errorf("could not catch the error")
	} else {
		fmt.Printf("Gottca error: %s\n", err)
	}

	fmt.Fprintf(os.Stdout, "Testing errors_test\n\n")
}
