# Design system

Scope: `frontend/`'s visual language — colors, type, spacing, motion,
components, and the accessibility commitments that constrain all of it.
This is the reference for "does this new screen look like it belongs,"
written down instead of left as tribal knowledge in whoever built the
last page. Source of truth for the actual values is always
`frontend/app/globals.css`; this document explains what's there and why.

## Philosophy

Three things this product is, in order of how much they should win a
disagreement:

1. **A healthcare and money tool a stressed person is reading.** Not a
   marketing site, not a dashboard for its own sake. Every screen should
   answer "what's happening, and what do I do" faster than it impresses.
2. **Minimal in the way Apple's own interfaces are minimal** — not
   "sparse," but *considered*: nothing on screen that isn't earning its
   place, real hierarchy instead of decoration standing in for it, and
   motion/materials used the way Apple actually uses them (see
   "Motion & materials" below) rather than as surface polish bolted on
   at the end.
3. **Honest about what it knows and doesn't.** The visual language has
   to be able to say "we're confident" and "this needs a person to look
   at it" differently, because the app itself has to say that
   constantly (see "Color" below — this is the whole reason tier color
   is treated as sacred).

If a design choice serves (2) at the expense of (1) or (3) — a
translucent panel that makes real content harder to read at a glance, a
color used decoratively that could be mistaken for a tier — it's wrong,
regardless of how good it looks in isolation.

## Color

### The rule everything else follows

**Color means a tier, or it means nothing.** Five outcomes
(green/amber/red/mixed/handoff) are the only thing in this app that
color is allowed to *communicate* — everything else (chrome, nav,
buttons, links) is deliberately as close to neutral as it can be while
still looking like a considered product, so it never competes with or
gets confused for tier meaning. This is why the primary brand color is
"ink," not a hue.

### Ink — primary, chrome, interactive

```
ink-50   #f4f5f6      ink-500  #5c6570
ink-100  #e7e9eb      ink-600  #3d444d   ← primary buttons, links, icon strokes
ink-200  #cdd1d6      ink-700  #2f353c
ink-300  #a8afb8      ink-800  #24282e
ink-400  #7c8590      ink-900  #1a1d21
                      ink-950  #0f1113
```

A near-monochrome graphite, just warm enough at the light end to sit on
the sand neutrals without reading as a cold, foreign gray. This
replaced an earlier deep teal that — worth being direct about — still
read as a shade of green in practice, which fought with tier-green for
the same part of the color wheel. The fix isn't a different accent hue;
it's removing hue from the chrome almost entirely, the same restraint
Apple's own interfaces lean on (status gets color, navigation mostly
doesn't).

### Marigold — the one sparing accent

```
marigold-50   #fdf8ec      marigold-600  #a86b0c
marigold-100  #f9ecc9      marigold-700  #86530d
marigold-200  #f3d68d      marigold-800  #6d4312
marigold-300  #ecbc55      marigold-900  #5b3813
marigold-400  #e0a12a
marigold-500  #cc8712
```

An "ember," not a craft-supply yellow. Small highlights and the
occasional badge only — never a large surface. If you're reaching for
marigold on anything bigger than an icon chip or a badge, that's a sign
the thing you're building actually wants a tier color, or wants no
color at all.

### Sand — the neutral base

```
sand-25   #fdfcfa      sand-500  #8c8169
sand-50   #faf8f3      sand-600  #6c6252
sand-100  #f3f0e8      sand-700  #554c40
sand-200  #e6e0d2      sand-800  #403a31
sand-300  #d3cab4      sand-900  #2a2621
sand-400  #b0a68b      sand-950  #1a1713
```

Warm paper, not cold gray. Backgrounds, borders, and the bulk of body
text all live on this scale — it's most of what's actually on screen at
any given time, so its warmth is what makes the app feel human rather
than clinical.

### Tier colors — the only meaningful color in the app

