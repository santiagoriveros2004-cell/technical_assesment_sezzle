package httpapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/santiagoriveros2004-cell/technical_assesment_sezzle/backend/internal/httpapi"
)

type problemBody struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
	Code   string `json:"code"`
	Field  string `json:"field"`
}

type request struct {
	method      string
	path        string
	contentType string
	body        string
	header      http.Header
}

func serve(t *testing.T, handler http.Handler, req request) *httptest.ResponseRecorder {
	t.Helper()
	method := req.method
	if method == "" {
		method = http.MethodPost
	}
	r := httptest.NewRequest(method, req.path, strings.NewReader(req.body))
	if req.contentType != "" {
		r.Header.Set("Content-Type", req.contentType)
	}
	for key, values := range req.header {
		r.Header[key] = values
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

func newHandler(t *testing.T) (http.Handler, *bytes.Buffer) {
	t.Helper()
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	return httpapi.NewHandler(logger), &logs
}

func post(operation, body string) request {
	return request{path: "/api/v1/operations/" + operation, contentType: "application/json", body: body}
}

func decodeProblem(t *testing.T, recorder *httptest.ResponseRecorder) problemBody {
	t.Helper()
	if got := recorder.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", got)
	}
	var problem problemBody
	decoder := json.NewDecoder(recorder.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&problem); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	return problem
}

func TestOperationSuccess(t *testing.T) {
	tests := []struct {
		name string
		req  request
		want string
	}{
		{name: "#1 add", req: post("add", `{"a":"0.1","b":"0.2"}`), want: `{"operation":"add","operands":["0.1","0.2"],"result":"0.3"}`},
		{name: "#2 subtract", req: post("subtract", `{"a":"5","b":"7.5"}`), want: `{"operation":"subtract","operands":["5","7.5"],"result":"-2.5"}`},
		{name: "#3 multiply", req: post("multiply", `{"a":"1.5","b":"-2"}`), want: `{"operation":"multiply","operands":["1.5","-2"],"result":"-3"}`},
		{name: "#4 divide", req: post("divide", `{"a":"1","b":"3"}`), want: `{"operation":"divide","operands":["1","3"],"result":"0.3333333333"}`},
		{name: "#8 power", req: post("power", `{"a":"2","b":"10"}`), want: `{"operation":"power","operands":["2","10"],"result":"1024"}`},
		{name: "#13 zero to zero", req: post("power", `{"a":"0","b":"0"}`), want: `{"operation":"power","operands":["0","0"],"result":"1"}`},
		{name: "#15 sqrt", req: post("sqrt", `{"a":"2"}`), want: `{"operation":"sqrt","operands":["2"],"result":"1.4142135624"}`},
		{name: "#17 percentage", req: post("percentage", `{"a":"20","b":"150"}`), want: `{"operation":"percentage","operands":["20","150"],"result":"30"}`},
		{name: "#26 operands echoed as received", req: post("multiply", `{"a":"-0","b":"5"}`), want: `{"operation":"multiply","operands":["-0","5"],"result":"0"}`},
		{name: "#29 echo keeps written form", req: post("power", `{"b":"3.0","a":"2.0"}`), want: `{"operation":"power","operands":["2.0","3.0"],"result":"8"}`},
		{name: "content type with charset", req: request{path: "/api/v1/operations/add", contentType: "application/json; charset=utf-8", body: `{"a":"1","b":"2"}`}, want: `{"operation":"add","operands":["1","2"],"result":"3"}`},
		{name: "content type is case insensitive", req: request{path: "/api/v1/operations/add", contentType: "Application/JSON", body: `{"a":"1","b":"2"}`}, want: `{"operation":"add","operands":["1","2"],"result":"3"}`},
		{name: "surrounding whitespace", req: post("add", " \n{\"a\":\"1\",\"b\":\"2\"}\n "), want: `{"operation":"add","operands":["1","2"],"result":"3"}`},
		{name: "escaped string operand", req: post("add", `{"a":"1","b":"2"}`), want: `{"operation":"add","operands":["1","2"],"result":"3"}`},
		{name: "body of exactly 1 KiB", req: post("add", `{"a":"1","b":"2"}`+strings.Repeat(" ", 1024-len(`{"a":"1","b":"2"}`))), want: `{"operation":"add","operands":["1","2"],"result":"3"}`},
	}

	handler, _ := newHandler(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := serve(t, handler, tc.req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body %s", recorder.Code, recorder.Body)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", got)
			}
			if got := strings.TrimSpace(recorder.Body.String()); got != tc.want {
				t.Fatalf("body = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestOperationErrors(t *testing.T) {
	tooLong := strings.Repeat("1", 33)
	oversized := `{"a":"1","b":"2"}` + strings.Repeat(" ", 1025-len(`{"a":"1","b":"2"}`))

	tests := []struct {
		name       string
		req        request
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{name: "#6 division by zero", req: post("divide", `{"a":"10","b":"0"}`), wantStatus: 422, wantCode: "division_by_zero"},
		{name: "#7 zero by zero", req: post("divide", `{"a":"0","b":"0"}`), wantStatus: 422, wantCode: "division_by_zero"},
		{name: "#10 exponent not integer", req: post("power", `{"a":"2","b":"0.5"}`), wantStatus: 422, wantCode: "exponent_not_integer", wantField: "b"},
		{name: "#11 exponent out of range", req: post("power", `{"a":"2","b":"1001"}`), wantStatus: 422, wantCode: "exponent_out_of_range", wantField: "b"},
		{name: "#12 zero to negative exponent", req: post("power", `{"a":"0","b":"-1"}`), wantStatus: 422, wantCode: "division_by_zero"},
		{name: "#14 result out of range", req: post("power", `{"a":"99999","b":"1000"}`), wantStatus: 422, wantCode: "result_out_of_range"},
		{name: "#16 negative square root", req: post("sqrt", `{"a":"-4"}`), wantStatus: 422, wantCode: "negative_square_root", wantField: "a"},
		{name: "#18 invalid a", req: post("add", `{"a":"1e5","b":"1"}`), wantStatus: 422, wantCode: "invalid_number", wantField: "a"},
		{name: "#18 invalid b", req: post("add", `{"a":"1","b":"+5"}`), wantStatus: 422, wantCode: "invalid_number", wantField: "b"},
		{name: "#18 empty operand", req: post("sqrt", `{"a":""}`), wantStatus: 422, wantCode: "invalid_number", wantField: "a"},
		{name: "#18 a reported before b", req: post("add", `{"a":"NaN","b":"Infinity"}`), wantStatus: 422, wantCode: "invalid_number", wantField: "a"},
		{name: "#19 number too long", req: post("add", `{"a":"1","b":"`+tooLong+`"}`), wantStatus: 422, wantCode: "number_too_long", wantField: "b"},
		{name: "#20 missing b", req: post("add", `{"a":"1"}`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "#20 missing a", req: post("add", `{"b":"1"}`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "#20 b on sqrt", req: post("sqrt", `{"a":"4","b":"2"}`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "#20 unknown field", req: post("add", `{"a":"1","b":"2","c":"3"}`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "#20 field names are case sensitive", req: post("add", `{"A":"1","b":"2"}`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "#21 malformed JSON", req: post("add", `{"a":"1","b":`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "#21 number operands", req: post("add", `{"a": 1, "b": 2}`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "#21 number b", req: post("add", `{"a":"1","b":2}`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "null operand", req: post("add", `{"a":null,"b":"2"}`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "boolean operand", req: post("sqrt", `{"a":true}`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "empty body", req: post("add", ``), wantStatus: 400, wantCode: "invalid_request"},
		{name: "null body", req: post("add", `null`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "array body", req: post("add", `["1","2"]`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "trailing object", req: post("add", `{"a":"1","b":"2"}{}`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "trailing garbage", req: post("add", `{"a":"1","b":"2"} x`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "#22 text content type", req: request{path: "/api/v1/operations/add", contentType: "text/plain", body: `{"a":"1","b":"2"}`}, wantStatus: 415, wantCode: "unsupported_media_type"},
		{name: "#22 missing content type", req: request{path: "/api/v1/operations/add", body: `{"a":"1","b":"2"}`}, wantStatus: 415, wantCode: "unsupported_media_type"},
		{name: "#22 malformed content type", req: request{path: "/api/v1/operations/add", contentType: "application/json; charset", body: `{"a":"1","b":"2"}`}, wantStatus: 415, wantCode: "unsupported_media_type"},
		{name: "#22 json suffix type", req: request{path: "/api/v1/operations/add", contentType: "application/problem+json", body: `{"a":"1","b":"2"}`}, wantStatus: 415, wantCode: "unsupported_media_type"},
		{name: "#23 body over 1 KiB", req: post("add", oversized), wantStatus: 413, wantCode: "payload_too_large"},
		{name: "#24 unknown operation", req: post("modulo", `{"a":"1","b":"2"}`), wantStatus: 404, wantCode: "unknown_operation"},
		{name: "#24 operation names are case sensitive", req: post("Add", `{"a":"1","b":"2"}`), wantStatus: 404, wantCode: "unknown_operation"},
		{name: "#25 GET on operation", req: request{method: http.MethodGet, path: "/api/v1/operations/add"}, wantStatus: 405, wantCode: "method_not_allowed"},
		{name: "#25 PUT on operation", req: request{method: http.MethodPut, path: "/api/v1/operations/sqrt", contentType: "application/json", body: `{"a":"4"}`}, wantStatus: 405, wantCode: "method_not_allowed"},
		{name: "#31 out of range", req: post("add", `{"a":"99999999999999999999999999999999","b":"1"}`), wantStatus: 422, wantCode: "result_out_of_range"},
		{name: "unknown path", req: request{method: http.MethodGet, path: "/unknown"}, wantStatus: 404, wantCode: "not_found"},
		{name: "root path", req: request{method: http.MethodGet, path: "/"}, wantStatus: 404, wantCode: "not_found"},
		{name: "operations collection", req: post("", `{}`), wantStatus: 404, wantCode: "not_found"},
		{name: "nested operation path", req: post("add/extra", `{"a":"1","b":"2"}`), wantStatus: 404, wantCode: "not_found"},
		{name: "unversioned path", req: request{path: "/api/operations/add", contentType: "application/json", body: `{"a":"1","b":"2"}`}, wantStatus: 404, wantCode: "not_found"},
		{name: "POST on health", req: request{path: "/healthz", contentType: "application/json", body: `{}`}, wantStatus: 404, wantCode: "not_found"},
		{name: "route checked before method", req: request{method: http.MethodGet, path: "/api/v1/operations/modulo"}, wantStatus: 404, wantCode: "unknown_operation"},
		{name: "method checked before content type", req: request{method: http.MethodDelete, path: "/api/v1/operations/add", contentType: "text/plain"}, wantStatus: 405, wantCode: "method_not_allowed"},
		{name: "content type checked before body size", req: request{path: "/api/v1/operations/add", contentType: "text/plain", body: oversized}, wantStatus: 415, wantCode: "unsupported_media_type"},
		{name: "body size checked before JSON shape", req: post("add", `{"a":`+strings.Repeat("x", 2048)), wantStatus: 413, wantCode: "payload_too_large"},
		{name: "JSON shape checked before operands", req: post("add", `{"a":"x"}`), wantStatus: 400, wantCode: "invalid_request"},
		{name: "operand format checked before constraints", req: post("divide", `{"a":"x","b":"0"}`), wantStatus: 422, wantCode: "invalid_number", wantField: "a"},
	}

	handler, _ := newHandler(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := serve(t, handler, tc.req)

			if recorder.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body %s", recorder.Code, tc.wantStatus, recorder.Body)
			}
			problem := decodeProblem(t, recorder)
			want := problemBody{
				Type:   "about:blank",
				Title:  problem.Title,
				Status: tc.wantStatus,
				Detail: problem.Detail,
				Code:   tc.wantCode,
				Field:  tc.wantField,
			}
			if problem != want {
				t.Fatalf("problem = %+v, want %+v", problem, want)
			}
			if problem.Title == "" || problem.Detail == "" {
				t.Fatalf("problem has empty title or detail: %+v", problem)
			}
		})
	}
}

func TestProblemTitles(t *testing.T) {
	tests := []struct {
		req       request
		wantTitle string
	}{
		{req: post("add", `{"a":"1"}`), wantTitle: "Bad Request"},
		{req: post("modulo", `{}`), wantTitle: "Not Found"},
		{req: request{method: http.MethodGet, path: "/api/v1/operations/add"}, wantTitle: "Method Not Allowed"},
		{req: post("add", strings.Repeat(" ", 1025)), wantTitle: "Content Too Large"},
		{req: request{path: "/api/v1/operations/add", contentType: "text/plain"}, wantTitle: "Unsupported Media Type"},
		{req: post("divide", `{"a":"1","b":"0"}`), wantTitle: "Unprocessable Content"},
	}

	handler, _ := newHandler(t)
	for _, tc := range tests {
		problem := decodeProblem(t, serve(t, handler, tc.req))
		if problem.Title != tc.wantTitle {
			t.Fatalf("title for status %d = %q, want %q", problem.Status, problem.Title, tc.wantTitle)
		}
	}
}

func TestProblemDetailsMatchContractExamples(t *testing.T) {
	handler, _ := newHandler(t)

	tests := []struct {
		req  request
		want string
	}{
		{
			req:  post("add", `{"a":"1"}`),
			want: `{"type":"about:blank","title":"Bad Request","status":400,"detail":"field \"b\" is required for operation add","code":"invalid_request"}`,
		},
		{
			req:  post("modulo", `{"a":"1","b":"2"}`),
			want: `{"type":"about:blank","title":"Not Found","status":404,"detail":"operation \"modulo\" is not supported","code":"unknown_operation"}`,
		},
		{
			req:  post("divide", `{"a":"1","b":"0"}`),
			want: `{"type":"about:blank","title":"Unprocessable Content","status":422,"detail":"division by zero","code":"division_by_zero"}`,
		},
		{
			req:  post("add", `{"a":"`+strings.Repeat("1", 33)+`","b":"1"}`),
			want: `{"type":"about:blank","title":"Unprocessable Content","status":422,"detail":"operand \"a\" exceeds 32 digits","code":"number_too_long","field":"a"}`,
		},
	}

	for _, tc := range tests {
		recorder := serve(t, handler, tc.req)
		if got := strings.TrimSpace(recorder.Body.String()); got != tc.want {
			t.Fatalf("body = %s, want %s", got, tc.want)
		}
	}
}

func TestMethodNotAllowedSetsAllowHeader(t *testing.T) {
	handler, _ := newHandler(t)
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions} {
		recorder := serve(t, handler, request{method: method, path: "/api/v1/operations/add"})
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s status = %d, want 405", method, recorder.Code)
		}
		if got := recorder.Header().Get("Allow"); got != http.MethodPost {
			t.Fatalf("%s Allow = %q, want POST", method, got)
		}
	}
}

func TestHealth(t *testing.T) {
	handler, _ := newHandler(t)
	recorder := serve(t, handler, request{method: http.MethodGet, path: "/healthz"})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	if got := strings.TrimSpace(recorder.Body.String()); got != `{"status":"ok"}` {
		t.Fatalf("body = %s, want {\"status\":\"ok\"}", got)
	}
}

func TestRequestID(t *testing.T) {
	handler, _ := newHandler(t)

	t.Run("generated when absent", func(t *testing.T) {
		first := serve(t, handler, request{method: http.MethodGet, path: "/healthz"}).Header().Get("X-Request-ID")
		second := serve(t, handler, request{method: http.MethodGet, path: "/healthz"}).Header().Get("X-Request-ID")
		if first == "" || second == "" || first == second {
			t.Fatalf("generated request IDs = %q, %q; want distinct non-empty values", first, second)
		}
	})

	t.Run("echoed when provided", func(t *testing.T) {
		req := post("add", `{"a":"1","b":"2"}`)
		req.header = http.Header{"X-Request-Id": {"client-id-123"}}
		if got := serve(t, handler, req).Header().Get("X-Request-ID"); got != "client-id-123" {
			t.Fatalf("X-Request-ID = %q, want client-id-123", got)
		}
	})

	t.Run("echoed on error responses", func(t *testing.T) {
		req := post("divide", `{"a":"1","b":"0"}`)
		req.header = http.Header{"X-Request-Id": {"failing-request"}}
		if got := serve(t, handler, req).Header().Get("X-Request-ID"); got != "failing-request" {
			t.Fatalf("X-Request-ID = %q, want failing-request", got)
		}
	})

	for name, unsafe := range map[string]string{
		"replaced when it contains spaces":        "two words",
		"replaced when it contains control bytes": "id\x00",
		"replaced when it is too long":            strings.Repeat("x", 129),
	} {
		t.Run(name, func(t *testing.T) {
			req := request{method: http.MethodGet, path: "/healthz", header: http.Header{"X-Request-Id": {unsafe}}}
			got := serve(t, handler, req).Header().Get("X-Request-ID")
			if got == "" || got == unsafe {
				t.Fatalf("X-Request-ID = %q, want a generated value", got)
			}
		})
	}
}

func TestAccessLog(t *testing.T) {
	handler, logs := newHandler(t)
	req := post("divide", `{"a":"1","b":"0"}`)
	req.header = http.Header{"X-Request-Id": {"logged-request"}}
	serve(t, handler, req)

	var entry struct {
		Msg       string `json:"msg"`
		Method    string `json:"method"`
		Path      string `json:"path"`
		Status    int    `json:"status"`
		Duration  *int64 `json:"duration"`
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(logs).Decode(&entry); err != nil {
		t.Fatalf("decode access log: %v", err)
	}
	if entry.Method != http.MethodPost || entry.Path != "/api/v1/operations/divide" ||
		entry.Status != http.StatusUnprocessableEntity || entry.Duration == nil || entry.RequestID != "logged-request" {
		t.Fatalf("access log = %+v", entry)
	}
	if rest, _ := io.ReadAll(logs); len(bytes.TrimSpace(rest)) != 0 {
		t.Fatalf("unexpected extra log lines: %s", rest)
	}
}
