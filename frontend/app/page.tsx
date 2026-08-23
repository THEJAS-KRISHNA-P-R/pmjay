import { Header } from "./components/Header";
import { IntakeForm } from "./components/IntakeForm";
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
              Point-of-Denial Assistance
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

          <div className="glass-panel-elevated p-6 sm:p-8">
            <IntakeForm />
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

        {/* 3-Step Process */}
        <HowItWorks />

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
