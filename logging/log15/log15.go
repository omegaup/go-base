package log15

import (
	"bytes"
	"fmt"
	"os"

	"github.com/go-stack/stack"
	log "github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/omegaup/go-base/v2/logging"
)

type log15Logger struct {
	l log.Logger
}

var _ logging.Logger = (*log15Logger)(nil)

// Wrap returns a new logging.Logger from an existing log.Logger.
func Wrap(l log.Logger) logging.Logger {
	return &log15Logger{l: l}
}

func (l *log15Logger) New(context map[string]interface{}) logging.Logger {
	return &log15Logger{l: l.l.New(log.Ctx(context))}
}

func (l *log15Logger) Error(msg string, context map[string]interface{}) {
	if l == nil {
		return
	}
	l.l.Error(msg, log.Ctx(context))
}

func (l *log15Logger) Warn(msg string, context map[string]interface{}) {
	if l == nil {
		return
	}
	l.l.Warn(msg, log.Ctx(context))
}

func (l *log15Logger) Info(msg string, context map[string]interface{}) {
	if l == nil {
		return
	}
	l.l.Info(msg, log.Ctx(context))
}

func (l *log15Logger) Debug(msg string, context map[string]interface{}) {
	if l == nil {
		return
	}
	l.l.Debug(msg, log.Ctx(context))
}

func (l *log15Logger) DebugEnabled() bool {
	// log15 doesn't have a good way of getting the current level.
	return false
}

// New opens a log15.Logger, and if it will be pointed to a real file,
// it installs a SIGHUP handler that will atomically reopen the file and
// redirect all future logging operations.
func New(level string, json bool) (logging.Logger, error) {
	l := log.New()
	var handler log.Handler
	if json {
		handler = log.StreamHandler(os.Stderr, log.JsonFormat())
	} else {
		handler = log.StderrHandler
	}

	// Don't log things that are chattier than level, but for errors also
	// include the stack trace.
	maxLvl, err := log.LvlFromString(level)
	if err != nil {
		return nil, err
	}
	l.SetHandler(errorCallerStackHandler(maxLvl, handler))
	return Wrap(l), nil
}

func rootCauseStackTrace(err error) errors.StackTrace {
	type causer interface {
		Cause() error
	}
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	var deepestStackTrace errors.StackTrace

	for err != nil {
		if stackTrace, ok := err.(stackTracer); ok {
			deepestStackTrace = stackTrace.StackTrace()
		}

		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return deepestStackTrace
}

// errorCallerStackHandler creates a handler that drops all logs that are less
// important than maxLvl, and also adds a stack trace to all events that are
// errors / critical, as well as the error values that have a stack trace.
func errorCallerStackHandler(maxLvl log.Lvl, handler log.Handler) log.Handler {
	callerStackHandler := log.FuncHandler(func(r *log.Record) error {
		// Get the stack trace of the call to log.Error/log.Crit.
		s := stack.Trace().TrimBelow(r.Call).TrimRuntime()
		if len(s) > 0 {
			var buf bytes.Buffer

			buf.WriteString("[")
			for i, pc := range s {
				if i > 0 {
					buf.WriteString(" ")
				}
				fmt.Fprintf(&buf, "%+n(%+v)", pc, pc)
			}
			buf.WriteString("]")

			r.Ctx = append(r.Ctx, "stack", buf.String())
		}

		// Get the stack trace of the first error value.
		for i := 1; i < len(r.Ctx); i += 2 {
			err, ok := r.Ctx[i].(error)
			if !ok {
				continue
			}

			stackTrace := rootCauseStackTrace(err)
			if stackTrace == nil {
				continue
			}

			r.Ctx = append(r.Ctx, "errstack", fmt.Sprintf("%+v", stackTrace))
			break
		}

		return handler.Log(r)
	})
	return log.FuncHandler(func(r *log.Record) error {
		if r.Lvl > maxLvl {
			return nil
		}
		if r.Lvl <= log.LvlError {
			return callerStackHandler.Log(r)
		}
		return handler.Log(r)
	})
}
