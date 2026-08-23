# Accessibility & Multimodal Feedback

## Reduced motion, transparency, and contrast

"Reduced" doesn't mean "removed." The goal is to keep the feedback that helps someone understand what happened, while dropping the parts that can cause discomfort (large parallax, spinning, elastic overshoot) for people sensitive to motion.

- **`prefers-reduced-motion: reduce`** — replace slides, springs, and parallax with a short opacity cross-fade or an instant state change. Drop bounce/overshoot entirely. Keep color and opacity changes that communicate state, since those aren't the problematic kind of motion.
- **`prefers-reduced-transparency: reduce`** — raise the background opacity of translucent surfaces (or drop the blur) so they read as closer to solid.
- **`prefers-contrast: more`** — move toward near-solid backgrounds with a clearly defined border rather than relying on subtle tonal differences.

Other things worth avoiding regardless of the setting: full-viewport backgrounds that are constantly moving, slow looping oscillation (roughly one cycle every few seconds tends to be the uncomfortable range), and abrupt brightness jumps — ease dark/light theme transitions rather than snapping between them.

```css
@media (prefers-reduced-motion: reduce) {
  .sheet { transition: opacity 200ms ease; transform: none !important; }
}
@media (prefers-reduced-transparency: reduce) {
  .toolbar { background: white; backdrop-filter: none; }
}
```

**Concrete contrast numbers, not just "make sure it's legible":** 4.5:1 minimum contrast for text against its background; 3:1 minimum for non-text UI (icon boundaries, a checked-vs-unchecked control state, a focus ring). Apple's own App Store accessibility evaluation criteria call out translucency specifically here — test with Increase Contrast on and Reduce Transparency off, *and* with both on, because a translucent background can pass a contrast check against a solid mockup color and still fail against real, busy content behind it. This is exactly the failure mode that shipped in iOS 26's Liquid Glass launch (see `materials-typography.md`) — it's a documented, real-world case of skipping this check, not a hypothetical.

## Dynamic Type: the actual scale, and how it breaks things

Dynamic Type is Apple's system-wide text-scaling preference (Settings → Accessibility → Display & Text Size). Roughly 30% of iOS users move it off the default — this isn't an edge case affecting a handful of people, it's a meaningful fraction of any real user base.

- **Twelve sizes, in two tiers.** Seven "standard" sizes (xSmall, Small, Medium — the default, Large, xLarge, xxLarge, xxxLarge) are always available. Five more "accessibility" sizes (AX1 through AX5) require the user to opt in via "Larger Accessibility Sizes," and scale text up to roughly 310% of its default size at the top end.
- **Text styles don't scale at the same rate.** Titles scale up slowly; body text scales at a moderate rate; captions scale up quickly but hit a floor (the smallest style, Caption 2, stops shrinking at 11pt and won't go below it). The counterintuitive consequence: at the largest accessibility sizes, Body text can become visually *larger* than Title 1. If you're building a custom type scale (native or web) and just multiplying every size by the same ratio, you're not actually replicating Apple's system — you're approximating it, and it'll feel subtly wrong at the extremes.
- **Touch targets don't get to stay small just because you're out of room.** The 44×44pt minimum (see `platforms.md`) holds regardless of text size — pad the hit area rather than letting a scaled-up label visually crowd out its own tap target.
- **The bugs that actually show up in production, in order of frequency:** single-line labels (`numberOfLines = 1` or a CSS `white-space: nowrap` equivalent) truncating instead of wrapping; fixed-height containers clipping scaled content instead of growing; horizontally-arranged elements overlapping or colliding as their labels grow (favor vertical stacking once space runs out); scrollable containers that don't actually scroll on the given platform by default and need it added explicitly.
- **A pragmatic, legitimate mitigation — not a cop-out if it's a deliberate choice:** many shipped apps cap scaling around AX1–AX3 rather than fully supporting AX5, because AX5 can break dense layouts throughout an app. That's a reasonable trade-off if you make it on purpose and test the cap works cleanly; it's a bug if it's just wherever your layout happened to stop holding together.
- **Test across the whole range, not just the default.** Preview or set the size explicitly rather than eyeballing default-size mockups — most Dynamic Type bugs are invisible until you actually push text size up. On the web, the closest equivalent test is 200%+ text-only zoom (not viewport resize) plus your OS-level text size setting, since some browsers now partially respect it.

