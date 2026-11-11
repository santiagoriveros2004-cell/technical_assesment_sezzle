package calculator_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/santiagoriveros2004-cell/technical_assesment_sezzle/backend/internal/calculator"
)

type calculation struct {
	name      string
	operation string
	operands  []string
	want      string
	wantCode  calculator.Code
	wantField string
}

func runCalculations(t *testing.T, cases []calculation) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			operation, ok := calculator.Lookup(tc.operation)
			if !ok {
				t.Fatalf("operation %q not registered", tc.operation)
			}

			got, err := operation.Calculate(tc.operands...)

			if tc.wantCode == "" {
				if err != nil {
					t.Fatalf("Calculate(%q) error = %v, want result %q", tc.operands, err, tc.want)
				}
				if got != tc.want {
					t.Fatalf("Calculate(%q) = %q, want %q", tc.operands, got, tc.want)
				}
				return
			}

			var domainErr *calculator.Error
			if !errors.As(err, &domainErr) {
				t.Fatalf("Calculate(%q) = %q, %v; want error code %q", tc.operands, got, err, tc.wantCode)
			}
			if domainErr.Code != tc.wantCode || domainErr.Field != tc.wantField {
				t.Fatalf("Calculate(%q) error = {%q, %q}, want {%q, %q}",
					tc.operands, domainErr.Code, domainErr.Field, tc.wantCode, tc.wantField)
			}
			if domainErr.Error() == "" {
				t.Fatal("error detail is empty")
			}
		})
	}
}

func TestBusinessRuleScenarios(t *testing.T) {
	runCalculations(t, []calculation{
		{name: "#1 add is exact", operation: "add", operands: []string{"0.1", "0.2"}, want: "0.3"},
		{name: "#2 subtract", operation: "subtract", operands: []string{"5", "7.5"}, want: "-2.5"},
		{name: "#3 multiply", operation: "multiply", operands: []string{"1.5", "-2"}, want: "-3"},
		{name: "#4 divide rounds down", operation: "divide", operands: []string{"1", "3"}, want: "0.3333333333"},
		{name: "#5 divide rounds up", operation: "divide", operands: []string{"2", "3"}, want: "0.6666666667"},
		{name: "#6 divide by zero", operation: "divide", operands: []string{"10", "0"}, wantCode: calculator.CodeDivisionByZero},
		{name: "#7 zero divided by zero", operation: "divide", operands: []string{"0", "0"}, wantCode: calculator.CodeDivisionByZero},
		{name: "#8 power", operation: "power", operands: []string{"2", "10"}, want: "1024"},
		{name: "#9 negative exponent", operation: "power", operands: []string{"2", "-2"}, want: "0.25"},
		{name: "#10 fractional exponent", operation: "power", operands: []string{"2", "0.5"}, wantCode: calculator.CodeExponentNotInteger, wantField: "b"},
		{name: "#11 exponent above range", operation: "power", operands: []string{"2", "1001"}, wantCode: calculator.CodeExponentOutOfRange, wantField: "b"},
		{name: "#12 zero to negative exponent", operation: "power", operands: []string{"0", "-1"}, wantCode: calculator.CodeDivisionByZero},
		{name: "#13 zero to zero", operation: "power", operands: []string{"0", "0"}, want: "1"},
		{name: "#14 huge power", operation: "power", operands: []string{"99999", "1000"}, wantCode: calculator.CodeResultOutOfRange},
		{name: "#15 square root", operation: "sqrt", operands: []string{"2"}, want: "1.4142135624"},
		{name: "#16 negative square root", operation: "sqrt", operands: []string{"-4"}, wantCode: calculator.CodeNegativeSquareRoot, wantField: "a"},
		{name: "#17 percentage", operation: "percentage", operands: []string{"20", "150"}, want: "30"},
		{name: "#19 operand with 33 digits", operation: "add", operands: []string{strings.Repeat("1", 33), "1"}, wantCode: calculator.CodeNumberTooLong, wantField: "a"},
		{name: "#26 negative zero", operation: "multiply", operands: []string{"-0", "5"}, want: "0"},
		{name: "#27 trailing zeros removed", operation: "add", operands: []string{"1.50", "1.50"}, want: "3"},
		{name: "#28 tiny result rounds to zero", operation: "divide", operands: []string{"0.00000000001", "1"}, want: "0"},
		{name: "#29 integral exponent with zero fraction", operation: "power", operands: []string{"2.0", "3.0"}, want: "8"},
		{name: "#30 32 significant digits kept", operation: "multiply", operands: []string{"12345678901234567890123456.123456", "1"}, want: "12345678901234567890123456.123456"},
		{name: "#31 sum exceeds 32 integer digits", operation: "add", operands: []string{strings.Repeat("9", 32), "1"}, wantCode: calculator.CodeResultOutOfRange},
	})
}

