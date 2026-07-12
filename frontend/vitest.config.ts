import { defineConfig } from "vitest/config";
import path from "node:path";

// Unit-test harness for the pure `src/lib` logic (the runner's mirror of the Go
// logic engine). Node environment — these tests exercise plain functions, no DOM.
// The dual-engine PARITY test (logicEngine.parity.test.ts) reads the SAME JSON
// fixture the Go test reads, so a divergence between the two engines goes red.
export default defineConfig({
  resolve: {
    alias: { "@": path.resolve(__dirname, "./src") },
  },
  test: {
    environment: "node",
    include: ["src/**/*.test.ts"],
    coverage: {
      provider: "v8",
      include: ["src/lib/logicEngine.ts"],
      reporter: ["text-summary"],
    },
  },
});
