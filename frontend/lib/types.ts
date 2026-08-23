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
  care_first_message: string;
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
