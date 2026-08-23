"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { getCase, ApiError } from "@/lib/api";
import type { CaseResponse } from "@/lib/types";
import { Header } from "@/app/components/Header";
import { CareFirstBanner } from "@/app/components/CareFirstBanner";
import { TierPanel } from "@/app/components/TierPanel";
import { ActionSteps } from "@/app/components/ActionSteps";
import { CopyableTextBox } from "@/app/components/CopyableTextBox";
import { EvidenceForm } from "@/app/components/EvidenceForm";
import { HandoffPanel } from "@/app/components/HandoffPanel";
import { CaseDocumentPanel } from "@/app/components/CaseDocumentPanel";

export default function CasePage() {
  const params = useParams<{ id: string }>();
  const [caseData, setCaseData] = useState<CaseResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    getCase(params.id)
      .then((data) => {
        if (!cancelled) setCaseData(data);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof ApiError ? err.message : "Could not load this case.");
      })
      .finally(() => {
        if (cancelled) return;
        setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [params.id]);

  return (
    <div className="min-h-screen flex flex-col">
      <Header />
      <main className="mx-auto w-full max-w-2xl flex-1 space-y-6 px-4 py-8 sm:px-6 sm:py-12 animate-enter">
        {loading && (
          <div className="space-y-5" role="status">
            <div className="flex items-center justify-center gap-2.5 text-center text-sm font-bold text-teal-900/70 py-2">
              <svg
                className="h-4 w-4 animate-spin text-teal-700"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                />
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
              <span>Looking into your situation…</span>
            </div>
            {/* Shimmer skeleton cards to eliminate layout shift */}
            <div className="h-24 rounded-2xl skeleton-shimmer border border-sand-200/50" />
            <div className="h-48 rounded-2xl skeleton-shimmer border border-sand-200/50" />
            <div className="h-32 rounded-2xl skeleton-shimmer border border-sand-200/50" />
          </div>
        )}

        {error && !loading && (
          <div 
            role="alert" 
            className="rounded-2xl border-2 border-tier-red-border bg-tier-red-bg/95 p-5 sm:p-7 shadow-sm backdrop-blur-md animate-enter"
          >
            <div className="flex items-start gap-3">
              <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-tier-red-strong text-white font-bold text-sm">
                !
              </span>
              <div>
                <p className="font-bold text-lg text-tier-red-text">{error}</p>
                <p className="mt-2 text-sm text-tier-red-text/90 leading-relaxed">
                  If this doesn&rsquo;t resolve, call the PMJAY helpline at{" "}
                  <a href="tel:14555" className="font-bold underline decoration-tier-red-strong">
                    14555
                  </a>
                  .
                </p>
              </div>
            </div>
          </div>
        )}

        {caseData && !loading && (
          <>
            <CareFirstBanner message={caseData.care_first_message} />

            <TierPanel
              outcome={caseData.outcome}
              message={caseData.tier_message}
              citation={caseData.citation}
            />

            {caseData.outcome === "handoff" && caseData.handoff_summary && (
              <HandoffPanel summary={caseData.handoff_summary} />
            )}

            <CaseDocumentPanel caseId={caseData.id} />

            {caseData.action_steps && caseData.action_steps.length > 0 && (
              <ActionSteps steps={caseData.action_steps} />
            )}

            {caseData.hospital_script && (
              <CopyableTextBox
                title="Exact words to use at the desk"
                text={caseData.hospital_script}
              />
            )}

            {caseData.complaint_text && (
              <CopyableTextBox
                title="Draft complaint, ready to review"
                helperText="Submit this yourself through the official Ayushman App when you're ready — this tool can't submit it for you."
                text={caseData.complaint_text}
              />
            )}

            {caseData.evidence_prompt && (
              <EvidenceForm
                caseId={caseData.id}
                prompt={caseData.evidence_prompt}
                initialEvidence={caseData.evidence ?? []}
              />
            )}
          </>
        )}
      </main>
    </div>
  );
}