func TestInvalidOperands(t *testing.T) {
	invalid := []string{"abc", "1e5", "NaN", "Infinity", "", ".5", "5.", "+5", " 1", "1 ", "1\n", "--1", "-", "-.5", "1.2.3", "0x10", "1,5", "١"}
	var cases []calculation
	for _, raw := range invalid {
		cases = append(cases,
			calculation{name: "#18 a=" + raw, operation: "add", operands: []string{raw, "1"}, wantCode: calculator.CodeInvalidNumber, wantField: "a"},
			calculation{name: "#18 b=" + raw, operation: "add", operands: []string{"1", raw}, wantCode: calculator.CodeInvalidNumber, wantField: "b"},
		)
	}
	runCalculations(t, cases)
}

func TestOperandLength(t *testing.T) {
	runCalculations(t, []calculation{
		{name: "32 integer digits", operation: "add", operands: []string{strings.Repeat("1", 32), "0"}, want: strings.Repeat("1", 32)},
		{name: "32 digits with sign and point", operation: "add", operands: []string{"-1234567890123456.1234567890123456", "0"}, want: "-1234567890123456.123456789"},
		{name: "33 digits with point", operation: "add", operands: []string{"1", "12345678901234567.1234567890123456"}, wantCode: calculator.CodeNumberTooLong, wantField: "b"},
		{name: "leading zeros count", operation: "add", operands: []string{"0" + strings.Repeat("1", 32), "0"}, wantCode: calculator.CodeNumberTooLong, wantField: "a"},
		{name: "trailing zeros count", operation: "add", operands: []string{"1." + strings.Repeat("0", 32), "0"}, wantCode: calculator.CodeNumberTooLong, wantField: "a"},
		{name: "leading zeros accepted", operation: "add", operands: []string{"007", "1"}, want: "8"},
	})
}

func TestOperandValidationOrder(t *testing.T) {
	runCalculations(t, []calculation{
		{name: "a reported before b", operation: "add", operands: []string{"x", "y"}, wantCode: calculator.CodeInvalidNumber, wantField: "a"},
		{name: "a length before b format", operation: "add", operands: []string{strings.Repeat("1", 33), "y"}, wantCode: calculator.CodeNumberTooLong, wantField: "a"},
		{name: "operand format before division by zero", operation: "divide", operands: []string{"x", "0"}, wantCode: calculator.CodeInvalidNumber, wantField: "a"},
		{name: "exponent format before base", operation: "power", operands: []string{"2", "1e3"}, wantCode: calculator.CodeInvalidNumber, wantField: "b"},
	})
}

func TestArithmetic(t *testing.T) {
	runCalculations(t, []calculation{
		{name: "add cancels to zero", operation: "add", operands: []string{"-0.1", "0.1"}, want: "0"},
		{name: "add aligns scales", operation: "add", operands: []string{"10000000000000000000000000000000", "0.0000000001"}, want: "10000000000000000000000000000000"},
		{name: "subtract zeros", operation: "subtract", operands: []string{"-0", "0"}, want: "0"},
		{name: "subtract to negative", operation: "subtract", operands: []string{"0.3", "0.5"}, want: "-0.2"},
		{name: "multiply large operands", operation: "multiply", operands: []string{"9999999999999999", "9999999999999999"}, want: "99999999999999980000000000000001"},
		{name: "multiply overflows", operation: "multiply", operands: []string{"10000000000000000", "10000000000000000"}, wantCode: calculator.CodeResultOutOfRange},
		{name: "multiply fractions", operation: "multiply", operands: []string{"0.00001", "0.00001"}, want: "0.0000000001"},
		{name: "divide exact", operation: "divide", operands: []string{"10", "4"}, want: "2.5"},
		{name: "divide negative", operation: "divide", operands: []string{"-1", "3"}, want: "-0.3333333333"},
		{name: "divide by negative", operation: "divide", operands: []string{"1", "-8"}, want: "-0.125"},
		{name: "divide repeating rounds up", operation: "divide", operands: []string{"1", "7"}, want: "0.1428571429"},
		{name: "divide by negative zero", operation: "divide", operands: []string{"1", "-0.0"}, wantCode: calculator.CodeDivisionByZero},
		{name: "divide below range rounds to zero", operation: "divide", operands: []string{"1", strings.Repeat("9", 32)}, want: "0"},
		{name: "divide above range", operation: "divide", operands: []string{"10000000000000000", "0.0000000000000001"}, wantCode: calculator.CodeResultOutOfRange},
		{name: "percentage fractional", operation: "percentage", operands: []string{"12.5", "8"}, want: "1"},
		{name: "percentage negative", operation: "percentage", operands: []string{"-50", "3"}, want: "-1.5"},
		{name: "percentage smallest step", operation: "percentage", operands: []string{"0.0001", "0.0001"}, want: "0.0000000001"},
		{name: "percentage by zero", operation: "percentage", operands: []string{"0", "-5"}, want: "0"},
		{name: "sqrt perfect square", operation: "sqrt", operands: []string{"16"}, want: "4"},
		{name: "sqrt fraction", operation: "sqrt", operands: []string{"0.0001"}, want: "0.01"},
		{name: "sqrt zero", operation: "sqrt", operands: []string{"0"}, want: "0"},
		{name: "sqrt negative zero", operation: "sqrt", operands: []string{"-0"}, want: "0"},
		{name: "sqrt small negative", operation: "sqrt", operands: []string{"-0.0001"}, wantCode: calculator.CodeNegativeSquareRoot, wantField: "a"},
		{name: "sqrt largest operand carries", operation: "sqrt", operands: []string{strings.Repeat("9", 32)}, want: "10000000000000000"},
	})
}

