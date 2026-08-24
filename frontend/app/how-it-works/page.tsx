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

      <main className="mx-auto w-full max-w-6xl flex-1 px-6 sm:px-8 lg:px-10 pt-8 sm:pt-12 lg:pt-16 pb-12 sm:pb-16 lg:pb-24 space-y-16 sm:space-y-20 animate-fade-in">
        <div className="max-w-2xl space-y-3.5">
          <p className="text-xs font-bold uppercase tracking-wider text-teal-700">Behind the answer</p>
          <h1 className="font-display text-3xl sm:text-4xl lg:text-5xl font-semibold leading-tight tracking-tight-display text-teal-950">
            How PMJAY Advocate Works
          </h1>
          <p className="text-base sm:text-lg leading-relaxed text-sand-700 font-medium">
            No guesswork, no vague reassurance — every answer traces back to a specific rule. Here&rsquo;s exactly
            what happens between typing your situation and getting a verdict.
          </p>
        </div>

        <HowItWorks />

        {/* What's actually running under the hood, stated plainly */}
        <section aria-labelledby="engine-title" className="card p-6 sm:p-10 space-y-6">
          <div className="max-w-2xl space-y-2">
            <h2 id="engine-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-sand-900">
              What&rsquo;s actually happening behind the scenes
            </h2>
            <p className="text-sm sm:text-base text-sand-600 leading-relaxed font-medium">
              We&rsquo;d rather explain this plainly than let it feel like a black box.
            </p>
          </div>

          <div className="grid gap-5 sm:grid-cols-3">
            <div className="space-y-2.5">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-teal-100 text-teal-700">
                <IconMessageText className="h-4 w-4" />
              </div>
              <h3 className="font-bold text-sand-900">1. Understanding your words</h3>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                An AI model reads your description and pulls out the clinical facts — what procedure, what the
                hospital said, what stage the dispute is at. It only extracts; it never decides your outcome.
              </p>
            </div>
            <div className="space-y-2.5">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-teal-100 text-teal-700">
                <IconClipboardList className="h-4 w-4" />
              </div>
              <h3 className="font-bold text-sand-900">2. Matching against real rates</h3>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                Those facts are checked against the 315-package HBP index — indicative rates and package
                definitions grounded in official NHA terminology, with unverified entries flagged, never hidden.
              </p>
            </div>
            <div className="space-y-2.5">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-teal-100 text-teal-700">
                <IconScale className="h-4 w-4" />
              </div>
              <h3 className="font-bold text-sand-900">3. A fixed rule decides, not the AI</h3>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                The actual Green/Amber/Red/Mixed/Handoff verdict is produced by deterministic logic, not the
                language model — the same facts always produce the same outcome, on purpose.
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
