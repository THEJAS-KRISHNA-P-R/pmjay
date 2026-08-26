import type { CaseResponse, ComplaintStatus, LocalCaseRecord, Outcome } from "./types";

/**
 * caseHistory.ts — this browser's own memory of the real cases it has
 * created, so a Dashboard can show "your cases" without a login.
 *
 * Why this exists: the backend has no accounts and no login by design
 * (ARCHITECTURE.md — a case is reachable by its unguessable ID alone,
 * on purpose, so nobody has to create an account at a hospital billing
 * desk under stress). That means the backend has no server-side concept
 * of "all of one family's cases" to hand back to a dashboard — there is
 * no user to attach a list to. The honest options were: build real
 * accounts (a much bigger, and arguably wrong, product decision to make
 * unilaterally on a redesign pass), or keep a local, per-browser record
 * of which real case IDs this browser has created, refreshed against
 * the real backend whenever it's convenient. This is the second option.
 *
 * What's real and what isn't: every record here started as an actual
 * CaseResponse from the actual backend — the id, outcome, and messages
 * are real data, just cached client-side. The only things invented
 * locally are bookkeeping (when it was last viewed) and the complaint
 * status, which the family sets by hand because nothing in this system
 * submits a complaint on their behalf (ARCHITECTURE.md again — the
 * family still submits the draft themselves via the official Ayushman
 * App). "Submitted" in this file means "the family told us they did
 * it," never "we verified it."
 *
 * Scope and limits, stated plainly rather than left implicit: this list
 * lives in one browser's localStorage. A different device, a cleared
 * browser, or private/incognito mode won't see it — the case itself
 * still exists on the backend (reachable directly at /cases/:id), just
 * not listed here. The Dashboard and Settings pages should say this
 * plainly rather than let it be a silent surprise.
 */

const STORAGE_KEY = "pmjay-advocate:case-history:v1";
const MAX_TRACKED_CASES = 200;

function isBrowser(): boolean {
  return typeof window !== "undefined" && typeof window.localStorage !== "undefined";
}

function readAll(): LocalCaseRecord[] {
  if (!isBrowser()) return [];
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed as LocalCaseRecord[];
  } catch {
    // Corrupt JSON, storage disabled, or a private-browsing quota wall —
    // treat exactly like "no history yet" rather than throwing and
    // breaking the Dashboard for a family who's already having a hard
    // enough day.
    return [];
  }
}

function writeAll(records: LocalCaseRecord[]): void {
  if (!isBrowser()) return;
  try {
    // Keep the most recently viewed cases if this ever grows past the
    // cap — a family working one real crisis is very unlikely to hit
    // 200 tracked cases, but a stable cap keeps localStorage bounded
    // regardless.
    const trimmed = [...records]
      .sort((a, b) => b.lastViewedAt.localeCompare(a.lastViewedAt))
      .slice(0, MAX_TRACKED_CASES);
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(trimmed));
  } catch {
    // Best-effort. A family who can't save history locally can still
    // use every real feature — they just won't see this case on the
    // Dashboard next time. Never let this throw into a form submission.
  }
}

/** All tracked cases, most recently viewed first. */
export function listCaseHistory(): LocalCaseRecord[] {
  return readAll().sort((a, b) => b.lastViewedAt.localeCompare(a.lastViewedAt));
}

export function getCaseHistoryEntry(id: string): LocalCaseRecord | undefined {
  return readAll().find((r) => r.id === id);
}

export function hasCaseHistory(): boolean {
  return readAll().length > 0;
}

/**
 * Upsert a case from a real CaseResponse. Call this right after a
 * successful createCase() or getCase() — never with invented data.
 * Preserves the existing complaintStatus and createdAt on repeat saves
 * so re-visiting a case doesn't reset progress the family already
 * tracked.
 */
export function saveCaseToHistory(caseData: CaseResponse): LocalCaseRecord {
  const all = readAll();
  const now = new Date().toISOString();
  const existing = all.find((r) => r.id === caseData.id);

  const updated: LocalCaseRecord = {
    id: caseData.id,
    description: caseData.description ?? existing?.description ?? "",
    outcome: caseData.outcome,
    tierMessage: caseData.tier_message,
    createdAt: existing?.createdAt ?? now,
    lastViewedAt: now,
    complaintStatus: existing?.complaintStatus ?? defaultComplaintStatus(caseData),
    lastVerifiedAt: now,
  };

  const next = [updated, ...all.filter((r) => r.id !== caseData.id)];
  writeAll(next);
  return updated;
}

/** A sensible starting status inferred only from what the backend
 * actually sent for a brand-new case — never a guess about what the
 * family has done. Most tiers start "not_started" (nothing to submit
 * yet, or nothing the tool drafted); a tier that came with a
 * ready-to-copy complaint starts at "draft" since that's literally what
 * it is the moment it's generated. */
function defaultComplaintStatus(caseData: CaseResponse): ComplaintStatus {
  return caseData.complaint_text ? "draft" : "not_started";
}

export function setComplaintStatus(id: string, status: ComplaintStatus): LocalCaseRecord | undefined {
  const all = readAll();
  const idx = all.findIndex((r) => r.id === id);
  const existing = all[idx];
  if (idx === -1 || !existing) return undefined;
  const updated: LocalCaseRecord = { ...existing, complaintStatus: status };
  all[idx] = updated;
  writeAll(all);
  return updated;
}

export function touchLastViewed(id: string): void {
  const all = readAll();
  const idx = all.findIndex((r) => r.id === id);
  const existing = all[idx];
  if (idx === -1 || !existing) return;
  all[idx] = { ...existing, lastViewedAt: new Date().toISOString() };
  writeAll(all);
}

export function removeCaseFromHistory(id: string): void {
  writeAll(readAll().filter((r) => r.id !== id));
}

/** Used by the Settings page's "forget this browser's case list"
 * action. Never deletes anything on the backend — a case ID typed or
 * bookmarked elsewhere still works exactly as before. */
export function clearAllHistory(): void {
  if (!isBrowser()) return;
  try {
    window.localStorage.removeItem(STORAGE_KEY);
  } catch {
    // Nothing to do — see writeAll's comment above.
  }
}

const OUTCOMES_NEEDING_ATTENTION: ReadonlySet<Outcome> = new Set(["handoff", "amber", "mixed"]);

/**
 * True when a case is something the Dashboard should surface as
 * "needs attention" rather than quietly resolved — either the tier
 * itself calls for another step (handoff/amber/mixed), or the family
 * has a complaint drafted but hasn't told us they've acted on it yet.
 * "green" and "red" are both, in their own way, settled answers with
 * nothing pending; this deliberately does not second-guess that.
 */
export function needsAttention(record: LocalCaseRecord): boolean {
  if (OUTCOMES_NEEDING_ATTENTION.has(record.outcome)) return true;
  return record.complaintStatus === "draft" || record.complaintStatus === "ready_to_submit";
}
