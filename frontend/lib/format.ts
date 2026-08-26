import type { ComplaintStatus, Outcome } from "./types";

/** "3 hours ago", "yesterday", "12 Aug" — deliberately coarse, never a
 * bare timestamp, since the exact minute rarely matters to a family
 * checking on a case and a hand-rolled version avoids adding a date
 * library for one small formatting job. */
export function formatRelativeTime(iso: string): string {
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "";
  const now = Date.now();
  const diffMs = now - then;
  const minute = 60_000;
  const hour = 60 * minute;
  const day = 24 * hour;

  if (diffMs < minute) return "just now";
  if (diffMs < hour) {
    const m = Math.round(diffMs / minute);
    return `${m} minute${m === 1 ? "" : "s"} ago`;
  }
  if (diffMs < day) {
    const h = Math.round(diffMs / hour);
    return `${h} hour${h === 1 ? "" : "s"} ago`;
  }
  if (diffMs < 2 * day) return "yesterday";
  if (diffMs < 6 * day) {
    const d = Math.round(diffMs / day);
    return `${d} days ago`;
  }
  return new Date(then).toLocaleDateString("en-IN", { day: "numeric", month: "short" });
}

export const COMPLAINT_STATUS_LABEL: Record<ComplaintStatus, string> = {
  not_started: "No complaint needed yet",
  draft: "Complaint drafted",
  ready_to_submit: "Ready to submit",
  submitted: "Submitted by you",
  awaiting_response: "Awaiting a response",
  resolved: "Resolved",
};

export const OUTCOME_SHORT_LABEL: Record<Outcome, string> = {
  green: "Covered",
  amber: "Needs a check",
  red: "Not covered",
  mixed: "Part covered",
  handoff: "Needs a person",
};

/** Trims a family's own description to a card-sized snippet without
 * cutting mid-word — used only for list previews; the full text is
 * always shown verbatim on the case workspace itself. */
export function snippet(text: string, maxLength = 140): string {
  const trimmed = text.trim();
  if (trimmed.length <= maxLength) return trimmed;
  const cut = trimmed.slice(0, maxLength);
  const lastSpace = cut.lastIndexOf(" ");
  return `${cut.slice(0, lastSpace > 40 ? lastSpace : maxLength)}…`;
}
