import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { createCase, getCase, addEvidence, caseDocumentUrl, ApiError } from "./api";

function mockFetchOnce(response: Partial<Response> & { jsonBody?: unknown }) {
  const { jsonBody, ...rest } = response;
  global.fetch = vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    json: async () => jsonBody,
    ...rest,
  } as Response);
}

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("createCase", () => {
  it("POSTs to /api/v1/cases with the description as JSON and returns the parsed body on success", async () => {
    mockFetchOnce({ ok: true, status: 201, jsonBody: { id: "case-1", outcome: "green" } });

    const result = await createCase("A description of what happened");

    expect(fetch).toHaveBeenCalledWith(
      "/api/v1/cases",
      expect.objectContaining({
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ description: "A description of what happened" }),
      }),
    );
    expect(result).toEqual({ id: "case-1", outcome: "green" });
  });

  it("on a JSON error response, throws ApiError carrying both the message and fallback_guidance — the frontend half of the documented 502 contract", async () => {
    mockFetchOnce({
      ok: false,
      status: 502,
      jsonBody: {
        error: "the system could not process this request right now",
        fallback_guidance: "Call the PMJAY helpline directly at 14555.",
      },
    });

    await expect(createCase("A description")).rejects.toMatchObject({
      name: "ApiError",
      message: "the system could not process this request right now",
      status: 502,
      fallbackGuidance: "Call the PMJAY helpline directly at 14555.",
    });
  });

  it("on a non-JSON error body (e.g. a proxy-level 502 HTML page), falls back to a generic message instead of throwing a confusing parse error", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 502,
      json: async () => {
        throw new SyntaxError("Unexpected token < in JSON");
      },
    } as unknown as Response);

    await expect(createCase("A description")).rejects.toMatchObject({
      name: "ApiError",
      status: 502,
      message: "Request failed with status 502",
    });
  });
});

describe("getCase", () => {
  it("GETs /api/v1/cases/{id} with the id URL-encoded", async () => {
    mockFetchOnce({ ok: true, status: 200, jsonBody: { id: "case-1" } });

    await getCase("case with spaces");

    expect(fetch).toHaveBeenCalledWith("/api/v1/cases/case%20with%20spaces");
  });

  it("on 404, throws ApiError with the not-found message", async () => {
    mockFetchOnce({ ok: false, status: 404, jsonBody: { error: "case not found" } });

    await expect(getCase("missing-id")).rejects.toMatchObject({
      status: 404,
      message: "case not found",
    });
  });
});

describe("addEvidence", () => {
  it("POSTs to /api/v1/cases/{id}/evidence with the entry as JSON", async () => {
    mockFetchOnce({ ok: true, status: 200, jsonBody: { id: "case-1", evidence: [] } });

    await addEvidence("case-1", { note: "Said card would not be accepted" });

    expect(fetch).toHaveBeenCalledWith(
      "/api/v1/cases/case-1/evidence",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ note: "Said card would not be accepted" }),
      }),
    );
  });
});

describe("ApiError", () => {
  it("is a real Error subclass, so existing try/catch and error-boundary handling still works", () => {
    const err = new ApiError("something failed", 500);
    expect(err).toBeInstanceOf(Error);
    expect(err.name).toBe("ApiError");
  });

  it("fallbackGuidance is undefined, not a crash, when the backend didn't send one", () => {
    const err = new ApiError("plain failure", 500);
    expect(err.fallbackGuidance).toBeUndefined();
  });
});

describe("caseDocumentUrl", () => {
  it("builds the /v1/cases/{id}/document URL with the id URL-encoded", () => {
    expect(caseDocumentUrl("case with spaces")).toBe("/api/v1/cases/case%20with%20spaces/document");
  });

  it("does not call fetch — it's a plain URL builder for a browser navigation, not a request wrapper", () => {
    caseDocumentUrl("case-1");
    expect(fetch).not.toHaveBeenCalled();
  });
});
