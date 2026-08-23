import { Header } from "./components/Header";
import { IntakeForm } from "./components/IntakeForm";
import { Features } from "./components/landing/Features";
import { Comparison } from "./components/landing/Comparison";
import { HowItWorks } from "./components/landing/HowItWorks";
import { ScenarioGrid } from "./components/landing/ScenarioGrid";
import { SafetyPledge } from "./components/landing/SafetyPledge";
import { FaqSection } from "./components/landing/FaqSection";
import { Footer } from "./components/landing/Footer";

export default function HomePage() {
  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      <Header />
      
      {/* Hero & Intake Form Section */}
      <main className="mx-auto w-full max-w-4xl flex-1 px-4 sm:px-6 pt-8 sm:pt-14 pb-12 space-y-16 animate-enter">
        <div className="space-y-8">
          <div className="space-y-3.5 max-w-2xl">
            <div className="inline-flex items-center rounded-full bg-teal-100/70 px-4 py-1.5 text-xs font-bold uppercase tracking-wider text-teal-800 shadow-[inset_0_1px_0_0_rgba(255,255,255,0.7)] backdrop-blur-sm">
              Hospital Point-of-Denial Assistance
            </div>
            
            <h1 className="font-display text-3xl sm:text-5xl font-semibold leading-tight tracking-tight-display text-teal-950">
              Is the hospital right to deny your PMJAY claim?
            </h1>
            
            <p className="text-base sm:text-lg leading-relaxed tracking-body text-sand-900/80 font-medium">
              Free, instant help figuring out whether a hospital&rsquo;s coverage denial is correct —
              right when you&rsquo;re standing at the billing desk. This is not a government service
              and it does not replace the official complaint process, but it can help you understand
              what&rsquo;s happening and what to say next.
            </p>
          </div>

          {/* Intake Card with Browser Mockup Header */}
          <div className="glass-panel-elevated overflow-hidden">
            <div className="flex items-center justify-between px-5 py-3.5 bg-sand-100/70 border-b border-sand-200/40">
              <div className="flex items-center gap-2">
                <span className="h-2.5 w-2.5 rounded-full bg-rose-400" />
                <span className="h-2.5 w-2.5 rounded-full bg-amber-400" />
                <span className="h-2.5 w-2.5 rounded-full bg-emerald-400" />
                <span className="ml-2 font-mono text-[11px] font-bold text-sand-900/50">
                  pmjay-advocate.org / intake-evaluator
                </span>
              </div>
              <span className="text-[10px] font-bold uppercase tracking-wider text-teal-800 bg-teal-50 px-2.5 py-0.5 rounded-full">
                Active Desk Mode
              </span>
            </div>
            <div className="p-6 sm:p-8">
              <IntakeForm />
            </div>
          </div>

          {/* Quick Trust / Scope Metrics Bar */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            {[
              { label: "HBP Packages", value: "1,949+", sub: "NHA 2022 Schedule" },
              { label: "Access Cost", value: "100% Free", sub: "Public Interest" },
              { label: "Evaluation", value: "Instant", sub: "Zero Auth Required" },
              { label: "Legal Escalation", value: "NALSA 15100", sub: "Toll-Free Support" },
            ].map((stat) => (
              <div
                key={stat.label}
                className="glass-panel p-4 text-center space-y-0.5"
              >
                <p className="font-display text-lg sm:text-xl font-bold text-teal-950">
                  {stat.value}
                </p>
                <p className="text-xs font-bold text-teal-800">
                  {stat.label}
                </p>
                <p className="text-[10px] text-sand-900/60 font-medium">
                  {stat.sub}
                </p>
              </div>
            ))}
          </div>
        </div>

        {/* Platform Capabilities (yourfee.in style feature grid) */}
        <Features />

        {/* 3-Step Process */}
        <HowItWorks />

        {/* Detailed Comparison Table (Advocate vs Typical Alternatives) */}
        <Comparison />

        {/* Common Denial Scenarios & Tiers */}
        <ScenarioGrid />

        {/* Safety & Care-First Philosophy */}
        <SafetyPledge />

        {/* FAQ Section */}
        <FaqSection />
      </main>

      {/* Footer */}
      <Footer />
    </div>
  );
}
