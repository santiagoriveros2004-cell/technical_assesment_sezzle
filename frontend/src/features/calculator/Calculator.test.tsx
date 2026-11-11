import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent, { type UserEvent } from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { BUSINESS_ERROR_CODES } from "../../api/calculatorClient";
import {
  createCalculatorService,
  OPERATIONS_ENDPOINT,
  problem,
  type StubbedOperation,
} from "../../test/calculatorService";
import { server } from "../../test/server";
import { Calculator } from "./Calculator";
import { DIGITS } from "./calculatorReducer";
import { ERROR_MESSAGES } from "./errorMessages";

const KEY_NAMES: Record<string, string> = {
  ".": "Decimal point",
  "±": "Toggle sign",
  "⌫": "Backspace",
  "+": "Add",
  "−": "Subtract",
  "×": "Multiply",
  "÷": "Divide",
  "^": "Power",
  "%": "Percentage",
  "√": "Square root",
  "=": "Equals",
  C: "Clear",
};

function renderCalculator(stubs: readonly StubbedOperation[] = []) {
  const service = createCalculatorService(stubs);
  server.use(service.handler);
  const user = userEvent.setup();
  render(<Calculator />);
  return { user, requests: service.requests };
}

async function clickKeys(user: UserEvent, keys: string) {
  for (const key of keys.replaceAll(" ", "")) {
    await user.click(screen.getByRole("button", { name: KEY_NAMES[key] ?? key }));
  }
}

function displayedValue() {
  return screen.getByRole("status");
}

function deferred() {
  let resolveResponse: (() => void) | undefined;
  const released = new Promise<void>((resolve) => {
    resolveResponse = resolve;
  });
  return {
    released,
    release: () => {
      resolveResponse?.();
    },
  };
}

