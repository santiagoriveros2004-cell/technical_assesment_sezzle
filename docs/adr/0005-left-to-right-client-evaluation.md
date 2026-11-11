# 0005. Left-to-right evaluation driven by the client

## Context

The UI is a classic calculator keypad. Handheld calculators evaluate immediately, left to right: `2 + 3 × 4 =` yields `20`. Supporting operator precedence would require either buffering a full expression in the client or an expression endpoint in the service.

## Decision

The frontend keeps the keypad state (left operand, pending operator, current entry) in a pure reducer. Each time a binary operation becomes complete, it sends that single operation to the service and uses the result as the next left operand.

## Consequences

- The service stays stateless and operation-oriented.
- Behavior matches user expectations for a keypad calculator and is fully specified in the business rules.
- One request per completed operation; requests are serialized because the keypad is locked while one is in flight.
- Operator precedence is intentionally not supported.
