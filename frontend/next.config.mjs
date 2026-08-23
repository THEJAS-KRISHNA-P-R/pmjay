/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,

  // Case pages are reachable only by knowing a case's unguessable
  // UUIDv4 ID — there's no login wall, by design (see
  // ../ARCHITECTURE.md), which means the ID itself is the only thing
  // standing between a stranger and one family's health/denial details.
  // These headers are defense in depth around that, not a substitute
  // for the ID's own unguessability:
  //   - X-Robots-Tag: a case URL should never be indexed or have its
  //     links followed, in the event one is ever discovered by a crawler
  //     (shared publicly by mistake, leaked via an analytics tool, etc).
  //     robots.txt (public/robots.txt) asks well-behaved crawlers not to
  //     visit at all; this is the header-level backstop for crawlers
  //     that fetch anyway or don't respect robots.txt.
  //   - Referrer-Policy: without this, following any link *out* of a
  //     case page would leak the full case URL — including the ID — to
  //     whatever site that link pointed to, via the Referer header.
  //     strict-origin-when-cross-origin sends only the origin (not the
  //     path) on cross-origin navigation, so the case ID never leaves
  //     this app.
  //   - X-Content-Type-Options / X-Frame-Options: standard hardening
  //     against MIME-sniffing and clickjacking; costs nothing to set.
  async headers() {
    return [
      {
        source: "/case/:path*",
        headers: [
          { key: "X-Robots-Tag", value: "noindex, nofollow" },
        ],
      },
      {
        source: "/:path*",
        headers: [
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "X-Frame-Options", value: "DENY" },
        ],
      },
    ];
  },
};

export default nextConfig;
