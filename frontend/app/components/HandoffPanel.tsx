import { IconUsers, IconPhone } from "./icons";

export function HandoffPanel({ summary }: { summary: string }) {
  return (
    <section className="rounded-2xl border-2 border-tier-handoff-border bg-tier-handoff-bg p-5 sm:p-7 shadow-sm animate-fade-in-up">
      <div className="flex items-center gap-2.5">
        <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-tier-handoff-icon text-tier-handoff-text">
          <IconUsers className="h-4 w-4" />
        </div>
        <h2 className="text-lg sm:text-xl font-bold tracking-tight text-tier-handoff-text">
          Free legal help, right now
        </h2>
      </div>

      <p className="mt-3 text-sm sm:text-base leading-relaxed text-sand-800">
        NALSA (the National Legal Services Authority) provides free legal help to families who
        can&rsquo;t afford a lawyer, including for exactly this kind of situation. A Para Legal
        Volunteer can help you in person or over the phone, at no cost.
      </p>

      <div className="mt-4">
        <a
          href="tel:15100"
          className="tap-target inline-flex items-center gap-2 rounded-xl bg-tier-handoff-text px-5 py-3 text-sm sm:text-base font-bold text-white transition-all hover:opacity-90 active:scale-95"
        >
          <IconPhone className="h-4 w-4" />
          <span>Call 15100 (free, toll-free)</span>
        </a>
      </div>

      <div className="mt-5 rounded-xl border border-tier-handoff-border bg-white/70 p-4 sm:p-5">
        <p className="text-xs sm:text-sm font-bold text-tier-handoff-text">
          What we&rsquo;ll have ready to share, so you don&rsquo;t have to repeat it
        </p>
        <p className="mt-2 whitespace-pre-line text-xs sm:text-sm leading-relaxed text-sand-800 font-medium">
          {summary}
        </p>
      </div>
    </section>
  );
}
