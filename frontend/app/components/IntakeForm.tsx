"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { createCase, ApiError } from "@/lib/api";

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
          <p className="mt-1 text-xs sm:text-sm leading-relaxed text-sand-900/70">
            Write it however comes naturally — English, Malayalam, or a mix of both. There&rsquo;s
            no wrong way to describe it.
          </p>
          <div className="relative mt-3">
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={5}
              className="w-full rounded-2xl bg-sand-100/60 p-4 sm:p-5 text-base leading-relaxed text-sand-900 shadow-[inset_0_2px_4px_rgba(0,0,0,0.03)] placeholder:text-sand-900/40 transition focus:bg-white focus:shadow-[0_0_0_3px_rgba(47,127,115,0.25)] focus:outline-none"
              placeholder="For example: My father needs an operation, the hospital says our Ayushman card won't cover it..."
            />
          </div>
        </div>

        {error && (
          <div 
            role="alert" 
            className="rounded-2xl bg-tier-red-bg p-4 sm:p-5 shadow-xs backdrop-blur-sm animate-enter"
          >
            <p className="font-bold text-tier-red-text">{error}</p>
            {fallbackGuidance && (
              <p className="mt-2 text-xs sm:text-sm text-tier-red-text/90 leading-relaxed pt-2 opacity-90">{fallbackGuidance}</p>
            )}
          </div>
        )}

        <div>
          <button
            type="submit"
            disabled={isPending}
            className="group relative inline-flex w-full sm:w-auto items-center justify-center gap-2.5 rounded-2xl bg-teal-800 px-8 py-4 text-base sm:text-lg font-bold text-white shadow-md transition touch-spring hover:bg-teal-900 hover:shadow-lg active:scale-[0.98] disabled:opacity-60 focus-visible:outline-teal-600"
          >
            {isPending ? (
              <>
                <svg
                  className="h-5 w-5 animate-spin text-white"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                  />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
                <span>Looking into it…</span>
              </>
            ) : (
              <>
                <span>Get help now</span>
                <span className="text-white/80 transition-transform group-hover:translate-x-0.5" aria-hidden="true">
                  →
                </span>
              </>
            )}
          </button>
        </div>
      </form>

      <div className="pt-3">
        <p className="text-xs sm:text-sm font-bold text-sand-900/70">Or start from something similar:</p>
        <div className="mt-3 flex flex-wrap gap-2.5">
          {EXAMPLE_PROMPTS.map((prompt) => (
            <button
              key={prompt}
              type="button"
              onClick={() => setDescription(prompt)}
              className="rounded-full bg-sand-100/75 px-4 py-2.5 text-left text-xs sm:text-sm text-sand-900/85 shadow-xs transition touch-spring hover:bg-white hover:text-teal-950 hover:shadow-sm active:scale-95 focus-visible:outline-teal-600"
            >
              {prompt}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
