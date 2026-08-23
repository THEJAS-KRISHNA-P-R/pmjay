import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "path";

// Separate from next.config.mjs deliberately: Next's own build/dev
// pipeline (Turbopack) doesn't run Vitest, and Vitest doesn't need
// Next's config -- these are two independent tools that happen to
// share this one repo. Path alias below is hand-kept in sync with
// tsconfig.json's "@/*" mapping (same reasoning lib/types.ts gives for
// not code-generating the DTO types: small, stable surface, not worth
// a sync-checking step for two config files that rarely change).
export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    setupFiles: ["./vitest.setup.ts"],
    globals: true,
    css: false,
  },
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "."),
    },
  },
});
