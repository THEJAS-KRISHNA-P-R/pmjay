"use client";

import { useState } from "react";
import { IconCopy, IconCheck } from "./icons";

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
      // Clipboard API fallback
    }
  }

  return (
    <section className="card p-6 sm:p-8 animate-fade-in-up">
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3.5 sm:gap-4">
        <div className="space-y-1">
          <h2 className="text-lg sm:text-xl font-bold tracking-tight text-sand-900">{title}</h2>
          {helperText && (
            <p className="text-xs sm:text-sm leading-relaxed text-sand-600">{helperText}</p>
          )}
        </div>
        <button
          type="button"
          onClick={handleCopy}
          className={`tap-target shrink-0 inline-flex items-center justify-center gap-2 rounded-2xl px-5 py-2.5 text-sm font-bold transition-all active:scale-95 ${
            copied
              ? "bg-tier-green-text text-white"
              : "bg-teal-700 text-white hover:bg-teal-800"
          }`}
          aria-live="polite"
        >
          {copied ? <IconCheck className="h-4 w-4" /> : <IconCopy className="h-4 w-4" />}
          {copied ? "Copied" : "Copy"}
        </button>
      </div>

      <div className="relative mt-5">
        <pre className="max-h-72 overflow-auto whitespace-pre-wrap rounded-2xl bg-sand-100 border border-sand-200/70 p-4 sm:p-5 font-sans text-xs sm:text-sm leading-relaxed text-sand-800">
          {text}
        </pre>
      </div>
    </section>
  );
}
