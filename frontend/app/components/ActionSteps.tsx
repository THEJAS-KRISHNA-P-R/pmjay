import { IconCheck } from "./icons";

export function ActionSteps({ steps }: { steps: string[] }) {
  if (steps.length === 0) return null;
  return (
    <section
      aria-labelledby="action-steps-heading"
      className="card p-6 sm:p-8 animate-fade-in-up"
    >
      <div className="flex items-center gap-2.5">
        <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-sand-100 border border-sand-200/60 text-sand-700 shadow-[inset_0_1px_1px_rgba(255,255,255,0.9),0_1px_2px_rgba(42,38,33,0.04)]">
          <IconCheck className="h-4 w-4" />
        </div>
        <h2 id="action-steps-heading" className="text-lg sm:text-xl font-bold tracking-tight text-sand-900">
          What to do right now
        </h2>
      </div>

      <ol className="mt-5 space-y-3">
        {steps.map((step, i) => (
          <li
            key={i}
            className="flex items-start gap-3.5 rounded-2xl bg-white border border-sand-200/80 p-4 sm:p-5 text-sm sm:text-base transition-colors hover:bg-sand-50 shadow-[0_1px_4px_rgba(42,38,33,0.05),inset_0_1px_1px_rgba(255,255,255,0.9)]"
          >
            <span
              aria-hidden="true"
              className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-ink-700 text-xs font-bold text-white mt-0.5 shadow-[0_1px_3px_rgba(42,38,33,0.2)]"
            >
              {i + 1}
            </span>
            <span className="leading-relaxed text-sand-800 font-medium">{step}</span>
          </li>
        ))}
      </ol>
    </section>
  );
}
