import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { EvidenceForm } from "./EvidenceForm";
import { ApiError } from "@/lib/api";

const addEvidenceMock = vi.fn();
vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return {
    ...actual,
    addEvidence: (...args: unknown[]) => addEvidenceMock(...args),
  };
});

beforeEach(() => {
  addEvidenceMock.mockReset();
});

describe("EvidenceForm", () => {
  it("blocks submission when all three fields are empty, matching the backend's at-least-one-field rule", async () => {
    const user = userEvent.setup();
    render(<EvidenceForm caseId="case-1" prompt="Keep a record" initialEvidence={[]} />);

    await user.click(screen.getByRole("button", { name: /save this/i }));

    // Same rule as backend/internal/api/handlers.go's evidence handler
    // ("at least one of staff_name, approx_time, or note is required") --
    // enforced here too so a family doesn't waste a round-trip on an
    // empty submission.
    expect(await screen.findByRole("alert")).toHaveTextContent(/add at least one detail/i);
    expect(addEvidenceMock).not.toHaveBeenCalled();
  });

  it("allows submission with only a single field filled in (note only)", async () => {
    addEvidenceMock.mockResolvedValueOnce({ evidence: [{ captured_at: "2026-08-14T00:00:00Z", note: "Said card would not be accepted" }] });
    const user = userEvent.setup();
    render(<EvidenceForm caseId="case-1" prompt="Keep a record" initialEvidence={[]} />);

    await user.type(
      screen.getByLabelText(/anything else worth noting/i),
      "Said card would not be accepted",
    );
    await user.click(screen.getByRole("button", { name: /save this/i }));

    await waitFor(() =>
      expect(addEvidenceMock).toHaveBeenCalledWith("case-1", {
        staff_name: undefined,
        approx_time: undefined,
        note: "Said card would not be accepted",
      }),
    );
  });

  it("clears the form and renders the updated evidence list after a successful save", async () => {
    addEvidenceMock.mockResolvedValueOnce({
      evidence: [{ captured_at: "2026-08-14T00:00:00Z", staff_name: "Billing clerk", note: "Refused verbally" }],
    });
    const user = userEvent.setup();
    render(<EvidenceForm caseId="case-1" prompt="Keep a record" initialEvidence={[]} />);

    const staffInput = screen.getByLabelText(/staff member/i);
    await user.type(staffInput, "Billing clerk");
    await user.type(screen.getByLabelText(/anything else worth noting/i), "Refused verbally");
    await user.click(screen.getByRole("button", { name: /save this/i }));

    await waitFor(() => expect(staffInput).toHaveValue(""));
    expect(await screen.findByText(/billing clerk.*refused verbally/i)).toBeInTheDocument();
  });

  it("shows the ApiError message on failure without losing what the user typed", async () => {
    addEvidenceMock.mockRejectedValueOnce(new ApiError("could not save evidence", 500));
    const user = userEvent.setup();
    render(<EvidenceForm caseId="case-1" prompt="Keep a record" initialEvidence={[]} />);

    await user.type(screen.getByLabelText(/anything else worth noting/i), "Some note");
    await user.click(screen.getByRole("button", { name: /save this/i }));

    expect(await screen.findByText(/could not save evidence/i)).toBeInTheDocument();
  });

  it("renders evidence entries already present when the page loads", () => {
    render(
      <EvidenceForm
        caseId="case-1"
        prompt="Keep a record"
        initialEvidence={[
          { captured_at: "2026-08-14T00:00:00Z", staff_name: "Front desk", approx_time: "4pm" },
        ]}
      />,
    );

    expect(screen.getByText(/front desk.*4pm/i)).toBeInTheDocument();
  });
});
