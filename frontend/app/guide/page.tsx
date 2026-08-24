import type { Metadata } from "next";
import { Header } from "../components/Header";
import { Footer } from "../components/landing/Footer";
import { IconCheck, IconX, IconPhone } from "../components/icons";

export const metadata: Metadata = {
  title: "Your Rights — PMJAY Advocate",
  description:
    "A plain-language guide to what PMJAY covers, what hospitals aren't allowed to do, and what to do if your coverage is denied.",
};

const SECTIONS = [
  { id: "what-is-covered", label: "What PMJAY covers" },
  { id: "your-rights", label: "Your core rights" },
  { id: "not-allowed", label: "What hospitals can't do" },
  { id: "watch-for", label: "Common denial patterns" },
  { id: "what-to-do", label: "What to do at the desk" },
  { id: "escalate", label: "When to escalate" },
];

export default function GuidePage() {
  return (
    <div className="min-h-screen flex flex-col bg-sand-50">
      <Header />

      <main className="mx-auto w-full max-w-6xl flex-1 px-6 sm:px-8 lg:px-10 pt-8 sm:pt-12 lg:pt-16 pb-12 sm:pb-16 lg:pb-24 animate-fade-in">
        <div className="max-w-3xl space-y-3.5 mb-10 sm:mb-12">
          <p className="text-xs font-bold uppercase tracking-wider text-teal-700">Know before you go</p>
          <h1 className="font-display text-3xl sm:text-4xl lg:text-5xl font-semibold leading-tight tracking-tight-display text-teal-950">
            Your Rights Under PMJAY
          </h1>
          <p className="text-base sm:text-lg leading-relaxed text-sand-700 font-medium">
            General guidance, written plainly — not a substitute for the official scheme rules, but enough to
            know when something at the billing desk doesn&rsquo;t add up.
          </p>
        </div>

        <div className="lg:grid lg:grid-cols-[220px_1fr] lg:gap-12 xl:gap-16">
          {/* Sticky table of contents — desktop/widescreen only */}
          <aside className="hidden lg:block">
            <nav aria-label="Guide sections" className="sticky top-24 space-y-1">
              {SECTIONS.map((s) => (
                <a
                  key={s.id}
                  href={`#${s.id}`}
                  className="block rounded-lg px-3 py-2 text-sm font-bold text-sand-600 hover:bg-sand-100 hover:text-teal-800 transition-colors"
                >
                  {s.label}
                </a>
              ))}
            </nav>
          </aside>

          <div className="space-y-12 sm:space-y-14 min-w-0 max-w-3xl">
            <section id="what-is-covered" className="scroll-mt-24 space-y-3">
              <h2 className="font-display text-xl sm:text-2xl font-bold text-sand-900">
                What PMJAY covers
              </h2>
              <p className="text-sm sm:text-base leading-relaxed text-sand-700 font-medium">
                The Pradhan Mantri Jan Arogya Yojana (PMJAY, part of Ayushman Bharat) provides cashless,
                paperless access to a defined list of secondary and tertiary care procedures — the Health
                Benefit Packages (HBP) — at empanelled hospitals, for eligible families. Each package has its
                own indicative rate and eligibility criteria set by the National Health Authority. PMJAY
                Advocate checks your situation against a 315-package index of these (300 independently
                verified) to see whether the procedure you&rsquo;ve described is likely on that list.
              </p>
            </section>

            <section id="your-rights" className="scroll-mt-24 space-y-3">
              <h2 className="font-display text-xl sm:text-2xl font-bold text-sand-900">
                Your core rights at an empanelled hospital
              </h2>
              <ul className="space-y-2.5">
                {[
                  "100% cashless treatment for procedures covered under your card's package list — no upfront payment for the covered portion.",
                  "The hospital must have a dedicated Ayushman Mitra / PMJAY kiosk to help verify your eligibility and process your claim.",
                  "Emergency treatment cannot be delayed or refused while eligibility or pre-authorisation is being verified.",
                  "You're entitled to a clear, specific reason if a claim or procedure is refused — not just a verbal 'not covered.'",
                ].map((point, i) => (
                  <li key={i} className="flex items-start gap-3 text-sm sm:text-base leading-relaxed text-sand-700 font-medium">
                    <IconCheck className="h-4 w-4 sm:h-5 sm:w-5 shrink-0 text-teal-600 mt-0.5" />
                    <span>{point}</span>
                  </li>
                ))}
              </ul>
            </section>

            <section id="not-allowed" className="scroll-mt-24 space-y-3">
              <h2 className="font-display text-xl sm:text-2xl font-bold text-sand-900">
                What hospitals are not allowed to do
              </h2>
              <ul className="space-y-2.5">
                {[
                  "Demand an advance cash deposit before admitting you for a covered procedure.",
                  "Bill you out-of-pocket for a package that's on the HBP list, then call it 'not covered' without a written reason.",
                  "Refuse to discharge you, or hold your documents, over a billing dispute.",
                  "Delay emergency care because a pre-authorisation request is still pending.",
                ].map((point, i) => (
                  <li key={i} className="flex items-start gap-3 text-sm sm:text-base leading-relaxed text-sand-700 font-medium">
                    <IconX className="h-4 w-4 sm:h-5 sm:w-5 shrink-0 text-tier-red-text mt-0.5" />
                    <span>{point}</span>
                  </li>
                ))}
              </ul>
            </section>

            <section id="watch-for" className="scroll-mt-24 space-y-3">
              <h2 className="font-display text-xl sm:text-2xl font-bold text-sand-900">
                Common denial patterns worth double-checking
              </h2>
              <div className="grid gap-3 sm:grid-cols-2">
                <div className="card p-4 sm:p-5">
                  <p className="text-sm font-bold text-sand-900">
                    &ldquo;Our portal is down, pay cash for now&rdquo;
                  </p>
                  <p className="mt-1.5 text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                    A technical delay isn&rsquo;t grounds to bill you directly. Ask them to log the request and
                    proceed with treatment.
                  </p>
                </div>
                <div className="card p-4 sm:p-5">
                  <p className="text-sm font-bold text-sand-900">
                    &ldquo;This isn&rsquo;t in the package list&rdquo;
                  </p>
                  <p className="mt-1.5 text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                    Package names vary by hospital phrasing. Worth checking against the official list before
                    accepting this at face value.
                  </p>
                </div>
                <div className="card p-4 sm:p-5">
                  <p className="text-sm font-bold text-sand-900">
                    Bundling covered and uncovered items
                  </p>
                  <p className="mt-1.5 text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                    A genuinely uncovered add-on (e.g. a cosmetic extra) shouldn&rsquo;t make the hospital bill
                    the whole procedure out-of-pocket.
                  </p>
                </div>
                <div className="card p-4 sm:p-5">
                  <p className="text-sm font-bold text-sand-900">
                    Vague or verbal-only refusal
                  </p>
                  <p className="mt-1.5 text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                    You&rsquo;re entitled to ask for the refusal in writing, with a specific reason attached.
                  </p>
                </div>
              </div>
            </section>

            <section id="what-to-do" className="scroll-mt-24 space-y-3">
              <h2 className="font-display text-xl sm:text-2xl font-bold text-sand-900">
                What to do at the desk, in order
              </h2>
              <ol className="space-y-3">
                {[
                  "Ask calmly for the specific reason for denial, in writing if possible.",
                  "Ask to speak with the hospital's Ayushman Mitra / PMJAY desk specifically, not just general billing staff.",
                  "Describe your situation to PMJAY Advocate to check it against the official package list.",
                  "If it looks wrongly denied, use the generated counter-script and ask for it to be escalated.",
                  "Log the staff name and time of any refusal — it strengthens a formal complaint later.",
                  "If it's still unresolved and treatment is urgent, accept care first and dispute the billing afterward.",
                ].map((step, i) => (
                  <li key={i} className="flex items-start gap-3.5 text-sm sm:text-base leading-relaxed text-sand-700 font-medium">
                    <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-teal-700 text-xs font-bold text-white mt-0.5">
                      {i + 1}
                    </span>
                    <span>{step}</span>
                  </li>
                ))}
              </ol>
            </section>

            <section id="escalate" className="scroll-mt-24 space-y-3">
              <h2 className="font-display text-xl sm:text-2xl font-bold text-sand-900">
                When to escalate to free legal aid
              </h2>
              <p className="text-sm sm:text-base leading-relaxed text-sand-700 font-medium">
                For anything beyond a straightforward billing disagreement — illegal detention of a patient or
                their documents, repeated refusal to provide anything in writing, or outright extortion — NALSA
                (National Legal Services Authority) provides genuinely free legal help, no cost, to families who
                qualify.
              </p>
              <a
                href="tel:15100"
                className="tap-target inline-flex items-center gap-2 rounded-xl bg-tier-handoff-text px-5 py-3 text-sm sm:text-base font-bold text-white transition-all hover:opacity-90 active:scale-95"
              >
                <IconPhone className="h-4 w-4" />
                <span>Call NALSA: 15100 (free, toll-free)</span>
              </a>
            </section>
          </div>
        </div>
      </main>

      <Footer />
    </div>
  );
}
