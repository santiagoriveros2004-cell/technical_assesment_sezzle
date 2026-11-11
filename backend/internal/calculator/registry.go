package calculator

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

type Operation struct {
	name    string
	arity   int
	compute func(operands []*apd.Decimal) (*apd.Decimal, error)
}

var registry = indexByName(
	binary("add", add),
	binary("subtract", subtract),
	binary("multiply", multiply),
	binary("divide", divide),
	binary("power", power),
	unary("sqrt", squareRoot),
	binary("percentage", percentage),
)

var operandFields = [...]string{"a", "b"}

func Lookup(name string) (Operation, bool) {
	operation, ok := registry[name]
	return operation, ok
}

func (o Operation) Name() string {
	return o.name
}

func (o Operation) Arity() int {
	return o.arity
}

func (o Operation) Calculate(operands ...string) (string, error) {
	if len(operands) != o.arity {
		return "", fmt.Errorf("operation %s expects %d operands, got %d", o.name, o.arity, len(operands))
	}

	values := make([]*apd.Decimal, len(operands))
	for i, raw := range operands {
		value, err := parseOperand(operandFields[i], raw)
		if err != nil {
			return "", err
		}
		values[i] = value
	}

	result, err := o.compute(values)
	if err != nil {
		return "", err
	}
	return normalize(result)
}

func unary(name string, apply func(a *apd.Decimal) (*apd.Decimal, error)) Operation {
	return Operation{
		name:  name,
		arity: 1,
		compute: func(operands []*apd.Decimal) (*apd.Decimal, error) {
			return apply(operands[0])
		},
	}
}

func binary(name string, apply func(a, b *apd.Decimal) (*apd.Decimal, error)) Operation {
	return Operation{
		name:  name,
		arity: 2,
		compute: func(operands []*apd.Decimal) (*apd.Decimal, error) {
			return apply(operands[0], operands[1])
		},
	}
}

func indexByName(operations ...Operation) map[string]Operation {
	index := make(map[string]Operation, len(operations))
	for _, operation := range operations {
		index[operation.name] = operation
	}
	return index
}
