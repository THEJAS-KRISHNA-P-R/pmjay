import type { CaseResponse, ErrorResponse, AddEvidenceRequest } from "./types";

// Resolved at build time for a static deploy target (e.g. Vercel), or at
// request time in the browser for a same-origin reverse-proxy setup —
// see docs/DEPLOYMENT.md for both options. Defaults to same-origin /api
// so local development behind a proxy works with zero configuration.
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "/api";

/**
 * ApiError carries the backend's fallback_guidance through to the UI
 * layer even on failure. This matters specifically because of Section
 * 10's care-first rule: an infrastructure failure must not leave a
 * family with nothing actionable, so the backend always includes a
 * fallback message on error responses (the care-first text plus the
 * PMJAY helpline number) — see backend/internal/api/handlers.go's
 * handleIntake. Throwing this custom error type, rather than a bare
 * Error, is what lets components render that fallback rather than
 * discarding it.
 */
export class ApiError extends Error {
  fallbackGuidance?: string;
  status: number;

  constructor(message: string, status: number, fallbackGuidance?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.fallbackGuidance = fallbackGuidance;
  }
}

export class RateLimitError extends Error {
  retryAfterSeconds: number;

  constructor(message: string, retryAfterSeconds: number) {
    super(message);
    this.name = "RateLimitError";
    this.retryAfterSeconds = retryAfterSeconds;
  }
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    if (res.status === 429) {
      const retryAfterStr = res.headers.get("Retry-After");
      const retryAfter = retryAfterStr ? parseInt(retryAfterStr, 10) : 10;
      let body: ErrorResponse = { error: "Too many requests. Please wait." };
      try {
        body = (await res.json()) as ErrorResponse;
      } catch {}
      throw new RateLimitError(body.error, isNaN(retryAfter) ? 10 : retryAfter);
    }

    let body: ErrorResponse = { error: `Request failed with status ${res.status}` };
    try {
      body = (await res.json()) as ErrorResponse;
    } catch {
      // Response body wasn't JSON (e.g. a proxy-level 502 HTML page) —
      // fall back to the generic message set above rather than throwing
      // a confusing parse error on top of the original failure.
    }
    throw new ApiError(body.error, res.status, body.fallback_guidance);
  }
  return res.json() as Promise<T>;
}

export async function createCase(description: string): Promise<CaseResponse> {
  const res = await fetch(`${API_BASE_URL}/v1/cases`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ description }),
  });
  return handleResponse<CaseResponse>(res);
}

export async function getCase(id: string): Promise<CaseResponse> {
  const res = await fetch(`${API_BASE_URL}/v1/cases/${encodeURIComponent(id)}`);
  return handleResponse<CaseResponse>(res);
}

export async function addEvidence(
  id: string,
  entry: AddEvidenceRequest,
): Promise<CaseResponse> {
  const res = await fetch(`${API_BASE_URL}/v1/cases/${encodeURIComponent(id)}/evidence`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(entry),
  });
  return handleResponse<CaseResponse>(res);
}

/**
 * caseDocumentUrl builds the URL for a case's downloadable/printable PDF
 * (backend/internal/api/handlers.go's handleGetCaseDocument). Deliberately
 * a plain URL builder, not a fetch wrapper like the functions above —
 * callers hand this straight to an <a href> for the browser to navigate
 * to directly, so the browser's native PDF viewer (with its own
 * print/save/share controls) handles the response. That also sidesteps
 * CORS entirely: a top-level navigation isn't subject to it the way a
 * script-initiated fetch/XHR would be, so this works unmodified whether
 * API_BASE_URL resolves to a same-origin path or a fully different origin.
 */
export function caseDocumentUrl(id: string): string {
  return `${API_BASE_URL}/v1/cases/${encodeURIComponent(id)}/document`;
}
