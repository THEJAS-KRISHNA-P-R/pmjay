import Link from "next/link";

export function Header() {
  return (
    <header className="sticky top-0 z-50 bg-white/75 backdrop-blur-2xl shadow-[0_4px_20px_-4px_rgba(13,54,52,0.04)] transition-all">
      <div className="mx-auto flex max-w-2xl items-center justify-between px-4 py-3.5 sm:px-6 sm:py-4">
        <Link 
          href="/" 
          className="group flex items-center gap-2.5 transition touch-spring focus-visible:outline-teal-600"
        >
          <span className="flex h-7 w-7 items-center justify-center rounded-xl bg-teal-800 text-white font-display text-sm font-bold shadow-xs">
            P
          </span>
          <span className="font-display text-xl font-semibold tracking-tight-display text-teal-950 group-hover:text-teal-700 transition-colors">
            PMJAY Advocate
          </span>
        </Link>
        <a
          href="tel:14555"
          className="inline-flex items-center gap-2 rounded-full bg-teal-50 px-4 py-2 text-xs sm:text-sm font-bold text-teal-800 transition touch-spring hover:bg-teal-100 shadow-[inset_0_1px_0_0_rgba(255,255,255,0.8)] focus-visible:outline-teal-600"
        >
          <span className="h-2 w-2 rounded-full bg-teal-600 animate-pulse" aria-hidden="true" />
          <span>Helpline: 14555</span>
        </a>
      </div>
    </header>
  );
}
