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
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [params.id]);

  return (
    <div className="min-h-screen">
      <Header />
      <main className="mx-auto max-w-2xl space-y-5 px-4 py-8 sm:px-6 sm:py-12">
        {loading && (
          <p className="text-center text-sand-900/60" role="status">
            Looking into your situation…
          </p>
        )}

        {error && !loading && (
          <div role="alert" className="rounded-xl border border-tier-red-border bg-tier-red-bg p-5">
            <p className="font-bold text-tier-red-text">{error}</p>
            <p className="mt-2 text-sm text-tier-red-text/90">
              If this doesn&rsquo;t resolve, call the PMJAY helpline at{" "}
              <a href="tel:14555" className="font-bold underline">
                14555
              </a>
              .
            </p>
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
