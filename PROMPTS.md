# Prompts

I used an AI coding called CLaude AI Code throughout this assignment, acting as a pair programmer under my direction. I set the working rules, made every product and design decision, reviewed the output and pushed back where I disagreed.


## 1. Role and working rules

> Act as a full-stack developer with 10 years of experience building web applications and microservice-based systems with React on the frontend and Go on the backend.
> - Don't assume information or business logic. Ask me everything you need.
> - Don't write any code before we have a plan and have validated the scenarios together.

## 2. Architecture and project structure

> Here is the assignment: *(full assignment text)*
>
> Write a document that explains the proposed architecture, with diagrams, and set up the main structure of the application. Add the documentation a real project would need. So I can understand how we go trough the project. Validate everything with me before implementing.

## 3. Decisions

The assistant came back with open questions instead of assumptions. These were my answers.

| Topic | Decision |
|---|---|
| User interface | Classic calculator keypad, with physical keyboard support |
| API design | One REST endpoint per operation |
| Number precision | Exact decimal arithmetic; numbers sent as strings so `0.1 + 0.2 = 0.3` |
| Optional operations | All three: exponentiation, square root and percentage |
| Percentage | `a % of b`, that is `a × b ÷ 100` |
| Rounding | At most 10 decimal places |
| Exponent | Integers only, between -1000 and 1000 |
| Input size | Up to 32 digits per number |
| Overflow | Reject results larger than 32 integer digits with a clear error |
| Deployment | Two containers (web and calculator service) orchestrated with Docker Compose |
| Language | English for the code, UI and documentation |
| Frontend stack | Vite, React, TypeScript, CSS Modules, Vitest, React Testing Library, MSW |
| Backend stack | Go standard library `net/http` with a single decimal library |
| Tooling | Plain npm and Go commands, documented in the README |

## 4. Parallel implementation

> The plan looks good. One change: use two agents, one for the backend and one for the frontend. Once we have all the plan done and set the rules we can continue with the implementing.

## 5. Agent briefs (summary)

Both agents received the same ground rules: follow the documents in `docs/` as the single source of truth, keep each change small and focused, write tests for every business scenario, and stop and report anything the documentation does not cover instead of deciding it.

**Backend.** Build the calculator service in Go. Validate every input, return clear and consistent error responses, keep the arithmetic exact, and make sure every result can be reused in the next operation. Package it in a small, secure container image.

**Frontend.** Build the calculator UI in React and TypeScript. It should behave like a handheld calculator, work with both mouse and keyboard, handle service errors gracefully, and be usable on a phone. The frontend never does arithmetic itself; it always asks the service. Package it behind nginx.

## 6. Keypad behavior review

After implementation the agents reported edge cases that the documentation did not cover. I reviewed each one:

| Case | My decision |
|---|---|
| Pressing `√` right after an operator (`9 + √`) | Ignore the key |
| Typing a digit after a square root (`9 + √16`, then `5`) | Start a fresh calculation |
| `±` on a result | Toggle its sign; ignore it right after an operator |
| When the "Service unavailable" message disappears | On the next key press |

While reviewing, I questioned whether `9 + √16` was being calculated correctly, since the display showed `4` instead of `13`. It turned out to be correct: `4` is the intermediate value of the square root, and `=` then gives `13`, the same as on a handheld calculator.

## 7. End-to-end verification

> Docker Desktop is running. Run the complete test.

## 8. Architecture review

> If this is supposed to be a microservice architecture, shouldn't the frontend and the backend run as separate instances on different ports, behind a load balancer? Right now everything goes through the same port, and that doesn't follow the idea of one instance per responsibility. Correct me if I'm wrong.
