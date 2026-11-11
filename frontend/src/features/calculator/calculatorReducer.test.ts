import { describe, expect, it } from "vitest";
import type { BusinessErrorCode, CalculationOutcome } from "../../api/calculatorClient";
import {
  calculatorReducer,
  initialCalculatorState,
  MAX_OPERAND_DIGITS,
  selectDisplay,
  type CalculatorInput,
  type CalculatorState,
} from "./calculatorReducer";

const KEY_INPUTS: Record<string, CalculatorInput> = {
  ".": { type: "decimalPoint" },
  "±": { type: "toggleSign" },
  "⌫": { type: "backspace" },
  "+": { type: "operator", operator: "add" },
  "−": { type: "operator", operator: "subtract" },
  "×": { type: "operator", operator: "multiply" },
  "÷": { type: "operator", operator: "divide" },
  "^": { type: "operator", operator: "power" },
  "%": { type: "operator", operator: "percentage" },
  "√": { type: "squareRoot" },
  "=": { type: "equals" },
  C: { type: "clear" },
};

function inputFor(key: string): CalculatorInput {
  if (/^\d$/.test(key)) {
    return { type: "digit", digit: key as Extract<CalculatorInput, { type: "digit" }>["digit"] };
  }
  const input = KEY_INPUTS[key];
  if (!input) {
    throw new Error(`Unknown key ${key}`);
  }
  return input;
}

function press(keys: string, state: CalculatorState = initialCalculatorState): CalculatorState {
  return Array.from(keys.replaceAll(" ", "")).reduce(
    (current, key) => calculatorReducer(current, inputFor(key)),
    state,
  );
}

function settle(state: CalculatorState, outcome: CalculationOutcome): CalculatorState {
  return calculatorReducer(state, { type: "requestSettled", outcome });
}

function succeed(state: CalculatorState, result: string): CalculatorState {
  return settle(state, { kind: "success", result });
}

function reject(state: CalculatorState, code: BusinessErrorCode): CalculatorState {
  return settle(state, { kind: "rejected", code });
}

function requestOf(state: CalculatorState) {
  return state.phase === "calculating" ? state.request : null;
}

function valueOf(state: CalculatorState) {
  const display = selectDisplay(state);
  return display.kind === "value" ? display.value : null;
}

function expressionOf(state: CalculatorState) {
  const display = selectDisplay(state);
  return display.kind === "value" ? display.expression : null;
}

