package pget

import (
	"testing"

	"github.com/pkg/errors"
)

func TestErrors(t *testing.T) {
	err := errors.New("first")
	err = errors.Wrap(err, "second")
	err = errors.Wrap(err, "third")

	err = errTop(err)
	if err.Error() != "first" {
		t.Errorf("could not get top message")
	}
}