```
green    bg #eaf6ee  border #b7ddc2  icon-bg #d9efe0  text #1e6b3d
amber    bg #fdf5e4  border #efd99a  icon-bg #f8e8c2  text #8f5f09
red      bg #fbeeec  border #e7beb6  icon-bg #f4dcd6  text #953f2d
mixed    bg #f4eef8  border #d8c2e7  icon-bg #ecdff3  text #693e8c
handoff  bg #eaf1fb  border #b8cfee  icon-bg #dbe8f9  text #23568f
```

Deliberately muted — none of these should read as a stoplight or a
generic error state, because they're carrying real news to someone
under real stress. **Never the only signal**: every tier also gets its
own icon and label (`TierBadge.tsx`) so color reinforces, it doesn't
carry meaning alone. Do not introduce a sixth "meaningful" color, and
do not reuse a tier hue for anything that isn't that tier — that's the
one hard rule in this whole document.

## Typography

- **Atkinson Hyperlegible** (`--font-sans`) — every piece of functional
  text. Designed by the Braille Institute for maximum
  character-to-character legibility, including for low-vision readers.
  This is an accessibility decision, not a style choice, and it isn't
  swapped out for a trendier grotesk lightly — a cleaner, more
  "Apple-like" geometric sans would look sharper in a screenshot and
  read worse for exactly the people this product is for.
- **Fraunces** (`--font-display`) — sparingly, for the handful of large
  display moments (hero heading, section titles, page H1s). Never body
  copy.
- **Tracking**: large display type gets tightened tracking
  (`.tracking-tight-display`, -0.02em) — Apple's own type guidance
  (tight at large sizes, neutral-to-loose at small sizes prevents small
  text from looking cramped) and it's already how this scale behaves;
  don't add letter-spacing overrides on body text.

## Spacing & layout

The page container is `mx-auto w-full max-w-6xl px-6 sm:px-8 lg:px-10`
— **this exact string, unchanged**, on every page, marketing or app.
It's the one layout constant every redesign since this product's first
version has preserved on purpose; a new page that doesn't use it will
visibly drift out of alignment with every other page the moment
they're viewed back to back.

`AppShell`'s sidebar (product pages only) sits *outside* that
container, as its own column — it's new chrome, not a change to the
container, so it doesn't violate the rule above.

## Corner radius & shadow

- Cards: `1.25rem`. Buttons/fields: `0.75rem`–`1rem`. Soft, but
  deliberately short of "every element is a pill" — the brief this
  system follows explicitly warns against rounding everything to death,
  and a flatter radius on buttons vs. cards is itself a hierarchy cue
  (cards are containers, buttons are actions).
- Default card shadow is a near-invisible two-layer shadow
  (`0 1px 2px`, `0 1px 3px`, both under 5% opacity) — present for
  separation, not for drama.
- **Neumorphic accents** (`--shadow-soft-raised` / `--shadow-soft-pressed`,
  `.tile-soft`, buttons, the active nav item) — a deliberately light
  touch, scoped to interactive chrome only. Full neumorphism (same-tone
  surfaces, borders replaced by shadow alone) trades away contrast and
  clear affordances that matter for people reading this under stress,
  on a washed-out phone screen, or with low vision — so real content
  surfaces keep their borders and contrast; this is a highlight/shadow
  pair layered on top of that, never a replacement for it. If you're
  tempted to remove a card's border because the soft shadow "looks like
  enough" — it isn't, for this app; don't.

## Motion & materials

- **Easing**: `--ease-out-quick` (`cubic-bezier(0.16, 1, 0.3, 1)`)
  everywhere. It's a snappy-then-settling curve close in feel to
  Apple's own UI springs — deliberately not a linear or generic
  `ease-in-out`, which reads as mechanical by comparison.
- **Apple's actual rule for motion** ("start from wherever the UI
  currently is, carry momentum, stay interruptible") is written for
  gesture-driven native interfaces; this is a server-rendered web app
  with no drag gestures to carry momentum from, so the honest scope
  here is: fast, considered `fade-in`/`fade-in-up` on content arriving,
  nothing that overshoots or bounces without a reason, and never a
  fixed-duration animation so long it makes someone wait to read real
  information.
