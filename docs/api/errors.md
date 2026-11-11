# Error catalog

Every error response uses `Content-Type: application/problem+json` and follows [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457):

```json
{
  "type": "about:blank",
  "title": "Unprocessable Content",
  "status": 422,
  "detail": "operand \"a\" exceeds 32 digits",
  "code": "number_too_long",
  "field": "a"
}
```

`code` is the stable contract for clients. `detail` is for humans and may change. `field` is present only when a specific operand caused the error.

| Status | Code | When |
|---|---|---|
| 400 | `invalid_request` | Malformed JSON, trailing data after the JSON object, operands sent as JSON numbers, a required operand missing, `b` sent to `sqrt`, or an unknown field |
| 404 | `not_found` | The path does not exist |
| 404 | `unknown_operation` | `/api/v1/operations/{operation}` with an unsupported operation |
| 405 | `method_not_allowed` | Any method other than `POST` on an operation endpoint; the `Allow: POST` header is set |
| 413 | `payload_too_large` | Request body larger than 1 KiB |
| 415 | `unsupported_media_type` | `Content-Type` is not `application/json` (parameters such as `charset=utf-8` are accepted) |
| 422 | `invalid_number` | An operand does not match `^-?\d+(\.\d+)?$` |
| 422 | `number_too_long` | An operand has more than 32 digits |
| 422 | `division_by_zero` | Division by zero, or zero raised to a negative exponent |
| 422 | `negative_square_root` | Square root of a negative number |
| 422 | `exponent_not_integer` | `power` exponent has a non-zero fractional part |
| 422 | `exponent_out_of_range` | `power` exponent outside `-1000..1000` |
| 422 | `result_out_of_range` | The result's integer part exceeds 32 digits |
| 500 | `internal_error` | Unexpected failure; details are logged, never returned |

## Validation order

1. Route and method (404, 405)
2. Content type (415)
3. Body size (413)
4. JSON shape (400)
5. Operand format and length, `a` before `b` (422)
6. Operation constraints and result range (422)
