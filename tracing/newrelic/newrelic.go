package newrelic

import (
	"net/http"

	"github.com/omegaup/go-base/v3/logging"
	"github.com/omegaup/go-base/v3/tracing"

	nr "github.com/newrelic/go-agent/v3/newrelic"
)

type provider struct {
	app *nr.Application
}

// New returns a new tracing.Provider that can interact with New Relic.
func New(app *nr.Application) tracing.Provider {
	return &provider{
		app: app,
	}
}

func (p *provider) StartTransaction(name string, args ...tracing.Arg) tracing.Transaction {
	txn := &transaction{txn: p.app.StartTransaction(name)}
	txn.AddAttributes(args...)
	return txn
}

func (p *provider) StartWebTransaction(name string, w http.ResponseWriter, r *http.Request, args ...tracing.Arg) (tracing.Transaction, http.ResponseWriter, *http.Request) {
	txn := p.StartTransaction(name, args...).(*transaction)

	txn.txn.SetWebRequestHTTP(r)
	w = txn.txn.SetWebResponse(w)
	r = r.WithContext(tracing.NewContext(r.Context(), txn))

	return txn, w, r
}

func (p *provider) WrapHandle(pattern string, handler http.Handler) (string, http.Handler) {
	if p.app == nil {
		return pattern, handler
	}
	return pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var txn tracing.Transaction
		txn, w, r = p.StartWebTransaction(r.Method+" "+pattern, w, r)
		defer txn.End()

		handler.ServeHTTP(w, r)
	})
}

type transaction struct {
	txn *nr.Transaction
}

func (t *transaction) WithMetadata(log logging.Logger) logging.Logger {
	lm := t.txn.GetLinkingMetadata()
	context := make(map[string]interface{}, 6)
	addField := func(key, val string) {
		if val == "" {
			return
		}
		context[key] = val
	}
	// These constants come from
	// https://pkg.go.dev/github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin
	addField("trace.id", lm.TraceID)
	addField("span.id", lm.SpanID)
	addField("entity.name", lm.EntityName)
	addField("entity.type", lm.EntityType)
	addField("entity.guid", lm.EntityGUID)
	addField("hostname", lm.Hostname)
	return log.New(context)
}

func (t *transaction) SetName(name string) {
	t.txn.SetName(name)
}

func (t *transaction) AddAttributes(args ...tracing.Arg) {
	for _, arg := range args {
		t.txn.AddAttribute(arg.Name, arg.Value)
	}
}

func (t *transaction) StartSegment(name string) tracing.Segment {
	s := nr.StartSegment(t.txn, name)
	return &segment{s: s}
}

func (t *transaction) NoticeError(err error) {
	t.txn.NoticeError(err)
}

func (t *transaction) InsertDistributedTraceHeaders(h http.Header) {
	t.txn.InsertDistributedTraceHeaders(h)
}

func (t *transaction) AcceptDistributedTraceHeaders(tt tracing.TransportType, h http.Header) {
	t.txn.AcceptDistributedTraceHeaders(nr.TransportType(string(tt)), h)
}

func (t *transaction) End() {
	t.txn.End()
}

type segment struct {
	s *nr.Segment
}

func (s *segment) End() {
	s.s.End()
}
