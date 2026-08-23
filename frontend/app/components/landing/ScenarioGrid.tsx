export function ScenarioGrid() {
  const tiers = [
    {
      badge: "Green Outcome",
      badgeClass: "bg-emerald-100 text-emerald-800",
      title: "Covered Procedures",
      scenario: "Hospital demands ₹35,000 cash for laparoscopic gallbladder surgery or cataract removal.",
      resolution: "Matches official HBP package. Generates written dispute script and CGRMS grievance draft.",
    },
    {
      badge: "Amber Outcome",
      badgeClass: "bg-amber-100 text-amber-900",
      title: "Pre-Auth Pending",
      scenario: "Hospital claims the insurance pre-authorisation portal is delayed or server is down.",
      resolution: "Guides you to verify pre-auth submission without panicking or creating false disputes.",
    },
    {
      badge: "Red Outcome",
      badgeClass: "bg-rose-100 text-rose-900",
      title: "Scheme Exclusions",
      scenario: "Hospital bills out-of-pocket for cosmetic surgery or non-medical aesthetic treatments.",
      resolution: "Provides honest, calm confirmation that this is an authentic national scheme exclusion.",
    },
    {
      badge: "Handoff Outcome",
      badgeClass: "bg-teal-100 text-teal-900",
      title: "Legal Aid Escalation",
      scenario: "Complex illegal detention of patient records or severe unresolvable hospital extortion.",
      resolution: "Prepares structured case file and connects directly with NALSA Para Legal Volunteers on 15100.",
    },
  ];

  return (
    <section aria-labelledby="scenarios-title" className="space-y-6">
      <div className="text-center space-y-2">
        <p className="text-xs font-bold uppercase tracking-wider text-teal-800">
          Clarity Across Every Scenario
        </p>
        <h2 id="scenarios-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-teal-950">
          How Different Cases Are Handled
        </h2>
        <p className="text-sm sm:text-base text-sand-900/70 max-w-xl mx-auto font-medium">
          We never generate fake grievances for valid exclusions, and we never let covered care get wrongfully billed.
        </p>
      </div>

      <div className="grid gap-5 sm:grid-cols-2 pt-2">
        {tiers.map((t) => (
          <div
            key={t.title}
            className="glass-panel p-6 sm:p-7 flex flex-col justify-between space-y-4"
          >
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className={`inline-flex rounded-full px-3 py-1 text-xs font-bold ${t.badgeClass}`}>
                  {t.badge}
                </span>
              </div>
              <h3 className="font-display text-lg font-bold text-teal-950">
                {t.title}
              </h3>
              <p className="text-xs sm:text-sm text-sand-900/75 leading-relaxed font-medium">
                <strong className="text-sand-900">Hospital Situation: </strong>
                {t.scenario}
              </p>
            </div>
            <div className="rounded-xl bg-sand-100/60 p-3.5 text-xs text-sand-900/80 leading-relaxed font-medium">
              <strong className="text-teal-900">Advocate Action: </strong>
              {t.resolution}
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
