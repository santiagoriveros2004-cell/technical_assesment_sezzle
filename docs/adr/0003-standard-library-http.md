# 0003. Go standard library HTTP server

## Context

The service has two routes and a handful of middleware concerns. Since Go 1.22, `net/http.ServeMux` supports method matching and path wildcards.

## Decision

Use `net/http` for routing and serving, `log/slog` for structured logging, and small hand-written middleware for request IDs, access logs and panic recovery. No web framework.

## Consequences

- Minimal dependency surface and idiomatic code any Go engineer can read.
- Handlers are plain `http.Handler` values, tested with `net/http/httptest`.
- Middleware that frameworks provide for free (request ID, recovery) is written explicitly, which is a few dozen lines.
