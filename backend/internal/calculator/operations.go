package calculator

import (
	"github.com/cockroachdb/apd/v3"
)

const (
	workingPrecision = 100
	maxExponent      = 1000
)

var workContext = func() apd.Context {
	ctx := apd.BaseContext.WithPrecision(workingPrecision)
	ctx.Rounding = apd.Round05Up
	return *ctx
}()

var (
	one     = apd.New(1, 0)
	hundred = apd.New(100, 0)
)

func add(a, b *apd.Decimal) (*apd.Decimal, error) {
	var sum apd.Decimal
	_, err := workContext.Add(&sum, a, b)
	return &sum, err
}

func subtract(a, b *apd.Decimal) (*apd.Decimal, error) {
	var difference apd.Decimal
	_, err := workContext.Sub(&difference, a, b)
	return &difference, err
}

func multiply(a, b *apd.Decimal) (*apd.Decimal, error) {
	var product apd.Decimal
	_, err := workContext.Mul(&product, a, b)
	return &product, err
}

func divide(a, b *apd.Decimal) (*apd.Decimal, error) {
	if b.IsZero() {
		return nil, errDivisionByZero()
	}
	var quotient apd.Decimal
	_, err := workContext.Quo(&quotient, a, b)
	return &quotient, err
}

func percentage(a, b *apd.Decimal) (*apd.Decimal, error) {
	product, err := multiply(a, b)
	if err != nil {
		return nil, err
	}
	return divide(product, hundred)
}

func squareRoot(a *apd.Decimal) (*apd.Decimal, error) {
	if a.Sign() < 0 {
		return nil, &Error{
			Code:   CodeNegativeSquareRoot,
			Field:  "a",
			Detail: "square root of a negative number",
		}
	}
	var root apd.Decimal
	_, err := workContext.Sqrt(&root, a)
	return &root, err
}

func power(base, exponent *apd.Decimal) (*apd.Decimal, error) {
	n, err := integerExponent(exponent)
	if err != nil {
		return nil, err
	}

	switch {
	case n == 0:
		return apd.New(1, 0), nil
	case base.IsZero() && n < 0:
		return nil, errDivisionByZero()
	case base.IsZero():
		return apd.New(0, 0), nil
	case powerExceedsResultRange(base, n):
		return nil, errResultOutOfRange()
	}

	magnitude := exactPower(base, abs(n))
	if n > 0 {
		return magnitude, nil
	}
	var reciprocal apd.Decimal
	_, err = workContext.Quo(&reciprocal, one, magnitude)
	return &reciprocal, err
}

func integerExponent(exponent *apd.Decimal) (int64, error) {
	var reduced apd.Decimal
	reduced.Reduce(exponent)
	if reduced.Exponent < 0 {
		return 0, &Error{
			Code:   CodeExponentNotInteger,
			Field:  "b",
			Detail: "exponent must be an integer",
		}
	}
	if reduced.Cmp(apd.New(-maxExponent, 0)) < 0 || reduced.Cmp(apd.New(maxExponent, 0)) > 0 {
		return 0, &Error{
			Code:   CodeExponentOutOfRange,
			Field:  "b",
			Detail: "exponent must be between -1000 and 1000",
		}
	}
	return reduced.Int64()
}

func powerExceedsResultRange(base *apd.Decimal, n int64) bool {
	order := base.NumDigits() + int64(base.Exponent) - 1
	if n > 0 {
		return order*n >= maxDigits
	}
	return (order+1)*n >= maxDigits
}

func exactPower(base *apd.Decimal, n int64) *apd.Decimal {
	var result apd.Decimal
	result.Coeff.Exp(&base.Coeff, apd.NewBigInt(n), nil)
	result.Exponent = base.Exponent * int32(n)
	result.Negative = base.Negative && n%2 == 1
	return &result
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
