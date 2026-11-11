import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/api": {
        target: process.env.VITE_API_PROXY_TARGET ?? "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    restoreMocks: true,
    coverage: {
      provider: "v8",
      include: ["src/**/*.{ts,tsx}"],
      exclude: ["src/main.tsx", "src/test/**", "src/**/*.test.{ts,tsx}"],
      reporter: ["text", "html", "lcov"],
      thresholds: {
        "src/features/**": { lines: 90, functions: 90, statements: 90, branches: 85 },
        "src/api/**": { lines: 90, functions: 90, statements: 90, branches: 85 },
      },
    },
  },
});
