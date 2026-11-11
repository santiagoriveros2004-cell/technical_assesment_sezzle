export type BinaryOperation = "add" | "subtract" | "multiply" | "divide" | "power" | "percentage";

export type OperationRequest =
  | { operation: BinaryOperation; a: string; b: string }
  | { operation: "sqrt"; a: string };

export const BUSINESS_ERROR_CODES = [
  "invalid_number",
  "number_too_long",
  "division_by_zero",
  "negative_square_root",
  "exponent_not_integer",
  "exponent_out_of_range",
  "result_out_of_range",
] as const;

export type BusinessErrorCode = (typeof BUSINESS_ERROR_CODES)[number];

export type CalculationOutcome =
  | { kind: "success"; result: string }
  | { kind: "rejected"; code: BusinessErrorCode }
  | { kind: "unavailable" };

export const REQUEST_TIMEOUT_MS = 5_000;

const DECIMAL_PATTERN = /^-?\d+(\.\d+)?$/;
const UNPROCESSABLE_CONTENT = 422;
const UNAVAILABLE: CalculationOutcome = { kind: "unavailable" };

export async function calculate(
  request: OperationRequest,
  signal?: AbortSignal,
): Promise<CalculationOutcome> {
  const controller = new AbortController();
  const abort = () => {
    controller.abort();
  };
  const timeoutId = setTimeout(abort, REQUEST_TIMEOUT_MS);
  signal?.addEventListener("abort", abort, { once: true });

  const { operation, ...operands } = request;

  try {
    const response = await fetch(`/api/v1/operations/${operation}`, {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify(operands),
      signal: controller.signal,
    });
    return await readOutcome(response);
  } catch {
    return UNAVAILABLE;
  } finally {
    clearTimeout(timeoutId);
    signal?.removeEventListener("abort", abort);
  }
}

async function readOutcome(response: Response): Promise<CalculationOutcome> {
  if (response.ok) {
    const body: unknown = await response.json();
    return isOperationResult(body) ? { kind: "success", result: body.result } : UNAVAILABLE;
  }

  if (response.status === UNPROCESSABLE_CONTENT) {
    const body: unknown = await response.json();
    return isBusinessProblem(body) ? { kind: "rejected", code: body.code } : UNAVAILABLE;
  }

  return UNAVAILABLE;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isOperationResult(body: unknown): body is { result: string } {
  return isRecord(body) && typeof body.result === "string" && DECIMAL_PATTERN.test(body.result);
}

function isBusinessProblem(body: unknown): body is { code: BusinessErrorCode } {
  return (
    isRecord(body) &&
    BUSINESS_ERROR_CODES.some((code) => code === body.code)
  );
}
