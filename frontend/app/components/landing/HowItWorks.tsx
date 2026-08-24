export function HowItWorks() {
  const steps = [
    {
      num: "01",
      title: "Describe the situation",
      desc: "Tell us what the hospital billing staff is asking for in English, native Malayalam script (മലയാളം), or Hindi. (Note: Transliterated Malayalam typed in English letters is not supported).",
    },
    {
      num: "02",
      title: "Instant HBP verification",
      desc: "Our engine cross-checks your procedure against the 315-package PMJAY Health Benefit index, indicative rates, and exclusion criteria in real time.",
    },
    {
      num: "03",
      title: "Get exact words & action steps",
      desc: "Receive clear hospital counter-scripts, ready-to-file CGRMS grievance drafts, and a downloadable PDF document to hand to hospital authorities.",
    },
  ];

  return (
    <section aria-labelledby="how-it-works-title" className="space-y-8">
      <div className="text-center space-y-2">
        <p className="text-xs font-bold uppercase tracking-wider text-teal-700">
          Simple 3-Step Process
        </p>
        <h2 id="how-it-works-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-sand-900">
          How PMJAY Advocate Protects You
        </h2>
        <p className="text-sm sm:text-base text-sand-600 max-w-xl mx-auto font-medium">
          Designed specifically for stressful moments at hospital billing desks when you need immediate clarity.
        </p>
      </div>

      <div className="grid gap-5 sm:grid-cols-3">
        {steps.map((step, i) => (
          <div key={step.num} className="relative">
            {/* Connecting line between steps, desktop only */}
            {i < steps.length - 1 && (
              <div
                aria-hidden="true"
                className="hidden sm:block absolute top-[1.6rem] left-[calc(50%+1.5rem)] right-[calc(-50%+1.5rem)] h-px bg-sand-200"
              />
            )}
            <div className="card relative p-6 sm:p-7 h-full flex flex-col gap-3">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-teal-700 text-white font-mono text-xs font-bold">
                {step.num}
              </span>
              <h3 className="font-display text-lg font-bold text-sand-900">
                {step.title}
              </h3>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                {step.desc}
              </p>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
