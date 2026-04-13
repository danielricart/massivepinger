package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newTestServer(t *testing.T) *http.Server {
	t.Helper()
	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGoCollector())
	return New(9999, reg)
}

func doMetricsRequest(t *testing.T, srv *http.Server) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)
	return rec
}

func TestMetricsReturns200(t *testing.T) {
	rec := doMetricsRequest(t, newTestServer(t))

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestMetricsContentTypeIsTextPlain(t *testing.T) {
	rec := doMetricsRequest(t, newTestServer(t))

	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected Content-Type to start with text/plain, got %q", ct)
	}
}

func TestMetricsBodyIsNonEmpty(t *testing.T) {
	rec := doMetricsRequest(t, newTestServer(t))

	if rec.Body.Len() == 0 {
		t.Error("expected non-empty response body")
	}
}
