import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { IntakeForm } from "./IntakeForm";
import { ApiError } from "@/lib/api";

const pushMock = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock }),
}));

const createCaseMock = vi.fn();
vi.mock("@/lib/api", async () => {
  const actual = await vi.importActual<typeof import("@/lib/api")>("@/lib/api");
  return {
    ...actual,
    createCase: (...args: unknown[]) => createCaseMock(...args),
  };
});

beforeEach(() => {
  pushMock.mockReset();
  createCaseMock.mockReset();
  window.localStorage.clear();
});

describe("IntakeForm", () => {
  it("blocks submission client-side under the same 5-character threshold the backend enforces, and never calls the API", async () => {
    const user = userEvent.setup();
    render(<IntakeForm />);

    await user.type(screen.getByLabelText(/happening at the hospital/i), "hi");
    await user.click(screen.getByRole("button", { name: /get help now/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      /describe a bit more/i,
    );
    // The whole point of client-side validation duplicating
    // backend/internal/api/handlers.go's 5-character minimum is to save
    // a wasted, paid API round-trip -- this assertion is what actually
    // catches a regression where that duplication silently breaks and
    // every short description starts hitting the network again.
    expect(createCaseMock).not.toHaveBeenCalled();
  });

  it("trims whitespace before checking length, matching backend behavior", async () => {
    const user = userEvent.setup();
    render(<IntakeForm />);

    // Five non-space characters padded with enough whitespace that the
    // *untrimmed* length would clear 5 either way -- what actually
    // matters here is that leading/trailing spaces around a short
    // real description don't fool the check into accepting it.
    await user.type(screen.getByLabelText(/happening at the hospital/i), "   hi   ");
    await user.click(screen.getByRole("button", { name: /get help now/i }));

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(createCaseMock).not.toHaveBeenCalled();
  });

  it("submits a valid description and navigates to the resulting case page", async () => {
    createCaseMock.mockResolvedValueOnce({
      id: "case-123",
      outcome: "green",
      description: "My mother needs her gallbladder removed, hospital says our card won't cover it.",
      care_first_message: "Get treatment first. Dispute the money after. Always.",
      disclaimer: "This is guidance, not a legal or medical ruling.",
      tier_message: "This looks like a covered package.",
    });
    const user = userEvent.setup();
    render(<IntakeForm />);

    const description =
      "My mother needs her gallbladder removed, hospital says our card won't cover it.";
    await user.type(screen.getByLabelText(/happening at the hospital/i), description);
    await user.click(screen.getByRole("button", { name: /get help now/i }));

    // Case workspace now lives at /cases/:id (see app/cases/[id]/page.tsx);
    // /case/:id redirects there for old links rather than hosting the
    // page itself — see app/case/[id]/page.tsx.
    await waitFor(() => expect(pushMock).toHaveBeenCalledWith("/cases/case-123"));
    expect(createCaseMock).toHaveBeenCalledWith(description);
  });

  it("saves the created case to this browser's local history so it appears on the Dashboard", async () => {
    createCaseMock.mockResolvedValueOnce({
      id: "case-456",
      outcome: "amber",
      description: "A description long enough to pass client-side validation.",
      care_first_message: "Get treatment first. Dispute the money after. Always.",
      disclaimer: "This is guidance, not a legal or medical ruling.",
      tier_message: "Needs one more check.",
    });
    const user = userEvent.setup();
    render(<IntakeForm />);

    await user.type(
      screen.getByLabelText(/happening at the hospital/i),
      "A description long enough to pass client-side validation.",
    );
    await user.click(screen.getByRole("button", { name: /get help now/i }));

    await waitFor(() => expect(pushMock).toHaveBeenCalled());
    const stored = JSON.parse(window.localStorage.getItem("pmjay-advocate:case-history:v1") ?? "[]");
    expect(stored).toHaveLength(1);
    expect(stored[0]).toMatchObject({ id: "case-456", outcome: "amber" });
  });

  it("on an ApiError with fallback guidance, shows both the error and the fallback text — never silently drops the fallback", async () => {
    createCaseMock.mockRejectedValueOnce(
      new ApiError(
        "the system could not process this request right now",
        502,
        "Call the PMJAY helpline directly at 14555.",
      ),
    );
    const user = userEvent.setup();
    render(<IntakeForm />);

    await user.type(
      screen.getByLabelText(/happening at the hospital/i),
      "A description long enough to pass client-side validation.",
    );
    await user.click(screen.getByRole("button", { name: /get help now/i }));

    // This is the frontend half of docs/API.md's documented 502 +
    // fallback_guidance contract -- if a future edit reads err.message
    // but forgets err.fallbackGuidance, this is the test that catches it.
    expect(await screen.findByText(/could not process this request/i)).toBeInTheDocument();
    expect(await screen.findByText(/call the pmjay helpline directly at 14555/i)).toBeInTheDocument();
  });

  it("on a non-ApiError failure, shows a generic message rather than crashing or showing nothing", async () => {
    createCaseMock.mockRejectedValueOnce(new TypeError("network failure"));
    const user = userEvent.setup();
    render(<IntakeForm />);

    await user.type(
      screen.getByLabelText(/happening at the hospital/i),
      "A description long enough to pass client-side validation.",
    );
    await user.click(screen.getByRole("button", { name: /get help now/i }));

    expect(await screen.findByText(/something went wrong/i)).toBeInTheDocument();
  });

  it("clicking an example prompt fills the textarea with that exact prompt", async () => {
    const user = userEvent.setup();
    render(<IntakeForm />);

    const exampleButton = screen.getByRole("button", {
      name: /hospital says our card won't cover it/i,
    });
    await user.click(exampleButton);

    const textarea = screen.getByLabelText(/happening at the hospital/i);
    expect(textarea).toHaveValue(
      "My mother needs gallbladder surgery, hospital says our card won't cover it",
    );
  });
});
