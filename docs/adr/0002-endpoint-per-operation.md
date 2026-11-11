# 0002. One REST endpoint per operation

## Context

Operations differ in arity (`sqrt` is unary) and in constraints (`power` requires an integer exponent). The API can expose a single generic endpoint that receives the operation in the body, one endpoint per operation, or an expression evaluator.

## Decision

Expose `POST /api/v1/operations/{operation}` with a request body of `{"a", "b"}` for binary operations and `{"a"}` for `sqrt`. The operation is a resource identified by the path; an operation registry in the domain maps each name to its arity and implementation.

An expression endpoint was rejected: it requires a parser and precedence rules that the keypad does not need.

## Consequences

- Each operation has a precise contract, documented and validated independently.
- Unsupported operations are a clean `404 unknown_operation`.
- Adding an operation means adding one registry entry and its tests; routing does not change.
- Path versioning (`/v1`) leaves room for breaking changes.
