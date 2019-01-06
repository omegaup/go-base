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

	if cause := UnwrapCauseFromErrorCategory(ErrCategory, ErrCategory); nil != cause {
		t.Errorf(
			"mismatched UnwrapCauseFromErrorCategory(x, x) "+
				"expected nil, got %s",
			cause,
		)
	}

	rootCauseError := errors.New("foo")
	categorizedError := ErrorWithCategory(ErrCategory, rootCauseError)
	if !HasErrorCategory(categorizedError, ErrCategory) {
		t.Errorf("HasErrorCategory(ErrorWithCategory(x, y), x) unexpectedly returned false")
	}

	if cause := UnwrapCauseFromErrorCategory(categorizedError, ErrCategory); rootCauseError != cause {
		t.Errorf(
			"mismatched UnwrapCauseFromErrorCategory(ErrorWithCategory(x, y), x) "+
				"expected %s, got %s",
			rootCauseError,
			cause,
		)
	}

	wrappedError := errors.Wrap(categorizedError, "bar")
	if !HasErrorCategory(wrappedError, ErrCategory) {
		t.Errorf("HasErrorCategory(errors.Wrap(ErrorWithCategory(x, y), z), x) unexpectedly returned false")
	}

	if cause := UnwrapCauseFromErrorCategory(wrappedError, ErrCategory); rootCauseError != cause {
		t.Errorf(
			"mismatched UnwrapCauseFromErrorCategory(errors.Wrap(ErrorWithCategory(x, y), z), x) "+
				"expected %s, got %s",
			rootCauseError,
			cause,
		)
	}
}
