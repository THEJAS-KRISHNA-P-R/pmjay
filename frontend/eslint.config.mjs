import js from "@eslint/js";
import tseslint from "typescript-eslint";
import nextPlugin from "@next/eslint-plugin-next";

// A native ESLint 9 flat config, deliberately NOT built through
// eslint-config-next's FlatCompat bridge: that bridge re-validates the
// resolved config via JSON.stringify, and the specific combination of
// packages resolved for this project (see package.json) produces a
// circular reference in eslint-plugin-react's flat-config export that
// crashes that validation step. Configuring the Next.js plugin directly,
// as done below, is both a real fix and — per ESLint's own current
// documentation — the more idiomatic way to configure flat config in the
// first place; the string-based `extends: ["next/core-web-vitals"]` API
// this replaces is itself a compatibility shim for pre-flat-config
// tooling.
export default tseslint.config(
  { ignores: [".next/**", "node_modules/**"] },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    plugins: { "@next/next": nextPlugin },
    rules: {
      ...nextPlugin.configs.recommended.rules,
      ...nextPlugin.configs["core-web-vitals"].rules,
    },
  },
);
