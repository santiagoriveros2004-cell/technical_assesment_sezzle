import type {
  BinaryOperation,
  BusinessErrorCode,
  CalculationOutcome,
  OperationRequest,
} from "../../api/calculatorClient";

export type Digit = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9";

export type CalculatorInput =
  | { type: "digit"; digit: Digit }
  | { type: "decimalPoint" }
  | { type: "toggleSign" }
  | { type: "backspace" }
  | { type: "operator"; operator: BinaryOperation }
  | { type: "squareRoot" }
  | { type: "equals" }
  | { type: "clear" };

export type CalculatorAction = CalculatorInput | { type: "requestSettled"; outcome: CalculationOutcome };

interface PendingOperation {
  left: string;
  operator: BinaryOperation;
}

type IdleState =
  | { phase: "entering"; operand: string; pending: PendingOperation | null }
  | { phase: "awaitingOperand"; pending: PendingOperation }
  | { phase: "result"; value: string; pending: PendingOperation | null }
  | { phase: "error"; code: BusinessErrorCode };

type Continuation =
  | { kind: "chain"; operator: BinaryOperation }
  | { kind: "equals" }
  | { kind: "squareRoot"; pending: PendingOperation | null };

interface CalculatingState {
  phase: "calculating";
  request: OperationRequest;
  continuation: Continuation;
  previous: IdleState;
}

export type CalculatorState = IdleState | CalculatingState;

export type CalculatorDisplay =
  | { kind: "value"; value: string; expression: string }
  | { kind: "error"; code: BusinessErrorCode };

export const MAX_OPERAND_DIGITS = 32;

const OPERATOR_SYMBOLS: Record<BinaryOperation, string> = {
  add: "+",
  subtract: "−",
  multiply: "×",
  divide: "÷",
  power: "^",
  percentage: "%",
};

export const initialCalculatorState: CalculatorState = {
  phase: "entering",
  operand: "0",
  pending: null,
};

export function calculatorReducer(state: CalculatorState, action: CalculatorAction): CalculatorState {
  if (action.type === "requestSettled") {
    return state.phase === "calculating" ? settle(state, action.outcome) : state;
  }

  switch (state.phase) {
    case "calculating":
      return state;
    case "entering":
      return handleEntering(state, action);
    case "awaitingOperand":
      return handleAwaitingOperand(state, action);
    case "result":
      return handleResult(state, action);
    case "error":
      return handleError(state, action);
  }
}

export function selectDisplay(state: CalculatorState): CalculatorDisplay {
  const shown = state.phase === "calculating" ? state.previous : state;

  switch (shown.phase) {
    case "entering":
      return { kind: "value", value: shown.operand, expression: formatExpression(shown.pending) };
    case "awaitingOperand":
      return { kind: "value", value: shown.pending.left, expression: formatExpression(shown.pending) };
    case "result":
      return { kind: "value", value: shown.value, expression: formatExpression(shown.pending) };
    case "error":
      return { kind: "error", code: shown.code };
  }
}

function toRequestOperand(operand: string): string {
  const withoutTrailingPoint = operand.endsWith(".") ? operand.slice(0, -1) : operand;
  return withoutTrailingPoint === "-0" ? "0" : withoutTrailingPoint;
}

type EnteringState = Extract<IdleState, { phase: "entering" }>;
type AwaitingOperandState = Extract<IdleState, { phase: "awaitingOperand" }>;
type ResultState = Extract<IdleState, { phase: "result" }>;
type ErrorState = Extract<IdleState, { phase: "error" }>;

function handleEntering(state: EnteringState, input: CalculatorInput): CalculatorState {
  switch (input.type) {
    case "digit":
      return withOperand(state, appendDigit(state.operand, input.digit));
    case "decimalPoint":
      return withOperand(state, appendDecimalPoint(state.operand));
    case "toggleSign":
      return withOperand(state, toggleSign(state.operand));
    case "backspace":
      return withOperand(state, removeLastCharacter(state.operand));
    case "operator":
      return state.pending
        ? calculatePending(state, state.pending, state.operand, { kind: "chain", operator: input.operator })
        : { phase: "awaitingOperand", pending: { left: toRequestOperand(state.operand), operator: input.operator } };
    case "squareRoot":
      return calculateSquareRoot(state, state.operand, state.pending);
    case "equals":
      return state.pending ? calculatePending(state, state.pending, state.operand, { kind: "equals" }) : state;
    case "clear":
      return initialCalculatorState;
  }
}

