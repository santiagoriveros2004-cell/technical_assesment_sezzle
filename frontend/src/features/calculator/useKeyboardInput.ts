import { useEffect } from "react";
import { DIGITS, type CalculatorInput } from "./calculatorReducer";

const KEY_BINDINGS: ReadonlyMap<string, CalculatorInput> = new Map<string, CalculatorInput>([
  ...DIGITS.map((digit) => [digit, { type: "digit", digit }] as const),
  [".", { type: "decimalPoint" }],
  [",", { type: "decimalPoint" }],
  ["+", { type: "operator", operator: "add" }],
  ["-", { type: "operator", operator: "subtract" }],
  ["*", { type: "operator", operator: "multiply" }],
  ["/", { type: "operator", operator: "divide" }],
  ["^", { type: "operator", operator: "power" }],
  ["%", { type: "operator", operator: "percentage" }],
  ["Enter", { type: "equals" }],
  ["=", { type: "equals" }],
  ["Backspace", { type: "backspace" }],
  ["Escape", { type: "clear" }],
]);

export function useKeyboardInput(onInput: (input: CalculatorInput) => void) {
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.ctrlKey || event.metaKey || event.altKey) {
        return;
      }
      const input = KEY_BINDINGS.get(event.key);
      if (!input) {
        return;
      }
      event.preventDefault();
      onInput(input);
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [onInput]);
}
