"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { createCase, ApiError } from "@/lib/api";
import { IconSpinner, IconArrowRight } from "./icons";

const EXAMPLE_PROMPTS = [
  "My mother needs gallbladder surgery, hospital says our card won't cover it",
  "They already did the operation but now say the insurance hasn't cleared it",
  "Doctor wants a cosmetic procedure alongside a covered one, hospital is billing both",
];

export function IntakeForm() {
  const router = useRouter();
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [fallbackGuidance, setFallbackGuidance] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setFallbackGuidance(null);

    const trimmed = description.trim();
    if (trimmed.length < 5) {
      setError("Please describe a bit more of what's happening.");
      return;
    }

    startTransition(async () => {
      try {
        const result = await createCase(trimmed);
        router.push(`/case/${result.id}`);
      } catch (err) {
        if (err instanceof ApiError) {
          setError(err.message);
          setFallbackGuidance(err.fallbackGuidance ?? null);
        } else {
          setError("Something went wrong. Please try again.");
        }
      }
    });
  }

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label htmlFor="description" className="block text-base sm:text-lg font-bold text-sand-900">
            What&rsquo;s happening at the hospital?
          </label>
          <p className="mt-1 text-xs sm:text-sm leading-relaxed text-sand-600">
            Describe your situation in standard English, native Malayalam script (മലയാളം), or Hindi.
            <span className="block text-sand-500 text-xs mt-0.5 font-normal">
              Note: Transliterated Malayalam typed in English letters (Manglish) is not supported.
            </span>
          </p>
          <div className="relative mt-3">
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={5}
              className="field p-4 sm:p-5 text-base leading-relaxed resize-none"
              placeholder="For example: My father needs an operation, the hospital says our Ayushman card won't cover it..."
            />
          </div>
        </div>

        {error && (
          <div
            role="alert"
            className="rounded-2xl border border-tier-red-border bg-tier-red-bg p-4 sm:p-5 animate-fade-in-up"
          >
            <p className="font-bold text-tier-red-text">{error}</p>
            {fallbackGuidance && (
              <p className="mt-2 text-xs sm:text-sm text-tier-red-text leading-relaxed pt-2 opacity-90">{fallbackGuidance}</p>
            )}
          </div>
        )}

        <div>
          <button
            type="submit"
            disabled={isPending}
            className="group relative inline-flex w-full sm:w-auto items-center justify-center gap-2.5 rounded-2xl bg-teal-700 px-8 py-4 text-base sm:text-lg font-bold text-white shadow-md transition-all hover:bg-teal-800 hover:shadow-lg active:scale-[0.98] disabled:opacity-60 disabled:active:scale-100"
          >
            {isPending ? (
              <>
                <IconSpinner className="h-5 w-5 animate-spin text-white" />
                <span>Looking into it…</span>
              </>
            ) : (
              <>
                <span>Get help now</span>
                <IconArrowRight className="h-4 w-4 text-white/80 transition-transform group-hover:translate-x-0.5" />
              </>
            )}
          </button>
        </div>
      </form>

      <div className="pt-3">
        <p className="text-xs sm:text-sm font-bold text-sand-600">Or start from something similar:</p>
        <div className="mt-3 flex flex-wrap gap-2.5">
          {EXAMPLE_PROMPTS.map((prompt) => (
            <button
              key={prompt}
              type="button"
              onClick={() => setDescription(prompt)}
              className="rounded-xl border border-sand-200 bg-white px-4 py-2.5 text-left text-xs sm:text-sm text-sand-700 transition-all hover:border-teal-300 hover:text-teal-800 hover:shadow-sm active:scale-95"
            >
              {prompt}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
