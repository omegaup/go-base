package base

import (
	"fmt"
	"github.com/pkg/errors"
)

// These two interfaces are not exported by errors, but they are part of its
// stable interface.
type stackTracer interface {
	StackTrace() errors.StackTrace
}

type causer interface {
	Cause() error
}

// categorizer allows obtaining a category from an error, if provided.
type categorizer interface {
	// Category returns the canonical error category for an error.
	Category() error
}

// withCategory is a wrapped error (the cause) with a category and a stack.
type withCategory struct {
	cause    error
	category error
	stack    errors.StackTrace
}

var _ categorizer = &withCategory{}
var _ causer = &withCategory{}
var _ error = &withCategory{}
var _ stackTracer = &withCategory{}

// ErrorWithCategory is similar to errors.Wrap, but instead of creating a new
// error as the wrapping message, a sentinel error is provided as a category.
// This category can then be inspected with HasErrorCategory.
func ErrorWithCategory(category error, cause error) error {
	if cause == nil {
		return category
	}
	var stack errors.StackTrace
	if originalStack, ok := cause.(stackTracer); ok {
		stack = originalStack.StackTrace()
	} else {
		// The error message is not important, we just want the stack trace.
		// We'll also skip the current stack frame because we want the caller's
		// instead.
		stack = errors.New("").(stackTracer).StackTrace()[1:]
	}
	return &withCategory{
		cause:    cause,
		category: category,
		stack:    stack,
	}
}

func (c *withCategory) Error() string {
	return fmt.Sprintf("%s: %s", c.category.Error(), c.cause.Error())
}

func (c *withCategory) Cause() error { return c.cause }

func (c *withCategory) Category() error { return c.category }

func (c *withCategory) StackTrace() errors.StackTrace { return c.stack }

// HasErrorCategory returns whether the provided error belongs to the provided
// category.
func HasErrorCategory(err error, category error) bool {
	if err == category {
		return true
	}
	for err != nil {
		cat, ok := err.(categorizer)
		if ok && cat.Category() == category {
			return true
		}
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return false
}
