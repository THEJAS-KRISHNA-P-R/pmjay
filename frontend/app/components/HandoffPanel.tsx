export function HandoffPanel({ summary }: { summary: string }) {
  return (
    <section className="rounded-xl border-2 border-tier-handoff-border bg-tier-handoff-bg p-5 sm:p-6">
      <h2 className="text-lg font-bold text-tier-handoff-strong">Free legal help, right now</h2>
      <p className="mt-2 leading-relaxed text-sand-900">
        NALSA (the National Legal Services Authority) provides free legal help to families who
        can&rsquo;t afford a lawyer, including for exactly this kind of situation. A Para Legal
        Volunteer can help you in person or over the phone, at no cost.
      </p>

      <a
        href="tel:15100"
        className="mt-4 inline-flex items-center gap-2 rounded-lg bg-tier-handoff-strong px-5 py-3 font-bold text-white transition hover:opacity-90"
      >
        Call 15100 (free, toll-free)
      </a>

      <div className="mt-5 rounded-lg border border-tier-handoff-border bg-white/60 p-4">
        <p className="text-sm font-bold text-tier-handoff-text">
          What we&rsquo;ll have ready to share, so you don&rsquo;t have to repeat it
        </p>
        <p className="mt-2 whitespace-pre-line text-sm leading-relaxed text-sand-900/90">
          {summary}
        </p>
      </div>
    </section>
  );
}
