import { SITE_DESCRIPTION, SITE_NAME, SITE_URL } from "@/lib/site";

/**
 * /llms-full.txt — the longer technical context, for an agent or model
 * that needs to reason about what this system can and cannot do before
 * answering a user's question about it.
 *
 * Generated from the same site constants as /llms.txt and the sitemap.
 * Every claim below is checkable against the repository: the tiering is
 * deterministic Go, the extraction is one model call over a shortlist,
 * and the dataset split between verified and unverified is enforced in
 * the build script and the packaging validator.
 */
export function GET() {
  const body = `# ${SITE_NAME} — full system context

> ${SITE_DESCRIPTION}

## Overview

${SITE_NAME} is an open-source, independent point-of-denial assistant for
beneficiaries of the Pradhan Mantri Jan Arogya Yojana (Ayushman Bharat /
PM-JAY), India's national public health assurance scheme. It is not
affiliated with or endorsed by the National Health Authority.

## The situation it is built for

A family reaches an empanelled hospital and the billing desk says one of:

- a covered procedure is "not covered" and cash is demanded up front;
- the "PM-JAY server is down, pay now and claim later";
- the procedure name does not match their package list;
- the pre-authorisation is "pending", and the family is told that means
  no;
- discharge papers are being held until the bill is paid.

The tool's job is narrow: decide whether that refusal looks correct, and
say what to do next in language a frightened person can use at a counter.

## Technical shape

### 1. Intake

One free-text description, in English, Malayalam, Hindi, or mixed code.
No account, no card number, no Aadhaar number. The request is bounded by
a server-side timeout so a slow model provider returns this system's own
care-first message instead of a dropped connection.

### 2. Retrieval before any model call

A dependency-free keyword pass narrows roughly 300 packages to a short
list. The model never sees the full dataset; that is a cost control and
also keeps the prompt small enough to stay accurate.

### 3. Extraction

A single tool-use call returns a structured match: candidate package
codes with confidence, exclusion matches, and a pending-vs-denied signal.
If no provider key is configured, the server still runs and serves
package search and health, while intake returns 503 with the helpline
number.

### 4. Deterministic tiering

The five-way decision is made by fixed rules in plain Go, not by the
model:

- **Green** — a specific package matches with enough confidence. Shows
  the code and, if verified, the indicative rate.
- **Amber** — ambiguous, low confidence, or a pending pre-authorisation.
  Asks one clarifying question instead of guessing.
- **Red** — a confirmed exclusion. Says the denial is correct and points
  to alternatives.
- **Mixed** — part of the bill is a covered package and part is a real
  exclusion. Splits the bill rather than answering yes or no.
- **Handoff** — detention, coercion, or a situation too tangled for a
  guided flow. Routes to NALSA on 15100 with a written summary.

### 5. Data honesty

The dataset ships in two files:

- 300 records with a code, name, specialty, and rate checked against a
  published government HBP schedule (verified).
- 15 records known by procedure name only, with **every rate field
  removed**. An agent may recognise the name; it may not quote a number
  that was never checked.

A packaging step refuses to build if an unverified record leaks a rate,
if the care-first rule is missing, or if frontend code appears in the
skill bundle.

## Privacy and retention

No accounts, no phone numbers, no passwords. The description is sent to
the configured AI provider (Anthropic, Groq, or Google) for extraction.
The resulting case is stored server-side under a random case ID so it can
be reopened; anyone holding the link can read it, because the ID is the
only credential. Cases are deleted after 30 days of inactivity and can be
deleted immediately from the case page. A browser-side list of case IDs
lives in localStorage.

This is designed around the Digital Personal Data Protection Act, 2023's
access, correction, and erasure principles, but it is an independent
public-interest project, not a registered data fiduciary with a grievance
officer. The privacy page spells out the limits rather than claiming
certification.

## Public routes

- ${SITE_URL}/ — intake and explanation.
- ${SITE_URL}/packages — deterministic package search. Works without an AI provider.
- ${SITE_URL}/how-it-works — pipeline explanation.
- ${SITE_URL}/guide — patient rights in plain language.
- ${SITE_URL}/faq — common denial questions.
- ${SITE_URL}/about — mission and independence.
- ${SITE_URL}/privacy — data handling, retention, deletion.
- ${SITE_URL}/terms — usage conditions and emergency clause.
- ${SITE_URL}/disclaimer — medical, legal, and financial limits.

Private and deliberately not indexed: /cases/[id] (a case ID is a
capability), /dashboard, /settings, /api.

## Authority links

- National Health Authority: https://nha.gov.in
- CGRMS grievance portal: https://cgrms.pmjay.gov.in
- NALSA: https://nalsa.gov.in
- PM-JAY helpline: 14555
`;

  return new Response(body, {
    headers: {
      "Content-Type": "text/plain; charset=utf-8",
      "Cache-Control": "public, max-age=3600",
    },
  });
}
