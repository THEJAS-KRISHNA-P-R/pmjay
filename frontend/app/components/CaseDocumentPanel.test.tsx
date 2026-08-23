import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { CaseDocumentPanel } from "./CaseDocumentPanel";

describe("CaseDocumentPanel", () => {
  it("links to the backend's PDF document endpoint for the given case id", () => {
    render(<CaseDocumentPanel caseId="case-123" />);
    const link = screen.getByRole("link", { name: /download pdf/i });
    expect(link).toHaveAttribute("href", "/api/v1/cases/case-123/document");
  });

  it("URL-encodes the case id, same as every other case-scoped link in this app", () => {
    render(<CaseDocumentPanel caseId="case with spaces" />);
    const link = screen.getByRole("link", { name: /download pdf/i });
    expect(link).toHaveAttribute("href", "/api/v1/cases/case%20with%20spaces/document");
  });

  it("opens in a new tab via a real navigation, not a fetch/blob download", () => {
    // target=_blank + rel=noopener is what lets the browser's own PDF
    // viewer (with its print/save/share controls) handle the response
    // — see this component's doc comment for why that's deliberate.
    render(<CaseDocumentPanel caseId="case-123" />);
    const link = screen.getByRole("link", { name: /download pdf/i });
    expect(link).toHaveAttribute("target", "_blank");
    expect(link).toHaveAttribute("rel", expect.stringContaining("noopener"));
  });

  it("explains what the document is for, not just that one exists", () => {
    render(<CaseDocumentPanel caseId="case-123" />);
    expect(screen.getByText(/print, save, or hand to hospital staff/i)).toBeInTheDocument();
  });

  it("renders without crashing given an empty case id", () => {
    // Defensive, matching this codebase's established pattern (see e.g.
    // HandoffPanel.test.tsx) — CasePage never actually renders this with
    // an empty id in practice, but a rendering bug here shouldn't be
    // able to take the rest of the result page down with it.
    expect(() => render(<CaseDocumentPanel caseId="" />)).not.toThrow();
  });
});
