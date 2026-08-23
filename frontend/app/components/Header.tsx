import Link from "next/link";

export function Header() {
  return (
    <header className="border-b border-sand-200 bg-white">
      <div className="mx-auto flex max-w-2xl items-center justify-between px-4 py-4 sm:px-6">
        <Link href="/" className="flex items-baseline gap-2">
          <span className="font-display text-xl font-semibold text-teal-800">
            PMJAY Advocate
          </span>
        </Link>
        <a
          href="tel:14555"
          className="rounded-lg border border-teal-700 px-3 py-1.5 text-sm font-bold text-teal-800 transition hover:bg-teal-50"
        >
          Helpline: 14555
        </a>
      </div>
    </header>
  );
}
