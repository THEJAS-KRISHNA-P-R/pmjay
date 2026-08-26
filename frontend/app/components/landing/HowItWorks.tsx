export function HowItWorks() {
  const steps = [
    {
      num: "1",
      title: "Describe the situation",
      desc: "Tell us what the hospital billing staff is asking for in English, your own regional script, or transliterated letters. Mixing languages in one message is fine.",
    },
    {
      num: "2",
      title: "Instant HBP verification",
      desc: "Our engine cross-checks your procedure against the 315-package PMJAY Health Benefit index, indicative rates, and exclusion criteria in real time.",
    },
    {
      num: "3",
      title: "Get exact words & action steps",
      desc: "Receive clear hospital counter-scripts, ready-to-file CGRMS grievance drafts, and a downloadable PDF document to hand to hospital authorities.",
    },
  ];

  return (
    <section aria-labelledby="how-it-works-title" className="space-y-6 sm:space-y-8">
      <div className="text-center space-y-2">
        <p className="text-xs font-bold uppercase tracking-wider text-sand-500">
          Simple 3-Step Process
        </p>
        <h2 id="how-it-works-title" className="font-display text-2xl sm:text-3xl lg:text-4xl font-semibold tracking-tight-display text-sand-900">
          How PMJAY Advocate Protects You
        </h2>
        <p className="text-sm sm:text-base text-sand-600 max-w-xl mx-auto font-medium">
          Designed specifically for stressful moments at hospital billing desks when you need immediate clarity.
        </p>
      </div>

      {/* Mobile: Unified, clean connected timeline (avoids repetitive boxy cards) */}
      <div className="sm:hidden card p-5 divide-y divide-sand-100">
        {steps.map((step, index) => (
          <div key={step.num} className={`flex gap-4 ${index > 0 ? "pt-4" : ""} ${index < steps.length - 1 ? "pb-4" : ""}`}>
            <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-emerald-100/80 text-emerald-800 font-bold text-xs">
              {step.num}
            </span>
            <div className="space-y-1">
              <h3 className="font-display text-base font-bold text-sand-900">
                {step.title}
              </h3>
              <p className="text-xs text-sand-600 leading-relaxed font-medium">
                {step.desc}
              </p>
            </div>
          </div>
        ))}
      </div>

      {/* Desktop (sm+): Interactive 3-column card grid */}
      <div className="hidden sm:grid gap-5 sm:grid-cols-3">
        {steps.map((step) => (
          <div
            key={step.num}
            className="card card-interactive group relative p-6 sm:p-7 flex flex-col justify-between overflow-hidden min-h-[180px]"
          >
            <div className="relative space-y-2.5 z-10">
              <h3 className="font-display text-lg sm:text-xl font-bold text-sand-900 leading-snug pr-8">
                {step.title}
              </h3>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                {step.desc}
              </p>
            </div>

            {/* Lower watermark number */}
            <span
              aria-hidden="true"
              className="absolute -bottom-3 right-3 font-display text-6xl sm:text-7xl font-extrabold text-sand-200/50 select-none group-hover:text-emerald-700/15 group-hover:scale-105 transition-all duration-300 pointer-events-none"
            >
              {step.num}
            </span>
          </div>
        ))}
      </div>
    </section>
  );
}
