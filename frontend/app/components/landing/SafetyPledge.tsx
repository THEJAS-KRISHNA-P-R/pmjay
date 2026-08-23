export function SafetyPledge() {
  const pledges = [
    {
      title: "Care Comes First, Always",
      desc: "If medical treatment is urgent, never delay or refuse admission over financial disputes. Settle the billing afterward.",
    },
    {
      title: "Zero Personal Data Exploitation",
      desc: "No phone numbers, Aadhaar numbers, or logins required to check coverage. Your clinical queries stay private.",
    },
    {
      title: "Grounded in Official NHA HBP Data",
      desc: "All recommendations are strictly verified against the National Health Authority Health Benefit Package 2022 schedule.",
    },
  ];

  return (
    <section 
      aria-labelledby="safety-pledge-title"
      className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-teal-900 via-teal-900 to-teal-950 p-7 sm:p-10 text-white shadow-xl backdrop-blur-2xl"
    >
      <div className="relative space-y-6">
        <div className="max-w-xl space-y-2">
          <div className="inline-flex items-center rounded-full bg-teal-800/80 px-3.5 py-1 text-xs font-bold uppercase tracking-wider text-teal-300">
            Safety & Ethical Commitment
          </div>
          <h2 id="safety-pledge-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-white">
            Our Public-Interest Guarantee
          </h2>
          <p className="text-sm sm:text-base text-teal-100/80 leading-relaxed font-medium">
            Standing at a billing desk when a loved one is unwell is distressing. We ensure you get calm, factual answers without alarmism.
          </p>
        </div>

        <div className="grid gap-4 sm:grid-cols-3 pt-2">
          {pledges.map((p) => (
            <div
              key={p.title}
              className="rounded-2xl bg-white/10 p-5 backdrop-blur-md shadow-xs space-y-2"
            >
              <h3 className="font-display text-base font-bold text-teal-150 text-white">
                {p.title}
              </h3>
              <p className="text-xs text-teal-100/75 leading-relaxed font-medium">
                {p.desc}
              </p>
            </div>
          ))}
        </div>

        <div className="flex flex-wrap items-center justify-between gap-4 border-t border-teal-800/80 pt-6 text-xs sm:text-sm text-teal-200/90 font-medium">
          <p>National PMJAY Toll-Free Helpline: <strong className="text-white">14555</strong></p>
          <p>NALSA Free Legal Services: <strong className="text-white">15100</strong></p>
        </div>
      </div>
    </section>
  );
}
