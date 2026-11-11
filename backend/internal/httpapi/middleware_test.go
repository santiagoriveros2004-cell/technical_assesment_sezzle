package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/santiagoriveros2004-cell/technical_assesment_sezzle/backend/internal/calculator"
)

func newLogger() (*slog.Logger, *bytes.Buffer) {
	var logs bytes.Buffer
	return slog.New(slog.NewJSONHandler(&logs, nil)), &logs
}

func chain(logger *slog.Logger, handler http.Handler) http.Handler {
	return withRequestID(withAccessLog(logger, withPanicRecovery(logger, handler)))
}

func TestPanicRecovery(t *testing.T) {
	logger, logs := newLogger()
	handler := chain(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("secret database password")
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/operations/add", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", got)
	}
	if recorder.Header().Get(requestIDHeader) == "" {
		t.Fatal("X-Request-ID missing on recovered response")
	}
	body := recorder.Body.String()
	if strings.Contains(body, "secret") {
		t.Fatalf("response leaks panic details: %s", body)
	}
	var got problem
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	if got.Code != codeInternalError || got.Status != http.StatusInternalServerError || got.Title != "Internal Server Error" {
		t.Fatalf("problem = %+v", got)
	}
	if !strings.Contains(logs.String(), "secret database password") || !strings.Contains(logs.String(), `"status":500`) {
		t.Fatalf("logs do not record the panic and the 500 status: %s", logs)
	}
}

func TestPanicRecoveryPropagatesAbortHandler(t *testing.T) {
	logger, _ := newLogger()
	handler := withPanicRecovery(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	defer func() {
		if recovered := recover(); recovered != http.ErrAbortHandler {
			t.Fatalf("recovered = %v, want http.ErrAbortHandler", recovered)
		}
	}()
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestAccessLogDefaultsToOK(t *testing.T) {
	logger, logs := newLogger()
	handler := chain(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(logs.String(), `"status":200`) {
		t.Fatalf("access log = %s, want status 200", logs)
	}
}

func TestStatusRecorderKeepsFirstStatus(t *testing.T) {
	recorder := &statusRecorder{ResponseWriter: httptest.NewRecorder()}
	if _, err := recorder.Write([]byte("ok")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	recorder.WriteHeader(http.StatusTeapot)

	if recorder.statusOrOK() != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.statusOrOK())
	}
	if recorder.Unwrap() == nil {
		t.Fatal("Unwrap() returned nil")
	}
}

func TestWriteErrorHidesUnexpectedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "plain error", err: errors.New("disk on fire")},
		{name: "domain error with unmapped code", err: &calculator.Error{Code: "disk_on_fire", Detail: "disk on fire"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger, logs := newLogger()
			recorder := httptest.NewRecorder()

			writeError(recorder, httptest.NewRequest(http.MethodPost, "/", nil), logger, tc.err)

			if recorder.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500", recorder.Code)
			}
			if strings.Contains(recorder.Body.String(), "disk on fire") {
				t.Fatalf("response leaks error details: %s", recorder.Body)
			}
			if !strings.Contains(logs.String(), "disk on fire") {
				t.Fatalf("error not logged: %s", logs)
			}
		})
	}
}
