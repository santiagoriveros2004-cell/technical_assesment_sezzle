# Business rules

This document is the single source of truth for calculator behavior. The backend enforces the arithmetic rules; the frontend enforces the keypad rules. Both are tested against the scenarios listed here.

## 1. Operations

| Operation | Endpoint | Arity | Definition |
|---|---|---|---|
| Addition | `add` | binary | `a + b` |
| Subtraction | `subtract` | binary | `a - b` |
| Multiplication | `multiply` | binary | `a × b` |
| Division | `divide` | binary | `a ÷ b` |
| Exponentiation | `power` | binary | `a ^ b`, integer exponent only |
| Square root | `sqrt` | unary | `√a` |
| Percentage | `percentage` | binary | `a % of b`, i.e. `a × b ÷ 100` |

## 2. Number model

All numbers are exact decimals transported as JSON strings. Binary floating point is never used, so `0.1 + 0.2` is exactly `0.3`.

### 2.1 Operand format

- An operand must match `^-?\d+(\.\d+)?$`: optional minus sign, at least one integer digit, optional fractional part with at least one digit.
- Rejected forms include `+5`, `.5`, `5.`, `1e5`, `NaN`, `Infinity`, empty strings, whitespace and JSON numbers.
- An operand may contain at most **32 digits**, counting every digit character as written (integer and fractional), excluding the sign and the decimal point.

### 2.2 Result format

Results are computed exactly (or with enough working precision for square roots and negative exponents) and then normalized:

1. Rounded to at most **10 fractional digits** using **half-even** rounding.
2. If the integer part has more than **32 digits**, the operation fails with `result_out_of_range`.
3. If integer digits plus fractional digits exceed 32, fractional digits are rounded away (half-even) until the total is 32. Rule 2 is re-checked after this rounding.
4. Trailing fractional zeros and a trailing decimal point are removed.
5. Negative zero is normalized to `0`.
6. A non-zero result whose magnitude is below `0.00000000005` rounds to `0`.

These rules guarantee that every result is a valid operand, so results can always be chained into the next operation.

### 2.3 Operation constraints

| Rule | Error code |
|---|---|
| Divisor is zero, including `0 ÷ 0` | `division_by_zero` |
| Square root of a negative number | `negative_square_root` |
| Exponent has a non-zero fractional part | `exponent_not_integer` |
| Exponent outside `-1000 ≤ b ≤ 1000` | `exponent_out_of_range` |
| Zero raised to a negative exponent | `division_by_zero` |
| `0 ^ 0` | defined as `1` |

An exponent written with a zero fractional part, such as `2.0`, is an integer.

## 3. Arithmetic scenarios

| # | Request | Expected response |
|---|---|---|
| 1 | `add` `0.1`, `0.2` | 200 `"0.3"` |
| 2 | `subtract` `5`, `7.5` | 200 `"-2.5"` |
| 3 | `multiply` `1.5`, `-2` | 200 `"-3"` |
| 4 | `divide` `1`, `3` | 200 `"0.3333333333"` |
| 5 | `divide` `2`, `3` | 200 `"0.6666666667"` |
| 6 | `divide` `10`, `0` | 422 `division_by_zero` |
| 7 | `divide` `0`, `0` | 422 `division_by_zero` |
| 8 | `power` `2`, `10` | 200 `"1024"` |
| 9 | `power` `2`, `-2` | 200 `"0.25"` |
| 10 | `power` `2`, `0.5` | 422 `exponent_not_integer` |
| 11 | `power` `2`, `1001` | 422 `exponent_out_of_range` |
| 12 | `power` `0`, `-1` | 422 `division_by_zero` |
| 13 | `power` `0`, `0` | 200 `"1"` |
| 14 | `power` `99999`, `1000` | 422 `result_out_of_range` |
| 15 | `sqrt` `2` | 200 `"1.4142135624"` |
| 16 | `sqrt` `-4` | 422 `negative_square_root` |
| 17 | `percentage` `20`, `150` | 200 `"30"` |
| 18 | any operand `"abc"`, `"1e5"`, `"NaN"`, `"Infinity"`, `""`, `".5"`, `"5."`, `"+5"` | 422 `invalid_number` |
| 19 | any operand with 33 digits | 422 `number_too_long` |
| 20 | missing `b` on a binary operation, `b` present on `sqrt`, or an unknown field | 400 `invalid_request` |
| 21 | malformed JSON or a JSON number operand such as `{"a": 1, "b": 2}` | 400 `invalid_request` |
| 22 | `Content-Type` other than `application/json` | 415 `unsupported_media_type` |
| 23 | request body larger than 1 KiB | 413 `payload_too_large` |
| 24 | `POST /api/v1/operations/modulo` | 404 `unknown_operation` |
| 25 | `GET /api/v1/operations/add` | 405 `method_not_allowed` |
| 26 | `multiply` `-0`, `5` | 200 `"0"` |

