import { IconCheck, IconX } from "../icons";

export function Comparison() {
  const rows = [
    {
      capability: "Point-of-Denial Guidance",
      advocate: "Instant verification against the 315-package indicative HBP rate index",
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
      advocate: "100% free public service, zero login, no phone number, no tracking",
      typical: "Data-harvesting commercial aggregators requiring logins and selling leads",
    },
  ];

  return (
    <section aria-labelledby="comparison-title" className="space-y-8">
      <div className="text-center space-y-2">
        <p className="text-xs font-bold uppercase tracking-wider text-teal-700">
          Why PMJAY Advocate
        </p>
        <h2 id="comparison-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-sand-900">
          Built Specifically for the Point of Denial
        </h2>
        <p className="text-sm sm:text-base text-sand-600 max-w-xl mx-auto font-medium">
          When a hospital billing desk refuses your Ayushman card, generic search engines don&rsquo;t help. You need instant, package-specific facts.
        </p>
      </div>

      <div className="card overflow-hidden">
        <div className="overflow-x-auto">
          <div className="min-w-[640px]">
            {/* Table Header */}
            <div className="grid grid-cols-[1.2fr_1.8fr_1.6fr] bg-sand-50 p-4 sm:px-6 text-xs font-bold uppercase tracking-wider text-sand-500 border-b border-sand-200">
              <div>Capability</div>
              <div className="text-teal-800 font-extrabold">PMJAY Advocate</div>
              <div>Typical Experience</div>
            </div>

            {/* Table Rows */}
            <div className="divide-y divide-sand-200/70">
              {rows.map((row, idx) => (
                <div
                  key={idx}
                  className="grid grid-cols-[1.2fr_1.8fr_1.6fr] p-4 sm:px-6 sm:py-5 text-xs sm:text-sm items-start gap-4 transition-colors hover:bg-sand-50/70"
                >
                  <div className="font-bold text-sand-900">
                    {row.capability}
                  </div>
                  <div className="flex items-start gap-2.5 text-sand-800 font-medium bg-teal-50 p-2.5 rounded-xl">
                    <IconCheck className="h-4 w-4 shrink-0 text-teal-700 mt-0.5" />
                    <span>{row.advocate}</span>
                  </div>
                  <div className="flex items-start gap-2 text-sand-500 font-medium p-2.5">
                    <IconX className="h-3.5 w-3.5 shrink-0 text-sand-400 mt-0.5" />
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
