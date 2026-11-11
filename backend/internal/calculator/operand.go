package calculator

import (
	"fmt"
	"regexp"

	"github.com/cockroachdb/apd/v3"
)

const maxDigits = 32

var operandPattern = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

func parseOperand(field, raw string) (*apd.Decimal, error) {
	if !operandPattern.MatchString(raw) {
		return nil, &Error{
			Code:   CodeInvalidNumber,
			Field:  field,
			Detail: fmt.Sprintf("operand %q is not a plain decimal number", field),
		}
	}
	if countDigits(raw) > maxDigits {
		return nil, &Error{
			Code:   CodeNumberTooLong,
			Field:  field,
			Detail: fmt.Sprintf("operand %q exceeds %d digits", field, maxDigits),
		}
	}
	value, _, err := apd.NewFromString(raw)
	if err != nil {
		return nil, fmt.Errorf("parse operand %q: %w", field, err)
	}
	return value, nil
}

func countDigits(raw string) int {
	digits := 0
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	return digits
}
