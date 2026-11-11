import { delay, http, HttpResponse } from "msw";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createCalculatorService, OPERATIONS_ENDPOINT, problem } from "../test/calculatorService";
import { server } from "../test/server";
import { BUSINESS_ERROR_CODES, calculate, REQUEST_TIMEOUT_MS } from "./calculatorClient";

describe("calculate", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("posts both operands of a binary operation as JSON strings and returns the result", async () => {
    const service = createCalculatorService([
      { operation: "add", operands: ["0.1", "0.2"], result: "0.3" },
    ]);
    server.use(service.handler);

    const outcome = await calculate({ operation: "add", a: "0.1", b: "0.2" });

    expect(outcome).toEqual({ kind: "success", result: "0.3" });
    expect(service.requests).toEqual([
      { operation: "add", body: { a: "0.1", b: "0.2" }, contentType: "application/json" },
    ]);
  });

  it("posts only the a operand for a square root", async () => {
    const service = createCalculatorService([
      { operation: "sqrt", operands: ["2"], result: "1.4142135624" },
    ]);
    server.use(service.handler);

    const outcome = await calculate({ operation: "sqrt", a: "2" });

    expect(outcome).toEqual({ kind: "success", result: "1.4142135624" });
    expect(service.requests[0]?.body).toEqual({ a: "2" });
  });

  it.each(BUSINESS_ERROR_CODES)("maps a 422 problem with code %s to a rejection", async (code) => {
    server.use(http.post(OPERATIONS_ENDPOINT, () => problem(422, code)));

    const outcome = await calculate({ operation: "divide", a: "10", b: "0" });

    expect(outcome).toEqual({ kind: "rejected", code });
  });

  it("treats a 422 response with a code outside the business catalog as unavailable", async () => {
    server.use(http.post(OPERATIONS_ENDPOINT, () => problem(422, "something_new")));

    const outcome = await calculate({ operation: "divide", a: "10", b: "0" });

    expect(outcome).toEqual({ kind: "unavailable" });
  });

  it.each([
    [400, "invalid_request"],
    [404, "unknown_operation"],
    [413, "payload_too_large"],
    [415, "unsupported_media_type"],
    [500, "internal_error"],
  ])("treats a %i response as unavailable", async (status, code) => {
    server.use(http.post(OPERATIONS_ENDPOINT, () => problem(status, code)));

    const outcome = await calculate({ operation: "add", a: "1", b: "2" });

    expect(outcome).toEqual({ kind: "unavailable" });
  });

  it("treats a gateway error without a problem body as unavailable", async () => {
    server.use(
      http.post(OPERATIONS_ENDPOINT, () => new HttpResponse("Bad Gateway", { status: 502 })),
    );

    const outcome = await calculate({ operation: "add", a: "1", b: "2" });

    expect(outcome).toEqual({ kind: "unavailable" });
  });

  it("treats a network error as unavailable", async () => {
    server.use(http.post(OPERATIONS_ENDPOINT, () => HttpResponse.error()));

    const outcome = await calculate({ operation: "add", a: "1", b: "2" });

    expect(outcome).toEqual({ kind: "unavailable" });
  });

  it.each([
    ["a body that is not JSON", () => new HttpResponse("<html>", { status: 200 })],
    ["a body without a result", () => HttpResponse.json({ operation: "add" })],
    ["a numeric result", () => HttpResponse.json({ result: 3 })],
    ["a result that is not a decimal", () => HttpResponse.json({ result: "1e5" })],
    ["a null body", () => HttpResponse.json(null)],
  ])("treats a successful response with %s as unavailable", async (_description, resolver) => {
    server.use(http.post(OPERATIONS_ENDPOINT, resolver));

    const outcome = await calculate({ operation: "add", a: "1", b: "2" });

    expect(outcome).toEqual({ kind: "unavailable" });
  });

  it("treats a 422 response with a malformed body as unavailable", async () => {
    server.use(
      http.post(OPERATIONS_ENDPOINT, () => new HttpResponse("not json", { status: 422 })),
    );

    const outcome = await calculate({ operation: "add", a: "1", b: "2" });

    expect(outcome).toEqual({ kind: "unavailable" });
  });

  it("gives up after the timeout and reports the service as unavailable", async () => {
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
    server.use(
      http.post(OPERATIONS_ENDPOINT, async () => {
        await delay("infinite");
        return HttpResponse.json({ result: "3" });
      }),
    );
    const settled = vi.fn();

    const outcome = calculate({ operation: "add", a: "1", b: "2" }).then((value) => {
      settled(value);
      return value;
    });

    await vi.advanceTimersByTimeAsync(REQUEST_TIMEOUT_MS - 1);
    expect(settled).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(1);
    await expect(outcome).resolves.toEqual({ kind: "unavailable" });
  });

  it("stops waiting when the caller aborts", async () => {
    server.use(
      http.post(OPERATIONS_ENDPOINT, async () => {
        await delay("infinite");
        return HttpResponse.json({ result: "3" });
      }),
    );
    const controller = new AbortController();

    const outcome = calculate({ operation: "add", a: "1", b: "2" }, controller.signal);
    controller.abort();

    await expect(outcome).resolves.toEqual({ kind: "unavailable" });
  });
});
