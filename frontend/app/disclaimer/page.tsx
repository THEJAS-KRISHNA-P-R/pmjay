import type { Metadata } from "next";
import { Header } from "../components/Header";
import { Footer } from "../components/landing/Footer";
import { IconAlertTriangle, IconShieldCheck, IconPhone, IconCheck } from "../components/icons";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Legal & Medical Disclaimer — PMJAY Advocate",
  description:
    "Important legal and medical disclaimers regarding the use of PMJAY Advocate. This is an informational tool and does not provide formal legal or medical advice.",
};

export default function DisclaimerPage() {
  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      <Header />

      <main className="mx-auto w-full max-w-4xl flex-1 px-4 sm:px-8 lg:px-10 pt-12 sm:pt-16 pb-16 space-y-10 animate-fade-in">
        {/* Header */}
        <div className="space-y-3">
          <div className="inline-flex items-center gap-2">
            <span className="text-xs sm:text-sm font-bold uppercase tracking-wider text-emerald-700">
              Disclaimers &amp; Disclosures
            </span>
          </div>
          <h1 className="font-display text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight-display text-ink-950">
            Legal &amp; Medical Disclaimer
          </h1>
          <p className="text-sm sm:text-base text-sand-600 font-medium leading-relaxed">
            Please read these disclaimers carefully. PMJAY Advocate is a public-interest informational tool and does not substitute for qualified clinical judgment, formal legal counsel, or official government determinations.
          </p>
        </div>

        {/* Emergency Medical Care Callout */}
        <div className="card p-6 sm:p-7 border-2 border-tier-red-border bg-tier-red-bg space-y-3 shadow-xs">
          <div className="flex items-start gap-3">
            <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-tier-red-icon text-tier-red-text shrink-0 mt-0.5">
              <IconAlertTriangle className="h-4.5 w-4.5" />
            </span>
            <div className="space-y-1.5">
              <h2 className="text-base font-bold text-tier-red-text">
                Medical Emergency Warning: Care Comes First
              </h2>
              <p className="text-xs sm:text-sm text-tier-red-text leading-relaxed font-medium">
                <strong>PMJAY Advocate is not an emergency response service and cannot provide clinical medical advice.</strong> If a patient is facing an acute health emergency, injury, or critical illness, obtain medical treatment immediately. Do not delay emergency hospital care while researching package codes or drafting dispute letters.
              </p>
            </div>
          </div>
        </div>

        {/* Disclaimer 1: Not Medical Advice */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              1
            </span>
            Not Medical or Clinical Advice
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              The content, package descriptions, and rate comparisons provided by this application do not constitute medical diagnoses, treatment recommendations, or clinical evaluations.
            </p>
            <ul className="space-y-2 pl-4 list-disc marker:text-emerald-700">
              <li>Always seek the advice of a qualified physician or healthcare professional regarding any medical condition.</li>
              <li>Clinical decisions regarding whether a surgery, diagnostic test, or admission is medically necessary rest solely with the treating medical team.</li>
            </ul>
          </div>
        </section>

        {/* Disclaimer 2: Not Legal Advice */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              2
            </span>
            Not Formal Legal Advice or Representation
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              Although this application generates counter-scripts and draft grievance letters for convenience, <strong>it does not provide legal representation or create an advocate-client relationship.</strong>
            </p>
            <p>
              The generated materials are automated educational templates based on public government guidelines. They are not legal pleadings and have not been tailored by a licensed advocate to your specific legal rights under the Consumer Protection Act or Clinical Establishments Act.
            </p>
            <p>
              For formal legal aid and representation at no cost, beneficiaries are entitled to free assistance through the National Legal Services Authority (NALSA) helpline at <a href="tel:15100" className="font-bold text-emerald-700 underline">15100</a>.
            </p>
          </div>
        </section>

        {/* Disclaimer 3: Scheme Data & Rate Accuracy */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              3
            </span>
            Government Scheme Data &amp; Package Currency
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              Package names, codes, rates, and pre-authorisation criteria are indexed from published National Health Authority (NHA) Health Benefit Package schedules (including HBP 2.1 and HBP 2022 master lists).
            </p>
            <p>
              Individual states (e.g., AB-PMJAY Karunya in Kerala, AB-PMJAY MJPJAY in Maharashtra) may maintain state-specific top-up packages, modified reserve prices, or specific pre-auth mandates. Final coverage determinations are made exclusively through official SHA/NHA IT portals (Transaction Management System - TMS).
            </p>
          </div>
        </section>

        {/* Disclaimer 4: Official Help Channels */}
        <section className="card p-6 sm:p-8 space-y-4">
          <h2 className="text-lg sm:text-xl font-bold text-sand-900 flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-sand-100 text-sand-800 text-xs font-bold">
              4
            </span>
            Official Government Channels
          </h2>
          <div className="space-y-3 text-xs sm:text-sm text-sand-700 leading-relaxed font-medium">
            <p>
              If an empanelled hospital refuses treatment or demands illicit cash payments, lodge your grievance directly with statutory authorities:
            </p>
            <div className="grid sm:grid-cols-2 gap-3 pt-1">
              <div className="p-4 rounded-xl bg-sand-50 border border-sand-200 space-y-1">
                <p className="font-bold text-sand-900">National Health Authority (NHA)</p>
                <p className="text-xs text-sand-600">Toll-Free Helpline: <a href="tel:14555" className="font-bold text-emerald-700 underline">14555</a></p>
                <p className="text-xs text-sand-600">Grievance Portal: <span className="font-mono text-[11px] text-sand-800">cgrms.pmjay.gov.in</span></p>
              </div>
              <div className="p-4 rounded-xl bg-sand-50 border border-sand-200 space-y-1">
                <p className="font-bold text-sand-900">National Legal Services Authority</p>
                <p className="text-xs text-sand-600">Legal Aid Helpline: <a href="tel:15100" className="font-bold text-emerald-700 underline">15100</a></p>
                <p className="text-xs text-sand-600">Portal: <span className="font-mono text-[11px] text-sand-800">nalsa.gov.in</span></p>
              </div>
            </div>
          </div>
        </section>
      </main>

      <Footer />
    </div>
  );
}
