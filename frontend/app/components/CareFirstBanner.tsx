/**
 * CareFirstBanner renders Section 10's non-negotiable rule. It is placed
 * at the top of every response view, before any tier-specific content,
 * deliberately mirroring the backend's own structural guarantee
 * (backend/internal/response/builder.go: the care-first message is set
 * before the tier switch even runs). The backend can never send a
 * response without this text, but a frontend bug could still render it
 * last, small, or easy to miss — so this component exists as a single,
 * un-skippable place every case page renders through, rather than each
 * page remembering to include a care banner itself.
 */
export function CareFirstBanner({ message }: { message: string }) {
  return (
    <div
      role="alert"
      className="rounded-xl border-2 border-teal-600 bg-teal-800 px-5 py-4 text-white shadow-sm"
    >
      <p className="text-base font-bold leading-snug sm:text-lg">{message}</p>
    </div>
  );
}
