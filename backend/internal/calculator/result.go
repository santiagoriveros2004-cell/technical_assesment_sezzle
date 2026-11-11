package calculator

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

const maxFractionDigits = 10

var resultContext = func() apd.Context {
	ctx := apd.BaseContext.WithPrecision(workingPrecision)
	ctx.Rounding = apd.RoundHalfEven
	return *ctx
}()

func normalize(value *apd.Decimal) (string, error) {
	integerDigits := integerDigitCount(value)
	if integerDigits > maxDigits {
		return "", errResultOutOfRange()
	}

	fractionDigits := min(maxFractionDigits, maxDigits-integerDigits)
	var rounded apd.Decimal
	if _, err := resultContext.Quantize(&rounded, value, -int32(fractionDigits)); err != nil {
		return "", fmt.Errorf("round result: %w", err)
	}
	if integerDigitCount(&rounded) > maxDigits {
		return "", errResultOutOfRange()
	}

	rounded.Reduce(&rounded)
	if rounded.IsZero() {
		rounded.Negative = false
	}
	return rounded.Text('f'), nil
}

func integerDigitCount(value *apd.Decimal) int64 {
	return max(value.NumDigits()+int64(value.Exponent), 1)
}
