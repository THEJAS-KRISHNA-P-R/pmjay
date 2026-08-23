export function HowItWorks() {
  const steps = [
    {
      num: "01",
      title: "Describe the situation",
      desc: "Tell us what the hospital billing staff is asking for in natural language — English, Malayalam, Hindi, or any mix. No technical medical terms needed.",
    },
    {
      num: "02",
      title: "Instant HBP verification",
      desc: "Our engine cross-checks your procedure against 1,949+ official PMJAY Health Benefit Packages, indicative rates, and exclusion criteria in real time.",
    },
    {
      num: "03",
      title: "Get exact words & action steps",
      desc: "Receive clear hospital counter-scripts, ready-to-file CGRMS grievance drafts, and a downloadable PDF document to hand to hospital authorities.",
    },
  ];

  return (
    <section aria-labelledby="how-it-works-title" className="space-y-6">
      <div className="text-center space-y-2">
        <p className="text-xs font-bold uppercase tracking-wider text-teal-800">
          Simple 3-Step Process
        </p>
        <h2 id="how-it-works-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-teal-950">
          How PMJAY Advocate Protects You
        </h2>
        <p className="text-sm sm:text-base text-sand-900/70 max-w-xl mx-auto font-medium">
          Designed specifically for stressful moments at hospital billing desks when you need immediate clarity.
        </p>
      </div>

      <div className="grid gap-5 sm:grid-cols-3 pt-2">
        {steps.map((step) => (
          <div
            key={step.num}
            className="glass-panel p-6 sm:p-7 flex flex-col justify-between transition-transform hover:-translate-y-0.5"
          >
            <div className="space-y-3">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-teal-800 text-white font-mono text-xs font-bold shadow-xs">
                {step.num}
              </span>
              <h3 className="font-display text-lg font-bold text-teal-950">
                {step.title}
              </h3>
              <p className="text-xs sm:text-sm text-sand-900/75 leading-relaxed font-medium">
                {step.desc}
              </p>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
