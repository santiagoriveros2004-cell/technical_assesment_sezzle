# 0004. Numbers transported as JSON strings

## Context

JSON numbers are parsed as IEEE 754 doubles by JavaScript and by many decoders, which loses precision for values beyond 15–17 significant digits and for most decimal fractions. The calculator supports operands of up to 32 digits.

## Decision

Operands and results are JSON strings in plain decimal notation (`^-?\d+(\.\d+)?$`). JSON numbers are rejected with `400 invalid_request`. The frontend treats values as strings and never performs arithmetic.

## Consequences

- Precision is preserved end to end.
- The frontend keeps a single source of truth for results: the service.
- Clients must quote numbers, which is documented in the OpenAPI contract and the README examples.
