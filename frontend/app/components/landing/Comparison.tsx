export function Comparison() {
  const rows = [
    {
      capability: "Point-of-Denial Guidance",
      advocate: "Instant verification of 1,949+ official NHA packages with indicative ceiling rates",
      typical: "Uncertainty at the desk, forced into taking high-interest medical loans",
    },
    {
      capability: "Hospital Counter-Script",
      advocate: "Exact polite, legally grounded words to say to billing staff to request written refusal",
      typical: "Vague arguments with receptionists with no official scheme citations",
    },
    {
      capability: "Official Grievance Draft",
      advocate: "Ready-to-file CGRMS complaint draft formatted for the official Ayushman App",
      typical: "Tedious manual drafting without required package codes or evidence logs",
    },
    {
      capability: "Emergency Protocol",
      advocate: "Care-First priority banner: 'Get treatment first, dispute money after. Always.'",
      typical: "Treatment delays caused by billing disputes while patient condition worsens",
    },
    {
      capability: "Legal Aid Escalation",
      advocate: "Structured case brief and direct connection to NALSA Para Legal Volunteers (15100)",
      typical: "Private legal consultations costing thousands of rupees out of pocket",
    },
    {
      capability: "Privacy & Access",
      advocate: "100% Free public service, zero login, no phone number, no tracking",
      typical: "Data-harvesting commercial aggregators requiring logins and selling leads",
    },
  ];

  return (
    <section aria-labelledby="comparison-title" className="space-y-6">
      <div className="text-center space-y-2">
        <p className="text-xs font-bold uppercase tracking-wider text-teal-800">
          Why PMJAY Advocate
        </p>
        <h2 id="comparison-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-teal-950">
          Built Specifically for the Point of Denial
        </h2>
        <p className="text-sm sm:text-base text-sand-900/70 max-w-xl mx-auto font-medium">
          When a hospital billing desk refuses your Ayushman card, generic search engines don&rsquo;t help. You need instant, package-specific facts.
        </p>
      </div>

      <div className="glass-panel overflow-hidden">
        <div className="overflow-x-auto">
          <div className="min-w-[640px]">
            {/* Table Header */}
            <div className="grid grid-cols-[1.2fr_1.8fr_1.6fr] bg-sand-100/70 p-4 sm:px-6 text-xs font-bold uppercase tracking-wider text-sand-900/70">
              <div>Capability</div>
              <div className="text-teal-900 font-extrabold">PMJAY Advocate</div>
              <div>Typical Experience</div>
            </div>

            {/* Table Rows */}
            <div className="divide-y divide-sand-200/40">
              {rows.map((row, idx) => (
                <div
                  key={idx}
                  className="grid grid-cols-[1.2fr_1.8fr_1.6fr] p-4 sm:px-6 sm:py-5 text-xs sm:text-sm items-start gap-4 transition hover:bg-white/60"
                >
                  <div className="font-bold text-teal-950">
                    {row.capability}
                  </div>
                  <div className="flex items-start gap-2.5 text-sand-900 font-medium bg-teal-50/50 p-2.5 rounded-xl">
                    <span className="flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-teal-700 text-white font-bold text-[10px] mt-0.5 shadow-2xs">
                      ✓
                    </span>
                    <span>{row.advocate}</span>
                  </div>
                  <div className="flex items-start gap-2 text-sand-900/60 font-medium p-2.5">
                    <span className="text-sand-900/40 text-xs mt-0.5">✕</span>
                    <span>{row.typical}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
