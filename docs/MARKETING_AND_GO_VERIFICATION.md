# Marketing site & Go verification — handover

Two concrete, unfinished pieces of work from the 26 August 2026 restructure
(`../HANDOVER.md`'s addendum of that date has the full writeup of what
*did* land). This document is the "how," not another status narrative —
read the addendum first for context, this for the actual next steps.

---

## Part 1 — Verifying the two backend changes

### What changed and why it's unverified

Two files, both edited without a Go toolchain available (a sandboxed
environment with no `go` binary) — verified by careful inspection only,
never compiled or run:

- **`backend/internal/api/dto.go`** — `CaseResponse` gained a
  `Description` field, populated from `store.CaseRecord.FamilyDescriptionRaw`
  (already persisted, just not previously exposed). Purely additive: one
  new optional JSON field, one new line in `caseRecordToResponse`. No
  existing test asserts an exact response body (checked — none do), so
  this is low-risk, but "low-risk by inspection" and "verified" are not
  the same thing.
- **`backend/internal/extract/prompt.go`** — the language paragraph of
  `systemPrompt` was rewritten. It previously scoped input to "English,
  native Malayalam script, or Hindi," with transliterated Malayalam
  (Manglish) explicitly called unsupported. It now welcomes all ten
  languages the product targets (English, Hindi, Malayalam, Tamil,
  Telugu, Kannada, Bengali, Marathi, Gujarati, Punjabi), each in native
  script, each transliterated into Latin letters, and any mix of the
  above within one message — while generalizing the existing safety
  rule ("score low confidence or UNSPECIFIED rather than hallucinate")
  to apply regardless of language or script, rather than special-casing
  it to Malayalam alone. Pure string content inside one raw string
  literal (backtick-delimited, no backticks in the new text — checked)
  — essentially zero syntax risk, but the actual thing this change is
  *for* (does a real LLM provider genuinely understand romanized Tamil,
  code-mixed Hindi/English, and so on) was never checked against a real
  provider. See "1.2" below — that's the part of this that actually
  matters.

### 1.1 — Mechanical verification (minutes, not judgment calls)

```bash
cd backend
go build ./...
go vet ./...
go test ./... -coverpkg=./...
```

