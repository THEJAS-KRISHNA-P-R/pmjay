import type { Outcome } from "@/lib/types";

interface TierStyle {
  label: string;
  bgClass: string;
  borderClass: string;
  textClass: string;
  strongClass: string;
  badgeBg: string;
  icon: string;
  description: string;
}

export const TIER_STYLES: Record<Outcome, TierStyle> = {
  green: {
    label: "This looks covered",
    bgClass: "bg-emerald-50/80",
    borderClass: "border-transparent",
    textClass: "text-emerald-900",
    strongClass: "text-emerald-800",
    badgeBg: "bg-emerald-100 text-emerald-800",
    icon: "✓",
    description: "The scheme should be paying for this.",
  },
  amber: {
    label: "Needs one more check",
    bgClass: "bg-amber-50/85",
    borderClass: "border-transparent",
    textClass: "text-amber-950",
    strongClass: "text-amber-900",
    badgeBg: "bg-amber-100 text-amber-900",
    icon: "?",
    description: "Not a clear yes or no yet — here's what to ask.",
  },
  red: {
    label: "This is correctly not covered",
    bgClass: "bg-rose-50/80",
    borderClass: "border-transparent",
    textClass: "text-rose-950",
    strongClass: "text-rose-900",
    badgeBg: "bg-rose-100 text-rose-900",
    icon: "i",
    description: "Not a dispute — a straight answer about the rule.",
  },
  mixed: {
    label: "Part covered, part not",
    bgClass: "bg-gradient-to-br from-emerald-50/90 to-amber-50/90",
    borderClass: "border-transparent",
    textClass: "text-amber-950",
    strongClass: "text-amber-900",
    badgeBg: "bg-amber-100 text-amber-900",
    icon: "½",
    description: "Two different parts of the same bill, treated separately.",
  },
  handoff: {
    label: "Let's get you a person",
    bgClass: "bg-teal-50/90",
    borderClass: "border-transparent",
    textClass: "text-teal-950",
    strongClass: "text-teal-900",
    badgeBg: "bg-teal-100 text-teal-900",
    icon: "→",
    description: "This needs a human — connecting you with free legal help.",
  },
};

export function TierBadge({ outcome }: { outcome: Outcome }) {
  const style = TIER_STYLES[outcome];
  return (
    <div className="flex items-center gap-3.5 sm:gap-4">
      <span
        aria-hidden="true"
        className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl text-xl font-bold shadow-xs transition-transform ${style.badgeBg}`}
      >
        {style.icon}
      </span>
      <div>
        <p className={`text-lg sm:text-xl font-bold tracking-tight ${style.strongClass}`}>
          {style.label}
        </p>
        <p className="text-xs sm:text-sm text-sand-900/75 leading-tight mt-0.5 font-medium">
          {style.description}
        </p>
      </div>
    </div>
  );
}
