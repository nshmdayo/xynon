package plugin

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	e1 := &ErrTimeout{}
	if e1.Error() == "" {
		t.Fail()
	}
	
	e5 := &ErrInvalidHandle{Handle: 1}
	if e5.Error() == "" {
		t.Fail()
	}

	e2 := &ErrExecution{Cause: errors.New("err")}
	if e2.Error() == "" {
		t.Fail()
	}
	if e2.Unwrap() == nil {
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
