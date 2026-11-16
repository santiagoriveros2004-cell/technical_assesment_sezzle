# Architecture

The system has two deployable units: a static React client served by nginx and a stateless Go microservice that owns all arithmetic. Both are reached through a single origin, so the browser never makes cross-origin requests.

## 1. System context

```mermaid
flowchart LR
    user([User]) -- "keypad clicks and keystrokes" --> web["Calculator web app"]
    web -- "JSON over HTTP" --> svc["calculator-service"]
```

## 2. Containers

```mermaid
flowchart LR
    browser["Browser<br/>React SPA"]
    subgraph compose["docker compose"]
        web["web<br/>nginx<br/>:8080"]
        svc["calculator-service<br/>Go net/http<br/>:8080"]
    end
    browser -- "GET /, /assets/*" --> web
    browser -- "POST /api/v1/operations/{op}" --> web
    web -- "proxy /api/*" --> svc
```

| Container | Technology | Responsibility |
|---|---|---|
| `web` | nginx serving the Vite production build | Serves static assets with long-lived cache for hashed files, falls back to `index.html`, reverse-proxies `/api/` to `calculator-service` |
| `calculator-service` | Go, standard library HTTP server, `cockroachdb/apd` | Validates input, performs exact decimal arithmetic, returns JSON or problem details |

During local development the Vite dev server replaces nginx and proxies `/api` to `localhost:8080`.

## 3. Backend components

```mermaid
flowchart TB
    main["cmd/calculator-service<br/>config, wiring, graceful shutdown"]
    subgraph httpapi["internal/httpapi"]
        mw["Middleware<br/>request ID, access log, panic recovery"]
        router["Router<br/>/api/v1/operations/{operation}, /healthz"]
        handler["Operation handler<br/>content type, body limit, strict JSON decoding"]
        problem["Problem writer<br/>domain error → status + code"]
    end
    subgraph calculator["internal/calculator"]
        registry["Operation registry<br/>name → arity + function"]
        operand["Operand parsing<br/>format and 32-digit limit"]
        ops["Operations<br/>add, subtract, multiply, divide, power, sqrt, percentage"]
        result["Result normalization<br/>10 decimals, half-even, range check"]
        errs["Typed domain errors"]
    end
    config["internal/config<br/>environment variables"]

    main --> config
    main --> mw --> router --> handler
    handler --> registry
    handler --> problem
    registry --> operand --> ops --> result
    ops --> errs
    problem --> errs
```

- `internal/calculator` has no HTTP knowledge. It exposes the operation registry and returns typed errors that carry a stable code.
- `internal/httpapi` translates HTTP into domain calls and domain errors into RFC 9457 responses. The mapping from error to status code lives in one place.
- `cmd/calculator-service` reads configuration, builds the server with explicit timeouts, and shuts down gracefully on `SIGINT`/`SIGTERM`.

### Configuration

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | Listen port |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `SHUTDOWN_TIMEOUT` | `10s` | Grace period for in-flight requests |

## 4. Frontend components

```mermaid
flowchart TB
    app["App"]
    subgraph feature["features/calculator"]
        calc["Calculator"]
        display["Display<br/>value + pending expression"]
        keypad["Keypad → Key"]
        hook["useCalculator<br/>reducer + request side effects"]
        keyboard["useKeyboardInput<br/>physical key mapping"]
        reducer["calculatorReducer<br/>pure state machine"]
        messages["errorMessages<br/>code → user message"]
    end
    client["api/calculatorClient<br/>fetch, 5 s timeout, problem parsing"]

    app --> calc
    calc --> display
    calc --> keypad
    calc --> hook
    calc --> keyboard
    hook --> reducer
    hook --> client
    calc --> messages
```

- `calculatorReducer` is pure and fully unit-tested. It decides *when* a request is needed and records it as pending; it never performs arithmetic.
- `useCalculator` executes the pending request through `calculatorClient` and dispatches the outcome back to the reducer.
- `calculatorClient` returns a discriminated union: a result, a business error with its code, or an unavailable outcome.

## 5. Request flows

### Successful chained calculation

```mermaid
sequenceDiagram
    actor U as User
    participant R as React app
    participant N as nginx
    participant S as calculator-service

    U->>R: 2 + 3
    U->>R: ×
    R->>N: POST /api/v1/operations/add {"a":"2","b":"3"}
    N->>S: proxy
    S-->>N: 200 {"result":"5"}
    N-->>R: 200
    R-->>U: display 5, pending "5 ×"
    U->>R: 4 =
    R->>N: POST /api/v1/operations/multiply {"a":"5","b":"4"}
    N->>S: proxy
    S-->>R: 200 {"result":"20"}
    R-->>U: display 20
```

### Business error

```mermaid
sequenceDiagram
    actor U as User
    participant R as React app
    participant S as calculator-service

    U->>R: 10 ÷ 0 =
    R->>S: POST /api/v1/operations/divide {"a":"10","b":"0"}
    S-->>R: 422 problem+json code=division_by_zero
    R-->>U: display "Cannot divide by zero"
```

### Service unavailable

```mermaid
sequenceDiagram
    actor U as User
    participant R as React app
    participant S as calculator-service

    U->>R: 7 + 1 =
    R-xS: POST /api/v1/operations/add
    Note over R: no response within 5 s
    R-->>U: "Service unavailable", state kept as "7 + 1"
    U->>R: =
    R->>S: POST /api/v1/operations/add {"a":"7","b":"1"}
    S-->>R: 200 {"result":"8"}
    R-->>U: display 8
```

## 6. Cross-cutting concerns

| Concern | Approach |
|---|---|
| Precision | Exact decimal arithmetic in the backend only; numbers are strings end to end |
| Validation | Defense in depth: the keypad cannot produce invalid operands, and the backend validates everything independently |
| Error contract | RFC 9457 problem details with a stable `code` |
| Observability | Structured JSON logs via `log/slog`, one access log line per request with method, path, status, duration and request ID; `X-Request-ID` is accepted (1–128 visible ASCII characters) or assigned by the service and echoed back |
| Resilience | Server read, write and idle timeouts; 1 KiB body limit; panic recovery; graceful shutdown; client-side 5 s timeout |
| Security | No CORS surface (same origin); containers run as non-root; distroless runtime image for the service |
| Statelessness | No storage; the service scales horizontally without coordination |

## 7. Decisions

Significant decisions are recorded as [architecture decision records](adr/README.md).
