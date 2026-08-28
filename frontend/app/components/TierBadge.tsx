import type { ComponentType } from "react";
import type { Outcome } from "@/lib/types";
import {
  IconShieldCheck,
  IconClipboardList,
  IconX,
  IconHalfCoverage,
  IconUser,
  type IconProps,
} from "./icons";

interface TierStyle {
  label: string;
  description: string;
  /** The icon component itself — swappable, testable by reference. */
  Icon: ComponentType<IconProps>;
  /** Stable string id for the icon, purely for tests/analytics — never rendered as text. */
  iconName: string;
  panelBg: string;
  panelBorder: string;
  iconBg: string;
  iconText: string;
}

// Every tier maps to its own muted --color-tier-* token (defined in
// globals.css) rather than a stock Tailwind palette color. That's
// deliberate: this app is often the first thing someone reads after a
// hospital billing desk has just told them "no," so none of these five
// states is allowed to look like a stoplight or a form-validation error.
// Color is reinforcement only — every state also gets its own icon and
// its own label, so the meaning survives grayscale, colorblindness, or
// a washed-out phone screen in direct sun.
export const TIER_STYLES: Record<Outcome, TierStyle> = {
  green: {
    label: "This looks covered",
    description: "The scheme should be paying for this.",
    Icon: IconShieldCheck,
    iconName: "shield-check",
    panelBg: "bg-tier-green-bg",
    panelBorder: "border-tier-green-border",
    iconBg: "bg-tier-green-icon",
    iconText: "text-tier-green-text",
  },
  amber: {
    label: "Needs one more check",
    description: "Not a clear yes or no yet: here's what to ask.",
    Icon: IconClipboardList,
    iconName: "clipboard-list",
    panelBg: "bg-tier-amber-bg",
    panelBorder: "border-tier-amber-border",
    iconBg: "bg-tier-amber-icon",
    iconText: "text-tier-amber-text",
  },
  red: {
    label: "This is correctly not covered",
    description: "Not a dispute: a straight answer about the rule.",
    Icon: IconX,
    iconName: "x",
    panelBg: "bg-tier-red-bg",
    panelBorder: "border-tier-red-border",
    iconBg: "bg-tier-red-icon",
    iconText: "text-tier-red-text",
  },
  mixed: {
    label: "Part covered, part not",
    description: "Two different parts of the same bill, treated separately.",
    Icon: IconHalfCoverage,
    iconName: "half-coverage",
    panelBg: "bg-tier-mixed-bg",
    panelBorder: "border-tier-mixed-border",
    iconBg: "bg-tier-mixed-icon",
    iconText: "text-tier-mixed-text",
  },
  handoff: {
    label: "Let's get you a person",
    description: "This needs a human (connecting you with free legal help).",
    Icon: IconUser,
    iconName: "user",
    panelBg: "bg-tier-handoff-bg",
    panelBorder: "border-tier-handoff-border",
    iconBg: "bg-tier-handoff-icon",
    iconText: "text-tier-handoff-text",
  },
};

export function TierBadge({ outcome }: { outcome: Outcome }) {
  const style = TIER_STYLES[outcome];
  const { Icon } = style;
  return (
    <div className="flex items-center gap-3.5 sm:gap-4">
      <span
        data-testid="tier-icon"
        data-icon={style.iconName}
        aria-hidden="true"
        className={`flex h-11 w-11 sm:h-12 sm:w-12 shrink-0 items-center justify-center rounded-2xl ${style.iconBg} ${style.iconText}`}
      >
        <Icon className="h-5 w-5 sm:h-6 sm:w-6" />
      </span>
      <div>
        <p className={`text-lg sm:text-xl font-bold tracking-tight ${style.iconText}`}>
          {style.label}
        </p>
        <p className="text-xs sm:text-sm text-sand-700 leading-snug mt-0.5 font-medium">
          {style.description}
        </p>
      </div>
    </div>
  );
}
