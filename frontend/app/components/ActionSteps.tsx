export function ActionSteps({ steps }: { steps: string[] }) {
  if (steps.length === 0) return null;
  return (
    <section 
      aria-labelledby="action-steps-heading" 
      className="glass-panel p-6 sm:p-8 animate-enter"
    >
      <div className="flex items-center gap-2.5">
        <div className="flex h-6 w-6 items-center justify-center rounded-lg bg-teal-100 text-teal-800">
          <svg
            className="h-3.5 w-3.5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="3"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <polyline points="20 6 9 17 4 12" />
          </svg>
        </div>
        <h2 id="action-steps-heading" className="text-lg sm:text-xl font-bold tracking-tight text-teal-950">
          What to do right now
        </h2>
      </div>
      
      <ol className="mt-5 space-y-3">
        {steps.map((step, i) => (
          <li 
            key={i} 
            className="flex items-start gap-3.5 rounded-2xl bg-sand-100/60 p-4 sm:p-5 text-sm sm:text-base transition hover:bg-sand-100/90 shadow-2xs"
          >
            <span
              aria-hidden="true"
              className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-teal-800 text-xs font-bold text-white shadow-xs mt-0.5"
            >
              {i + 1}
            </span>
            <span className="leading-relaxed text-sand-900 font-medium">{step}</span>
          </li>
        ))}
      </ol>
    </section>
  );
}
