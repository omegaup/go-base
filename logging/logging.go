package logging

import (
	"context"
	"io"
	"time"

	"github.com/go-logfmt/logfmt"
)

// Logger is an interface compatible with newrelic.Logger
type Logger interface {
	New(context map[string]interface{}) Logger
	NewContext(ctx context.Context) Logger
	Error(msg string, context map[string]interface{})
	Warn(msg string, context map[string]interface{})
	Info(msg string, context map[string]interface{})
	Debug(msg string, context map[string]interface{})
	DebugEnabled() bool
}

func mergeContexts(a, b map[string]interface{}) map[string]interface{} {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	newContext := make(map[string]interface{}, len(a)+len(b))
	for k, v := range a {
		newContext[k] = v
	}
	for k, v := range b {
		newContext[k] = v
	}
	return newContext
}

type multiLogger struct {
	l       []Logger
	context map[string]interface{}
}

// NewMultiLogger returns a composed Logger that forwards all calls to all
// provided Loggers.
func NewMultiLogger(l ...Logger) Logger {
	return &multiLogger{l: l}
}

func (l *multiLogger) New(context map[string]interface{}) Logger {
	return &multiLogger{
		l:       l.l,
		context: mergeContexts(l.context, context),
	}
}

func (l *multiLogger) NewContext(ctx context.Context) Logger {
	loggers := make([]Logger, len(l.l))
	for i := 0; i < len(l.l); i++ {
		loggers[i] = l.l[i].NewContext(ctx)
	}
	return &multiLogger{
		l:       loggers,
		context: l.context,
	}
}

func (l *multiLogger) Error(msg string, context map[string]interface{}) {
	newContext := mergeContexts(l.context, context)
	for _, ll := range l.l {
		ll.Error(msg, newContext)
	}
}

func (l *multiLogger) Warn(msg string, context map[string]interface{}) {
	newContext := mergeContexts(l.context, context)
	for _, ll := range l.l {
		ll.Warn(msg, newContext)
	}
}

func (l *multiLogger) Info(msg string, context map[string]interface{}) {
	newContext := mergeContexts(l.context, context)
	for _, ll := range l.l {
		ll.Info(msg, newContext)
	}
}

func (l *multiLogger) Debug(msg string, context map[string]interface{}) {
	newContext := mergeContexts(l.context, context)
	for _, ll := range l.l {
		ll.Debug(msg, newContext)
	}
}

func (l *multiLogger) DebugEnabled() bool {
	for _, ll := range l.l {
		if ll.DebugEnabled() {
			return true
		}
	}
	return false
}

type inMemoryLogfmtLogger struct {
	w       *logfmt.Encoder
	context map[string]interface{}
}

// NewInMemoryLogfmtLogger writes logfmt-formatted records to the provided writer.
func NewInMemoryLogfmtLogger(w io.Writer) Logger {
	return &inMemoryLogfmtLogger{w: logfmt.NewEncoder(w)}
}

func (l *inMemoryLogfmtLogger) New(context map[string]interface{}) Logger {
	return &inMemoryLogfmtLogger{
		w:       l.w,
		context: mergeContexts(l.context, context),
	}
}

func (l *inMemoryLogfmtLogger) NewContext(ctx context.Context) Logger {
	return l
}

func (l *inMemoryLogfmtLogger) Error(msg string, context map[string]interface{}) {
	l.w.EncodeKeyvals(
		"t", time.Now().Format(time.RFC3339),
		"lvl", "eror",
		"msg", msg,
	)
	for k, v := range mergeContexts(l.context, context) {
		l.w.EncodeKeyval(k, v)
	}
	l.w.EndRecord()
}

func (l *inMemoryLogfmtLogger) Warn(msg string, context map[string]interface{}) {
	l.w.EncodeKeyvals(
		"t", time.Now().Format(time.RFC3339),
		"lvl", "warn",
		"msg", msg,
	)
	for k, v := range mergeContexts(l.context, context) {
		l.w.EncodeKeyval(k, v)
	}
	l.w.EndRecord()
}

func (l *inMemoryLogfmtLogger) Info(msg string, context map[string]interface{}) {
	l.w.EncodeKeyvals(
		"t", time.Now().Format(time.RFC3339),
		"lvl", "info",
		"msg", msg,
	)
	for k, v := range mergeContexts(l.context, context) {
		l.w.EncodeKeyval(k, v)
	}
	l.w.EndRecord()
}

func (l *inMemoryLogfmtLogger) Debug(msg string, context map[string]interface{}) {
	l.w.EncodeKeyvals(
		"t", time.Now().Format(time.RFC3339),
		"lvl", "dbug",
		"msg", msg,
	)
	for k, v := range mergeContexts(l.context, context) {
		l.w.EncodeKeyval(k, v)
	}
	l.w.EndRecord()
}

func (l *inMemoryLogfmtLogger) DebugEnabled() bool {
	return true
}