- **`prefers-reduced-motion`**: respected globally — all animation and
  transition durations collapse to near-zero. Never build a new
  animated component without checking it degrades through this.
- **Materials**: translucency is scoped to exactly two places —
  sticky navigation chrome (`bg-sand-25/90` + `backdrop-blur-md` +
  `backdrop-saturate-150`, Apple's own blur-plus-saturation recipe so
  content scrolling underneath stays legible) and the care-first
  banner's gradient. Nowhere else. Content someone is making a real
  decision from reads better flat, at full contrast, with nothing
  behind it competing for attention — see "Philosophy" above.

## Components (see `app/components/`, `app/globals.css`'s `@layer components`)

| Class | For | Notes |
|---|---|---|
| `.card` | Default container surface | Solid, bordered, never glass |
| `.badge` | Small status/category tag | Structural only — caller adds semantic color utilities |
| `.btn-primary` | The one primary action per screen | Soft raised→pressed shadow on interaction |
| `.btn-secondary` | Secondary actions | Bordered, white |
| `.btn-ghost` | Quiet, text-weight actions | No border, hover bg only |
| `.field` | Text input / textarea | One definition so focus/disabled/placeholder never drift between forms |
| `.app-nav-item` | Sidebar / bottom tab bar (product pages) | `data-active="true"` gets a soft pressed inset |
| `.section-nav-link` | In-page "jump to section" link | Case workspace's quick-nav — not an ARIA tabs widget, see `app/cases/[id]/README.md` |
| `.tile-soft` | Small raised summary tiles (Dashboard stats) | Off `.card` on purpose — case *content* never gets the soft-shadow treatment |
| `.tap-target` | Anything tappable under 44×44 | WCAG 2.5.8 |
| `.skeleton-shimmer` | Loading placeholders with known final height | Holds layout still, no content jump |

## Accessibility — non-negotiable, not a checklist to satisfy once

- One visible, high-contrast focus style everywhere
  (`outline: 2.5px solid var(--color-ink-600)`) — never removed, never
  swapped for a subtle box-shadow. Someone may be tabbing through this
  on a keyboard-only kiosk browser at a hospital.
- 44×44 minimum tap targets on anything interactive.
- Color is never the only signal (see "Tier colors" above).
- `prefers-reduced-motion` is handled globally, not per-component.
- Atkinson Hyperlegible is a legibility decision, not swappable for
  aesthetics — see "Typography."

## Patterns

- **Empty states** name the specific thing that's missing and give one
  primary action, not a generic "No data found." (`EmptyDashboard` in
  `app/dashboard/page.tsx` is the reference example.)
- **Loading states** use `role="status"` and, where the final layout is
  already known, `.skeleton-shimmer` blocks sized to match — never a
  bare spinner on an otherwise blank page.
- **Error states** use `role="alert"`, plain language, and — for
  anything on the case path — a `tel:14555` fallback, since a failed
  page load shouldn't be a dead end for someone who needs an answer.
- **Wait states that reflect real processing** (`IntakeForm`'s staged
  loading copy) are allowed to describe what's actually happening in
  roughly the right order; they are not allowed to fake precision (a
  percentage, a literal progress bar) the backend doesn't actually
  report.

## Explicit don'ts

Carried forward from the brief this redesign was built against, because
they're easy to erode one small PR at a time:

- Don't put every section in a floating card.
- Don't use gradients outside the two named deliberate spots.
- Don't add shadow for drama — shadow here means "this is slightly
  raised," never "look here."
- Don't round everything into a pill.
- Don't introduce a color that could be mistaken for a tier.
- Don't build a new animated or translucent element without checking
  it against `prefers-reduced-motion` and against a solid, opaque
  content surface underneath.

## Open / not yet decided

- No dark mode yet. If it's built, tier colors need their own dark
  palette pass (not just an inverted lightness) since "muted, never a
  stoplight" is a harder property to hold onto against a dark
  background.
- The `ink` scale hasn't been checked against a color-blindness
  simulator alongside the five tier colors as a full set — each tier
  already carries its own icon+label so no single check is
  safety-critical, but a real pass is worth doing before this scales
  much further.
