import { Header } from "./components/Header";
import { IntakeForm } from "./components/IntakeForm";
import { HowItWorks } from "./components/landing/HowItWorks";
import { ScenarioGrid } from "./components/landing/ScenarioGrid";
import { SafetyPledge } from "./components/landing/SafetyPledge";
import { FaqSection } from "./components/landing/FaqSection";
import { Footer } from "./components/landing/Footer";
import { IconShieldCheck, IconLock, IconScale } from "./components/icons";

export default function HomePage() {
  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      <Header />

      {/* Hero & intake form */}
      <main className="mx-auto w-full max-w-6xl flex-1 px-6 sm:px-8 lg:px-10 pt-8 sm:pt-12 lg:pt-14 pb-12 sm:pb-16 lg:pb-24 space-y-16 sm:space-y-20 animate-fade-in">
        <div className="space-y-8">
          {/* Eyebrow and headline */}
          <div className="space-y-3.5 max-w-2xl">
            <p className="text-xs font-bold uppercase tracking-wider text-teal-700">
              Point-of-denial assistance
            </p>
            <h1 className="font-display text-3xl sm:text-4xl lg:text-5xl font-semibold leading-tight tracking-tight-display text-teal-950">
              Is the hospital right to deny your PMJAY claim?
            </h1>
            <p className="text-base sm:text-lg leading-relaxed text-sand-700 font-medium">
              Free, instant help figuring out whether a hospital&rsquo;s coverage denial is correct —
              right when you&rsquo;re standing at the billing desk. This is not a government service
              and it does not replace the official complaint process, but it can help you understand
              what&rsquo;s happening and what to say next.
            </p>
          </div>

          {/* Primary intake form */}
          <div className="card p-6 sm:p-8 shadow-[0_2px_8px_rgba(42,38,33,0.06)]">
            <IntakeForm />
          </div>

          {/* Three trust / scope pillars */}
          <div className="grid gap-3 sm:grid-cols-3 pt-2">
            <div className="flex items-start gap-3 rounded-2xl bg-sand-100/70 p-4">
              <IconShieldCheck className="h-5 w-5 shrink-0 text-teal-700 mt-0.5" />
              <div>
                <p className="text-sm font-bold text-sand-900">300+ Verified Packages</p>
                <p className="mt-0.5 text-xs text-sand-600 leading-relaxed font-medium">
                  Grounded in official NHA rate definitions.
                </p>
              </div>
            </div>
            <div className="flex items-start gap-3 rounded-2xl bg-sand-100/70 p-4">
              <IconLock className="h-5 w-5 shrink-0 text-teal-700 mt-0.5" />
              <div>
                <p className="text-sm font-bold text-sand-900">100% Free &amp; Private</p>
                <p className="mt-0.5 text-xs text-sand-600 leading-relaxed font-medium">
                  No login, no phone number required.
                </p>
              </div>
            </div>
            <div className="flex items-start gap-3 rounded-2xl bg-sand-100/70 p-4">
              <IconScale className="h-5 w-5 shrink-0 text-teal-700 mt-0.5" />
              <div>
                <p className="text-sm font-bold text-sand-900">Legal Escalation Link</p>
                <p className="mt-0.5 text-xs text-sand-600 leading-relaxed font-medium">
                  Prepares summary for NALSA helpline (15100).
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Explanatory landing sections */}
        <HowItWorks />
        <ScenarioGrid />
        <SafetyPledge />
        <FaqSection />
      </main>

      <Footer />
    </div>
  );
}
