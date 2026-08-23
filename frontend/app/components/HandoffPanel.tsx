export function HandoffPanel({ summary }: { summary: string }) {
  return (
    <section className="rounded-2xl border-2 border-tier-handoff-border bg-tier-handoff-bg/90 p-5 sm:p-7 shadow-sm backdrop-blur-md animate-enter">
      <div className="flex items-center gap-2.5">
        <div className="flex h-6 w-6 items-center justify-center rounded-md bg-teal-100 text-teal-800">
          <svg
            className="h-3.5 w-3.5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
            <circle cx="9" cy="7" r="4" />
            <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
            <path d="M16 3.13a4 4 0 0 1 0 7.75" />
          </svg>
        </div>
        <h2 className="text-lg sm:text-xl font-bold tracking-tight text-tier-handoff-strong">
          Free legal help, right now
        </h2>
      </div>
      
      <p className="mt-3 text-sm sm:text-base leading-relaxed text-sand-900">
        NALSA (the National Legal Services Authority) provides free legal help to families who
        can&rsquo;t afford a lawyer, including for exactly this kind of situation. A Para Legal
        Volunteer can help you in person or over the phone, at no cost.
      </p>

      <div className="mt-4">
        <a
          href="tel:15100"
          className="inline-flex items-center gap-2 rounded-xl bg-tier-handoff-strong px-5 py-3 text-sm sm:text-base font-bold text-white shadow-sm transition touch-spring hover:opacity-95 active:scale-95 focus-visible:outline-teal-600"
        >
          <svg
            className="h-4 w-4"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z" />
          </svg>
          <span>Call 15100 (free, toll-free)</span>
        </a>
      </div>

      <div className="mt-5 rounded-xl border border-tier-handoff-border/80 bg-white/80 p-4 sm:p-5 shadow-2xs">
        <p className="text-xs sm:text-sm font-bold text-tier-handoff-text">
          What we&rsquo;ll have ready to share, so you don&rsquo;t have to repeat it
        </p>
        <p className="mt-2 whitespace-pre-line text-xs sm:text-sm leading-relaxed text-sand-900/90 font-medium">
          {summary}
        </p>
      </div>
    </section>
  );
}