func TestPower(t *testing.T) {
	runCalculations(t, []calculation{
		{name: "negative base odd exponent", operation: "power", operands: []string{"-2", "3"}, want: "-8"},
		{name: "negative base even exponent", operation: "power", operands: []string{"-2", "4"}, want: "16"},
		{name: "negative base negative exponent", operation: "power", operands: []string{"-2", "-3"}, want: "-0.125"},
		{name: "fractional base negative exponent", operation: "power", operands: []string{"0.5", "-3"}, want: "8"},
		{name: "fractional base", operation: "power", operands: []string{"1.5", "2"}, want: "2.25"},
		{name: "repeating reciprocal", operation: "power", operands: []string{"3", "-1"}, want: "0.3333333333"},
		{name: "zero to positive exponent", operation: "power", operands: []string{"0", "5"}, want: "0"},
		{name: "zero to zero with fraction", operation: "power", operands: []string{"0", "0.0"}, want: "1"},
		{name: "negative zero to negative exponent", operation: "power", operands: []string{"-0", "-1"}, wantCode: calculator.CodeDivisionByZero},
		{name: "largest in range power of ten", operation: "power", operands: []string{"10", "31"}, want: "10000000000000000000000000000000"},
		{name: "power of ten out of range", operation: "power", operands: []string{"10", "32"}, wantCode: calculator.CodeResultOutOfRange},
		{name: "largest in range power of two", operation: "power", operands: []string{"2", "106"}, want: "81129638414606681695789005144064"},
		{name: "power of two out of range", operation: "power", operands: []string{"2", "107"}, wantCode: calculator.CodeResultOutOfRange},
		{name: "maximum exponent", operation: "power", operands: []string{"1", "1000"}, want: "1"},
		{name: "minimum exponent", operation: "power", operands: []string{"2", "-1000"}, want: "0"},
		{name: "exponent below range", operation: "power", operands: []string{"2", "-1001"}, wantCode: calculator.CodeExponentOutOfRange, wantField: "b"},
		{name: "fraction checked before range", operation: "power", operands: []string{"2", "1000.5"}, wantCode: calculator.CodeExponentNotInteger, wantField: "b"},
		{name: "integral exponent with fraction zeros", operation: "power", operands: []string{"2", "-1000.000"}, want: "0"},
		{name: "near one to maximum exponent", operation: "power", operands: []string{"1.0000000001", "1000"}, want: "1.0000001"},
		{name: "reciprocal of tenth reaches range", operation: "power", operands: []string{"0.1", "-31"}, want: "10000000000000000000000000000000"},
		{name: "reciprocal of tenth exceeds range", operation: "power", operands: []string{"0.1", "-32"}, wantCode: calculator.CodeResultOutOfRange},
		{name: "reciprocal of hundredth exceeds range", operation: "power", operands: []string{"0.01", "-16"}, wantCode: calculator.CodeResultOutOfRange},
		{name: "tiny base with minimum exponent", operation: "power", operands: []string{"0.0000000000000000000000000000001", "-1000"}, wantCode: calculator.CodeResultOutOfRange},
		{name: "tiny base with maximum exponent", operation: "power", operands: []string{"0.0000000000000000000000000000001", "1000"}, want: "0"},
		{name: "widest base to maximum exponent", operation: "power", operands: []string{"1.0000000000000000000000000000001", "1000"}, want: "1"},
		{name: "widest base to minimum exponent", operation: "power", operands: []string{"1.0000000000000000000000000000001", "-1000"}, want: "1"},
		{name: "widest large base to maximum exponent", operation: "power", operands: []string{"9.9999999999999999999999999999999", "1000"}, wantCode: calculator.CodeResultOutOfRange},
		{name: "widest large base to minimum exponent", operation: "power", operands: []string{"9.9999999999999999999999999999999", "-1000"}, want: "0"},
		{name: "negative base with maximum exponent", operation: "power", operands: []string{"-1.0000000001", "999"}, want: "-1.0000000999"},
	})
}

