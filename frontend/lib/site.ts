/**
 * site.ts — the one place the public origin is decided.
 *
 * Every absolute URL this app emits (canonical tags, Open Graph, the
 * sitemap, robots.txt) reads from here. The origin comes from
 * NEXT_PUBLIC_SITE_URL because this repo is deployable to any host and
 * hardcoding a domain would either be wrong on the next deploy or, worse,
 * point canonical tags at a domain the project does not control.
 *
 * The localhost fallback is only correct in development. A production
 * deploy that forgets to set the variable will emit localhost URLs in its
 * sitemap — which is exactly the kind of silent SEO failure that is worth
 * failing loudly for instead. See docs/DEPLOYMENT.md.
 */

export const SITE_NAME = "PMJAY Advocate";

export const SITE_DESCRIPTION =
  "Free, independent help understanding whether a hospital's Ayushman Bharat (PM-JAY) coverage denial is correct, plus a searchable index of 300 source-verified Health Benefit Packages.";

function normalise(url: string): string {
  return url.replace(/\/+$/, "");
}

export const SITE_URL = normalise(
  process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000",
);

export function absoluteUrl(path = "/"): string {
  return `${SITE_URL}${path.startsWith("/") ? path : `/${path}`}`;
}
