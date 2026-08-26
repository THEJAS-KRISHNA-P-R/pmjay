import Link from "next/link";
import Image from "next/image";

const SITE_LINKS = [
  { href: "/how-it-works", label: "How It Works" },
  { href: "/guide", label: "Your Rights" },
  { href: "/faq", label: "FAQ" },
  { href: "/about", label: "About" },
];

const LEGAL_LINKS = [
  { href: "/privacy", label: "Privacy Policy" },
  { href: "/terms", label: "Terms of Service" },
  { href: "/disclaimer", label: "Legal & Medical Disclaimer" },
  { href: "/settings", label: "Privacy Settings" },
];

export function Footer() {
  return (
    <footer className="mt-16 sm:mt-24 border-t border-sand-200 bg-sand-25/80 pt-12 sm:pt-16 pb-12">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-8 lg:px-10 space-y-10">
        <div className="grid gap-8 sm:grid-cols-2 md:grid-cols-12">
          {/* Brand & Description (4 cols) */}
          <div className="space-y-3 sm:col-span-2 md:col-span-4">
            <Link href="/" className="inline-flex items-center gap-3">
              <Image
                src="/logo.svg"
                alt="PMJAY Logo"
                width={36}
                height={36}
                className="h-8 w-auto object-contain shrink-0"
              />
              <span className="font-display text-lg font-bold tracking-tight text-sand-900">
                PMJAY Advocate
              </span>
            </Link>
            <p className="text-xs sm:text-sm text-sand-600 leading-relaxed max-w-sm font-medium">
              A free, independent, public-interest software tool helping Indian families verify PMJAY / Ayushman Bharat coverage during hospital billing disputes.
            </p>
            <div className="pt-1">
              <span className="inline-flex items-center gap-1.5 text-[11px] font-bold text-emerald-800 bg-emerald-50 border border-emerald-200/70 px-2.5 py-1 rounded-lg">
                100% Free &amp; Open Access
              </span>
            </div>
          </div>

          {/* Site Pages (2.5 cols) */}
          <div className="space-y-2.5 md:col-span-3">
            <p className="text-xs font-bold uppercase tracking-wider text-sand-900">Explore</p>
            <ul className="space-y-2 text-xs sm:text-sm text-sand-600 font-medium">
              {SITE_LINKS.map((link) => (
                <li key={link.href}>
                  <Link href={link.href} className="hover:text-emerald-800 transition-colors">
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Legal & Trust (2.5 cols) */}
          <div className="space-y-2.5 md:col-span-3">
            <p className="text-xs font-bold uppercase tracking-wider text-sand-900">Legal &amp; Trust</p>
            <ul className="space-y-2 text-xs sm:text-sm text-sand-600 font-medium">
              {LEGAL_LINKS.map((link) => (
                <li key={link.href}>
                  <Link href={link.href} className="hover:text-emerald-800 transition-colors">
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Emergency & Numbers (2 cols) */}
          <div className="space-y-2.5 md:col-span-2">
            <p className="text-xs font-bold uppercase tracking-wider text-sand-900">Helplines</p>
            <ul className="space-y-2 text-xs sm:text-sm text-sand-600 font-medium">
              <li>
                PMJAY:{" "}
                <a href="tel:14555" className="font-bold text-emerald-800 hover:underline">
                  14555
                </a>
              </li>
              <li>
                NALSA Legal Aid:{" "}
                <a href="tel:15100" className="font-bold text-emerald-800 hover:underline">
                  15100
                </a>
              </li>
              <li>
                Emergency:{" "}
                <a href="tel:112" className="font-bold text-sand-900 hover:underline">
                  112
                </a>
              </li>
            </ul>
          </div>
        </div>

        {/* Bottom Legal Disclaimer & Discovery bar */}
        <div className="border-t border-sand-200 pt-6 space-y-4 text-xs text-sand-500 leading-relaxed font-medium">
          <p className="text-center max-w-4xl mx-auto">
            <strong>Legal Notice:</strong> PMJAY Advocate is an independent public-interest resource and is not affiliated with the National Health Authority (NHA), State Health Agencies (SHAs), or any government department. In any medical emergency, seek urgent clinical treatment immediately regardless of billing disputes.
          </p>
          <div className="flex flex-wrap items-center justify-between gap-4 pt-2 text-[11px] border-t border-sand-200/50">
            <p>© {new Date().getFullYear()} PMJAY Advocate. Open-source public interest initiative.</p>
            <div className="flex items-center gap-4">
              <Link href="/llms.txt" className="hover:underline text-sand-600">
                llms.txt
              </Link>
              <Link href="/privacy" className="hover:underline text-sand-600">
                DPDP Act Compliant
              </Link>
              <Link href="/terms" className="hover:underline text-sand-600">
                Terms
              </Link>
              <Link href="/disclaimer" className="hover:underline text-sand-600">
                Disclaimers
              </Link>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}
