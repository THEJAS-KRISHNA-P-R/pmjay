import { render, screen, waitFor, within } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import CasePage from "./page";
import type { CaseResponse } from "@/lib/types";
import { ApiError } from "@/lib/api";

// This is the one test in the suite that renders the real, full page
// component tree (Header + CareFirstBanner + TierPanel + conditionally
// HandoffPanel/ActionSteps/CopyableTextBox/EvidenceForm) exactly as
// app/case/[id]/page.tsx composes them, with only the network boundary
// mocked. Every other test in this suite exercises one component in
// isolation; this one exists specifically to catch a bug in how they're
// wired together, which isolated tests structurally cannot catch --
// e.g. a prop passed under the wrong name, or a conditional that hides
// a component that should be showing. See docs/TESTING.md for why this
// stands in for a true browser-driven e2e test in this environment.

vi.mock("next/navigation", () => ({
  useParams: () => ({ id: "case-123" }),
}));

const getCaseMock = vi.fn();
vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return {
    ...actual,
    getCase: (...args: unknown[]) => getCaseMock(...args),
  };
});

function baseCase(overrides: Partial<CaseResponse> = {}): CaseResponse {
  return {
    id: "case-123",
    outcome: "green",
    care_first_message: "Get treatment first. Dispute the money after. Always.",
    tier_message: "This looks like a covered package.",
    ...overrides,
  };
}

beforeEach(() => {
  getCaseMock.mockReset();
});

describe("CasePage (integration)", () => {
  it("shows a loading state, then the loaded case, without ever showing both at once", async () => {
    getCaseMock.mockResolvedValue(baseCase());
    render(<CasePage />);

    expect(screen.getByRole("status")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText(/This looks like a covered package/)).toBeInTheDocument());
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });

  it("always renders the care-first message, regardless of outcome — the one structural guarantee this whole tool exists to keep", async () => {
    // Mirrors backend/internal/response/builder_test.go's
    // TestBuild_CareFirstMessageIsAlwaysPresent_EveryOutcome, but for the
    // actual page a family sees, not just the JSON the backend sends.
    for (const outcome of ["green", "amber", "red", "mixed", "handoff"] as const) {
      getCaseMock.mockResolvedValue(
        baseCase({ outcome, handoff_summary: outcome === "handoff" ? "Summary" : undefined }),
      );
      const { unmount } = render(<CasePage />);
      await waitFor(() =>
        expect(
          screen.getByText("Get treatment first. Dispute the money after. Always."),
        ).toBeInTheDocument(),
      );
      unmount();
    }
  });

  it("renders HandoffPanel only for a handoff outcome, never for the other four", async () => {
    for (const outcome of ["green", "amber", "red", "mixed"] as const) {
      getCaseMock.mockResolvedValue(baseCase({ outcome }));
      const { unmount } = render(<CasePage />);
      await waitFor(() => expect(screen.getByText(/This looks like a covered package|tier message/i)).toBeTruthy());
      expect(screen.queryByText("Free legal help, right now")).not.toBeInTheDocument();
      unmount();
    }

    getCaseMock.mockResolvedValue(
      baseCase({ outcome: "handoff", handoff_summary: "Family's case summary for NALSA." }),
    );
    render(<CasePage />);
    await waitFor(() =>
      expect(screen.getByText("Free legal help, right now")).toBeInTheDocument(),
    );
    expect(screen.getByText("Family's case summary for NALSA.")).toBeInTheDocument();
  });

  it("renders action steps and copyable boxes only when the backend actually sent them", async () => {
    getCaseMock.mockResolvedValue(baseCase());
    render(<CasePage />);
    await waitFor(() => expect(screen.getByText(/This looks like a covered package/)).toBeInTheDocument());
    expect(screen.queryByText("What to do right now")).not.toBeInTheDocument();
    expect(screen.queryByText("Exact words to use at the desk")).not.toBeInTheDocument();
    expect(screen.queryByText("Draft complaint, ready to review")).not.toBeInTheDocument();
  });

  it("renders action steps, hospital script, and complaint text when the backend sends all three", async () => {
    getCaseMock.mockResolvedValue(
      baseCase({
        action_steps: ["Ask for the itemised bill", "Call the PMJAY helpline"],
        hospital_script: "Please note this on my file: I am invoking my PMJAY entitlement.",
        complaint_text: "To whom it may concern, I am writing to formally dispute...",
      }),
    );
    render(<CasePage />);

    await waitFor(() => expect(screen.getByText("What to do right now")).toBeInTheDocument());
    expect(screen.getByText("Ask for the itemised bill")).toBeInTheDocument();
    expect(screen.getByText("Exact words to use at the desk")).toBeInTheDocument();
    expect(screen.getByText("Draft complaint, ready to review")).toBeInTheDocument();
    // The complaint box carries a helper line the hospital-script box
    // does not (see app/case/[id]/page.tsx) -- this is the one thing
    // that actually distinguishes the two CopyableTextBox instances from
    // each other, so it's worth its own assertion rather than assuming
    // two matching titles means both are correctly configured.
    expect(screen.getByText(/Submit this yourself/)).toBeInTheDocument();
  });

  it("renders the case document download link with the correct case id wired through", async () => {
    // The specific class of bug this whole integration suite exists to
    // catch (see the file-level comment above): CaseDocumentPanel takes
    // caseId as a prop, and an isolated component test (CaseDocumentPanel.test.tsx)
    // can only confirm the component behaves correctly given *some* id
    // — it can't catch CasePage itself passing the wrong one, or none
    // at all. This is the one test that actually exercises that wiring.
    getCaseMock.mockResolvedValue(baseCase({ id: "case-456" }));
    render(<CasePage />);
    await waitFor(() => expect(screen.getByText(/This looks like a covered package/)).toBeInTheDocument());

    const link = screen.getByRole("link", { name: /download pdf/i });
    expect(link).toHaveAttribute("href", "/api/v1/cases/case-456/document");
  });

  it("shows an error state with a PMJAY helpline fallback when the case fails to load, and no tier content", async () => {
    getCaseMock.mockRejectedValue(new ApiError("Case not found.", 404));
    render(<CasePage />);

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("Case not found.");
    // Header renders its own "Helpline: 14555" link too, so this query
    // is deliberately scoped to inside the error alert -- an unscoped
    // query here is exactly the ambiguity that caught a real gap when
    // this test was first written (both links legitimately match
    // /14555/, and getByRole correctly refused to guess between them).
    expect(within(alert).getByRole("link", { name: /14555/ })).toHaveAttribute(
      "href",
      "tel:14555",
    );
    expect(screen.queryByText(/Based on:/)).not.toBeInTheDocument();
  });
});
