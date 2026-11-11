import styles from "./Calculator.module.css";
import { Display } from "./Display";
import { ERROR_MESSAGES, SERVICE_UNAVAILABLE_MESSAGE } from "./errorMessages";
import { Keypad } from "./Keypad";
import { useCalculator } from "./useCalculator";
import { useKeyboardInput } from "./useKeyboardInput";

export function Calculator() {
  const { display, isBusy, isServiceUnavailable, press } = useCalculator();
  useKeyboardInput(press);

  return (
    <div className={styles.calculator} aria-busy={isBusy}>
      {display.kind === "error" ? (
        <Display value={ERROR_MESSAGES[display.code]} expression="" />
      ) : (
        <Display value={display.value} expression={display.expression} />
      )}
      {isServiceUnavailable && (
        <p className={styles.alert} role="alert">
          {SERVICE_UNAVAILABLE_MESSAGE}
        </p>
      )}
      <Keypad disabled={isBusy} onPress={press} />
    </div>
  );
}
