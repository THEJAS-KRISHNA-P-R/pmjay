import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, afterEach } from "vitest";
import { CopyableTextBox } from "./CopyableTextBox";

function mockClipboard(writeText = vi.fn().mockResolvedValue(undefined)) {
  Object.defineProperty(navigator, "clipboard", {
    value: { writeText },
    configurable: true,
  });
  return writeText;
}

describe("CopyableTextBox", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("renders the title and the exact text content", () => {
    mockClipboard();
    render(<CopyableTextBox title="Draft complaint" text="Dear Sir/Madam," />);
    expect(screen.getByText("Draft complaint")).toBeInTheDocument();
    expect(screen.getByText("Dear Sir/Madam,")).toBeInTheDocument();
  });

  it("renders helper text only when it's provided", () => {
    mockClipboard();
    const { rerender } = render(
      <CopyableTextBox title="X" text="Y" helperText="Submit this yourself." />,
    );
    expect(screen.getByText("Submit this yourself.")).toBeInTheDocument();

    rerender(<CopyableTextBox title="X" text="Y" />);
    expect(screen.queryByText("Submit this yourself.")).not.toBeInTheDocument();
  });

  it("copies exactly the displayed text to the clipboard on click, nothing more or less", async () => {
    const writeText = mockClipboard();
    render(<CopyableTextBox title="X" text="the exact text to copy" />);
    fireEvent.click(screen.getByRole("button", { name: /copy/i }));
    await waitFor(() => expect(writeText).toHaveBeenCalledWith("the exact text to copy"));
    expect(writeText).toHaveBeenCalledTimes(1);
  });

  it("shows a confirmation state after copying, then reverts to 'Copy' once the window closes", async () => {
    vi.useFakeTimers();
    mockClipboard();
    render(<CopyableTextBox title="X" text="Y" />);
    const button = screen.getByRole("button", { name: /copy/i });

    await act(async () => {
      fireEvent.click(button);
      // Flush the microtask queue for `await navigator.clipboard.writeText`
      // -- unaffected by the faked *timers* below, since Promise
      // resolution isn't a timer.
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(button).toHaveTextContent("Copied ✓");

    act(() => {
      vi.advanceTimersByTime(2500);
    });
    expect(button).toHaveTextContent("Copy");
    expect(button).not.toHaveTextContent("Copied");
  });

  it("degrades gracefully if the Clipboard API rejects, rather than crashing the page", async () => {
    // Source comment explains why this matters: older browsers or a
    // denied permission can make writeText reject. The text is still
    // fully readable/selectable in the box below regardless, so an
    // unhandled rejection here would be a strictly worse outcome than
    // simply not showing the confirmation state.
    const writeText = vi.fn().mockRejectedValue(new Error("denied"));
    mockClipboard(writeText);
    render(<CopyableTextBox title="X" text="Y" />);
    const button = screen.getByRole("button", { name: /copy/i });

    await act(async () => {
      fireEvent.click(button);
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(button).toHaveTextContent("Copy");
    expect(button).not.toHaveTextContent("Copied");
  });

  it("marks the button aria-live=polite so the confirmation state change is announced", () => {
    mockClipboard();
    render(<CopyableTextBox title="X" text="Y" />);
    expect(screen.getByRole("button", { name: /copy/i })).toHaveAttribute(
      "aria-live",
      "polite",
    );
  });
});