func TestRounding(t *testing.T) {
	runCalculations(t, []calculation{
		{name: "tie rounds down to even zero", operation: "add", operands: []string{"0.00000000005", "0"}, want: "0"},
		{name: "negative tie rounds to zero", operation: "multiply", operands: []string{"0.00000000005", "-1"}, want: "0"},
		{name: "tie rounds up to even", operation: "add", operands: []string{"0.00000000015", "0"}, want: "0.0000000002"},
		{name: "tie stays at even", operation: "add", operands: []string{"0.00000000025", "0"}, want: "0.0000000002"},
		{name: "above tie rounds up", operation: "add", operands: []string{"0.000000000050001", "0"}, want: "0.0000000001"},
		{name: "computed tie rounds to even", operation: "divide", operands: []string{"0.0000000003", "2"}, want: "0.0000000002"},
		{name: "exact square root tie", operation: "sqrt", operands: []string{"0.0000000000000000000025"}, want: "0"},
		{name: "exact square root tie rounds up", operation: "sqrt", operands: []string{"0.0000000000000000000225"}, want: "0.0000000002"},
		{name: "carry into new integer digit", operation: "add", operands: []string{"9.99999999995", "0"}, want: "10"},
		{name: "negative carry into new integer digit", operation: "add", operands: []string{"-99.999999999999", "0"}, want: "-100"},
		{name: "tie with 32 integer digits rounds to even", operation: "add", operands: []string{"99999999999999999999999999999998", "0.5"}, want: "99999999999999999999999999999998"},
		{name: "carry beyond 32 integer digits", operation: "add", operands: []string{strings.Repeat("9", 32), "0.5"}, wantCode: calculator.CodeResultOutOfRange},
		{name: "total digits capped at 32", operation: "divide", operands: []string{"10000000000000000000000000", "3"}, want: "3333333333333333333333333.3333333"},
		{name: "capped digits round up", operation: "divide", operands: []string{"20000000000000000000000000", "3"}, want: "6666666666666666666666666.6666667"},
		{name: "capping rounds once from the exact value", operation: "add", operands: []string{"12345678901234567890123.123456789", "0.00000000046"}, want: "12345678901234567890123.123456789"},
		{name: "capping to one fractional digit", operation: "add", operands: []string{"1234567890123456789012345678901.5", "0.06"}, want: "1234567890123456789012345678901.6"},
		{name: "capping to integer rounds half to even", operation: "add", operands: []string{"12345678901234567890123456789013", "0.5"}, want: "12345678901234567890123456789014"},
	})
}

func TestLookup(t *testing.T) {
	for name, arity := range map[string]int{
		"add": 2, "subtract": 2, "multiply": 2, "divide": 2, "power": 2, "percentage": 2, "sqrt": 1,
	} {
		operation, ok := calculator.Lookup(name)
		if !ok {
			t.Fatalf("Lookup(%q) not found", name)
		}
		if operation.Name() != name || operation.Arity() != arity {
			t.Fatalf("Lookup(%q) = {%q, %d}, want {%q, %d}", name, operation.Name(), operation.Arity(), name, arity)
		}
	}

	for _, name := range []string{"modulo", "", "ADD", "Add"} {
		if _, ok := calculator.Lookup(name); ok {
			t.Fatalf("Lookup(%q) found an operation", name)
		}
	}
}

func TestCalculateRejectsWrongOperandCount(t *testing.T) {
	sqrt, _ := calculator.Lookup("sqrt")
	add, _ := calculator.Lookup("add")

	for _, call := range []func() (string, error){
		func() (string, error) { return sqrt.Calculate("4", "2") },
		func() (string, error) { return add.Calculate("4") },
	} {
		_, err := call()
		var domainErr *calculator.Error
		if err == nil || errors.As(err, &domainErr) {
			t.Fatalf("error = %v, want a non-domain error", err)
		}
	}
}
