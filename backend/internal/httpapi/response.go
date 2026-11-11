package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/santiagoriveros2004-cell/technical_assesment_sezzle/backend/internal/calculator"
)

const (
	codeInvalidRequest       = "invalid_request"
	codeNotFound             = "not_found"
	codeUnknownOperation     = "unknown_operation"
	codeMethodNotAllowed     = "method_not_allowed"
	codePayloadTooLarge      = "payload_too_large"
	codeUnsupportedMediaType = "unsupported_media_type"
	codeInternalError        = "internal_error"
)

type problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
	Code   string `json:"code"`
	Field  string `json:"field,omitempty"`
}

type requestError struct {
	status int
	code   string
	detail string
}

func (e *requestError) Error() string {
	return e.detail
}

func invalidRequest(detail string) *requestError {
	return &requestError{status: http.StatusBadRequest, code: codeInvalidRequest, detail: detail}
}

func writeError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	if reqErr, ok := errors.AsType[*requestError](err); ok {
		writeProblem(w, reqErr.status, reqErr.code, reqErr.detail, "")
		return
	}

	if domainErr, ok := errors.AsType[*calculator.Error](err); ok {
		if status, known := domainStatus(domainErr.Code); known {
			writeProblem(w, status, string(domainErr.Code), domainErr.Detail, domainErr.Field)
			return
		}
	}

	logger.LogAttrs(r.Context(), slog.LevelError, "request failed",
		slog.String("error", err.Error()),
		slog.String("request_id", requestIDFrom(r.Context())),
	)
	writeInternalError(w)
}

func domainStatus(code calculator.Code) (int, bool) {
	switch code {
	case calculator.CodeInvalidNumber,
		calculator.CodeNumberTooLong,
		calculator.CodeDivisionByZero,
		calculator.CodeNegativeSquareRoot,
		calculator.CodeExponentNotInteger,
		calculator.CodeExponentOutOfRange,
		calculator.CodeResultOutOfRange:
		return http.StatusUnprocessableEntity, true
	}
	return 0, false
}

func writeInternalError(w http.ResponseWriter) {
	writeProblem(w, http.StatusInternalServerError, codeInternalError, "an unexpected error occurred", "")
}

func writeProblem(w http.ResponseWriter, status int, code, detail, field string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem{
		Type:   "about:blank",
		Title:  statusTitle(status),
		Status: status,
		Detail: detail,
		Code:   code,
		Field:  field,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func statusTitle(status int) string {
	switch status {
	case http.StatusRequestEntityTooLarge:
		return "Content Too Large"
	case http.StatusUnprocessableEntity:
		return "Unprocessable Content"
	}
	return http.StatusText(status)
}
