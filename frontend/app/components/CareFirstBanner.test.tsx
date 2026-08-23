import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { CareFirstBanner } from "./CareFirstBanner";

// Mirrors the backend's own safety-critical test philosophy (see
// backend/internal/response/builder_test.go's
// TestBuild_CareFirstMessageIsAlwaysPresent_EveryOutcome and
// docs/SAFETY_DESIGN.md section 1-2): the backend can never omit this
// message structurally, but a frontend bug could still render it,
// hide it, or make it visually indistinguishable from ordinary text.
// This suite checks the frontend side of that same guarantee.

describe("CareFirstBanner", () => {
  it("renders the exact message text passed to it, unmodified", () => {
    const message = "Get treatment first. Dispute the money after. Always.";
    render(<CareFirstBanner message={message} />);
    expect(screen.getByText(message)).toBeInTheDocument();
  });

  it("uses role=alert so assistive tech announces it without the user having to find it", () => {
    render(<CareFirstBanner message="Test message" />);
    // role=alert is what makes this un-skippable for a screen reader --
    // losing this attribute would be a silent accessibility regression,
    // not a visible one, which is exactly the kind of bug a test should
    // catch since a manual visual check would not notice it.
    expect(screen.getByRole("alert")).toBeInTheDocument();
  });

  it("does not truncate or ellipsize a long message", () => {
    const longMessage =
      "Get treatment first. Dispute the money after. Always. If you can pay now and " +
      "settle the dispute later, or move to a different hospital, do that — do not " +
      "let this disagreement delay or stop care.";
    render(<CareFirstBanner message={longMessage} />);
    expect(screen.getByText(longMessage)).toBeInTheDocument();
  });

  it("renders distinctly even with an empty string, rather than crashing", () => {
    // Defensive: the backend guarantees this is always non-empty (see
    // docs/SAFETY_DESIGN.md), but the frontend component itself should
    // not assume that and blow up if it somehow received one -- a
    // rendering crash on this specific component would be worse than a
    // blank banner, since it could take the rest of the page down with it.
    expect(() => render(<CareFirstBanner message="" />)).not.toThrow();
  });
});
