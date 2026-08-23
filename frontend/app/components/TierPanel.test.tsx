import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { TierPanel } from "./TierPanel";
import { TIER_STYLES } from "./TierBadge";
import type { Outcome } from "@/lib/types";

const ALL_OUTCOMES: Outcome[] = ["green", "amber", "red", "mixed", "handoff"];

describe("TierPanel", () => {
  it.each(ALL_OUTCOMES)(
    "for outcome '%s', renders the matching TierBadge label and the message text",
    (outcome) => {
      render(<TierPanel outcome={outcome} message="This is the tier message." />);
      expect(screen.getByText(TIER_STYLES[outcome].label)).toBeInTheDocument();
      expect(screen.getByText("This is the tier message.")).toBeInTheDocument();
    },
  );

  it("renders the citation, prefixed 'Based on:', when one is provided", () => {
    render(
      <TierPanel
        outcome="green"
        message="Message text"
        citation="Cataract Surgery (Ophthalmology), listed with an indicative rate of ₹7,500"
      />,
    );
    expect(screen.getByText(/Based on:/)).toBeInTheDocument();
    expect(
      screen.getByText(/Cataract Surgery \(Ophthalmology\)/),
    ).toBeInTheDocument();
  });

  it("renders no citation section at all when citation is omitted", () => {
    // Amber/handoff decisions often have no single matched package to
    // cite (see backend/internal/response/templates.go's amberMessage
    // default case) -- an empty or "Based on: " with nothing after it
    // would look like a rendering bug, not an intentional omission.
    render(<TierPanel outcome="amber" message="Message text" />);
    expect(screen.queryByText(/Based on:/)).not.toBeInTheDocument();
  });

  it("preserves line breaks in the message", () => {
    const { container } = render(
      <TierPanel outcome="mixed" message={"First part.\n\nSecond part."} />,
    );
    // whitespace-pre-line is what makes the \n\n visually meaningful;
    // this asserts the raw text (with the break) actually reached the
    // DOM rather than being collapsed by the JSX/HTML layer before the
    // CSS class ever gets a chance to act on it.
    expect(container).toHaveTextContent(/First part\.\s+Second part\./);
  });

  it("associates the panel with its heading via aria-labelledby", () => {
    render(<TierPanel outcome="red" message="Message text" />);
    const heading = document.getElementById("tier-heading");
    expect(heading).not.toBeNull();
    const section = screen.getByText("Message text").closest("section");
    expect(section).toHaveAttribute("aria-labelledby", "tier-heading");
  });
});
