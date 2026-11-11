import type { BinaryOperation } from "../../api/calculatorClient";
import type { CalculatorInput, Digit } from "./calculatorReducer";
import { Key, type KeySpan, type KeyTone } from "./Key";
import styles from "./Keypad.module.css";

interface KeyDefinition {
  label: string;
  name: string;
  tone: KeyTone;
  input: CalculatorInput;
  span?: KeySpan;
}

const digitKey = (digit: Digit, span?: KeySpan): KeyDefinition => ({
  label: digit,
  name: digit,
  tone: "digit",
  input: { type: "digit", digit },
  span,
});

const operatorKey = (label: string, name: string, operator: BinaryOperation): KeyDefinition => ({
  label,
  name,
  tone: "operator",
  input: { type: "operator", operator },
});

const KEYS: readonly KeyDefinition[] = [
  { label: "C", name: "Clear", tone: "function", input: { type: "clear" } },
  { label: "⌫", name: "Backspace", tone: "function", input: { type: "backspace" } },
  operatorKey("%", "Percentage", "percentage"),
  operatorKey("÷", "Divide", "divide"),
  { label: "√", name: "Square root", tone: "function", input: { type: "squareRoot" } },
  operatorKey("xʸ", "Power", "power"),
  { label: "±", name: "Toggle sign", tone: "function", input: { type: "toggleSign" } },
  operatorKey("×", "Multiply", "multiply"),
  digitKey("7"),
  digitKey("8"),
  digitKey("9"),
  operatorKey("−", "Subtract", "subtract"),
  digitKey("4"),
  digitKey("5"),
  digitKey("6"),
  operatorKey("+", "Add", "add"),
  digitKey("1"),
  digitKey("2"),
  digitKey("3"),
  { label: "=", name: "Equals", tone: "equals", input: { type: "equals" }, span: "tall" },
  digitKey("0", "wide"),
  { label: ".", name: "Decimal point", tone: "digit", input: { type: "decimalPoint" } },
];

interface KeypadProps {
  disabled: boolean;
  onPress: (input: CalculatorInput) => void;
}

export function Keypad({ disabled, onPress }: KeypadProps) {
  return (
    <div className={styles.keypad} role="group" aria-label="Keypad">
      {KEYS.map(({ input, ...key }) => (
        <Key
          key={key.name}
          {...key}
          disabled={disabled}
          onPress={() => {
            onPress(input);
          }}
        />
      ))}
    </div>
  );
}