Additional normalization scenarios:

| # | Request | Expected response |
|---|---|---|
| 27 | `add` `1.50`, `1.50` | 200 `"3"` |
| 28 | `divide` `0.00000000001`, `1` | 200 `"0"` |
| 29 | `power` `2.0`, `3.0` | 200 `"8"` |
| 30 | `multiply` `12345678901234567890123456.123456`, `1` | 200 `"12345678901234567890123456.123456"` |
| 31 | `add` `99999999999999999999999999999999`, `1` | 422 `result_out_of_range` |

When more than one operand is invalid, operand `a` is reported first.

## 4. Keypad behavior

The calculator UI behaves like a classic handheld calculator. All arithmetic is delegated to the backend; the frontend only edits strings.

### 4.1 Keys

| Key | Label | Physical keyboard |
|---|---|---|
| Digits | `0`–`9` | `0`–`9` |
| Decimal point | `.` | `.` and `,` |
| Sign toggle | `±` | none |
| Add, subtract, multiply, divide | `+ − × ÷` | `+ - * /` |
| Power | `xʸ` | `^` |
| Percentage | `%` | `%` |
| Square root | `√` | none |
| Equals | `=` | `Enter` and `=` |
| Backspace | `⌫` | `Backspace` |
| Clear | `C` | `Escape` |

### 4.2 Evaluation model

- Operations are evaluated **left to right**, without operator precedence: `2 + 3 × 4 =` displays `20`.
- The display shows the operand being typed or the last result. A secondary line shows the pending expression, for example `2 +`.
- Pressing a binary operator:
  - with no pending operator, stores the displayed value as the left operand;
  - after a new operand has been typed, sends the pending operation to the backend, shows the result and stores it as the left operand of the new operator;
  - immediately after another operator, replaces the pending operator.
- `√` applies immediately to the displayed value. The result replaces the displayed value and any pending binary operator is kept: `9 + √16 =` displays `13`.
- `=` sends the pending operation when a right operand has been typed. Pressing `=` with no pending operator, with no right operand typed, or repeatedly has no effect.
- After a result is shown, typing a digit or `.` starts a new operand.

### 4.3 Entry rules

- The typed operand can never exceed 32 digits; further digit keys are ignored.
- A second decimal point is ignored. Pressing `.` when starting an operand produces `0.`.
- A leading `0` is replaced by the next digit, so `0` then `7` displays `7`.
- `±` toggles the sign of the displayed value; it has no effect on `0`.
- `⌫` removes the last character of an operand being typed; removing the last digit leaves `0`. It has no effect on a result.
- `C` resets the calculator to `0` with no pending operation.
- Before sending, an operand ending in `.` is sent without the point (`5.` → `5`), and `-0` is sent as `0`.

### 4.4 Errors and in-flight requests

- While a request is in flight, all keys are ignored.
- A business error (HTTP 422) shows a message in the display. After that, a digit or `.` starts a fresh calculation, `C` resets, and every other key is ignored.

| Error code | Message |
|---|---|
| `division_by_zero` | Cannot divide by zero |
| `negative_square_root` | Invalid input for square root |
| `exponent_not_integer` | Exponent must be an integer |
| `exponent_out_of_range` | Exponent must be between -1000 and 1000 |
| `result_out_of_range` | Result is too large |
| `invalid_number`, `number_too_long` | Invalid number |

- A network failure, a timeout after 5 seconds, or any other non-422 response shows `Service unavailable` outside the display and keeps the calculator state unchanged, so pressing the same key again retries the operation.

### 4.5 Layout

- Responsive key grid, usable from a 320 px wide viewport.
- The display shrinks its font size so a 32-digit value with sign and decimal point fits without overflow.
