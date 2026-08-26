"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { getCase, ApiError } from "@/lib/api";
import { saveCaseToHistory, touchLastViewed, getCaseHistoryEntry } from "@/lib/caseHistory";
import type { CaseResponse } from "@/lib/types";
import { AppShell } from "@/app/components/AppShell";
import { CareFirstBanner } from "@/app/components/CareFirstBanner";
import { DisclaimerNote } from "@/app/components/DisclaimerNote";
import { TierPanel } from "@/app/components/TierPanel";
import { ActionSteps } from "@/app/components/ActionSteps";
import { CopyableTextBox } from "@/app/components/CopyableTextBox";
import { EvidenceForm } from "@/app/components/EvidenceForm";
import { HandoffPanel } from "@/app/components/HandoffPanel";
import { CaseDocumentPanel } from "@/app/components/CaseDocumentPanel";
import { ComplaintStatusTracker } from "@/app/components/ComplaintStatusTracker";
import { IconSpinner, IconAlertTriangle, IconMessageText } from "@/app/components/icons";

interface Section {
  id: string;
  label: string;
}

export default function CaseWorkspacePage() {
  const params = useParams<{ id: string }>();
  const [caseData, setCaseData] = useState<CaseResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeSection, setActiveSection] = useState<string>("overview");

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    getCase(params.id)
      .then((data) => {
        if (cancelled) return;
        setCaseData(data);
        saveCaseToHistory(data);
        touchLastViewed(data.id);
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

  const sections = useMemo<Section[]>(() => {
    if (!caseData) return [];
    const s: Section[] = [{ id: "overview", label: "Overview" }];
    if (caseData.description) s.push({ id: "your-story", label: "Your story" });
    if (caseData.action_steps && caseData.action_steps.length > 0) s.push({ id: "next-steps", label: "Next steps" });
    s.push({ id: "documents", label: "Documents & letters" });
    if (caseData.complaint_text) s.push({ id: "escalation", label: "Track your complaint" });
    if (caseData.evidence_prompt) s.push({ id: "evidence", label: "Evidence" });
    return s;
  }, [caseData]);

  // Track active section on scroll for sidebar highlight
  useEffect(() => {
    if (
      typeof window === "undefined" ||
      typeof IntersectionObserver === "undefined" ||
      sections.length === 0
    ) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setActiveSection(entry.target.id);
          }
        }
      },
      { rootMargin: "-20% 0% -60% 0%" }
    );

    for (const section of sections) {
      const el = document.getElementById(section.id);
      if (el) observer.observe(el);
    }

    return () => observer.disconnect();
  }, [sections]);

  return (
    <AppShell>
      {loading && (
        <div className="space-y-5" role="status">
          <div className="flex items-center justify-center gap-2.5 text-center text-sm font-bold text-ink-800 py-2">
            <IconSpinner className="h-4 w-4 animate-spin text-ink-700" />
            <span>Looking into your situation…</span>
          </div>
          <div className="h-24 rounded-2xl skeleton-shimmer border border-sand-200/50" />
          <div className="h-48 rounded-2xl skeleton-shimmer border border-sand-200/50" />
          <div className="h-32 rounded-2xl skeleton-shimmer border border-sand-200/50" />
        </div>
      )}

      {error && !loading && (
        <div
          role="alert"
          className="rounded-2xl border-2 border-tier-red-border bg-tier-red-bg p-5 sm:p-7 shadow-sm animate-fade-in-up"
        >
          <div className="flex items-start gap-3">
            <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-tier-red-icon text-tier-red-text">
              <IconAlertTriangle className="h-5 w-5" />
            </span>
            <div>
              <p className="font-bold text-lg text-tier-red-text">{error}</p>
              <p className="mt-2 text-sm text-tier-red-text leading-relaxed opacity-90">
                If this doesn&rsquo;t resolve, call the PMJAY helpline at{" "}
                <a href="tel:14555" className="font-bold underline decoration-tier-red-border">
                  14555
                </a>
                .
              </p>
            </div>
          </div>
        </div>
      )}

      {caseData && !loading && (
        <div className="relative">
          {/* Desktop in-page fixed clean table of contents — 100% FIXED from pixel 0, ZERO initial shift */}
          <aside className="hidden lg:block fixed top-[88px] w-48 xl:w-52 z-20">
            <nav aria-label="Case sections" className="space-y-1">
              {sections.map((s) => {
                const isCurrent = activeSection === s.id;
                return (
                  <a
                    key={s.id}
                    href={`#${s.id}`}
                    onClick={() => setActiveSection(s.id)}
                    className={`block px-3.5 py-2 rounded-xl text-sm font-semibold transition-colors duration-150 border ${
                      isCurrent
                        ? "bg-sand-100 text-sand-950 border-sand-200 font-bold"
                        : "text-sand-600 hover:text-sand-900 hover:bg-sand-100 border-transparent"
                    }`}
                  >
                    {s.label}
                  </a>
                );
              })}
            </nav>
          </aside>

          {/* Mobile: 100% FIXED horizontal pill navigation directly under top header — ZERO shift, zero clipping */}
          <nav
            aria-label="Case sections"
            className="lg:hidden fixed top-[60px] inset-x-0 z-40 h-[48px] flex items-center gap-2 overflow-x-auto px-4 sm:px-8 bg-sand-50/95 backdrop-blur-md border-b border-sand-200/80 shadow-xs scrollbar-none"
          >
            {sections.map((s) => {
              const isCurrent = activeSection === s.id;
              return (
                <a
                  key={s.id}
                  href={`#${s.id}`}
                  onClick={() => setActiveSection(s.id)}
                  className={`shrink-0 rounded-full px-3.5 py-1.5 text-xs font-semibold transition-colors ${
                    isCurrent
                      ? "bg-emerald-700 text-white shadow-xs"
                      : "border border-sand-200/80 bg-white text-sand-700 hover:bg-sand-100"
                  }`}
                >
                  {s.label}
                </a>
              );
            })}
          </nav>

          {/* Main case content */}
          <div className="min-w-0 w-full pt-8 lg:pt-0 lg:pl-56 xl:pl-60 space-y-6 sm:space-y-8">
            <div className="space-y-3">
              <CareFirstBanner message={caseData.care_first_message} />
              {caseData.disclaimer && <DisclaimerNote text={caseData.disclaimer} />}
            </div>

            <section id="overview" className="scroll-mt-24 space-y-6">
              <TierPanel outcome={caseData.outcome} message={caseData.tier_message} citation={caseData.citation} />
              {caseData.outcome === "handoff" && caseData.handoff_summary && (
                <HandoffPanel summary={caseData.handoff_summary} />
              )}
            </section>

            {caseData.description && (
              <section id="your-story" className="scroll-mt-24 card p-6 sm:p-8 animate-fade-in-up">
                <div className="flex items-center gap-2.5">
                  <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-ink-100 text-ink-800">
                    <IconMessageText className="h-4 w-4" />
                  </div>
                  <h2 className="text-lg sm:text-xl font-bold tracking-tight text-sand-900">Your story</h2>
                </div>
                <p className="mt-2 text-xs sm:text-sm text-sand-600 font-medium">
                  Exactly what you told us (unchanged, in your own words).
                </p>
                <p className="mt-4 whitespace-pre-wrap rounded-2xl bg-sand-50 border border-sand-200/70 p-4 sm:p-5 text-sm sm:text-base leading-relaxed text-sand-800">
                  {caseData.description}
                </p>
              </section>
            )}

            {caseData.action_steps && caseData.action_steps.length > 0 && (
              <section id="next-steps" className="scroll-mt-24">
                <ActionSteps steps={caseData.action_steps} />
              </section>
            )}

            <section id="documents" className="scroll-mt-24 space-y-4">
              <div className="space-y-1 px-1">
                <h2 className="text-lg sm:text-xl font-bold tracking-tight text-sand-900">Documents &amp; letters</h2>
                <p className="text-xs sm:text-sm text-sand-600 font-medium">
                  Everything you can download, print, or copy for the hospital desk or your complaint.
                </p>
              </div>
              <div className="space-y-6 sm:space-y-8">
                <CaseDocumentPanel caseId={caseData.id} />
                {caseData.hospital_script && (
                  <CopyableTextBox title="Exact words to use at the desk" text={caseData.hospital_script} />
                )}
                {caseData.complaint_text && (
                  <CopyableTextBox
                    title="Draft complaint, ready to review"
                    helperText="Submit this yourself through the official Ayushman App when you're ready (this tool can't submit it for you)."
                    text={caseData.complaint_text}
                  />
                )}
              </div>
            </section>

            {caseData.complaint_text && (
              <section id="escalation" className="scroll-mt-24">
                <ComplaintStatusTracker
                  caseId={caseData.id}
                  initialStatus={getCaseHistoryEntry(caseData.id)?.complaintStatus ?? "draft"}
                />
              </section>
            )}

            {caseData.evidence_prompt && (
              <section id="evidence" className="scroll-mt-24">
                <EvidenceForm
                  caseId={caseData.id}
                  prompt={caseData.evidence_prompt}
                  initialEvidence={caseData.evidence ?? []}
                />
              </section>
            )}
          </div>
        </div>
      )}
    </AppShell>
  );
}