describe("Calculator", () => {
  it("renders every key as an accessible button", () => {
    renderCalculator();

    const names = [
      ...DIGITS,
      "Decimal point",
      "Toggle sign",
      "Add",
      "Subtract",
      "Multiply",
      "Divide",
      "Power",
      "Percentage",
      "Square root",
      "Equals",
      "Backspace",
      "Clear",
    ];

    for (const name of names) {
      expect(screen.getByRole("button", { name })).toBeEnabled();
    }
    expect(displayedValue()).toHaveTextContent("0");
  });

  it("evaluates a chained calculation left to right with clicks", async () => {
    const { user, requests } = renderCalculator([
      { operation: "add", operands: ["2", "3"], result: "5" },
      { operation: "multiply", operands: ["5", "4"], result: "20" },
    ]);

    await clickKeys(user, "2 + 3 ×");

    expect(await screen.findByText("5 ×")).toBeInTheDocument();
    expect(displayedValue()).toHaveTextContent(/^5$/);

    await clickKeys(user, "4 =");

    await waitFor(() => {
      expect(displayedValue()).toHaveTextContent(/^20$/);
    });
    expect(requests.map(({ operation, body }) => ({ operation, body }))).toEqual([
      { operation: "add", body: { a: "2", b: "3" } },
      { operation: "multiply", body: { a: "5", b: "4" } },
    ]);
  });

  it("shows the pending expression while the right operand is typed", async () => {
    const { user } = renderCalculator();

    await clickKeys(user, "1 2 ÷ 3 . 5");

    expect(screen.getByText("12 ÷")).toBeInTheDocument();
    expect(displayedValue()).toHaveTextContent(/^3\.5$/);
  });

  it("keeps a pending operator across a square root: 9 + √16 = displays 13", async () => {
    const { user, requests } = renderCalculator([
      { operation: "sqrt", operands: ["16"], result: "4" },
      { operation: "add", operands: ["9", "4"], result: "13" },
    ]);

    await clickKeys(user, "9 + 1 6 √");
    await waitFor(() => {
      expect(displayedValue()).toHaveTextContent(/^4$/);
    });
    expect(screen.getByText("9 +")).toBeInTheDocument();

    await clickKeys(user, "=");

    await waitFor(() => {
      expect(displayedValue()).toHaveTextContent(/^13$/);
    });
    expect(requests.map(({ body }) => body)).toEqual([{ a: "16" }, { a: "9", b: "4" }]);
  });

  it("sends JSON with string operands", async () => {
    const { user, requests } = renderCalculator([
      { operation: "percentage", operands: ["20", "150"], result: "30" },
    ]);

    await clickKeys(user, "2 0 % 1 5 0 =");

    await waitFor(() => {
      expect(displayedValue()).toHaveTextContent(/^30$/);
    });
    expect(requests).toEqual([
      { operation: "percentage", body: { a: "20", b: "150" }, contentType: "application/json" },
    ]);
  });

  it("drives the calculator from the physical keyboard", async () => {
    const { user, requests } = renderCalculator([
      { operation: "multiply", operands: ["12.5", "2"], result: "25" },
      { operation: "power", operands: ["25", "2"], result: "625" },
    ]);

    await user.keyboard("12,5*2{Enter}");

    await waitFor(() => {
      expect(displayedValue()).toHaveTextContent(/^25$/);
    });

    await user.keyboard("^2=");

    await waitFor(() => {
      expect(displayedValue()).toHaveTextContent(/^625$/);
    });
    expect(requests.map(({ operation }) => operation)).toEqual(["multiply", "power"]);
  });

  it.each([
    ["+", "add", "+"],
    ["-", "subtract", "−"],
    ["/", "divide", "÷"],
    ["%", "percentage", "%"],
  ])("maps the %s key to the %s operator", async (key, _operation, symbol) => {
    const { user } = renderCalculator();

    await user.keyboard(`8${key}`);

    expect(screen.getByText(`8 ${symbol}`)).toBeInTheDocument();
  });

  it("edits and clears from the keyboard", async () => {
    const { user } = renderCalculator();

    await user.keyboard("123.{Backspace}{Backspace}");
    expect(displayedValue()).toHaveTextContent(/^12$/);

    await user.keyboard("+4{Escape}");
    expect(displayedValue()).toHaveTextContent(/^0$/);
    expect(screen.queryByText("12 +")).not.toBeInTheDocument();
  });

  it("ignores keys pressed together with a modifier", async () => {
    const { user } = renderCalculator();

    await user.keyboard("{Control>}5{/Control}{Meta>}6{/Meta}{Alt>}7{/Alt}");

    expect(displayedValue()).toHaveTextContent(/^0$/);
  });

  it("ignores keys without a binding and keys with no physical shortcut", async () => {
    const { user } = renderCalculator();

    await user.keyboard("5a{ArrowUp} x");

    expect(displayedValue()).toHaveTextContent(/^5$/);
    expect(screen.queryByText(/5 /)).not.toBeInTheDocument();
  });

  it("does not trigger a focused key when Enter is pressed", async () => {
    const { user } = renderCalculator();

    await clickKeys(user, "7");
    await user.keyboard("{Enter}");

    expect(displayedValue()).toHaveTextContent(/^7$/);
  });

  it.each(BUSINESS_ERROR_CODES)("shows the message for %s in the display", async (code) => {
    const { user } = renderCalculator([{ operation: "divide", operands: ["1", "2"], error: code }]);

    await clickKeys(user, "1 ÷ 2 =");

    await waitFor(() => {
      expect(displayedValue()).toHaveTextContent(ERROR_MESSAGES[code]);
    });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("recovers from a business error with a fresh calculation", async () => {
    const { user } = renderCalculator([
      { operation: "divide", operands: ["10", "0"], error: "division_by_zero" },
    ]);

    await clickKeys(user, "1 0 ÷ 0 =");
    await waitFor(() => {
      expect(displayedValue()).toHaveTextContent("Cannot divide by zero");
    });

    await clickKeys(user, "+ √ ±");
    expect(displayedValue()).toHaveTextContent("Cannot divide by zero");

    await clickKeys(user, "7");
    expect(displayedValue()).toHaveTextContent(/^7$/);
  });

  it("reports an unavailable service outside the display and retries the same operation", async () => {
    const { user, requests } = renderCalculator([{ operation: "add", operands: ["7", "1"], result: "8" }]);
    server.use(http.post(OPERATIONS_ENDPOINT, () => problem(500, "internal_error"), { once: true }));

    await clickKeys(user, "7 + 1 =");

    expect(await screen.findByRole("alert")).toHaveTextContent("Service unavailable");
    expect(displayedValue()).toHaveTextContent(/^1$/);
    expect(screen.getByText("7 +")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Equals" })).toBeEnabled();

    await clickKeys(user, "=");

    await waitFor(() => {
      expect(displayedValue()).toHaveTextContent(/^8$/);
    });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(requests.map(({ body }) => body)).toEqual([{ a: "7", b: "1" }]);
  });

  it("reports an unavailable service on a network error", async () => {
    const { user } = renderCalculator();
    server.use(http.post(OPERATIONS_ENDPOINT, () => HttpResponse.error()));

    await clickKeys(user, "4 √");

    expect(await screen.findByRole("alert")).toHaveTextContent("Service unavailable");
    expect(displayedValue()).toHaveTextContent(/^4$/);
  });

  it("disables the keys and ignores the keyboard while a request is in flight", async () => {
    const { user } = renderCalculator();
    const response = deferred();
    server.use(
      http.post(OPERATIONS_ENDPOINT, async () => {
        await response.released;
        return HttpResponse.json({ operation: "add", operands: ["2", "3"], result: "5" });
      }),
    );

    await clickKeys(user, "2 + 3 =");

    for (const button of screen.getAllByRole("button")) {
      expect(button).toBeDisabled();
    }

    await user.keyboard("9{Escape}");
    expect(displayedValue()).toHaveTextContent(/^3$/);
    expect(screen.getByText("2 +")).toBeInTheDocument();

    response.release();

    await waitFor(() => {
      expect(displayedValue()).toHaveTextContent(/^5$/);
    });
    for (const button of screen.getAllByRole("button")) {
      expect(button).toBeEnabled();
    }
  });

  it("aborts an in-flight request when unmounted", async () => {
    const { user } = renderCalculator();
    const response = deferred();
    let aborted = false;
    server.use(
      http.post(OPERATIONS_ENDPOINT, async ({ request }) => {
        request.signal.addEventListener("abort", () => {
          aborted = true;
        });
        await response.released;
        return HttpResponse.json({ operation: "sqrt", operands: ["9"], result: "3" });
      }),
    );

    await clickKeys(user, "9 √");
    cleanup();

    await waitFor(() => {
      expect(aborted).toBe(true);
    });
    response.release();
  });
});
