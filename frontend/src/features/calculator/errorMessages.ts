import type { BusinessErrorCode } from "../../api/calculatorClient";

export const ERROR_MESSAGES: Record<BusinessErrorCode, string> = {
  division_by_zero: "Cannot divide by zero",
  negative_square_root: "Invalid input for square root",
  exponent_not_integer: "Exponent must be an integer",
  exponent_out_of_range: "Exponent must be between -1000 and 1000",
  result_out_of_range: "Result is too large",
  invalid_number: "Invalid number",
  number_too_long: "Invalid number",
};

export const SERVICE_UNAVAILABLE_MESSAGE = "Service unavailable";