## VoiceOver and screen-reader structure

Getting the visuals right and skipping this is only half of "Apple-style" — Apple's own accessibility chapter is one of the most operational, specific parts of the HIG, not an afterthought.

- **Every interactive element needs a label describing what it *is*, not just how it looks.** An icon-only button needs an explicit accessible name; a decorative image that adds nothing needs to be explicitly hidden from the accessibility tree (`accessibilityHidden` / `aria-hidden="true"`) rather than read aloud as noise or given a redundant label that repeats visible text next to it.
- **Hints are for what happens next, used sparingly.** Only add a hint when the action genuinely isn't obvious from the label and surrounding context — hints on every element train people to skip them, the same way over-used haptics train people to ignore all of them.
- **Traits/roles matter as much as labels.** Marking something as a button, header, or link (rather than leaving it as an untyped generic element) is what lets VoiceOver announce it correctly and what makes Rotor-based navigation (jump by heading, by link, by form control) actually work. A `<div onClick>` styled to look like a button provides none of this for free — use the real semantic element, or add the role explicitly if you truly can't.
- **Group what reads as one thing, separate what doesn't.** A card with an image, title, subtitle, and price should usually read as a single swipe-stop for VoiceOver, not four separate stops — unless a piece of it (a "favorite" button inside the card) needs independent interaction, in which case it stays a separate stop nested appropriately.
- **Accessible reading order should match visual order**, or be set explicitly when it can't (a CSS grid or flexbox `order` that reflows visual position without reflowing DOM/reading order is a common way this quietly breaks).
- **Content that changes without the user navigating to it needs an explicit announcement** — a toast, a loading state resolving, a form error appearing. On the web, an appropriately-scoped `aria-live` region; natively, a posted accessibility notification. Silent state changes are invisible to someone who isn't looking at the screen.

## Right-to-left layout: what mirrors, and what deliberately doesn't

Arabic, Hebrew, Persian, and Urdu are read right-to-left, and Apple's platforms mirror the interface automatically when one of these is the active language — but "mirror everything" is the wrong mental model and produces visible mistakes.

**Mirrors:** overall reading/layout direction and text alignment · navigation elements (a "back" affordance points right, not left) · progress bars and sliders · breadcrumbs · directional icons that mean "forward"/"back" (use direction-aware icon names — `chevron.forward`/`chevron.backward` — rather than literal ones like `chevron.right`/`chevron.left`, so the correct mirrored glyph is chosen automatically) · icon+label ordering · scrollbar position (moves to the left edge).

**Does not mirror:** media playback controls and scrubbers (play still points right — it's a tape/timeline convention, not a reading-direction one) · clocks and clockwise-motion icons like refresh/redo · checkmarks and musical notation · numerals (digits stay left-to-right even inside RTL text flow — though which numeral system is conventional varies by locale: Hebrew text conventionally uses Western Arabic numerals, while Arabic text may use Western or Eastern Arabic numerals depending on region, so that's a locale decision, not a mirroring one) · logos · code snippets.

A quick sanity check for any icon you're not sure about: does it represent spatial/reading direction (mirror it), or does it represent a real-world convention unrelated to text direction, like tape transport or clock rotation (leave it alone)?

## Combining motion, sound, and haptics

When multiple feedback channels are involved, three things keep them from working against each other:

1. **Causality** — it should be obvious what caused the feedback. Trigger it on the actual causal event (the toggle actually flipping, the item actually snapping into place), and match its character to how physical that action is — a light tap gets a light haptic, a heavier commit gets a heavier one.
2. **Harmony** — visual, sound, and haptic should land on the same frame. Any lag between them (a CSS transition finishing a beat after a sound plays, say) breaks the illusion that they're one event.
3. **Utility** — only add feedback where it earns its place. Reserve haptics and sound for genuinely meaningful moments (success, error, a commit, a snap) — feedback on everything trains people to tune out all of it, including the moments that actually mattered.