describe("calculatorReducer", () => {
  describe("initial state", () => {
    it("shows 0 with no pending expression", () => {
      expect(selectDisplay(initialCalculatorState)).toEqual({ kind: "value", value: "0", expression: "" });
    });
  });

  describe("digit entry", () => {
    it("appends digits to the operand", () => {
      expect(valueOf(press("123"))).toBe("123");
    });

    it("replaces a leading zero with the next digit", () => {
      expect(valueOf(press("07"))).toBe("7");
      expect(valueOf(press("00"))).toBe("0");
    });

    it("keeps a zero after the decimal point", () => {
      expect(valueOf(press("0.05"))).toBe("0.05");
    });

    it(`ignores digits beyond ${String(MAX_OPERAND_DIGITS)}`, () => {
      const maximum = "9".repeat(MAX_OPERAND_DIGITS);

      expect(valueOf(press(`${maximum}1`))).toBe(maximum);
    });

    it("counts fractional digits toward the limit", () => {
      const integerPart = "1".repeat(20);
      const fractionalPart = "2".repeat(MAX_OPERAND_DIGITS - 20);

      expect(valueOf(press(`${integerPart}.${fractionalPart}3`))).toBe(`${integerPart}.${fractionalPart}`);
    });

    it("does not count the sign or the decimal point toward the limit", () => {
      const remainingDigits = "3".repeat(MAX_OPERAND_DIGITS - 2);

      expect(valueOf(press(`5 ± 1.${remainingDigits}4`))).toBe(`-51.${remainingDigits}`);
    });

    it("still accepts the decimal point when the digit limit is reached", () => {
      const maximum = "9".repeat(MAX_OPERAND_DIGITS);

      expect(valueOf(press(`${maximum}.`))).toBe(`${maximum}.`);
    });

    it("replaces the zero of a negative zero with the next digit", () => {
      const negativeZero = press("0.5 ± ⌫ ⌫");

      expect(valueOf(negativeZero)).toBe("-0");
      expect(valueOf(press("7", negativeZero))).toBe("-7");
    });
  });

  describe("decimal point", () => {
    it("produces 0. when starting an operand", () => {
      expect(valueOf(press("."))).toBe("0.");
    });

    it("ignores a second decimal point", () => {
      expect(valueOf(press("1.2.3"))).toBe("1.23");
    });

    it("produces 0. when starting the right operand", () => {
      const state = press("2 + .");

      expect(valueOf(state)).toBe("0.");
      expect(expressionOf(state)).toBe("2 +");
    });
  });

  describe("sign toggle", () => {
    it("toggles the sign of the operand being typed", () => {
      expect(valueOf(press("5 ±"))).toBe("-5");
      expect(valueOf(press("5 ± ±"))).toBe("5");
    });

    it("has no effect on 0", () => {
      expect(press("±")).toBe(initialCalculatorState);
    });

    it("keeps typing after the sign is toggled", () => {
      expect(valueOf(press("5 ± 2"))).toBe("-52");
    });

    it("toggles the sign of a displayed result", () => {
      const result = succeed(press("2 + 3 ="), "5");

      expect(valueOf(press("±", result))).toBe("-5");
    });

    it("has no effect on a result of 0", () => {
      const result = succeed(press("2 − 2 ="), "0");

      expect(press("±", result)).toBe(result);
    });

    it("has no effect while waiting for the right operand", () => {
      const awaiting = press("2 +");

      expect(press("±", awaiting)).toBe(awaiting);
    });
  });

  describe("backspace", () => {
    it("removes the last character of the operand", () => {
      expect(valueOf(press("123 ⌫"))).toBe("12");
      expect(valueOf(press("1.5 ⌫"))).toBe("1.");
      expect(valueOf(press("1. ⌫"))).toBe("1");
    });

    it("leaves 0 after removing the last digit", () => {
      expect(valueOf(press("5 ⌫"))).toBe("0");
      expect(valueOf(press("5 ± ⌫"))).toBe("0");
      expect(valueOf(press("⌫"))).toBe("0");
    });

    it("has no effect on a result", () => {
      const result = succeed(press("2 + 3 ="), "5");

      expect(press("⌫", result)).toBe(result);
    });

    it("has no effect while waiting for the right operand", () => {
      const awaiting = press("2 +");

      expect(press("⌫", awaiting)).toBe(awaiting);
    });
  });

  describe("binary operators", () => {
    it("stores the displayed value as the left operand when nothing is pending", () => {
      const state = press("2 +");

      expect(requestOf(state)).toBeNull();
      expect(selectDisplay(state)).toEqual({ kind: "value", value: "2", expression: "2 +" });
    });

    it("uses 0 as the left operand on a fresh calculator", () => {
      expect(expressionOf(press("×"))).toBe("0 ×");
    });

    it("replaces the pending operator when pressed right after another operator", () => {
      const state = press("2 + × ÷");

      expect(requestOf(state)).toBeNull();
      expect(expressionOf(state)).toBe("2 ÷");
    });

    it.each([
      ["+", "add", "+"],
      ["−", "subtract", "−"],
      ["×", "multiply", "×"],
      ["÷", "divide", "÷"],
      ["^", "power", "^"],
      ["%", "percentage", "%"],
    ])("sends %s as the %s operation", (key, operation, symbol) => {
      expect(expressionOf(press(`8 ${key}`))).toBe(`8 ${symbol}`);
      expect(requestOf(press(`8 ${key} 2 =`))).toEqual({ operation, a: "8", b: "2" });
    });

    it("sends the pending operation when an operator follows a typed operand", () => {
      const calculating = press("2 + 3 ×");

      expect(requestOf(calculating)).toEqual({ operation: "add", a: "2", b: "3" });

      const chained = succeed(calculating, "5");

      expect(selectDisplay(chained)).toEqual({ kind: "value", value: "5", expression: "5 ×" });
    });

    it("evaluates left to right: 2 + 3 × 4 = displays 20", () => {
      const afterAdd = succeed(press("2 + 3 ×"), "5");
      const multiplying = press("4 =", afterAdd);

      expect(requestOf(multiplying)).toEqual({ operation: "multiply", a: "5", b: "4" });

      const result = succeed(multiplying, "20");

      expect(selectDisplay(result)).toEqual({ kind: "value", value: "20", expression: "" });
    });

    it("uses a result as the left operand of the next operator", () => {
      const result = succeed(press("2 + 3 ="), "5");
      const state = press("−", result);

      expect(requestOf(state)).toBeNull();
      expect(expressionOf(state)).toBe("5 −");
    });
  });

  describe("equals", () => {
    it("sends the pending operation when a right operand has been typed", () => {
      expect(requestOf(press("7 + 1 ="))).toEqual({ operation: "add", a: "7", b: "1" });
    });

    it("shows the result with no pending expression", () => {
      const result = succeed(press("7 + 1 ="), "8");

      expect(selectDisplay(result)).toEqual({ kind: "value", value: "8", expression: "" });
    });

    it("has no effect without a pending operator", () => {
      const typed = press("42");

      expect(press("=", typed)).toBe(typed);
      expect(press("=")).toBe(initialCalculatorState);
    });

    it("has no effect when no right operand has been typed", () => {
      const awaiting = press("7 +");

      expect(press("=", awaiting)).toBe(awaiting);
    });

    it("has no effect when pressed repeatedly after a result", () => {
      const result = succeed(press("7 + 1 ="), "8");

      expect(press("= =", result)).toBe(result);
    });
  });

  describe("after a result", () => {
    it("starts a new operand when a digit is typed", () => {
      const result = succeed(press("2 + 3 ="), "5");

      expect(selectDisplay(press("7", result))).toEqual({ kind: "value", value: "7", expression: "" });
    });

    it("starts a new operand when the decimal point is typed", () => {
      const result = succeed(press("2 + 3 ="), "5");

      expect(valueOf(press(".", result))).toBe("0.");
    });
  });

  describe("square root", () => {
    it("applies immediately to the displayed value", () => {
      const calculating = press("16 √");

      expect(requestOf(calculating)).toEqual({ operation: "sqrt", a: "16" });
      expect(selectDisplay(succeed(calculating, "4"))).toEqual({ kind: "value", value: "4", expression: "" });
    });

    it("keeps the pending operator: 9 + √16 = displays 13", () => {
      const rooted = succeed(press("9 + 16 √"), "4");

      expect(selectDisplay(rooted)).toEqual({ kind: "value", value: "4", expression: "9 +" });

      const adding = press("=", rooted);

      expect(requestOf(adding)).toEqual({ operation: "add", a: "9", b: "4" });
      expect(valueOf(succeed(adding, "13"))).toBe("13");
    });

    it("uses the root as the right operand when an operator follows", () => {
      const rooted = succeed(press("9 + 16 √"), "4");
      const chaining = press("×", rooted);

      expect(requestOf(chaining)).toEqual({ operation: "add", a: "9", b: "4" });
      expect(expressionOf(succeed(chaining, "13"))).toBe("13 ×");
    });

    it("lets a newly typed operand replace the root while keeping the pending operator", () => {
      const rooted = succeed(press("9 + 16 √"), "4");

      expect(requestOf(press("5 =", rooted))).toEqual({ operation: "add", a: "9", b: "5" });
    });

    it("applies to the displayed left operand while waiting for the right operand", () => {
      const calculating = press("9 + √");

      expect(requestOf(calculating)).toEqual({ operation: "sqrt", a: "9" });

      const rooted = succeed(calculating, "3");

      expect(selectDisplay(rooted)).toEqual({ kind: "value", value: "3", expression: "9 +" });
      expect(requestOf(press("=", rooted))).toEqual({ operation: "add", a: "9", b: "3" });
    });

    it("applies to a displayed result", () => {
      const result = succeed(press("2 + 2 ="), "4");

      expect(requestOf(press("√", result))).toEqual({ operation: "sqrt", a: "4" });
    });
  });

  describe("operand normalization before sending", () => {
    it("sends an operand ending in a point without the point", () => {
      expect(requestOf(press("5. + 3 ="))).toEqual({ operation: "add", a: "5", b: "3" });
      expect(requestOf(press("2 + 5. ="))).toEqual({ operation: "add", a: "2", b: "5" });
      expect(requestOf(press("5. √"))).toEqual({ operation: "sqrt", a: "5" });
    });

    it("sends -0 as 0", () => {
      expect(requestOf(press("2 × 0.5 ± ⌫ ⌫ ="))).toEqual({ operation: "multiply", a: "2", b: "0" });
      expect(expressionOf(press("0.5 ± ⌫ ⌫ +"))).toBe("0 +");
    });

    it("sends -0. as 0", () => {
      expect(requestOf(press("0. ± √"))).toEqual({ operation: "sqrt", a: "0" });
    });
  });

  describe("business errors", () => {
    it("shows the error code returned by the service", () => {
      const failed = reject(press("10 ÷ 0 ="), "division_by_zero");

      expect(selectDisplay(failed)).toEqual({ kind: "error", code: "division_by_zero" });
    });

    it("starts a fresh calculation when a digit is typed", () => {
      const failed = reject(press("10 ÷ 0 ="), "division_by_zero");

      expect(selectDisplay(press("7", failed))).toEqual({ kind: "value", value: "7", expression: "" });
    });

    it("starts a fresh calculation when the decimal point is typed", () => {
      const failed = reject(press("10 ÷ 0 ="), "division_by_zero");

      expect(selectDisplay(press(".", failed))).toEqual({ kind: "value", value: "0.", expression: "" });
    });

    it.each(["+", "−", "×", "÷", "^", "%", "√", "=", "±", "⌫"])("ignores %s", (key) => {
      const failed = reject(press("16 ± √"), "negative_square_root");

      expect(press(key, failed)).toBe(failed);
    });

    it("resets on clear", () => {
      const failed = reject(press("10 ÷ 0 ="), "division_by_zero");

      expect(press("C", failed)).toBe(initialCalculatorState);
    });
  });

  describe("in-flight requests", () => {
    it.each(["1", ".", "±", "⌫", "+", "√", "=", "C"])("ignores %s while a request is in flight", (key) => {
      const calculating = press("2 + 3 =");

      expect(press(key, calculating)).toBe(calculating);
    });

    it("keeps showing the state from before the request", () => {
      expect(selectDisplay(press("2 + 3 ="))).toEqual({ kind: "value", value: "3", expression: "2 +" });
    });

    it("ignores an outcome when no request is in flight", () => {
      const typed = press("12");

      expect(succeed(typed, "99")).toBe(typed);
    });
  });

  describe("service unavailable", () => {
    it("restores the exact state from before the request", () => {
      const beforeRequest = press("7 + 1");
      const restored = settle(press("=", beforeRequest), { kind: "unavailable" });

      expect(restored).toBe(beforeRequest);
    });

    it("retries the same operation when the same key is pressed again", () => {
      const restored = settle(press("7 + 1 ="), { kind: "unavailable" });

      expect(requestOf(press("=", restored))).toEqual({ operation: "add", a: "7", b: "1" });
    });

    it("restores a chained operator so it can be retried", () => {
      const restored = settle(press("2 + 3 ×"), { kind: "unavailable" });

      expect(selectDisplay(restored)).toEqual({ kind: "value", value: "3", expression: "2 +" });
      expect(requestOf(press("×", restored))).toEqual({ operation: "add", a: "2", b: "3" });
    });
  });

  describe("clear", () => {
    it.each([
      ["while typing", "123"],
      ["while typing the right operand", "2 + 3"],
      ["while waiting for the right operand", "2 +"],
    ])("resets to 0 with no pending operation %s", (_description, keys) => {
      expect(press("C", press(keys))).toBe(initialCalculatorState);
    });

    it("resets after a result", () => {
      const result = succeed(press("2 + 3 ="), "5");

      expect(press("C", result)).toBe(initialCalculatorState);
    });
  });
});
