import type { Outcome } from "@/lib/types";
import { TIER_STYLES, TierBadge } from "./TierBadge";

export function TierPanel({
  outcome,
  message,
  citation,
}: {
  outcome: Outcome;
  message: string;
  citation?: string;
}) {
  const style = TIER_STYLES[outcome];
  return (
    <section
      className={`rounded-xl border-2 ${style.borderClass} ${style.bgClass} p-5 sm:p-6`}
      aria-labelledby="tier-heading"
    >
      <div id="tier-heading">
        <TierBadge outcome={outcome} />
      </div>
      <p className="mt-4 whitespace-pre-line text-base leading-relaxed text-sand-900">
        {message}
      </p>
      {citation && (
        <p className={`mt-4 border-t ${style.borderClass} pt-3 text-sm ${style.textClass}`}>
          <span className="font-bold">Based on: </span>
          {citation}
        </p>
      )}
    </section>
  );
}
