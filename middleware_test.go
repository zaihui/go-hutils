package hutils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
)

type Ping struct{}

func TestSkywalkingHTTPMiddlewareTestSuite(t *testing.T) {
	report, err := reporter.NewLogReporter()
	if err != nil {
		return
	}
	tracer, err := go2sky.NewTracer("test", go2sky.WithReporter(report))
	if err != nil {
		return
	}
	middleware, err := NewServerSkywalkingHTTPMiddleware(tracer)
	if err != nil {
		return
	}
	mux := http.NewServeMux()
	mux.Handle("/ping", middleware(&Ping{}))

	req := httptest.NewRequest("GET", "/ping", nil)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)

	response := rw.Body.String()
	if response != "OK" {
		t.Errorf("Response gotten was %q", response)
	}
}

func (p *Ping) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
