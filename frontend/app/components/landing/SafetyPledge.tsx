import { IconShieldCheck, IconLock, IconScale, type IconProps } from "../icons";
import type { ComponentType } from "react";

export function SafetyPledge() {
  const pledges: { title: string; desc: string; Icon: ComponentType<IconProps> }[] = [
    {
      title: "Care Comes First, Always",
      desc: "If medical treatment is urgent, never delay or refuse admission over financial disputes. Settle the billing afterward.",
      Icon: IconShieldCheck,
    },
    {
      title: "Zero Personal Data Exploitation",
      desc: "No phone numbers, Aadhaar numbers, or logins required to check coverage. Your clinical queries stay private.",
      Icon: IconLock,
    },
    {
      title: "Grounded in Official NHA HBP Data",
      desc: "Recommendations are checked against the National Health Authority's Health Benefit Package schedule, with unverified entries flagged rather than hidden.",
      Icon: IconScale,
    },
  ];

  return (
    <section
      aria-labelledby="safety-pledge-title"
      className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-ink-800 via-ink-900 to-ink-950 p-7 sm:p-10 text-white shadow-xl"
    >
      <div className="relative space-y-6">
        <div className="max-w-xl space-y-2">
          <div className="badge bg-white/10 border-white/15 text-ink-200">
            Safety &amp; Ethical Commitment
          </div>
          <h2 id="safety-pledge-title" className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-white">
            Our Public-Interest Guarantee
          </h2>
          <p className="text-sm sm:text-base text-ink-100/85 leading-relaxed font-medium">
            Standing at a billing desk when a loved one is unwell is distressing. We ensure you get calm, factual answers without alarmism.
          </p>
        </div>

        <div className="grid gap-4 sm:grid-cols-3">
          {pledges.map((p) => (
            <div key={p.title} className="rounded-2xl bg-white/10 border border-white/15 p-5 space-y-2.5 shadow-[inset_0_1px_1px_rgba(255,255,255,0.12)]">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-white/15 text-ink-100">
                <p.Icon className="h-4 w-4" />
              </div>
              <h3 className="font-display text-base font-bold text-white">
                {p.title}
              </h3>
              <p className="text-xs text-ink-100/80 leading-relaxed font-medium">
                {p.desc}
              </p>
            </div>
          ))}
        </div>

        <div className="flex flex-wrap items-center justify-between gap-4 border-t border-white/15 pt-6 text-xs sm:text-sm text-ink-100/90 font-medium">
          <p>National PMJAY Toll-Free Helpline: <strong className="text-white">14555</strong></p>
          <p>NALSA Free Legal Services: <strong className="text-white">15100</strong></p>
        </div>
      </div>
    </section>
  );
}
