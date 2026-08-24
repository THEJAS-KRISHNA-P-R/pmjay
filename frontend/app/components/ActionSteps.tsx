import { IconCheck } from "./icons";

export function ActionSteps({ steps }: { steps: string[] }) {
  if (steps.length === 0) return null;
  return (
    <section
      aria-labelledby="action-steps-heading"
      className="card p-6 sm:p-8 animate-fade-in-up"
    >
      <div className="flex items-center gap-2.5">
        <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-teal-100 text-teal-800">
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
            className="flex items-start gap-3.5 rounded-2xl bg-sand-50 border border-sand-200/70 p-4 sm:p-5 text-sm sm:text-base transition-colors hover:bg-teal-50/60 hover:border-teal-200"
          >
            <span
              aria-hidden="true"
              className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-teal-700 text-xs font-bold text-white mt-0.5"
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
