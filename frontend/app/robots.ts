import type { MetadataRoute } from "next";
import { SITE_URL } from "@/lib/site";

/**
 * Generated rather than static so the Sitemap line always matches the
 * origin this build is actually deployed at. A static robots.txt with a
 * hardcoded domain is one of the quietest ways to break a deploy: the
 * file exists, crawlers read it, and they are pointed somewhere else.
 *
 * The disallowed paths are the private ones:
 *   /cases/  a case ID is a capability; indexing it would publish one
 *            family's situation
 *   /case/   legacy redirect for old links, same rule
 *   /dashboard, /settings  personalised, no-login "your data lives in
 *            this browser" views
 *   /api/    the JSON API itself, including package search. Search pages
 *            are server-rendered by /packages; crawlers do not need the
 *            raw endpoint.
 */
export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: "/",
      disallow: ["/api/", "/case/", "/cases/", "/dashboard", "/settings"],
    },
    sitemap: `${SITE_URL}/sitemap.xml`,
    host: SITE_URL,
  };
}
