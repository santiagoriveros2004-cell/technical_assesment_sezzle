import { http, HttpResponse } from "msw";
import type { BusinessErrorCode, OperationRequest } from "../api/calculatorClient";

export const OPERATIONS_ENDPOINT = "/api/v1/operations/:operation";

type OperationName = OperationRequest["operation"];

export type StubbedOperation = {
  operation: OperationName;
  operands: readonly string[];
} & ({ result: string } | { error: BusinessErrorCode });

export interface RecordedRequest {
  operation: string;
  body: unknown;
  contentType: string | null;
}

export function problem(status: number, code: string) {
  return HttpResponse.json(
    { type: "about:blank", title: "Error", status, detail: code.replaceAll("_", " "), code },
    { status, headers: { "Content-Type": "application/problem+json" } },
  );
}

export function createCalculatorService(stubs: readonly StubbedOperation[]) {
  const requests: RecordedRequest[] = [];

  const handler = http.post(OPERATIONS_ENDPOINT, async ({ request, params }) => {
    const operation = String(params.operation);
    const body: unknown = await request.json();
    requests.push({ operation, body, contentType: request.headers.get("Content-Type") });

    const operands = operandsOf(body);
    const stub = stubs.find(
      (candidate) =>
        candidate.operation === operation &&
        candidate.operands.length === operands.length &&
        candidate.operands.every((operand, index) => operand === operands[index]),
    );

    if (!stub) {
      return problem(500, "internal_error");
    }
    if ("error" in stub) {
      return problem(422, stub.error);
    }
    return HttpResponse.json({ operation, operands, result: stub.result });
  });

  return { handler, requests };
}

function operandsOf(body: unknown): string[] {
  if (typeof body !== "object" || body === null) {
    return [];
  }
  return ["a", "b"]
    .map((field) => (body as Record<string, unknown>)[field])
    .filter((operand): operand is string => typeof operand === "string");
}
