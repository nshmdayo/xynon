package plugin

import (
	"errors"
	"testing"
)

// TestErrors tests the functionality of the respective component.
func TestErrors(t *testing.T) {
	e1 := &ErrTimeout{}
	if e1.Error() == "" {
		t.Fail()
	}
	
	e5 := &ErrInvalidHandle{Handle: 1}
	if e5.Error() == "" {
		t.Fail()
	}

	cause := errors.New("err")
	e2 := &ErrExecution{Cause: cause}
	if e2.Error() == "" {
		t.Fail()
	}
	if !errors.Is(e2, cause) {
		t.Fail()
	}

	e3 := &ErrABIVersion{Got: 1, Want: 2}
	if e3.Error() == "" {
		t.Fail()
	}

	e4 := &ErrMissingExport{Name: "n"}
	if e4.Error() == "" {
		t.Fail()
	}
}
