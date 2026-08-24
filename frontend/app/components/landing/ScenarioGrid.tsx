import { IconCheck, IconHelpCircle, IconInfo, IconLifeBuoy, type IconProps } from "../icons";
import type { ComponentType } from "react";

export function ScenarioGrid() {
  const tiers: {
    badge: string;
    Icon: ComponentType<IconProps>;
    iconBg: string;
    iconBorder: string;
    iconText: string;
    title: string;
    scenario: string;
    resolution: string;
  }[] = [
    {
      badge: "Green Outcome",
      Icon: IconCheck,
      iconBg: "bg-tier-green-icon",
      iconBorder: "border-tier-green-border",
      iconText: "text-tier-green-text",
      title: "Covered Procedures",
      scenario: "Hospital demands ₹35,000 cash for laparoscopic gallbladder surgery or cataract removal.",
      resolution: "Matches official HBP package. Generates written dispute script and CGRMS grievance draft.",
    },
    {
      badge: "Amber Outcome",
      Icon: IconHelpCircle,
      iconBg: "bg-tier-amber-icon",
      iconBorder: "border-tier-amber-border",
      iconText: "text-tier-amber-text",
      title: "Pre-Auth Pending",
      scenario: "Hospital claims the insurance pre-authorisation portal is delayed or server is down.",
      resolution: "Guides you to verify pre-auth submission without panicking or creating false disputes.",
    },
    {
      badge: "Red Outcome",
      Icon: IconInfo,
      iconBg: "bg-tier-red-icon",
      iconBorder: "border-tier-red-border",
      iconText: "text-tier-red-text",
      title: "Scheme Exclusions",
      scenario: "Hospital bills out-of-pocket for cosmetic surgery or non-medical aesthetic treatments.",
      resolution: "Provides honest, calm confirmation that this is an authentic national scheme exclusion.",
    },
    {
      badge: "Handoff Outcome",
      Icon: IconLifeBuoy,
      iconBg: "bg-tier-handoff-icon",
      iconBorder: "border-tier-handoff-border",
      iconText: "text-tier-handoff-text",
      title: "Legal Aid Escalation",
      scenario: "Complex illegal detention of patient records or severe unresolvable hospital extortion.",
      resolution: "Prepares structured case file and connects directly with NALSA Para Legal Volunteers on 15100.",
    },
  ];

  return (
    <section aria-labelledby="scenarios-title" className="space-y-8">
      <div className="text-center space-y-2">
        <p className="text-xs font-bold uppercase tracking-wider text-teal-700">
          Clarity Across Every Scenario
        </p>
        <h2 id="scenarios-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-sand-900">
          How Different Cases Are Handled
        </h2>
        <p className="text-sm sm:text-base text-sand-600 max-w-xl mx-auto font-medium">
          We never generate fake grievances for valid exclusions, and we never let covered care get wrongfully billed.
        </p>
      </div>

      <div className="grid gap-5 sm:grid-cols-2">
        {tiers.map((t) => (
          <div key={t.title} className="card p-6 sm:p-7 flex flex-col justify-between space-y-4">
            <div className="space-y-3">
              <span className={`badge ${t.iconBg} ${t.iconBorder} ${t.iconText}`}>
                <t.Icon className="h-3 w-3" />
                {t.badge}
              </span>
              <h3 className="font-display text-lg font-bold text-sand-900">
                {t.title}
              </h3>
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed font-medium">
                <strong className="text-sand-800">Hospital Situation: </strong>
                {t.scenario}
              </p>
            </div>
            <div className="rounded-xl bg-sand-50 border border-sand-200/70 p-3.5 text-xs text-sand-700 leading-relaxed font-medium">
              <strong className="text-teal-800">Advocate Action: </strong>
              {t.resolution}
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