function handleAwaitingOperand(state: AwaitingOperandState, input: CalculatorInput): CalculatorState {
  switch (input.type) {
    case "digit":
    case "decimalPoint":
      return startOperand(input, state.pending);
    case "operator":
      return { ...state, pending: { ...state.pending, operator: input.operator } };
    case "squareRoot":
      return calculateSquareRoot(state, state.pending.left, state.pending);
    case "toggleSign":
    case "backspace":
    case "equals":
      return state;
    case "clear":
      return initialCalculatorState;
  }
}

function handleResult(state: ResultState, input: CalculatorInput): CalculatorState {
  switch (input.type) {
    case "digit":
    case "decimalPoint":
      return startOperand(input, state.pending);
    case "toggleSign":
      return withValue(state, toggleSign(state.value));
    case "operator":
      return state.pending
        ? calculatePending(state, state.pending, state.value, { kind: "chain", operator: input.operator })
        : { phase: "awaitingOperand", pending: { left: state.value, operator: input.operator } };
    case "squareRoot":
      return calculateSquareRoot(state, state.value, state.pending);
    case "equals":
      return state.pending ? calculatePending(state, state.pending, state.value, { kind: "equals" }) : state;
    case "backspace":
      return state;
    case "clear":
      return initialCalculatorState;
  }
}

function handleError(state: ErrorState, input: CalculatorInput): CalculatorState {
  switch (input.type) {
    case "digit":
    case "decimalPoint":
      return startOperand(input, null);
    case "clear":
      return initialCalculatorState;
    default:
      return state;
  }
}

function settle(state: CalculatingState, outcome: CalculationOutcome): CalculatorState {
  switch (outcome.kind) {
    case "success":
      return applyResult(outcome.result, state.continuation);
    case "rejected":
      return { phase: "error", code: outcome.code };
    case "unavailable":
      return state.previous;
  }
}

function applyResult(result: string, continuation: Continuation): IdleState {
  switch (continuation.kind) {
    case "chain":
      return { phase: "awaitingOperand", pending: { left: result, operator: continuation.operator } };
    case "equals":
      return { phase: "result", value: result, pending: null };
    case "squareRoot":
      return { phase: "result", value: result, pending: continuation.pending };
  }
}

function calculatePending(
  previous: IdleState,
  pending: PendingOperation,
  rightOperand: string,
  continuation: Continuation,
): CalculatingState {
  return {
    phase: "calculating",
    request: { operation: pending.operator, a: pending.left, b: toRequestOperand(rightOperand) },
    continuation,
    previous,
  };
}

function calculateSquareRoot(
  previous: IdleState,
  operand: string,
  pending: PendingOperation | null,
): CalculatingState {
  return {
    phase: "calculating",
    request: { operation: "sqrt", a: toRequestOperand(operand) },
    continuation: { kind: "squareRoot", pending },
    previous,
  };
}

function startOperand(
  input: Extract<CalculatorInput, { type: "digit" | "decimalPoint" }>,
  pending: PendingOperation | null,
): EnteringState {
  return { phase: "entering", operand: input.type === "digit" ? input.digit : "0.", pending };
}

function withOperand(state: EnteringState, operand: string): EnteringState {
  return operand === state.operand ? state : { ...state, operand };
}

function withValue(state: ResultState, value: string): ResultState {
  return value === state.value ? state : { ...state, value };
}

function formatExpression(pending: PendingOperation | null): string {
  return pending ? `${pending.left} ${OPERATOR_SYMBOLS[pending.operator]}` : "";
}

function countDigits(operand: string): number {
  return operand.replace(/\D/g, "").length;
}

function appendDigit(operand: string, digit: Digit): string {
  if (operand === "0") {
    return digit;
  }
  if (operand === "-0") {
    return `-${digit}`;
  }
  return countDigits(operand) < MAX_OPERAND_DIGITS ? operand + digit : operand;
}

function appendDecimalPoint(operand: string): string {
  return operand.includes(".") ? operand : `${operand}.`;
}

function toggleSign(value: string): string {
  if (value === "0") {
    return value;
  }
  return value.startsWith("-") ? value.slice(1) : `-${value}`;
}

function removeLastCharacter(operand: string): string {
  const remaining = operand.slice(0, -1);
  return remaining === "" || remaining === "-" ? "0" : remaining;
}
