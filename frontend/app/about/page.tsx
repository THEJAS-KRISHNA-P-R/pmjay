import type { Metadata } from "next";
import { Header } from "../components/Header";
import { Footer } from "../components/landing/Footer";
import { SafetyPledge } from "../components/landing/SafetyPledge";
import { IconHeart, IconLock, IconScale } from "../components/icons";

export const metadata: Metadata = {
  title: "About — PMJAY Advocate",
  description:
    "Why PMJAY Advocate was built, how it operates as a free public-interest tool, and our core principles of privacy and safety.",
};

export default function AboutPage() {
  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      <Header />

      <main className="mx-auto w-full max-w-6xl flex-1 px-4 sm:px-8 lg:px-10 pt-6 sm:pt-12 lg:pt-16 pb-12 sm:pb-16 lg:pb-24 space-y-12 sm:space-y-16 animate-fade-in">
        <div className="max-w-2xl space-y-3.5">
          <p className="text-xs font-bold uppercase tracking-wider text-ink-700">Our purpose</p>
          <h1 className="font-display text-3xl sm:text-4xl lg:text-5xl font-semibold leading-tight tracking-tight-display text-ink-950">
            About PMJAY Advocate
          </h1>
          <p className="text-base sm:text-lg leading-relaxed text-sand-700 font-medium">
            Standing at a hospital billing desk during a family medical emergency is one of the most
            vulnerable moments a person can face. This tool exists to give families clarity, calm, and
            factual standing in that exact moment.
          </p>
        </div>

        {/* Core principles */}
        <section aria-labelledby="principles-title" className="space-y-6">
          <h2 id="principles-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-sand-900">
            Our Core Commitments
          </h2>
          <div className="grid gap-5 sm:grid-cols-3">
            <div className="card p-6 space-y-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-sand-100 border border-sand-200/60 text-sand-700 shadow-[inset_0_1px_1px_rgba(255,255,255,0.9),0_1px_3px_rgba(42,38,33,0.05)]">
                <IconHeart className="h-4 w-4" />
              </div>
              <h3 className="font-bold text-sand-900">Care First, Always</h3>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                We never advise delaying urgent medical care over a billing dispute. Emergency treatment comes
                first; financial and administrative disputes are handled afterward.
              </p>
            </div>

            <div className="card p-6 space-y-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-sand-100 border border-sand-200/60 text-sand-700 shadow-[inset_0_1px_1px_rgba(255,255,255,0.9),0_1px_3px_rgba(42,38,33,0.05)]">
                <IconLock className="h-4 w-4" />
              </div>
              <h3 className="font-bold text-sand-900">Zero Personal Data Required</h3>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                No phone number, Aadhaar number, or account creation is required to use this tool. Case details
                are held under an anonymous ID and are not tied to your identity.
              </p>
            </div>

            <div className="card p-6 space-y-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-sand-100 border border-sand-200/60 text-sand-700 shadow-[inset_0_1px_1px_rgba(255,255,255,0.9),0_1px_3px_rgba(42,38,33,0.05)]">
                <IconScale className="h-4 w-4" />
              </div>
              <h3 className="font-bold text-sand-900">Grounded in Real Rates</h3>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                All package checks are cross-referenced against the official National Health Authority Health
                Benefit Package (HBP) index. We never invent coverage rules or generate unfounded disputes.
              </p>
            </div>
          </div>
        </section>

        {/* Public interest / non-gov statement */}
        <section aria-labelledby="independence-title" className="card p-6 sm:p-10 space-y-4">
          <h2 id="independence-title" className="font-display text-xl sm:text-2xl font-bold text-sand-900">
            An Independent Public-Interest Resource
          </h2>
          <p className="text-sm sm:text-base text-sand-700 leading-relaxed font-medium">
            PMJAY Advocate is an independent public-interest tool, not a government body or official service
            of the National Health Authority (NHA). It is designed to assist patients and families in
            understanding their rights under published scheme guidelines. It does not replace the official
            grievance redressal process (CGRMS) or formal legal representation through bodies like NALSA.
          </p>
          <p className="text-sm sm:text-base text-sand-700 leading-relaxed font-medium">
            Official NHA grievance portal:{" "}
            <a
              href="https://cgrms.pmjay.gov.in"
              target="_blank"
              rel="noopener noreferrer"
              className="font-bold text-ink-800 underline hover:opacity-80"
            >
              cgrms.pmjay.gov.in
            </a>
          </p>
        </section>

        <SafetyPledge />
      </main>

      <Footer />
    </div>
  );
}
