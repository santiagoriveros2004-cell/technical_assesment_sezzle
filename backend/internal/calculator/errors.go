package calculator

type Code string

const (
	CodeInvalidNumber      Code = "invalid_number"
	CodeNumberTooLong      Code = "number_too_long"
	CodeDivisionByZero     Code = "division_by_zero"
	CodeNegativeSquareRoot Code = "negative_square_root"
	CodeExponentNotInteger Code = "exponent_not_integer"
	CodeExponentOutOfRange Code = "exponent_out_of_range"
	CodeResultOutOfRange   Code = "result_out_of_range"
)

type Error struct {
	Code   Code
	Field  string
	Detail string
}

func (e *Error) Error() string {
	return e.Detail
}

func errDivisionByZero() error {
	return &Error{Code: CodeDivisionByZero, Detail: "division by zero"}
}

func errResultOutOfRange() error {
	return &Error{Code: CodeResultOutOfRange, Detail: "result exceeds 32 integer digits"}
}
