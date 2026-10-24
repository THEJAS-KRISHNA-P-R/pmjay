import { SITE_DESCRIPTION, SITE_NAME, SITE_URL } from "@/lib/site";

/**
 * /llms.txt — a short, factual orientation for language models and
 * agents that are asked about this project.
 *
 * Generated rather than static so every URL matches the deployment's
 * real origin. The previous static file hardcoded a domain, which is the
 * same failure mode as a hardcoded sitemap: correct once, wrong after
 * the next move.
 */
export function GET() {
  const body = `# ${SITE_NAME} — independent PM-JAY denial help

> ${SITE_DESCRIPTION}

## What this is

A free, independent public-interest tool for Indian families who are told
at a hospital billing desk that a treatment is not covered under Ayushman
Bharat (PM-JAY). It checks the stated procedure against a reference index
of Health Benefit Package (HBP) data and explains what to do next.

It is not the National Health Authority, not a hospital, not an insurer,
and not a lawyer. It cannot check a live pre-authorisation, submit a
complaint on someone's behalf, or change a PM-JAY record.

## What it actually does

- Matches a plain-language description against 315 PM-JAY procedure names,
  300 of which carry a code and rate verified against a published HBP
  schedule; 15 are recognised by name with no verified code or rate.
- Sorts the situation into Green (looks covered), Amber (one more fact
  needed), Red (the denial is correct), Mixed (part covered, part
  excluded), or Handoff (route to free legal aid).
- Produces a desk script, a CGRMS complaint draft with placeholders, and
  an evidence checklist.
- Searches the package index directly, with no AI call, at /packages.
- Does not invent package codes, rates, or pre-authorisation statuses.

## The rule that outranks everything else

Get treatment first. Dispute the money after. The tool will not tell
anyone to delay, refuse, or cancel care while a billing question is open.

## Coverage, stated plainly

The full PM-JAY master list has roughly 1,900 procedures. This index
covers 315 names, 300 with verified rates. General surgery, orthopaedics,
obstetrics and gynaecology, ophthalmology, ENT, urology, and neurosurgery
are the known thin areas. A missing result means "not in this index",
never "not covered by PM-JAY".

## Pages

- [Home](${SITE_URL}/): Describe a denial and get a care-first reading.
- [Package search](${SITE_URL}/packages): Search verified HBP codes and rates. No login, no model.
- [How it works](${SITE_URL}/how-it-works): The retrieval, matching, and rule-based decision path.
- [Your rights](${SITE_URL}/guide): Plain-language guide to patient rights under PM-JAY.
- [FAQ](${SITE_URL}/faq): Cash deposits, pre-authorisation, portability, and common denials.
- [About](${SITE_URL}/about): Mission, independence, and how the dataset is sourced.
- [Privacy](${SITE_URL}/privacy): What is stored, who processes it, the 30-day retention window, and how to delete a case.
- [Terms](${SITE_URL}/terms): Usage conditions and liability limits.
- [Disclaimer](${SITE_URL}/disclaimer): Medical emergency priority and non-agency notice.

## Contacts

- PM-JAY national helpline: 14555
- NALSA free legal aid: 15100
- National emergency number: 112
- Official grievance portal: https://cgrms.pmjay.gov.in
`;

  return new Response(body, {
    headers: {
      "Content-Type": "text/plain; charset=utf-8",
      "Cache-Control": "public, max-age=3600",
    },
  });
}
