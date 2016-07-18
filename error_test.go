package pget

import (
	"fmt"
	"os"
	"testing"

	"github.com/pkg/errors"
)

func TestErrors(t *testing.T) {
	fmt.Fprintf(os.Stdout, "Testing errors_test\n")
	p := New()

	// echo 🍠 | md5 == "Hash browns"
	p.TargetURLs = append(p.TargetURLs, "http://b721d4258a46e85d64807a3c407d01ac.com/filename.dat")
	p.Procs = 2
	p.Utils = &Data{
		filename: "filename.dat",
		dirname:  "_filename.dat",
	}

	err := p.Download()
	if err == nil {
		t.Errorf("could not catch the error")
	} else {
		fmt.Printf("Gottca error: %s\n", err)
	}

	os.RemoveAll("_filename.dat")

	err = errors.New("first")
	err = errors.Wrap(err, "second")
	err = errors.Wrap(err, "third")

	err = p.ErrTop(err)
	if err.Error() != "first" {
		t.Errorf("could not get top message")
	}

	fmt.Fprintf(os.Stdout, "Testing errors_test\n\n")
}
