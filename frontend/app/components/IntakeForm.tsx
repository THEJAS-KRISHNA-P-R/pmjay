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
    <div>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="description" className="block text-lg font-bold text-sand-900">
            What&rsquo;s happening at the hospital?
          </label>
          <p className="mt-1 text-sm text-sand-900/70">
            Write it however comes naturally — English, Malayalam, or a mix of both. There&rsquo;s
            no wrong way to describe it.
          </p>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={6}
            className="mt-3 w-full rounded-xl border-2 border-sand-200 bg-white p-4 text-base leading-relaxed text-sand-900 focus:border-teal-500"
            placeholder="For example: My father needs an operation, the hospital says our Ayushman card won't cover it..."
          />
        </div>

        {error && (
          <div role="alert" className="rounded-lg border border-tier-red-border bg-tier-red-bg p-4">
            <p className="font-bold text-tier-red-text">{error}</p>
            {fallbackGuidance && (
              <p className="mt-2 text-sm text-tier-red-text/90">{fallbackGuidance}</p>
            )}
          </div>
        )}

        <button
          type="submit"
          disabled={isPending}
          className="w-full rounded-xl bg-teal-700 px-6 py-4 text-lg font-bold text-white shadow-sm transition hover:bg-teal-800 disabled:opacity-60 sm:w-auto"
        >
          {isPending ? "Looking into it…" : "Get help now"}
        </button>
      </form>

      <div className="mt-8">
        <p className="text-sm font-bold text-sand-900/70">Or start from something similar:</p>
        <div className="mt-2 flex flex-wrap gap-2">
          {EXAMPLE_PROMPTS.map((prompt) => (
            <button
              key={prompt}
              type="button"
              onClick={() => setDescription(prompt)}
              className="rounded-full border border-sand-200 bg-white px-4 py-2 text-left text-sm text-sand-900/80 transition hover:border-teal-400 hover:text-teal-800"
            >
              {prompt}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
