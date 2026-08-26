// Types mirroring backend/internal/api/dto.go exactly. Kept as a single
// hand-written file rather than code-generated from the Go structs —
// this API surface is small and stable enough that a codegen step would
// be more moving parts than it's worth (see ARCHITECTURE.md's general
// bias toward fewer dependencies and simpler build steps throughout this
// project).

export type Outcome = "green" | "amber" | "red" | "mixed" | "handoff";

export interface EvidenceEntry {
  captured_at: string;
  staff_name?: string;
  approx_time?: string;
  note?: string;
}

export interface CaseResponse {
  id: string;
  outcome: Outcome;
  citation?: string;
  /** The family's own words, verbatim, as originally submitted. Backend
   * has always stored this (store.CaseRecord.FamilyDescriptionRaw); this
   * field just exposes it, so a case reopened later can still show the
   * family their own story back rather than only derived output. */
  description?: string;
  care_first_message: string;
  /** Always present — the backend guarantees this the same structural
   * way it guarantees care_first_message (see response/types.go). Not
   * marked required here defensively, in case an older cached response
   * object is ever read without it. */
  disclaimer?: string;
  tier_message: string;
  action_steps?: string[];
  complaint_text?: string;
  hospital_script?: string;
  handoff_summary?: string;
  evidence_prompt?: string;
  evidence?: EvidenceEntry[];
}

export interface ErrorResponse {
  error: string;
  fallback_guidance?: string;
}

export interface AddEvidenceRequest {
  staff_name?: string;
  approx_time?: string;
  note?: string;
}

// ---------------------------------------------------------------------
// Local-only types (lib/caseHistory.ts)
// ---------------------------------------------------------------------
// Everything below this line describes data that lives in this browser
// only, never on the backend. This product has no accounts and no
// login (see ARCHITECTURE.md / the "100% Free & Private, no login"
// pillar on the homepage) — that's a deliberate product choice, not an
// oversight — so there is no server-side concept of "all of my cases."
// A dashboard that lists more than one case has to get that list from
// somewhere, and the honest answer for a product with no accounts is:
// this specific browser's own history of cases it created. See
// lib/caseHistory.ts's file header for the full reasoning.

/** Escalation/complaint status the family sets themselves. There is no
 * backend integration that submits or tracks this for them — the
 * complaint text is a draft the family submits by hand through the
 * official Ayushman App (see ARCHITECTURE.md's "what was deliberately
 * not built"), so "submitted" here means the family told us they did
 * it, not that we verified it. */
export type ComplaintStatus =
  | "not_started"
  | "draft"
  | "ready_to_submit"
  | "submitted"
  | "awaiting_response"
  | "resolved";

/** One case as tracked locally in this browser. Everything except the
 * status/timestamp bookkeeping is a snapshot of real data this browser
 * received from the real backend at some point — nothing here is
 * fabricated, it's cached so a dashboard can list many cases without
 * re-fetching each just to render a list. */
export interface LocalCaseRecord {
  id: string;
  description: string;
  outcome: Outcome;
  tierMessage: string;
  createdAt: string;
  lastViewedAt: string;
  complaintStatus: ComplaintStatus;
  /** Set once this browser has re-fetched the case from the real
   * backend since it was cached, confirming it still exists. */
  lastVerifiedAt?: string;
}
