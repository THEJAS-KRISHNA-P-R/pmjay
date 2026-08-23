"use client";

import { useState } from "react";

export function CopyableTextBox({
  title,
  helperText,
  text,
}: {
  title: string;
  helperText?: string;
  text: string;
}) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    } catch {
      // Clipboard API can fail (older browsers, permissions) — the text
      // is still fully selectable/readable in the box below, so this is
      // a degraded experience, not a broken one.
    }
  }

  return (
    <section className="rounded-xl border border-sand-200 bg-white p-5 sm:p-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-lg font-bold text-teal-800">{title}</h2>
          {helperText && <p className="mt-1 text-sm text-sand-900/70">{helperText}</p>}
        </div>
        <button
          type="button"
          onClick={handleCopy}
          className="shrink-0 rounded-lg border-2 border-teal-700 bg-white px-4 py-2 text-sm font-bold text-teal-800 transition hover:bg-teal-50 active:bg-teal-100"
          aria-live="polite"
        >
          {copied ? "Copied ✓" : "Copy"}
        </button>
      </div>
      <pre className="mt-4 max-h-72 overflow-auto whitespace-pre-wrap rounded-lg bg-sand-100 p-4 font-sans text-sm leading-relaxed text-sand-900">
        {text}
      </pre>
    </section>
  );
}
