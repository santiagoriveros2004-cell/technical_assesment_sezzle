# Testing strategy

Tests are organized around the business rules: every scenario in [business-rules.md](business-rules.md) has at least one test in the layer that owns it.

## Backend

| Package | Kind | Focus |
|---|---|---|
| `internal/calculator` | Table-driven unit tests | Every arithmetic scenario, operand validation, half-even rounding ties, rounding that carries into a new integer digit, total-digit capping, power range pre-checks |
| `internal/httpapi` | `net/http/httptest` handler tests | Status codes, problem details bodies, `Allow`, `Content-Type` and `X-Request-ID` headers, body size limit, strict JSON decoding, panic recovery, access logging |
| `internal/config` | Unit tests | Defaults, valid overrides and rejection of invalid values |

`cmd/calculator-service` only wires configuration, logger and server. It is exercised by running the binary and the container rather than by unit tests.

```sh
cd backend
gofmt -l .
go vet ./...
go test ./... -race -coverprofile=coverage.out
go tool cover -func=coverage.out
go tool cover -html=coverage.out
```

`-race` requires cgo and a C toolchain.

## Frontend

| Area | Kind | Focus |
|---|---|---|
| `calculatorReducer` | Pure unit tests | Every keypad rule: entry limits, decimal point, sign toggle, backspace, operator replacement, left-to-right chaining, square root, no-op equals, error and in-flight states, operand normalization |
| `calculatorClient` | Unit tests with MSW | Success, 422 code mapping, other statuses, network failure, 5 s timeout with fake timers, malformed bodies |
| `Calculator` and `App` | React Testing Library + user-event + MSW | Click and physical keyboard flows, error messages, service unavailable and retry, keys disabled while a request is in flight |

The MSW handler in `src/test/calculatorService.ts` follows the OpenAPI contract and records requests, so component tests assert on the exact operations sent to the service.

```sh
cd frontend
npm run lint
npm run typecheck
npm run test:coverage
```

Coverage thresholds are enforced in `vite.config.ts` for `src/features` and `src/api`. The HTML report is written to `frontend/coverage/index.html`.

## Coverage summary

| Layer | Scope | Statements | Branches |
|---|---|---|---|
| Backend | `internal/calculator` | 97.4% | n/a |
| Backend | `internal/httpapi` | 98.3% | n/a |
| Backend | `internal/config` | 100% | n/a |
| Frontend | `src/api`, `src/features`, `src/App.tsx` | 100% | 100% |

Go reports statement coverage only.

## End-to-end verification

With `docker compose up --build` running:

1. Run scenarios 1–17 from the keypad, with both clicks and the physical keyboard.
2. Run scenarios 18–31 with `curl` through `http://localhost:8080/api/v1/operations/...`.
3. Stop `calculator-service` with `docker compose stop calculator-service` and check that the UI shows `Service unavailable` and recovers after a restart.
4. Check the layout in a 320 px wide viewport.
