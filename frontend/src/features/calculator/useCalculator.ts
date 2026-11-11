import { useCallback, useEffect, useReducer, useState } from "react";
import { calculate } from "../../api/calculatorClient";
import {
  calculatorReducer,
  initialCalculatorState,
  selectDisplay,
  type CalculatorInput,
} from "./calculatorReducer";

export function useCalculator() {
  const [state, dispatch] = useReducer(calculatorReducer, initialCalculatorState);
  const [isServiceUnavailable, setServiceUnavailable] = useState(false);
  const request = state.phase === "calculating" ? state.request : null;

  useEffect(() => {
    if (!request) {
      return;
    }

    const controller = new AbortController();

    void calculate(request, controller.signal).then((outcome) => {
      if (controller.signal.aborted) {
        return;
      }
      setServiceUnavailable(outcome.kind === "unavailable");
      dispatch({ type: "requestSettled", outcome });
    });

    return () => {
      controller.abort();
    };
  }, [request]);

  const press = useCallback((input: CalculatorInput) => {
    setServiceUnavailable(false);
    dispatch(input);
  }, []);

  return {
    display: selectDisplay(state),
    isBusy: request !== null,
    isServiceUnavailable,
    press,
  };
}
