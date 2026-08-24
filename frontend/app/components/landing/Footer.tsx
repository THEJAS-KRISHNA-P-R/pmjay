import Link from "next/link";
import Image from "next/image";

const SITE_LINKS = [
  { href: "/how-it-works", label: "How It Works" },
  { href: "/guide", label: "Your Rights" },
  { href: "/faq", label: "FAQ" },
  { href: "/about", label: "About" },
];

export function Footer() {
  return (
    <footer className="mt-16 sm:mt-24 border-t border-sand-200 bg-sand-25/80 pt-12 sm:pt-16 pb-12">
      <div className="mx-auto w-full max-w-6xl px-6 sm:px-8 lg:px-10 space-y-8">
        <div className="grid gap-8 sm:grid-cols-2 md:grid-cols-4">
          <div className="space-y-3 sm:col-span-2">
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
            <p className="text-xs sm:text-sm text-sand-600 leading-relaxed max-w-md font-medium">
              A free, independent public-interest tool helping families verify PMJAY / Ayushman Bharat
              coverage when facing hospital billing disputes.
            </p>
          </div>

          <div className="space-y-2.5">
            <p className="text-xs font-bold uppercase tracking-wider text-sand-900">Site Pages</p>
            <ul className="space-y-2 text-xs sm:text-sm text-sand-600 font-medium">
              {SITE_LINKS.map((link) => (
                <li key={link.href}>
                  <Link href={link.href} className="hover:text-teal-800 transition-colors">
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div className="space-y-2.5">
            <p className="text-xs font-bold uppercase tracking-wider text-sand-900">Important Numbers</p>
            <ul className="space-y-2 text-xs sm:text-sm text-sand-600 font-medium">
              <li>
                PMJAY Helpline:{" "}
                <a href="tel:14555" className="font-bold text-teal-800 hover:underline">
                  14555
                </a>
              </li>
              <li>
                NALSA Free Legal Aid:{" "}
                <a href="tel:15100" className="font-bold text-teal-800 hover:underline">
                  15100
                </a>
              </li>
              <li>
                Emergency:{" "}
                <a href="tel:112" className="font-bold text-sand-800 hover:underline">
                  112
                </a>
              </li>
            </ul>
          </div>
        </div>

        <div className="border-t border-sand-200 pt-6 text-center text-xs text-sand-500 leading-relaxed font-medium">
          <p>
            PMJAY Advocate is an independent public-interest resource and is not affiliated with the
            National Health Authority or any government body. In any medical emergency, seek urgent
            treatment immediately regardless of billing disputes.
          </p>
        </div>
      </div>
    </footer>
  );
}
