import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { TierBadge, TIER_STYLES } from "./TierBadge";
import type { Outcome } from "@/lib/types";

const ALL_OUTCOMES: Outcome[] = ["green", "amber", "red", "mixed", "handoff"];

describe("TierBadge", () => {
  it.each(ALL_OUTCOMES)(
    "for outcome '%s', renders an icon, a label, AND a description together — never color alone",
    (outcome) => {
      render(<TierBadge outcome={outcome} />);
      const style = TIER_STYLES[outcome];
      // The component's own source comment (TierBadge.tsx) explains why:
      // meaning has to survive for a colorblind reader or a washed-out
      // phone screen in bright sunlight at a hospital entrance. This test
      // exists so a future edit that removes the label or description
      // text (leaving only a colored icon) fails loudly instead of only
      // being caught by someone re-reading that comment.
      expect(screen.getByText(style.label)).toBeInTheDocument();
      expect(screen.getByText(style.description)).toBeInTheDocument();
      expect(screen.getByText(style.icon)).toBeInTheDocument();
    },
  );

  it("marks the icon glyph aria-hidden, since the adjacent text already carries the meaning", () => {
    render(<TierBadge outcome="green" />);
    const icon = screen.getByText(TIER_STYLES.green.icon);
    expect(icon).toHaveAttribute("aria-hidden", "true");
  });

  it.each(ALL_OUTCOMES)(
    "outcome '%s' has a genuinely distinct icon from every other outcome",
    (outcome) => {
      // Two outcomes sharing an icon would defeat the "icon carries
      // meaning independent of color" design -- explicit pairwise check
      // rather than trusting the styles object was written correctly.
      const otherOutcomes = ALL_OUTCOMES.filter((o) => o !== outcome);
      for (const other of otherOutcomes) {
        expect(TIER_STYLES[outcome].icon).not.toBe(TIER_STYLES[other].icon);
      }
    },
  );

  it("every tier style has a non-empty label and description", () => {
    // Guards the data itself, not just one rendered instance -- catches
    // a future added outcome that forgot to fill these fields in.
    for (const outcome of ALL_OUTCOMES) {
      expect(TIER_STYLES[outcome].label.length).toBeGreaterThan(0);
      expect(TIER_STYLES[outcome].description.length).toBeGreaterThan(0);
    }
  });
});
