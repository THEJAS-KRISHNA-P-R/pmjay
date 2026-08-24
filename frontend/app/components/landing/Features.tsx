"use client";

import { useState } from "react";
import {
  IconClipboardList,
  IconScale,
  IconMessageText,
  IconFileText,
  IconPencilLine,
  IconUsers,
  IconX,
  IconCheck,
  type IconProps,
} from "../icons";
import type { ComponentType } from "react";

interface Feature {
  id: string;
  title: string;
  shortDesc: string;
  fullDesc: string;
  bullets: string[];
  Icon: ComponentType<IconProps>;
}

const features: Feature[] = [
  {
    id: "hbp-database",
    title: "315 HBP Package Index",
    shortDesc: "National Health Authority benefit schedule with indicative rates — 300 of 315 entries independently verified.",
    fullDesc: "Covers HBP packages across surgical, medical, and pediatric specialties with indicative rates and package definitions. A small number of entries are placeholders pending verification against the official master file — flagged rather than presented as settled.",
    bullets: [
      "Indicative procedure rates and package definitions",
      "Specialty-wise categorization (Cardiology, General Surgery, Ortho)",
      "Every entry marked verified or pending — never blurred together",
    ],
    Icon: IconClipboardList,
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
    Icon: IconScale,
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
    Icon: IconMessageText,
  },
  {
    id: "pdf-document",
    title: "Instant Case PDF Generator",
    shortDesc: "Zero-dependency PDF document generated on demand.",
    fullDesc: "Transforms your entire evaluation, package details, hospital script, and evidence log into a clean printable PDF to physically hand to hospital administrators.",
    bullets: [
      "Instant native download with zero browser dependencies",
      "Structured layout with emergency priority banner",
      "Ready for official hospital grievance submission",
    ],
    Icon: IconFileText,
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
    Icon: IconPencilLine,
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
    Icon: IconUsers,
  },
];

export function Features() {
  const [activeFeature, setActiveFeature] = useState<Feature | null>(null);

  return (
    <section aria-labelledby="features-title" className="space-y-8">
      <div className="text-center space-y-2">
        <p className="text-xs font-bold uppercase tracking-wider text-teal-700">
          Core Engine Capabilities
        </p>
        <h2 id="features-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-sand-900">
          Everything You Need at the Hospital Desk
        </h2>
        <p className="text-sm sm:text-base text-sand-600 max-w-xl mx-auto font-medium">
          Engineered for speed, privacy, and clinical-legal accuracy.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {features.map((f) => {
          const { Icon } = f;
          return (
            <button
              type="button"
              key={f.id}
              onClick={() => setActiveFeature(f)}
              className="card p-6 flex flex-col justify-between text-left transition-all hover:-translate-y-0.5 hover:shadow-md hover:border-teal-200"
            >
              <div className="space-y-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-teal-100 text-teal-700">
                  <Icon className="h-5 w-5" />
                </div>
                <h3 className="font-display text-base sm:text-lg font-bold text-sand-900">
                  {f.title}
                </h3>
                <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                  {f.shortDesc}
                </p>
              </div>
              <div className="pt-4 text-xs font-bold text-teal-700">
                Explore details →
              </div>
            </button>
          );
        })}
      </div>

      {/* Feature Detail Modal */}
      {activeFeature && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-sand-950/40 backdrop-blur-sm animate-fade-in"
          onClick={() => setActiveFeature(null)}
          role="dialog"
          aria-modal="true"
          aria-labelledby="feature-modal-title"
        >
          <div
            className="max-w-lg w-full rounded-3xl bg-white p-6 sm:p-8 space-y-5 shadow-2xl"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between">
              <span className="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-teal-700 text-white">
                <activeFeature.Icon className="h-5 w-5" />
              </span>
              <button
                type="button"
                onClick={() => setActiveFeature(null)}
                aria-label="Close"
                className="tap-target h-9 w-9 rounded-full bg-sand-100 text-sand-600 hover:text-sand-900 hover:bg-sand-200 flex items-center justify-center transition-colors"
              >
                <IconX className="h-4 w-4" />
              </button>
            </div>

            <div>
              <h3 id="feature-modal-title" className="font-display text-xl sm:text-2xl font-bold text-sand-900">
                {activeFeature.title}
              </h3>
              <p className="mt-2 text-sm text-sand-700 leading-relaxed font-medium">
                {activeFeature.fullDesc}
              </p>
            </div>

            <div className="rounded-2xl bg-sand-50 border border-sand-200/70 p-4 sm:p-5 space-y-2.5">
              <p className="text-xs font-bold uppercase tracking-wider text-teal-800">
                Key Invariants
              </p>
              <ul className="space-y-2 text-xs sm:text-sm text-sand-700 font-medium">
                {activeFeature.bullets.map((b, i) => (
                  <li key={i} className="flex items-start gap-2">
                    <IconCheck className="h-4 w-4 shrink-0 text-teal-600 mt-0.5" />
                    <span>{b}</span>
                  </li>
                ))}
              </ul>
            </div>

            <button
              type="button"
              onClick={() => setActiveFeature(null)}
              className="btn-primary w-full py-3"
            >
              Done
            </button>
          </div>
        </div>
      )}
    </section>
  );
}