Expected: clean build, clean vet, every existing test still passing at
whatever the current total is (268 backend tests as of the last count
that made it into `../HANDOVER.md` — 21 August; confirm the real
current number from this run rather than trusting that figure, since a
24 August session's work isn't reflected in `../HANDOVER.md` either —
see that document's 26 August addendum for the same observation).

If something fails, it will almost certainly be in one of these two
files specifically (nothing else changed) — and given how narrow both
edits are, a failure here would itself be a useful finding: it would
mean either an existing test was more tightly coupled to
`CaseResponse`'s exact shape or `systemPrompt`'s exact wording than a
`git grep` for `systemPrompt` and response-body assertions turned up
this session (that grep came back empty for both, but a sandbox without
`go` can't confirm a grep found everything a compiler would).

### 1.2 — The verification that actually matters: does the language claim hold up?

Compiling was never the risk here. The prompt now tells the model to
expect ten languages, three input forms each (native script,
transliterated, mixed), and previously explicitly said transliterated
Malayalam specifically *doesn't* work reliably enough to trust. That
line didn't get written for no reason — it's worth actually checking
whether removing it was correct, not just assuming a capable model
handles it because most do in general.

**Protocol**: run real `POST /v1/cases` requests (not the mocked test
clients — an actual configured provider, whichever of
Anthropic/Groq/Gemini is live in the target environment) with inputs
covering the full new surface. Suggested set, pulled directly from the
product brief that drove this change so the test matches what was
actually promised:

| Language | Form | Sample input |
|---|---|---|
| Malayalam | Native | എന്റെ അമ്മയുടെ കാൽ ഒടിഞ്ഞപ്പോൾ ആശുപത്രി സർജറിയുടെ ബിൽ ഞങ്ങളോട് അടയ്ക്കാൻ പറഞ്ഞു |
| Malayalam | Romanized | ente ammede kaal odinjappo hospital surgery de bill njanglaod adakkan paranji |
| Hindi | Native | मेरी माँ का पैर टूट गया था और अस्पताल ने सर्जरी का बिल हमें भरने के लिए कहा |
| Hindi | Romanized | meri maa ka pair toot gaya tha hospital ne surgery ka bill humko pay karne ko bola |
| Tamil | Native | என் அம்மாவின் கால் உடைந்தபோது மருத்துவமனை அறுவை சிகிச்சைக்கான பில் எங்களிடம் கட்டச் சொன்னார்கள் |
| Tamil | Romanized | en amma oda kaal odinjapo hospital surgery bill engakitta pay panna sonnanga |
| Hindi + English | Code-mixed | meri amma ka surgery hua but hospital ne PMJAY cover nahi hai bolke bill de diya |
| Telugu / Kannada / Bengali / Marathi / Gujarati / Punjabi | Native | No existing test history in this system at all — worth at least one real sample each, since these are the languages added this session with zero prior validation of any kind |
| Something genuinely unparseable | — | A short, ambiguous fragment with no clear clinical content — confirms the generalized low-confidence rule actually fires instead of a confident-sounding hallucinated match |

For each: does the tier/citation look sane given the input, does
`extracted_situation_summary` (internal — not on the API response,
check server logs or add a temporary log line) actually restate the
situation correctly, and — the one that matters most — does the
ambiguous/unparseable case correctly come back low-confidence or
UNSPECIFIED rather than a fabricated match?

Write up whatever this finds the way `docs/DATA_SOURCES.md` documents
each verified data point — real inputs, real outputs, dated. If it
holds up, that's worth a line in `../HANDOVER.md`'s next addendum, the
same way the multi-provider work got documented. If it doesn't hold up
for specific languages, the honest fix is narrowing the prompt's claim
back down for those, not tuning the wording until it sounds like it
works — and the marketing copy (Part 2) needs to match whatever's
actually true, not what was originally hoped.

### 1.3 — If it needs a partial revert

Both changes are independent and easy to roll back individually. If 1.2
finds the widened language claim doesn't hold up for some subset, the
targeted fix is narrowing `systemPrompt`'s language paragraph back to
whichever languages/forms actually verified well — not reverting to the
exact old wording, since the *generalized* low-confidence rule (applies
to any language, not just Malayalam) is a strict improvement over the
old Malayalam-specific version regardless of how many languages end up
supported. Keep that part regardless of what 1.2 finds.

---

## Part 2 — Finishing the marketing site

### What this session actually did here (small, deliberately)

Renamed the color system everywhere (teal → ink, mechanical), fixed the
two places that would have directly contradicted the widened language
prompt (the FAQ answer and How It Works' first step both used to say
Manglish wasn't supported), and added a conditional "My Cases" link to
`Header` for returning visitors. That's it — no content rewrite. Here's
what's actually left, checked against the original product brief's
Sections 8 (SEO), 9 (marketing site), and 18 (content quality), and
against what's *actually* already on each page rather than assumed.

### 2.1 — The clear, concrete gap: technical SEO

Checked this session — currently genuinely absent, not just
unverified:

- **No Open Graph or Twitter card metadata anywhere.** `app/layout.tsx`'s
  `metadata` export has `title`, `description`, and icons only. Every
  page inherits that same generic description regardless of what it's
  actually about.
- **No structured data (JSON-LD)** — the brief explicitly asks for this
  "where appropriate" (an `Organization` or `WebSite` block at minimum;
  possibly `FAQPage` on `/faq`, which would be a genuine, low-effort win
  given that page's content is already structured as Q&A pairs).
  Next.js's built-in JSON-LD support (a `<script type="application/ld+json">`
  in the relevant page, or the `metadata` API's newer structured-data
  helpers) fits this codebase's existing zero-extra-dependency posture.
- **No `sitemap.ts` or `robots.ts`** beyond the static `public/robots.txt`
  already covering the app routes (`next.config.mjs`'s security-headers
  comment explains that one). A generated sitemap covering the five
  marketing pages plus the new Languages page below is a small,
  mechanical Next.js App Router feature (`app/sitemap.ts`) and a clear,
  bounded task.
- Per-page `title`/`description` already exist on `/guide`, `/how-it-works`,
  `/about`, `/faq` (checked — they follow the established `"Page Name —
  PMJAY Advocate"` pattern already; don't restructure that, it's fine).
  What's missing is OG/Twitter tags *building on* those, not replacing
  them.

### 2.2 — Net new: the Languages page

Doesn't exist. The brief calls for this explicitly (Section 9) as a way
to make multilingual support a visible, checkable product strength
rather than a claim buried in a placeholder paragraph. Concrete shape:

- Route: `app/languages/page.tsx`, using `Header`/`Footer` like the
  other marketing pages (not `AppShell` — this is public, indexed
  content, the opposite of the noindex app pages).
- Content: the same honest framing already established in `IntakeForm`
  and `Settings` ("most reliable today in English, Hindi, or Malayalam,
  actively widening") — **do not** advertise all ten languages as
  equally solid until Part 1.2 actually confirms that. Show real
  example pairs (native script + romanized) for the languages that
  *have* been checked; if Part 1.2 hasn't run yet when this page gets
  built, say so plainly rather than implying it has.
- Add it to `Header`'s `NAV_LINKS` and to the new sitemap above. It
  should **not** get the `X-Robots-Tag: noindex` treatment
  `next.config.mjs` applies to `/case`, `/cases`, `/dashboard`,
  `/settings` — this page is exactly the kind of content that *should*
  be found by search.

### 2.3 — Homepage: align with the brief's suggested framing

The brief suggests specific hero language ("Understand your hospital
and PM-JAY situation. Know what to do next.") and a five-step workflow
visual (Tell us what happened → Understand the issue → Organize
evidence → Build your case → Take action). Check the current homepage
against both, and against the brief's explicit copy don'ts ("Unlock
powerful insights with our AI-powered platform" and its relatives —
leverage, empower, revolutionize, empty AI-marketing language). Also
confirm the homepage's own intake form still matches `IntakeForm`'s
current honest language copy (it should — same component — but confirm
after any homepage edit doesn't accidentally fork it).

### 2.4 — Targeted audits, not rewrites, for Guide / About / FAQ

Checked this session — these are in better shape than "needs a rewrite"
implies:

- **`/guide`** already covers what PMJAY covers, core rights, what
  hospitals can't do, denial patterns, what to do at the desk, and when
  to escalate — real overlap with the brief's suggested SEO topic list.
  Gap worth checking specifically: "PM-JAY empanelled hospitals" isn't
  an explicit section; "PM-JAY complaint" is only implicit inside
  "escalate." Audit against the brief's full topic list (Section 8) and
  add sections for genuine gaps — don't restructure what's already
  there.
- **`/about`** already has hero + a "principles" section + an
  "independence" section (checked headings only, not full copy this
  session). Audit against the brief's Trust checklist specifically —
  what the platform does, what it doesn't, how information is handled,
  limitations, disclaimers — and check it doesn't contradict
  `DisclaimerNote`'s on-screen text or `response/types.go`'s actual
  disclaimer guarantee.
- **`/faq`** — one answer fixed this session (language support). Read
  the rest for the same class of issue: any answer whose claim has
  drifted from what the code actually does.

### 2.5 — Constraints while doing any of this

- Everything here is public, indexed content — the opposite posture
  from the app pages this session focused on. Don't carry over the
  `noindex` habit by reflex.
- `../../DESIGN.md` (repo root) is the actual design-system reference
  now — ink color tokens, the `mx-auto w-full max-w-6xl px-6 sm:px-8
  lg:px-10` container on every page, Atkinson Hyperlegible for all
  functional text, translucency scoped to sticky chrome only. Read it
  before touching layout or color on any marketing page.
- If pulling real content for `/guide` or the Languages page from an
  official PMJAY source, paraphrase and cite — don't reproduce
  government text verbatim at length.
- Keep language claims exactly as strong as Part 1.2 actually verified,
  no stronger. This is the one place in this whole handover where
  sequencing genuinely matters: do 1.2 before writing Languages-page or
  homepage copy that talks about language support, or the copy risks
  overselling something nobody's actually checked.

### 2.6 — Suggested order

1. Part 1 (Go verification) — the marketing copy in 2.2/2.3 depends on
   knowing what's actually true.
2. Technical SEO (2.1) — small, mechanical, no content judgment calls,
   immediate win.
3. Languages page (2.2) — net new, directly serves the brief's most
   emphasized differentiator.
4. Homepage alignment (2.3).
5. Guide / About / FAQ audits (2.4) — lower urgency, already in decent
   shape.

Same closing note the rest of this project's handover documents end
on: this is what's actually left, not an invitation to invent more.
