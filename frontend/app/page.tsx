import Link from "next/link";
import { Header } from "./components/Header";
import { IntakeForm } from "./components/IntakeForm";
import { HowItWorks } from "./components/landing/HowItWorks";
import { ScenarioGrid } from "./components/landing/ScenarioGrid";
import { SafetyPledge } from "./components/landing/SafetyPledge";
import { FaqSection } from "./components/landing/FaqSection";
import { Footer } from "./components/landing/Footer";
import { IconShieldCheck, IconLock, IconScale, IconCheck, IconArrowRight } from "./components/icons";

export default function HomePage() {
  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      <Header />

      <main className="mx-auto w-full max-w-6xl flex-1 px-4 sm:px-8 lg:px-10 pb-16 sm:pb-24 space-y-16 sm:space-y-24 animate-fade-in">
        {/* High-Credibility Hero Section: Centered on mobile, 2-Column split on PC */}
        <section aria-label="Hero" className="pt-12 sm:pt-16 lg:pt-16 pb-4">
          <div className="grid lg:grid-cols-[1.15fr_0.85fr] gap-10 lg:gap-12 items-center">
            {/* Left Column (Desktop) / Centered Content (Mobile) */}
            <div className="space-y-6 text-center lg:text-left">
              {/* Eyebrow badge */}
              <div className="inline-flex items-center gap-2">
                <span className="text-xs sm:text-sm font-bold uppercase tracking-wider text-emerald-700">
                  Point-of-Denial Legal &amp; Scheme Verification
                </span>
              </div>

              {/* Two-tone headline */}
              <h1 className="font-display text-4xl sm:text-5xl lg:text-[3.5rem] font-bold tracking-tight-display text-ink-950 leading-[1.08]">
                Never pay when PMJAY{" "}
                <span className="text-emerald-700 block sm:inline">covers it.</span>
              </h1>

              {/* Subtitle */}
              <p className="text-base sm:text-lg text-sand-600 font-medium leading-relaxed max-w-xl mx-auto lg:mx-0">
                Verify any hospital refusal under Ayushman Bharat / PMJAY against official government rates in seconds. 100% free, instant, and based on your legal rights under PMJAY.
              </p>

              {/* Action CTAs: Primary emerald button + Underlined linked text */}
              <div className="flex items-center justify-center lg:justify-start gap-4 sm:gap-6 pt-1">
                <a
                  href="#check-coverage"
                  className="btn-primary tap-target px-6 sm:px-7 py-3.5 sm:py-4 text-sm sm:text-base"
                >
                  <span>Get Help Now</span>
                  <span aria-hidden="true" className="text-base">↓</span>
                </a>
                <Link
                  href="/how-it-works"
                  className="inline-flex items-center text-sm sm:text-base font-semibold text-sand-800 hover:text-emerald-700 underline underline-offset-4 decoration-sand-300 hover:decoration-emerald-600 transition-colors whitespace-nowrap py-2"
                >
                  <span>How It Works</span>
                </Link>
              </div>

              {/* Coverage Validation Trust Badges */}
              <div className="flex flex-wrap items-center justify-center lg:justify-start gap-x-6 gap-y-2 pt-3 text-xs sm:text-sm font-semibold text-sand-600 border-t border-sand-200/60">
                <span className="inline-flex items-center gap-1.5">
                  <IconCheck className="h-3.5 w-3.5 text-emerald-700" />
                  All PMJAY &amp; Ayushman Cards
                </span>
                <span className="inline-flex items-center gap-1.5">
                  <IconCheck className="h-3.5 w-3.5 text-emerald-700" />
                  Senior (70+) &amp; State Schemes
                </span>
                <span className="inline-flex items-center gap-1.5">
                  <IconShieldCheck className="h-3.5 w-3.5 text-emerald-700" />
                  Official 315-Package Rates
                </span>
              </div>
            </div>

            {/* Right Column: Live Verification Card */}
            <div className="max-w-md lg:max-w-none mx-auto lg:mx-0 w-full text-left">
              <div className="card p-5 sm:p-7 space-y-4">
                <div className="flex items-center justify-between border-b border-sand-100 pb-3 gap-2">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-emerald-50 text-emerald-700 shrink-0">
                      <IconShieldCheck className="h-4.5 w-4.5" />
                    </span>
                    <div className="min-w-0">
                      <p className="text-xs sm:text-sm font-bold text-sand-900 truncate">Ayushman Live Verification</p>
                      <p className="text-[11px] text-sand-500 font-medium">NHA HBP 2022 Master Schedule</p>
                    </div>
                  </div>
                  <span className="badge bg-emerald-50 border-emerald-200 text-emerald-800 text-[11px] font-bold shrink-0">
                    100% Cashless Covered
                  </span>
                </div>

                <div className="space-y-2.5 bg-sand-50/80 p-3.5 rounded-xl border border-sand-100/80">
                  <div className="flex justify-between items-center text-xs font-semibold">
                    <span className="text-sand-600">Hospital Stated Bill:</span>
                    <span className="text-tier-red-text line-through font-bold">₹85,000 (Demanded Cash)</span>
                  </div>
                  <div className="flex justify-between items-center text-xs font-semibold">
                    <span className="text-sand-600">Official Beneficiary Cost:</span>
                    <span className="text-emerald-700 font-extrabold text-sm">₹0 (Zero Upfront)</span>
                  </div>
                  <div className="flex justify-between items-center text-[11px] text-sand-500 pt-1 border-t border-sand-200/40">
                    <span>Matched Code: HP-104 (Angioplasty)</span>
                    <span className="text-emerald-800 font-bold">Pre-auth Approved</span>
                  </div>
                </div>

                <div className="rounded-xl bg-emerald-50/70 border border-emerald-200/60 p-3 text-xs text-emerald-950 font-medium leading-relaxed">
                  <p className="font-bold text-emerald-900 flex items-center gap-1.5 mb-0.5">
                    <IconCheck className="h-3.5 w-3.5 text-emerald-700 shrink-0" />
                    Legal Protection Guaranteed:
                  </p>
                  Empanelled hospitals cannot charge cash deposits or refuse admission for covered packages under National Health Authority regulations.
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Dedicated Intake Section below the Hero */}
        <section id="check-coverage" className="scroll-mt-24 max-w-3xl mx-auto space-y-6">
          <div className="text-center space-y-2">
            <h2 className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-ink-950">
              Check Your Case
            </h2>
            <p className="text-sm sm:text-base text-sand-600 font-medium">
              Tell us what the hospital billing desk is saying, in your own words.
            </p>
          </div>

          <div className="card p-6 sm:p-8">
            <IntakeForm />
          </div>

          {/* Three trust / scope pillars */}
          <div className="grid gap-3 sm:grid-cols-3 pt-4 text-left">
            <div className="card p-4 flex items-start gap-3">
              <IconShieldCheck className="h-5 w-5 shrink-0 text-emerald-800 mt-0.5" />
              <div>
                <p className="text-sm font-bold text-sand-900">315 Verified Packages</p>
                <p className="mt-0.5 text-xs text-sand-600 leading-relaxed font-medium">
                  Official NHA 2022 rate index.
                </p>
              </div>
            </div>
            <div className="card p-4 flex items-start gap-3">
              <IconLock className="h-5 w-5 shrink-0 text-emerald-800 mt-0.5" />
              <div>
                <p className="text-sm font-bold text-sand-900">100% Free &amp; Private</p>
                <p className="mt-0.5 text-xs text-sand-600 leading-relaxed font-medium">
                  No login, no phone number.
                </p>
              </div>
            </div>
            <div className="card p-4 flex items-start gap-3">
              <IconScale className="h-5 w-5 shrink-0 text-emerald-800 mt-0.5" />
              <div>
                <p className="text-sm font-bold text-sand-900">Official Dispute Draft</p>
                <p className="mt-0.5 text-xs text-sand-600 leading-relaxed font-medium">
                  Ready for CGRMS &amp; NALSA (15100).
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Supporting landing sections */}
        <HowItWorks />
        <ScenarioGrid />
        <SafetyPledge />
        <FaqSection />
      </main>

      <Footer />
    </div>
  );
}

