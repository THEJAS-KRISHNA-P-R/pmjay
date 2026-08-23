import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { Header } from "./Header";

describe("Header", () => {
  it("renders the brand name linking back to the home page", () => {
    render(<Header />);
    const brandLink = screen.getByRole("link", { name: /pmjay advocate/i });
    expect(brandLink).toHaveAttribute("href", "/");
  });

  it("renders a working tel: link to the PMJAY helpline, 14555", () => {
    // This is the one helpline number a family can reach before a case
    // even exists yet (the in-case fallback on error, and HandoffPanel's
    // 15100, are different numbers for different purposes — see
    // app/case/[id]/page.tsx and HandoffPanel.tsx). A wrong or missing
    // href here would be silent: it still looks like a working button.
    render(<Header />);
    const helplineLink = screen.getByRole("link", { name: /14555/ });
    expect(helplineLink).toHaveAttribute("href", "tel:14555");
  });
});
