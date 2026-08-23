import type { Outcome } from "@/lib/types";

interface TierStyle {
  label: string;
  bgClass: string;
  borderClass: string;
  textClass: string;
  strongClass: string;
  icon: string; // a plain-text/emoji glyph, never relying on color alone to convey meaning (Appendix K)
  description: string;
}

// Deliberately not just "green means good, red means bad" — every tier
// pairs its color with an icon AND a plain-language label, so the
// meaning survives for a colorblind reader or a low-contrast phone
// screen in bright sunlight at a hospital entrance.
export const TIER_STYLES: Record<Outcome, TierStyle> = {
  green: {
    label: "This looks covered",
    bgClass: "bg-tier-green-bg",
    borderClass: "border-tier-green-border",
    textClass: "text-tier-green-text",
    strongClass: "text-tier-green-strong",
    icon: "✓",
    description: "The scheme should be paying for this.",
  },
  amber: {
    label: "Needs one more check",
    bgClass: "bg-tier-amber-bg",
    borderClass: "border-tier-amber-border",
    textClass: "text-tier-amber-text",
    strongClass: "text-tier-amber-strong",
    icon: "?",
    description: "Not a clear yes or no yet — here's what to ask.",
  },
  red: {
    label: "This is correctly not covered",
    bgClass: "bg-tier-red-bg",
    borderClass: "border-tier-red-border",
    textClass: "text-tier-red-text",
    strongClass: "text-tier-red-strong",
    icon: "i",
    description: "Not a dispute — a straight answer about the rule.",
  },
  mixed: {
    label: "Part covered, part not",
    bgClass: "bg-gradient-to-br from-tier-green-bg to-tier-amber-bg",
    borderClass: "border-tier-amber-border",
    textClass: "text-tier-amber-text",
    strongClass: "text-tier-amber-strong",
    icon: "½",
    description: "Two different parts of the same bill, treated separately.",
  },
  handoff: {
    label: "Let's get you a person",
    bgClass: "bg-tier-handoff-bg",
    borderClass: "border-tier-handoff-border",
    textClass: "text-tier-handoff-text",
    strongClass: "text-tier-handoff-strong",
    icon: "→",
    description: "This needs a human — connecting you with free legal help.",
  },
};

export function TierBadge({ outcome }: { outcome: Outcome }) {
  const style = TIER_STYLES[outcome];
  return (
    <div className="flex items-center gap-3">
      <span
        aria-hidden="true"
        className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full border-2 text-lg font-bold ${style.borderClass} ${style.strongClass} bg-white`}
      >
        {style.icon}
      </span>
      <div>
        <p className={`text-lg font-bold ${style.strongClass}`}>{style.label}</p>
        <p className="text-sm text-sand-900/70">{style.description}</p>
      </div>
    </div>
  );
}
