import type { Metadata } from "next";
import { Header } from "../components/Header";
import { Footer } from "../components/landing/Footer";
import { HowItWorks } from "../components/landing/HowItWorks";
import { Comparison } from "../components/landing/Comparison";
import { ScenarioGrid } from "../components/landing/ScenarioGrid";
import { IconScale, IconClipboardList, IconMessageText } from "../components/icons";

export const metadata: Metadata = {
  title: "How It Works — PMJAY Advocate",
  description:
    "How PMJAY Advocate checks a hospital's coverage denial against the official Health Benefit Package index and decides what to do next.",
};

export default function HowItWorksPage() {
  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      <Header />

      <main className="mx-auto w-full max-w-6xl flex-1 px-4 sm:px-8 lg:px-10 pt-6 sm:pt-12 lg:pt-16 pb-12 sm:pb-16 lg:pb-24 space-y-12 sm:space-y-20 animate-fade-in">
        <div className="max-w-2xl space-y-3.5">
          <p className="text-xs font-bold uppercase tracking-wider text-sand-500">Behind the answer</p>
          <h1 className="font-display text-3xl sm:text-4xl lg:text-5xl font-semibold leading-tight tracking-tight-display text-ink-950">
            How PMJAY Advocate Works
          </h1>
          <p className="text-base sm:text-lg leading-relaxed text-sand-700 font-medium">
            No guesswork, no vague reassurance: every answer traces back to a specific rule. Here&rsquo;s exactly
            what happens between typing your situation and getting a verdict.
          </p>
        </div>

        <HowItWorks />

        {/* Engine Transparency */}
        <section aria-labelledby="engine-title" className="card p-5 sm:p-8 space-y-6">
          <div className="max-w-2xl space-y-1.5">
            <h2 id="engine-title" className="font-display text-xl sm:text-2xl lg:text-3xl font-semibold tracking-tight-display text-sand-900">
              What happens behind the scenes
            </h2>
            <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
              We explain this plainly rather than leaving it as a black box.
            </p>
          </div>

          <div className="grid gap-4 sm:gap-6 sm:grid-cols-3 divide-y sm:divide-y-0 divide-sand-100">
            <div className="space-y-2 pt-3 sm:pt-0">
              <div className="flex items-center gap-2.5">
                <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-sand-100 text-ink-700 shrink-0">
                  <IconMessageText className="h-4 w-4" />
                </span>
                <h3 className="font-bold text-sm sm:text-base text-sand-900">1. Fact Extraction</h3>
              </div>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                The model extracts clinical details (procedure name, billing demand, dispute stage). It never makes the final legal decision.
              </p>
            </div>

            <div className="space-y-2 pt-3 sm:pt-0">
              <div className="flex items-center gap-2.5">
                <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-sand-100 text-ink-700 shrink-0">
                  <IconClipboardList className="h-4 w-4" />
                </span>
                <h3 className="font-bold text-sm sm:text-base text-sand-900">2. NHA Index Match</h3>
              </div>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                Clinical facts are cross-referenced against the official 315 PMJAY package index and exclusion criteria in real time.
              </p>
            </div>

            <div className="space-y-2 pt-3 sm:pt-0">
              <div className="flex items-center gap-2.5">
                <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-sand-100 text-ink-700 shrink-0">
                  <IconScale className="h-4 w-4" />
                </span>
                <h3 className="font-bold text-sm sm:text-base text-sand-900">3. Deterministic Rules</h3>
              </div>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                Outcomes are decided by strict, deterministic Go logic, guaranteeing the same facts always produce the exact same verdict.
              </p>
            </div>
          </div>
        </section>

        <Comparison />
        <ScenarioGrid />
      </main>

      <Footer />
    </div>
  );
}
