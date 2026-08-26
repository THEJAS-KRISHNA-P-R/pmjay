import { IconAlertTriangle } from "./icons";

/**
 * CareFirstBanner renders Section 10's non-negotiable rule. It is placed
 * at the top of every response view, before any tier-specific content,
 * deliberately mirroring the backend's own structural guarantee
 * (backend/internal/response/builder.go: the care-first message is set
 * before the tier switch even runs).
 */
export function CareFirstBanner({ message }: { message: string }) {
  return (
    <div
      role="alert"
      className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-ink-800 via-ink-900 to-ink-950 px-5 py-4 sm:px-6 sm:py-5 text-white shadow-lg animate-fade-in-up"
    >
      <div className="relative flex items-start gap-3.5 sm:gap-4">
        <div className="flex h-9 w-9 sm:h-10 sm:w-10 shrink-0 items-center justify-center rounded-xl bg-white/15 text-ink-100">
          <IconAlertTriangle className="h-5 w-5" />
        </div>
        <div className="flex-1 pt-0.5">
          <p className="text-xs font-bold uppercase tracking-wider text-ink-300">
            Emergency Priority
          </p>
          <p className="mt-1 text-base sm:text-lg font-bold leading-snug sm:leading-normal text-white">
            {message}
          </p>
        </div>
      </div>
    </div>
  );
}
