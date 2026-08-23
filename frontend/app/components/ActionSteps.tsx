export function ActionSteps({ steps }: { steps: string[] }) {
  if (steps.length === 0) return null;
  return (
    <section aria-labelledby="action-steps-heading" className="rounded-xl border border-sand-200 bg-white p-5 sm:p-6">
      <h2 id="action-steps-heading" className="text-lg font-bold text-teal-800">
        What to do right now
      </h2>
      <ol className="mt-3 space-y-3">
        {steps.map((step, i) => (
          <li key={i} className="flex gap-3">
            <span
              aria-hidden="true"
              className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-teal-700 text-sm font-bold text-white"
            >
              {i + 1}
            </span>
            <span className="pt-0.5 leading-relaxed text-sand-900">{step}</span>
          </li>
        ))}
      </ol>
    </section>
  );
}
