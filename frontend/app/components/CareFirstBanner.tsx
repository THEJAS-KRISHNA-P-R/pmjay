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
      className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-teal-900 via-teal-900 to-teal-950 px-5 py-4 sm:px-6 sm:py-5 text-white shadow-lg backdrop-blur-xl animate-enter"
    >
      <div className="relative flex items-start gap-3.5 sm:gap-4">
        <div className="flex h-9 w-9 sm:h-10 sm:w-10 shrink-0 items-center justify-center rounded-xl bg-white/15 text-teal-100 shadow-inner">
          <svg 
            className="h-5 w-5" 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor" 
            strokeWidth="2.5" 
            strokeLinecap="round" 
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <path d="M12 9v4" />
            <path d="M12 17h.01" />
            <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
          </svg>
        </div>
        <div className="flex-1 pt-0.5">
          <p className="text-xs font-bold uppercase tracking-wider text-teal-300">
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
