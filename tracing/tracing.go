package tracing

import (
	"context"
	"net/http"

	"github.com/omegaup/go-base/v3/logging"
)

type key int

var transactionKey key

// An Arg represents a name-value pair.
type Arg struct {
	Name  string
	Value interface{}
}

// A Segment is a part of a transaction used to instrument
// functions, methods, and blocks of code.
type Segment interface {
	// End marks the segment as finished.
	End()
}

// A Transaction represents one logical unit of work: either an
// inbound web request or background task.
type Transaction interface {
	// WithMetadata wraps the logger with the transaction metadata.
	WithMetadata(log logging.Logger) logging.Logger

	// SetName sets the current transaction's name.
	SetName(name string)

	// AddAttributes adds any attributes after the transaction was created.
	AddAttributes(args ...Arg)

	// StartSegment creates a new Segment inside the transaction.
	StartSegment(name string) Segment

	// NoticeError associates an error with the transaction.
	NoticeError(err error)

	// End marks the transaction as finished.
	End()
}

// Provider is the interface that can create Transactions.
type Provider interface {
	// StartTransaction starts a new Transaction with the provided name and
	// arguments.
	StartTransaction(name string, args ...Arg) Transaction

	// StartWebTransaction starts a new web Transaction with the provided name
	// and arguments.
	StartWebTransaction(
		name string,
		w http.ResponseWriter,
		r *http.Request,
		args ...Arg,
	) (Transaction, http.ResponseWriter, *http.Request)

	// WrapHandle is a convenience function for wrapping an http.Handler and
	// adding all necessary tracing information.
	WrapHandle(pattern string, handler http.Handler) (string, http.Handler)
}

// NewContext associates the transaction with the provided context.
// FromContext can then be used to retrieve the transaction.
func NewContext(ctx context.Context, txn Transaction) context.Context {
	return context.WithValue(ctx, transactionKey, txn)
}

// FromContext returns the current Transaction associated with the Context.
func FromContext(ctx context.Context) Transaction {
	if ctx == nil {
		return &noopTransaction{}
	}
	txn, ok := ctx.Value(transactionKey).(Transaction)
	if !ok {
		return &noopTransaction{}
	}

	return txn
}

type noopProvider struct{}

// NewNoOpProvider returns a new provider that does nothing.
func NewNoOpProvider() Provider {
	return &noopProvider{}
}

func (p *noopProvider) StartTransaction(name string, args ...Arg) Transaction {
	return &noopTransaction{}
}
func (p *noopProvider) StartWebTransaction(name string, w http.ResponseWriter, r *http.Request, args ...Arg) (Transaction, http.ResponseWriter, *http.Request) {
	return &noopTransaction{}, w, r
}
func (p *noopProvider) TransactionFromContext(ctx context.Context) Transaction {
	return NewNoOpTransaction()
}
func (p *noopProvider) WrapHandle(pattern string, handler http.Handler) (string, http.Handler) {
	return pattern, handler
}

type noopTransaction struct{}

// NewNoOpTransaction returns a Transaction that does nothing. Useful for tests.
func NewNoOpTransaction() Transaction {
	return &noopTransaction{}
}

func (t *noopTransaction) WithMetadata(log logging.Logger) logging.Logger {
	return log
}
func (t *noopTransaction) SetName(name string)       {}
func (t *noopTransaction) AddAttributes(args ...Arg) {}
func (t *noopTransaction) StartSegment(name string) Segment {
	return &noopSegment{}
}
func (t *noopTransaction) NoticeError(err error) {}
func (t *noopTransaction) End()                  {}

type noopSegment struct{}

func (s *noopSegment) AddAttributes(args ...Arg) {}
func (s *noopSegment) End()                      {}
