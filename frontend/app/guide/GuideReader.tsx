"use client";

import { useState } from "react";
import { IconCheck, IconX, IconPhone, IconChevronRight } from "../components/icons";

export interface SectionDef {
  id: string;
  label: string;
}

export const SECTIONS: SectionDef[] = [
  { id: "what-is-covered", label: "What PMJAY covers" },
  { id: "your-rights", label: "Your core rights" },
  { id: "not-allowed", label: "What hospitals can't do" },
  { id: "watch-for", label: "Common denial patterns" },
  { id: "what-to-do", label: "What to do at the desk" },
  { id: "escalate", label: "When to escalate" },
];

export function GuideReader() {
  const [activeId, setActiveId] = useState<string>("what-is-covered");

  const currentIndex = SECTIONS.findIndex((s) => s.id === activeId);
  const nextSection = currentIndex < SECTIONS.length - 1 ? SECTIONS[currentIndex + 1] : null;
  const prevSection = currentIndex > 0 ? SECTIONS[currentIndex - 1] : null;

  return (
    <div className="space-y-6 lg:space-y-0 lg:grid lg:grid-cols-[240px_1fr] lg:gap-10 xl:gap-14 items-start">
      {/* Mobile Horizontal Section Tabs — consistent font weight, zero bold jump */}
      <div className="lg:hidden flex gap-2 overflow-x-auto pb-2 scrollbar-none">
        {SECTIONS.map((s) => {
          const isActive = s.id === activeId;
          return (
            <button
              key={s.id}
              type="button"
              onClick={() => setActiveId(s.id)}
              className={`whitespace-nowrap px-4 py-2 rounded-xl text-xs font-semibold transition-colors duration-150 shrink-0 border ${
                isActive
                  ? "bg-sand-100 text-sand-950 border-sand-200"
                  : "bg-transparent text-sand-600 hover:bg-sand-100 hover:text-sand-900 border-transparent"
              }`}
            >
              {s.label}
            </button>
          );
        })}
      </div>

      {/* Desktop Sidebar Navigation — uniform font-semibold, active highlighted pill */}
      <aside className="hidden lg:block">
        <nav aria-label="Guide sections" className="sticky top-24 space-y-1.5">
          {SECTIONS.map((s) => {
            const isActive = s.id === activeId;
            return (
              <button
                key={s.id}
                type="button"
                onClick={() => setActiveId(s.id)}
                className={`w-full text-left rounded-xl px-3.5 py-2.5 text-sm font-semibold transition-colors duration-150 flex items-center justify-between border ${
                  isActive
                    ? "bg-sand-100 text-sand-950 border-sand-200"
                    : "text-sand-600 hover:bg-sand-100 hover:text-sand-900 border-transparent"
                }`}
              >
                <span>{s.label}</span>
                {isActive && <IconChevronRight className="h-3.5 w-3.5 text-sand-700" />}
              </button>
            );
          })}
        </nav>
      </aside>

      {/* Active Section Content Area */}
      <div className="min-w-0 max-w-3xl space-y-4">
        {/* Previous / Next Section Quick Navigation — firmly positioned on TOP for 100% position consistency */}
        <div className="flex items-center justify-between gap-4 pb-1">
          <button
            type="button"
            onClick={() => prevSection && setActiveId(prevSection.id)}
            disabled={!prevSection}
            className={`inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs sm:text-sm font-semibold transition-all ${
              prevSection
                ? "bg-sand-100/90 text-sand-800 hover:bg-sand-200 hover:text-sand-950 active:scale-95 cursor-pointer"
                : "bg-transparent text-sand-400 opacity-25 cursor-not-allowed pointer-events-none"
            }`}
          >
            <span>← Previous Section</span>
          </button>

          <span className="text-xs font-semibold text-sand-500">
            {currentIndex + 1} of {SECTIONS.length}
          </span>

          <button
            type="button"
            onClick={() => nextSection && setActiveId(nextSection.id)}
            disabled={!nextSection}
            className={`inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs sm:text-sm font-semibold transition-all ${
              nextSection
                ? "bg-emerald-50 text-emerald-800 hover:bg-emerald-100 hover:text-emerald-950 border border-emerald-200/70 active:scale-95 cursor-pointer"
                : "bg-transparent text-sand-400 opacity-25 cursor-not-allowed pointer-events-none"
            }`}
          >
            <span>Next Section →</span>
          </button>
        </div>

        {/* Content Card with smooth slide-fade animation */}
        <div key={activeId} className="card p-6 sm:p-9 space-y-6 animate-smooth-fade min-h-[380px] sm:min-h-[360px]">
          <div>
            {activeId === "what-is-covered" && (
              <div className="space-y-4">
                <span className="text-xs font-bold uppercase tracking-wider text-sand-500">
                  Scheme Scope
                </span>
                <h2 className="font-display text-2xl sm:text-3xl font-bold text-sand-900">
                  What PMJAY covers
                </h2>
                <p className="text-sm sm:text-base leading-relaxed text-sand-700 font-medium">
                  The Pradhan Mantri Jan Arogya Yojana (PMJAY, part of Ayushman Bharat) provides cashless,
                  paperless access to secondary and tertiary inpatient care procedures across empanelled hospitals
                  for eligible families.
                </p>
                <p className="text-sm sm:text-base leading-relaxed text-sand-700 font-medium">
                  Each package has an indicative rate and pre-authorisation criteria set by the National Health Authority (NHA).
                  PMJAY Advocate cross-checks your procedure against the 315-package index to verify whether what the
                  hospital is billing for is covered.
                </p>
              </div>
            )}

            {activeId === "your-rights" && (
              <div className="space-y-4">
                <span className="text-xs font-bold uppercase tracking-wider text-emerald-800">
                  Guaranteed Rights
                </span>
                <h2 className="font-display text-2xl sm:text-3xl font-bold text-sand-900">
                  Your core rights at an empanelled hospital
                </h2>
                <ul className="space-y-3.5 pt-2">
                  {[
                    "100% cashless inpatient treatment for covered procedures under your card's package schedule (zero upfront cash deposit).",
                    "Dedicated Ayushman Mitra / PMJAY helpdesk at every empanelled hospital to verify beneficiary eligibility.",
                    "Emergency care cannot be delayed or refused while pre-authorisation or server verification is pending.",
                    "You are legally entitled to a written, specific reason if any procedure or claim is denied (not just a verbal refusal).",
                  ].map((point, i) => (
                    <li key={i} className="flex items-start gap-3 text-sm sm:text-base leading-relaxed text-sand-700 font-medium">
                      <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-emerald-800 mt-0.5">
                        <IconCheck className="h-3.5 w-3.5" />
                      </span>
                      <span>{point}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {activeId === "not-allowed" && (
              <div className="space-y-4">
                <span className="text-xs font-bold uppercase tracking-wider text-tier-red-text">
                  Illegal Practices
                </span>
                <h2 className="font-display text-2xl sm:text-3xl font-bold text-sand-900">
                  What hospitals are not allowed to do
                </h2>
                <ul className="space-y-3.5 pt-2">
                  {[
                    "Demand an advance cash deposit before admitting you for a covered procedure.",
                    "Bill you out-of-pocket for an HBP listed package and falsely claim it is 'not covered' without a formal written rejection.",
                    "Refuse to discharge a patient, or hold patient discharge summaries and original ID cards over a billing dispute.",
                    "Delay emergency care because pre-authorisation is pending or insurance servers are temporarily slow.",
                  ].map((point, i) => (
                    <li key={i} className="flex items-start gap-3 text-sm sm:text-base leading-relaxed text-sand-700 font-medium">
                      <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-tier-red-bg text-tier-red-text mt-0.5">
                        <IconX className="h-3.5 w-3.5" />
                      </span>
                      <span>{point}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {activeId === "watch-for" && (
              <div className="space-y-4">
                <span className="text-xs font-bold uppercase tracking-wider text-amber-800">
                  Denial Red Flags
                </span>
                <h2 className="font-display text-2xl sm:text-3xl font-bold text-sand-900">
                  Common denial patterns worth double-checking
                </h2>
                <div className="grid gap-3.5 sm:grid-cols-2 pt-2">
                  <div className="rounded-xl bg-sand-50/80 border border-sand-100 p-4">
                    <p className="text-sm font-bold text-sand-900">
                      &ldquo;Our portal is down, pay cash for now&rdquo;
                    </p>
                    <p className="mt-1.5 text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                      Technical downtime is not grounds to bill you. Ask them to log the case and proceed with admission.
                    </p>
                  </div>
                  <div className="rounded-xl bg-sand-50/80 border border-sand-100 p-4">
                    <p className="text-sm font-bold text-sand-900">
                      &ldquo;This isn&rsquo;t in the package list&rdquo;
                    </p>
                    <p className="mt-1.5 text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                      Hospitals often use internal clinical terms that differ from the NHA master list. Check the official package match.
                    </p>
                  </div>
                  <div className="rounded-xl bg-sand-50/80 border border-sand-100 p-4">
                    <p className="text-sm font-bold text-sand-900">
                      Bundling covered and uncovered items
                    </p>
                    <p className="mt-1.5 text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                      An uncovered minor add-on must never invalidate or convert the primary covered surgery into a private cash bill.
                    </p>
                  </div>
                  <div className="rounded-xl bg-sand-50/80 border border-sand-100 p-4">
                    <p className="text-sm font-bold text-sand-900">
                      Vague or verbal-only refusal
                    </p>
                    <p className="mt-1.5 text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                      You have the right to request the denial in writing on hospital letterhead with the doctor&rsquo;s or billing desk&rsquo;s stamp.
                    </p>
                  </div>
                </div>
              </div>
            )}

            {activeId === "what-to-do" && (
              <div className="space-y-4">
                <span className="text-xs font-bold uppercase tracking-wider text-sand-500">
                  Action Protocol
                </span>
                <h2 className="font-display text-2xl sm:text-3xl font-bold text-sand-900">
                  What to do at the billing desk, in order
                </h2>
                <ol className="space-y-3 pt-2">
                  {[
                    "Ask calmly for the specific reason for denial, in writing if possible.",
                    "Ask to speak with the hospital's dedicated Ayushman Mitra / PMJAY kiosk staff.",
                    "Describe your situation in PMJAY Advocate to verify against official package rates.",
                    "If wrongly denied, use our generated counter-script to request formal escalation.",
                    "Log staff names and timestamps: this strengthens formal CGRMS grievance filing.",
                    "If care is urgent, accept treatment first and dispute the financial bill afterward.",
                  ].map((step, i) => (
                    <li key={i} className="flex items-start gap-3.5 text-sm sm:text-base leading-relaxed text-sand-700 font-medium">
                      <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-sand-900 text-xs font-bold text-white mt-0.5">
                        {i + 1}
                      </span>
                      <span>{step}</span>
                    </li>
                  ))}
                </ol>
              </div>
            )}

            {activeId === "escalate" && (
              <div className="space-y-4">
                <span className="text-xs font-bold uppercase tracking-wider text-tier-handoff-text">
                  Legal Aid Escalation
                </span>
                <h2 className="font-display text-2xl sm:text-3xl font-bold text-sand-900">
                  When to escalate to free legal aid
                </h2>
                <p className="text-sm sm:text-base leading-relaxed text-sand-700 font-medium">
                  For situations beyond straightforward billing disputes (illegal detention of a patient or documents,
                  repeated refusal to give written reasons, or outright extortion), NALSA (National Legal Services Authority)
                  provides 100% free legal counsel.
                </p>
                <div className="pt-2">
                  <a
                    href="tel:15100"
                    className="tap-target inline-flex items-center gap-2 rounded-xl bg-tier-handoff-text px-5 py-3 text-sm sm:text-base font-bold text-white shadow-xs hover:opacity-90 active:scale-95 transition-all"
                  >
                    <IconPhone className="h-4 w-4" />
                    <span>Call NALSA Helpline: 15100 (Free, Toll-Free)</span>
                  </a>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
