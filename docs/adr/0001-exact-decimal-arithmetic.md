# 0001. Exact decimal arithmetic with cockroachdb/apd

## Context

Binary floating point cannot represent most decimal fractions: `0.1 + 0.2` evaluates to `0.30000000000000004` in `float64`. Users of a calculator expect decimal results, and in a payments domain silent rounding errors are unacceptable. The service also needs square roots, integer powers, a fixed rounding mode and detection of division by zero and overflow.

## Decision

Perform all arithmetic with `github.com/cockroachdb/apd/v3`, an arbitrary-precision implementation of the General Decimal Arithmetic specification. Results are rounded to 10 fractional digits with half-even rounding.

`shopspring/decimal` was considered. It has no square root, so a hand-written Newton iteration with its own convergence and rounding rules would be needed. `math/big.Float` is binary, so it does not remove representation error.

## Consequences

- Addition, subtraction, multiplication and percentage are exact before the final rounding step.
- Division, square root and negative exponents run with a working precision well above the 42 significant digits a result can have, so the final half-even rounding is correct.
- apd conditions (division by zero, invalid operation, overflow) map directly to domain errors.
- One third-party dependency in the service.
