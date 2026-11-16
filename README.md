# Calculator

Full-stack calculator: a React + TypeScript keypad application backed by `calculator-service`, a Go REST microservice that performs exact decimal arithmetic.

Supported operations: addition, subtraction, multiplication, division, exponentiation, square root and percentage.

## Repository layout

| Path | Contents |
|---|---|
| `backend/` | `calculator-service`, the Go REST microservice |
| `frontend/` | React + TypeScript web client served by nginx in production |
| `docs/` | Architecture, business rules, API contract and architecture decision records |
| `docker-compose.yml` | Runs the full stack |

## Quick start with Docker

Requirements: Docker with Compose v2.

```sh
docker compose up --build
```

Open http://localhost:8080. nginx serves the web app and proxies `/api/` to `calculator-service`; the service itself is not published to the host.

## Running locally

Requirements: Go 1.27+, Node.js 26+.

### Backend

```sh
cd backend
go run ./cmd/calculator-service
```

The service listens on `:8080`.

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | Listen port |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `SHUTDOWN_TIMEOUT` | `10s` | Grace period for in-flight requests on shutdown |

### Frontend

```sh
cd frontend
npm ci
npm run dev
```

Open http://localhost:5173. The Vite dev server proxies `/api` to `http://localhost:8080`; set `VITE_API_PROXY_TARGET` to point it elsewhere.

| Script | Purpose |
|---|---|
| `npm run dev` | Development server |
| `npm run build` | Type-check and production build into `dist/` |
| `npm run preview` | Serve the production build |
| `npm run lint` | ESLint |
| `npm run typecheck` | TypeScript project check |
| `npm test` | Unit and component tests |
| `npm run test:coverage` | Tests with coverage report in `coverage/` |

## API

Full contract: [`docs/api/openapi.yaml`](docs/api/openapi.yaml). Error codes: [`docs/api/errors.md`](docs/api/errors.md).

Operands and results are **decimal strings**, never JSON numbers.

```
POST /api/v1/operations/{add|subtract|multiply|divide|power|percentage}   {"a": "...", "b": "..."}
POST /api/v1/operations/sqrt                                               {"a": "..."}
GET  /healthz
```

Examples, against the backend on port 8080 or the Docker stack:

```sh
curl -s -X POST http://localhost:8080/api/v1/operations/add \
  -H 'Content-Type: application/json' -d '{"a":"0.1","b":"0.2"}'
# {"operation":"add","operands":["0.1","0.2"],"result":"0.3"}

curl -s -X POST http://localhost:8080/api/v1/operations/divide \
  -H 'Content-Type: application/json' -d '{"a":"1","b":"3"}'
# {"operation":"divide","operands":["1","3"],"result":"0.3333333333"}

curl -s -X POST http://localhost:8080/api/v1/operations/power \
  -H 'Content-Type: application/json' -d '{"a":"2","b":"-2"}'
# {"operation":"power","operands":["2","-2"],"result":"0.25"}

curl -s -X POST http://localhost:8080/api/v1/operations/sqrt \
  -H 'Content-Type: application/json' -d '{"a":"2"}'
# {"operation":"sqrt","operands":["2"],"result":"1.4142135624"}

curl -s -X POST http://localhost:8080/api/v1/operations/percentage \
  -H 'Content-Type: application/json' -d '{"a":"20","b":"150"}'
# {"operation":"percentage","operands":["20","150"],"result":"30"}
```

Errors use RFC 9457 problem details with a stable `code`:

```sh
curl -s -X POST http://localhost:8080/api/v1/operations/divide \
  -H 'Content-Type: application/json' -d '{"a":"10","b":"0"}'
# HTTP 422
# {"type":"about:blank","title":"Unprocessable Content","status":422,"detail":"division by zero","code":"division_by_zero"}
```

## Tests and coverage

```sh
cd backend && go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out
cd frontend && npm run test:coverage
```

| Layer | Package | Statement coverage |
|---|---|---|
| Backend | `internal/calculator` | 97.4% |
| Backend | `internal/httpapi` | 98.3% |
| Backend | `internal/config` | 100% |
| Frontend | `src/api`, `src/features` | 100% |

The testing strategy is described in [`docs/testing.md`](docs/testing.md).

## Design decisions

- **Exact decimal arithmetic.** The service uses `cockroachdb/apd`, so `0.1 + 0.2 = 0.3`. Results are rounded half-even to at most 10 decimals. ([ADR 0001](docs/adr/0001-exact-decimal-arithmetic.md))
- **Numbers as strings.** Operands of up to 32 digits travel as JSON strings to avoid IEEE 754 precision loss. The frontend never performs arithmetic. ([ADR 0004](docs/adr/0004-numbers-as-strings.md))
- **One endpoint per operation.** Each operation has its own contract and arity, and unknown operations return a clean 404. ([ADR 0002](docs/adr/0002-endpoint-per-operation.md))
- **Standard library HTTP.** `net/http` routing, `log/slog` structured logs, and small explicit middleware for request IDs, access logs and panic recovery. ([ADR 0003](docs/adr/0003-standard-library-http.md))
- **Handheld calculator semantics.** The keypad evaluates left to right (`2 + 3 × 4 = 20`). A pure reducer models the keypad as a state machine and sends one request per completed operation. ([ADR 0005](docs/adr/0005-left-to-right-client-evaluation.md))
- **Same origin.** nginx in production and Vite in development proxy `/api`, so there is no CORS surface and one build runs everywhere. ([ADR 0006](docs/adr/0006-same-origin-reverse-proxy.md))
- **Problem details.** Errors carry a stable `code` that the frontend maps to user messages. ([ADR 0007](docs/adr/0007-problem-details-errors.md))

### Assumptions

The exact behavior was agreed before implementation and is specified in [`docs/business-rules.md`](docs/business-rules.md):

- `percentage(a, b)` computes `a % of b`, that is `a × b ÷ 100`.
- `power` accepts only integer exponents between -1000 and 1000, and `0 ^ 0 = 1`.
- Operands are limited to 32 digits. A result whose integer part exceeds 32 digits is rejected, so every result can be chained into the next operation.
- Division by zero, the square root of a negative number and out-of-range results are business errors (HTTP 422).

## Documentation

- [Architecture](docs/architecture.md): context, container and component diagrams and request flows
- [Business rules](docs/business-rules.md): number model, rounding, keypad behavior and acceptance scenarios
- [API contract](docs/api/openapi.yaml) and [error catalog](docs/api/errors.md)
- [Architecture decision records](docs/adr/README.md)
- [Testing strategy](docs/testing.md)
