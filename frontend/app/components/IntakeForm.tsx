"use client";

import { useEffect, useRef, useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { createCase, ApiError, RateLimitError } from "@/lib/api";
import { saveCaseToHistory } from "@/lib/caseHistory";
import { IconSpinner, IconArrowRight, IconClock } from "./icons";

const EXAMPLE_PROMPTS = [
  "My mother needs gallbladder surgery, hospital says our card won't cover it",
  "They already did the operation but now say the insurance hasn't cleared it",
  "ammayude kaal odinjappo, hospital surgery de bill njangalod adakkan paranju",
  "meri maa ka pair toot gaya, hospital ne surgery ka bill humse maanga",
];

// How the real pipeline actually processes one submission (see
// ARCHITECTURE.md's "System shape" diagram): a cheap keyword pass, the
// one LLM call, deterministic tiering, then the care-first response
// gets built. This cycles through plain-language labels for those real
// stages while the single request is in flight — it is not a live
// progress feed (the backend answers in one response, it doesn't
// stream steps), so the labels rotate on a timer rather than claiming
// to track exact progress. Honest about being a wait state, not a
// fabricated one.
const PROCESSING_STAGES = [
  "Reading what you shared…",
  "Checking it against PMJAY coverage…",
  "Working out what this means for you…",
  "Preparing your explanation…",
];

export function IntakeForm() {
  const router = useRouter();
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [fallbackGuidance, setFallbackGuidance] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();
  const [stageIndex, setStageIndex] = useState(0);
  const stageTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  
  // Rate limiting countdown state
  const [rateLimitCountdown, setRateLimitCountdown] = useState<number | null>(null);
  const rateLimitTimer = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    if (isPending) {
      setStageIndex(0);
      stageTimer.current = setInterval(() => {
        setStageIndex((i) => Math.min(i + 1, PROCESSING_STAGES.length - 1));
      }, 1300);
    } else if (stageTimer.current) {
      clearInterval(stageTimer.current);
      stageTimer.current = null;
    }
    return () => {
      if (stageTimer.current) clearInterval(stageTimer.current);
    };
  }, [isPending]);

  // Clean up rate limit timer on unmount
  useEffect(() => {
    return () => {
      if (rateLimitTimer.current) clearInterval(rateLimitTimer.current);
    };
  }, []);

  function startRateLimitCountdown(seconds: number) {
    setRateLimitCountdown(seconds);
    if (rateLimitTimer.current) clearInterval(rateLimitTimer.current);
    
    rateLimitTimer.current = setInterval(() => {
      setRateLimitCountdown((prev) => {
        if (prev === null || prev <= 1) {
          if (rateLimitTimer.current) clearInterval(rateLimitTimer.current);
          setError(null);
          return null;
        }
        return prev - 1;
      });
    }, 1000);
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (rateLimitCountdown !== null) return;
    
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
        saveCaseToHistory(result);
        router.push(`/cases/${result.id}`);
      } catch (err) {
        if (err instanceof RateLimitError) {
          setError(err.message);
          startRateLimitCountdown(err.retryAfterSeconds);
        } else if (err instanceof ApiError) {
          setError(err.message);
          setFallbackGuidance(err.fallbackGuidance ?? null);
        } else {
          setError("Something went wrong. Please try again.");
        }
      }
    });
  }

  const isFormDisabled = isPending || rateLimitCountdown !== null;

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="description" className="block text-base sm:text-lg font-bold text-sand-900">
            What&rsquo;s happening at the hospital?
          </label>
          <p className="mt-1 text-xs sm:text-sm text-sand-600 leading-relaxed">
            Describe your situation in your own words (English, Malayalam, or Hindi). Mixing languages is completely fine.
          </p>
          <div className="relative mt-3">
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={4}
              className="field p-4 sm:p-5 text-base leading-relaxed resize-none focus:ring-2 focus:ring-emerald-600/20"
              placeholder="For example: My father needs an operation, the hospital says our Ayushman card won't cover it..."
              lang=""
              disabled={isFormDisabled}
            />
          </div>
        </div>

        {error && (
          <div
            role="alert"
            className={`rounded-2xl border p-4 sm:p-5 animate-fade-in-up ${
              rateLimitCountdown !== null 
                ? "border-tier-amber-border bg-tier-amber-bg text-tier-amber-text" 
                : "border-tier-red-border bg-tier-red-bg text-tier-red-text"
            }`}
          >
            <p className="font-bold">{error}</p>
            {fallbackGuidance && (
              <p className="mt-2 text-xs sm:text-sm leading-relaxed pt-2 opacity-90">{fallbackGuidance}</p>
            )}
          </div>
        )}

        <div className="flex items-center justify-between pt-1">
          <button
            type="submit"
            disabled={isFormDisabled}
            className="btn-primary tap-target w-full sm:w-auto px-7 py-3.5 text-sm sm:text-base disabled:opacity-70 disabled:cursor-not-allowed disabled:active:scale-100"
          >
            {isPending ? (
              <>
                <IconSpinner className="h-5 w-5 animate-spin text-white shrink-0" />
                <span aria-live="polite">{PROCESSING_STAGES[stageIndex]}</span>
              </>
            ) : rateLimitCountdown !== null ? (
              <>
                <IconClock className="h-4 w-4 text-white shrink-0" />
                <span aria-live="polite">Please wait {rateLimitCountdown}s</span>
              </>
            ) : (
              <>
                <span>Get help now</span>
                <IconArrowRight className="h-4 w-4 text-white/90 transition-transform group-hover:translate-x-0.5" />
              </>
            )}
          </button>
        </div>
      </form>

      <div className="pt-2 border-t border-sand-100/80">
        <p className="text-xs font-bold uppercase tracking-wider text-sand-500">Or start from an example:</p>
        <div className="mt-2.5 flex flex-wrap gap-2">
          {EXAMPLE_PROMPTS.map((prompt) => (
            <button
              key={prompt}
              type="button"
              onClick={() => setDescription(prompt)}
              disabled={isPending}
              className="btn-secondary px-3.5 py-2 text-left text-xs text-sand-700 disabled:opacity-60"
            >
              {prompt}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
