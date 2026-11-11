# 0007. RFC 9457 problem details for errors

## Context

Clients need to tell business errors (division by zero) from request errors (malformed JSON) and show specific messages. Free-form error strings are brittle to match.

## Decision

All error responses use `application/problem+json` (RFC 9457) with `type: about:blank`, the HTTP status text as `title`, a human-readable `detail`, a stable machine-readable `code`, and an optional `field` identifying the offending operand. Status codes follow HTTP semantics: 400 for malformed requests, 422 for well-formed requests that violate business rules.

## Consequences

- The frontend maps `code` to user messages and never parses `detail`.
- The error catalog in `docs/api/errors.md` is part of the API contract.
- Unexpected failures return `500 internal_error` without leaking internals; details go to the logs with the request ID.
