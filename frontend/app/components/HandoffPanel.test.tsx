import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { HandoffPanel } from "./HandoffPanel";

describe("HandoffPanel", () => {
  it("renders a working tel: link to NALSA's free helpline, 15100", () => {
    // Distinct from Header's 14555 (the PMJAY scheme helpline) --
    // 15100 is NALSA's free legal aid line, the actual human handoff
    // this tier promises. Confusing the two numbers here would send a
    // family calling the wrong desk for what this tier says it's doing.
    render(<HandoffPanel summary="Summary text" />);
    const link = screen.getByRole("link", { name: /15100/ });
    expect(link).toHaveAttribute("href", "tel:15100");
  });

  it("explicitly names NALSA and states the help is free", () => {
    // "free" is load-bearing here -- a family in a billing dispute is
    // exactly the audience that needs to not wonder whether calling
    // this number leads to another bill. Two elements legitimately
    // contain "free" (the explainer paragraph and the call-to-action
    // link's own text) -- getAllByText confirms both, rather than
    // picking one arbitrarily with getByText.
    render(<HandoffPanel summary="Summary text" />);
    expect(screen.getByText(/NALSA/)).toBeInTheDocument();
    expect(screen.getAllByText(/free/i).length).toBeGreaterThanOrEqual(2);
  });

  it("renders the case summary so the family doesn't have to repeat it on the call", () => {
    const summary = "Family describes a denied cataract surgery claim at a district hospital.";
    render(<HandoffPanel summary={summary} />);
    expect(screen.getByText(summary)).toBeInTheDocument();
  });

  it("renders without crashing even with an empty summary", () => {
    // Defensive, same reasoning as CareFirstBanner.test.tsx's equivalent
    // case: builder.go always populates HandoffSummary for a real
    // handoff decision, but this component shouldn't assume that and
    // take the rest of the page down if it somehow didn't.
    expect(() => render(<HandoffPanel summary="" />)).not.toThrow();
  });
});
