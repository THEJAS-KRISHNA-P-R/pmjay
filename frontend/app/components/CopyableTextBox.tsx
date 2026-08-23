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
      // Clipboard API fallback
    }
  }

  return (
    <section className="glass-panel p-6 sm:p-8 animate-enter">
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3.5 sm:gap-4">
        <div className="space-y-1">
          <h2 className="text-lg sm:text-xl font-bold tracking-tight text-teal-950">{title}</h2>
          {helperText && (
            <p className="text-xs sm:text-sm leading-relaxed text-sand-900/70">{helperText}</p>
          )}
        </div>
        <button
          type="button"
          onClick={handleCopy}
          className={`shrink-0 inline-flex items-center justify-center gap-2 rounded-2xl px-5 py-2.5 text-sm font-bold transition touch-spring shadow-xs focus-visible:outline-teal-600 ${
            copied
              ? "bg-emerald-700 text-white scale-105"
              : "bg-teal-800 text-white hover:bg-teal-900 active:scale-95"
          }`}
          aria-live="polite"
        >
          {copied ? "Copied ✓" : "Copy"}
        </button>
      </div>
      
      <div className="relative mt-5">
        <pre className="max-h-72 overflow-auto whitespace-pre-wrap rounded-2xl bg-sand-100/65 p-4 sm:p-5 font-sans text-xs sm:text-sm leading-relaxed text-sand-900 shadow-inner">
          {text}
        </pre>
      </div>
    </section>
  );
}
