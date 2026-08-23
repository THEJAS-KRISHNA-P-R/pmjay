import Link from "next/link";

export function Footer() {
  return (
    <footer className="bg-sand-100/50 pt-12 pb-8 border-t border-sand-200/50 mt-16">
      <div className="mx-auto max-w-4xl px-4 sm:px-6 space-y-8">
        <div className="grid gap-8 sm:grid-cols-3">
          <div className="space-y-3">
            <Link href="/" className="inline-flex items-center gap-2">
              <span className="flex h-7 w-7 items-center justify-center rounded-xl bg-teal-800 text-white font-display text-sm font-bold shadow-xs">
                P
              </span>
              <span className="font-display text-lg font-bold tracking-tight text-teal-950">
                PMJAY Advocate
              </span>
            </Link>
            <p className="text-xs text-sand-900/70 leading-relaxed font-medium">
              Free, independent point-of-denial guidance for Indian families navigating PMJAY Ayushman Bharat coverage disputes.
            </p>
          </div>

          <div className="space-y-2.5">
            <p className="text-xs font-bold uppercase tracking-wider text-teal-900">
              Official Portals
            </p>
            <ul className="space-y-1.5 text-xs text-sand-900/75 font-medium">
              <li>
                <a
                  href="https://pmjay.gov.in"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-teal-800 transition"
                >
                  National Health Authority (NHA)
                </a>
              </li>
              <li>
                <a
                  href="https://cgrms.pmjay.gov.in"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-teal-800 transition"
                >
                  CGRMS Grievance Portal
                </a>
              </li>
              <li>
                <a
                  href="https://nalsa.gov.in"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-teal-800 transition"
                >
                  NALSA Legal Services
                </a>
              </li>
            </ul>
          </div>

          <div className="space-y-2.5">
            <p className="text-xs font-bold uppercase tracking-wider text-teal-900">
              Helplines
            </p>
            <ul className="space-y-1.5 text-xs text-sand-900/75 font-medium">
              <li>
                PMJAY Helpline: <a href="tel:14555" className="font-bold text-teal-800 hover:underline">14555</a>
              </li>
              <li>
                NALSA Free Legal Aid: <a href="tel:15100" className="font-bold text-teal-800 hover:underline">15100</a>
              </li>
              <li>
                Emergency Ambulance: <a href="tel:108" className="font-bold text-teal-800 hover:underline">108</a> / <a href="tel:112" className="font-bold text-teal-800 hover:underline">112</a>
              </li>
            </ul>
          </div>
        </div>

        <div className="border-t border-sand-200/60 pt-6 text-center text-[11px] text-sand-900/60 leading-relaxed font-medium">
          <p>
            PMJAY Advocate is a public-interest project and is not affiliated with the National Health Authority or any government body. In any medical emergency, seek urgent treatment immediately regardless of billing disputes.
          </p>
        </div>
      </div>
    </footer>
  );
}
