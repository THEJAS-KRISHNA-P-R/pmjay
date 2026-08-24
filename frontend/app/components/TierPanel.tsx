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
      className={`rounded-3xl border ${style.panelBg} ${style.panelBorder} p-6 sm:p-8 shadow-sm animate-fade-in-up`}
      aria-labelledby="tier-heading"
    >
      <div id="tier-heading">
        <TierBadge outcome={outcome} />
      </div>
      <div className="mt-5 pt-4 border-t border-black/5">
        <p className="whitespace-pre-line text-base sm:text-lg leading-relaxed text-sand-900 font-medium">
          {message}
        </p>
      </div>
      {citation && (
        <div className="mt-5 card p-4 sm:p-5 text-xs sm:text-sm">
          <p className="leading-relaxed text-sand-800">
            <span className="font-bold text-teal-800">Based on: </span>
            {citation}
          </p>
        </div>
      )}
    </section>
  );
}
