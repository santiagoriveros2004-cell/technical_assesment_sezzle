package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"mime"
	"net/http"
	"slices"

	"github.com/santiagoriveros2004-cell/technical_assesment_sezzle/backend/internal/calculator"
)

const maxBodyBytes = 1 << 10

type operationResult struct {
	Operation string   `json:"operation"`
	Operands  []string `json:"operands"`
	Result    string   `json:"result"`
}

func (a *api) calculate(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("operation")
	operation, ok := calculator.Lookup(name)
	if !ok {
		writeProblem(w, http.StatusNotFound, codeUnknownOperation, fmt.Sprintf("operation %q is not supported", name), "")
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeProblem(w, http.StatusMethodNotAllowed, codeMethodNotAllowed, fmt.Sprintf("method %s is not allowed, use POST", r.Method), "")
		return
	}

	operands, err := readOperands(w, r, operation)
	if err != nil {
		writeError(w, r, a.logger, err)
		return
	}

	result, err := operation.Calculate(operands...)
	if err != nil {
		writeError(w, r, a.logger, err)
		return
	}

	writeJSON(w, http.StatusOK, operationResult{
		Operation: operation.Name(),
		Operands:  operands,
		Result:    result,
	})
}

func readOperands(w http.ResponseWriter, r *http.Request, operation calculator.Operation) ([]string, error) {
	if !isJSON(r.Header.Get("Content-Type")) {
		return nil, &requestError{
			status: http.StatusUnsupportedMediaType,
			code:   codeUnsupportedMediaType,
			detail: "Content-Type must be application/json",
		}
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		if _, tooLarge := errors.AsType[*http.MaxBytesError](err); tooLarge {
			return nil, &requestError{
				status: http.StatusRequestEntityTooLarge,
				code:   codePayloadTooLarge,
				detail: fmt.Sprintf("request body exceeds %d bytes", maxBodyBytes),
			}
		}
		return nil, invalidRequest("request body could not be read")
	}

	return decodeOperands(body, operation)
}

func isJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && mediaType == "application/json"
}

func decodeOperands(body []byte, operation calculator.Operation) ([]string, error) {
	var fields map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&fields); err != nil {
		return nil, invalidRequest("request body must be a JSON object")
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, invalidRequest("request body must contain a single JSON object")
	}

	names := operation.OperandNames()
	for _, field := range slices.Sorted(maps.Keys(fields)) {
		if !slices.Contains(names, field) {
			return nil, invalidRequest(fmt.Sprintf("field %q is not allowed for operation %s", field, operation.Name()))
		}
	}

	operands := make([]string, len(names))
	for i, name := range names {
		raw, present := fields[name]
		if !present {
			return nil, invalidRequest(fmt.Sprintf("field %q is required for operation %s", name, operation.Name()))
		}
		var operand *string
		if err := json.Unmarshal(raw, &operand); err != nil || operand == nil {
			return nil, invalidRequest(fmt.Sprintf("field %q must be a JSON string", name))
		}
		operands[i] = *operand
	}
	return operands, nil
}
