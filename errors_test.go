package base

import (
	stderrors "errors"
	"github.com/pkg/errors"
	"testing"
)

var (
	ErrCategory = stderrors.New("category")
)

func TestErrorCategories(t *testing.T) {
	if !HasErrorCategory(ErrCategory, ErrCategory) {
		t.Errorf("HasErrorCategory(x, x) unexpectedly returned false")
	}

	if !HasErrorCategory(ErrorWithCategory(ErrCategory, errors.New("foo")), ErrCategory) {
		t.Errorf("HasErrorCategory(ErrorWithCategory(x, y), x) unexpectedly returned false")
	}

	if !HasErrorCategory(errors.Wrap(ErrorWithCategory(ErrCategory, errors.New("foo")), "bar"), ErrCategory) {
		t.Errorf("HasErrorCategory(errors.Wrap(ErrorWithCategory(x, y), z), x) unexpectedly returned false")
	}
}
