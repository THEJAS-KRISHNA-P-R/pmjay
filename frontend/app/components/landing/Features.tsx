"use client";

import { useState } from "react";

interface Feature {
  id: string;
  title: string;
  shortDesc: string;
  fullDesc: string;
  bullets: string[];
}

const features: Feature[] = [
  {
    id: "hbp-database",
    title: "1,949+ HBP Package Index",
    shortDesc: "Complete official National Health Authority benefit schedule with rates.",
    fullDesc: "Includes all 2022 NHA Health Benefit Packages across surgical, medical, and pediatric specialties with official indicative rates and ceiling definitions.",
    bullets: [
      "Exact procedure codes and ceiling tariffs",
      "Specialty-wise categorization (Cardiology, General Surgery, Ortho)",
      "Standard national package inclusions and duration rules",
    ],
  },
  {
    id: "deterministic-tiering",
    title: "Deterministic Decision Engine",
    shortDesc: "Pure rule-based logic to guarantee zero hallucinations on rights.",
    fullDesc: "The LLM is restricted to extracting clinical details. Final tiering (Green, Amber, Red, Mixed, Handoff) is executed by deterministic Go logic to ensure safety.",
    bullets: [
      "Strict separation between extraction and decision logic",
      "Confidence-gap fallbacks for ambiguous inputs",
      "Guaranteed consistent outcomes for identical clinical facts",
    ],
  },
  {
    id: "counter-scripts",
    title: "Hospital Counter-Scripts",
    shortDesc: "Exact polite, legally grounded words to speak at the billing desk.",
    fullDesc: "Hospital billing desks often rely on families being intimidated by medical jargon. Our counter-scripts give families the exact words to ask for written reasons.",
    bullets: [
      "Calm, non-confrontational phrasing for families under stress",
      "Requests written refusal reasons with official scheme citation",
      "Clear guidance on holding empanelled hospitals accountable",
    ],
  },
  {
    id: "pdf-document",
    title: "Instant Case PDF Generator",
    shortDesc: "Zero-dependency %PDF-1.4 document generated on demand.",
    fullDesc: "Transforms your entire evaluation, package details, hospital script, and evidence log into a clean printable PDF to physically hand to hospital administrators.",
    bullets: [
      "Instant native download with zero browser dependencies",
      "Structured layout with emergency priority banner",
      "Ready for official hospital grievance submission",
    ],
  },
  {
    id: "evidence-logging",
    title: "Timestamped Evidence Log",
    shortDesc: "Keep a real-time record of billing desk interactions.",
    fullDesc: "Record staff names, refusal timestamps, and verbal statements on the spot. All logged entries are atomically saved and included in the case dossier.",
    bullets: [
      "Staff member name and desk identifier tracking",
      "Approximate timestamp recording for audit trails",
      "Permanent record for formal CGRMS complaint filings",
    ],
  },
  {
    id: "nalsa-legal-aid",
    title: "NALSA Free Legal Aid Link",
    shortDesc: "Direct escalation to toll-free Para Legal Volunteers on 15100.",
    fullDesc: "For complex extortion, illegal patient detention, or refusal to discharge, the advocate creates a structured brief for free government legal aid attorneys.",
    bullets: [
      "Toll-free NALSA hotline integration (15100)",
      "Pre-formatted case summary so families don't repeat trauma",
      "100% free legal representation for eligible citizens",
    ],
  },
];

export function Features() {
  const [activeFeature, setActiveFeature] = useState<Feature | null>(null);

  return (
    <section aria-labelledby="features-title" className="space-y-6">
      <div className="text-center space-y-2">
        <p className="text-xs font-bold uppercase tracking-wider text-teal-800">
          Core Engine Capabilities
        </p>
        <h2 id="features-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-teal-950">
          Everything You Need at the Hospital Desk
        </h2>
        <p className="text-sm sm:text-base text-sand-900/70 max-w-xl mx-auto font-medium">
          Engineered for speed, privacy, and clinical-legal accuracy.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 pt-2">
        {features.map((f) => (
          <div
            key={f.id}
            onClick={() => setActiveFeature(f)}
            className="glass-panel p-6 flex flex-col justify-between cursor-pointer transition-all hover:-translate-y-1 hover:shadow-md"
          >
            <div className="space-y-2.5">
              <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-teal-100 text-teal-800 text-xs font-bold shadow-xs">
                ✦
              </div>
              <h3 className="font-display text-base sm:text-lg font-bold text-teal-950">
                {f.title}
              </h3>
              <p className="text-xs sm:text-sm text-sand-900/75 leading-relaxed font-medium">
                {f.shortDesc}
              </p>
            </div>
            <div className="pt-4 flex items-center text-xs font-bold text-teal-800 gap-1">
              <span>Explore details</span>
              <span>→</span>
            </div>
          </div>
        ))}
      </div>

      {/* Feature Detail Modal */}
      {activeFeature && (
        <div 
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm animate-enter"
          onClick={() => setActiveFeature(null)}
        >
          <div 
            className="glass-panel-elevated max-w-lg w-full p-6 sm:p-8 space-y-5 bg-white/95"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-teal-800 text-white font-bold text-sm shadow-xs">
                ✦
              </span>
              <button
                type="button"
                onClick={() => setActiveFeature(null)}
                className="h-8 w-8 rounded-full bg-sand-100 text-sand-900/70 hover:text-sand-900 flex items-center justify-center text-sm font-bold transition"
              >
                ✕
              </button>
            </div>

            <div>
              <h3 className="font-display text-xl sm:text-2xl font-bold text-teal-950">
                {activeFeature.title}
              </h3>
              <p className="mt-2 text-sm text-sand-900/80 leading-relaxed font-medium">
                {activeFeature.fullDesc}
              </p>
            </div>

            <div className="rounded-2xl bg-sand-100/60 p-4 sm:p-5 space-y-2.5">
              <p className="text-xs font-bold uppercase tracking-wider text-teal-900">
                Key Invariants
              </p>
              <ul className="space-y-2 text-xs sm:text-sm text-sand-900/85 font-medium">
                {activeFeature.bullets.map((b, i) => (
                  <li key={i} className="flex items-start gap-2">
                    <span className="text-teal-700 font-bold">✓</span>
                    <span>{b}</span>
                  </li>
                ))}
              </ul>
            </div>

            <button
              type="button"
              onClick={() => setActiveFeature(null)}
              className="w-full rounded-xl bg-teal-800 py-3 text-sm font-bold text-white shadow-sm hover:bg-teal-900 transition touch-spring"
            >
              Done
            </button>
          </div>
        </div>
      )}
    </section>
  );
}
