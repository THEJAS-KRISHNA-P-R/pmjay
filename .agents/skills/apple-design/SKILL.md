---
name: apple-design
description: Apple's design language — motion, materials, typography, platform conventions, and design philosophy across Apple's platforms, drawn from the Human Interface Guidelines and WWDC talks. Covers Liquid Glass (mechanics, concentricity, documented accessibility failures), SF Symbols, Dynamic Type, semantic color, app icons, widgets, Live Activities/Dynamic Island, concrete component specs (touch targets, tab bars, corner radii), and Apple-style scroll storytelling (apple.com's frame-sequence technique) — not just high-level principles. Use whenever building, reviewing, or critiquing UI that should feel "Apple-like" — gesture-driven interactions, spring animation, sheets, translucent materials, dark mode, safe areas, tab bars, haptics, accessibility handling. Trigger for SwiftUI/UIKit/AppKit AND web/React work chasing Apple-style polish, even if the user just says "make this feel more native," "add Liquid Glass," or "make my landing page scroll like Apple's site," without naming Apple explicitly.
---

# Apple Design

This skill packages Apple's design language into something you can actually apply: not just "use rounded corners," but the underlying reasoning Apple's own designers give in the Human Interface Guidelines and WWDC talks, translated into concrete values, checklists, and code you can use on native platforms or the web.

Reach for this skill any time you're building, reviewing, or giving feedback on an interface that should feel considered, physical, and calm rather than mechanical — whether that's a native Swift app or a web app chasing that same polish.

## How Apple thinks about design

Apple's own current framework — laid out in "Principles of great design" at WWDC 2026 — names eight principles. When a decision doesn't have an obvious answer, come back to these:

**Purpose** (design with intention; deciding what *not* to build matters as much as what you build) · **Agency** (put people in control — real choices, plus forgiveness for mistakes) · **Responsibility** (privacy as a human right; anticipate misuse, especially in AI-driven features) · **Familiarity** (build on metaphors and conventions people already know; consistency beats novelty) · **Flexibility** (adapt to different devices, contexts, and abilities rather than one ideal user on one ideal screen) · **Simplicity** (removing friction, *not* minimalism — sometimes the simplifying move is to add context, not strip it) · **Craft** (uncompromising attention to typography, color, motion, and performance detail) · **Delight** (the *result* of getting the other seven right, never confetti bolted on at the end).

Full explanations and Apple's own illustrative examples for each live in `references/design-foundations.md` — read it before making any judgment call that isn't obviously covered elsewhere in this skill. That file also covers Apple's practical patterns (the four flavors of feedback, the four questions every screen should answer) and process notes.

## The current design language: Liquid Glass

As of iOS 26 / iPadOS 26 / macOS Tahoe 26 / watchOS 26 / tvOS 26 (shipped September 2025) and refined in iOS 27 (WWDC 2026), Apple's controls layer is built from a material called **Liquid Glass** — a translucent, refractive surface, not a simple blur, that floats above content as a distinct functional layer and morphs between states. When someone asks for "Apple-style" UI today, this is what they mean by default — not the flat iOS 7–18 look, unless they say otherwise.

The load-bearing caveat, learned the hard way: Liquid Glass shipped in 2025 with real, extensively documented legibility failures — text over busy or photographic backgrounds lost contrast, floating controls blended into content, Nielsen Norman Group published a detailed critique, and iOS 26 adoption visibly lagged prior years as a result. Apple's own iOS 27 response added a user-facing intensity slider specifically because there was no middle ground between full transparency and switching it off entirely. **The lesson generalizes past Apple's own product: glass/translucency is a hierarchy tool that has to earn its contrast against content — it isn't a default surface you reach for because it looks nice in a static mockup.** See `references/materials-typography.md` for the rendering mechanics, concentricity, the specific documented failure cases (and why each one failed), and how to get the aesthetic without repeating the mistake — on native platforms or the web.

This is a genuinely fast-moving area less than a year old at the time of writing. If precision on the current OS version's exact behavior matters for what you're building, verify with a web search rather than assuming this summary is still current.

## The core idea behind Apple's motion

Apple's interfaces feel "alive" rather than "operated" because of one rule, repeated everywhere: **motion should start from wherever the interface currently is, carry the user's momentum, and be interruptible at any instant.** A panel you're swiping away should be grabbable and reversible mid-flight, not committed the moment it starts animating.

This is why Apple's UI leans on physics-based springs instead of fixed-duration, fixed-curve animations — a spring can smoothly change its target without a visible seam, a scripted `ease-out` cannot.

Full detail, code, and concrete spring values live in `references/motion-and-gestures.md`. The short version:

- Give feedback the instant a finger or pointer touches something — never wait for release.
- Track drag gestures 1:1, respecting the exact point the user grabbed, not the element's center.
- Animate from the element's *current, on-screen* position/value, never from its "should be" value — this is what makes interruption possible.
- Default to a **critically damped** spring (no overshoot) for most UI; add a little bounce only when the motion is carrying real momentum from a flick or throw.
- When a drag ends, hand the release velocity straight into the follow-up animation so there's no seam between "being dragged" and "animating."
- At edges, resist progressively (rubber-banding) instead of stopping dead.

## Materials, depth, and typography

Apple builds hierarchy with translucent "material" layers (Liquid Glass, plus the older blur+tint vocabulary it extends) rather than flat panels, gives every corner radius a mathematical relationship to its container ("concentricity"), and shapes type differently at every size rather than using one fixed style. See `references/materials-typography.md` for the full mechanics: Liquid Glass's rendering model and where it's valid to use, concentric corner-radius math with CSS/SwiftUI examples, SF Symbols (nine weights, three scales, four rendering modes, and which to reach for), Apple's semantic/adaptive color system (and why hardcoding its hex values is a mistake), the system type families, and `backdrop-filter` recipes for the web.

## Platform conventions

iOS, iPadOS, macOS, watchOS, and visionOS each have their own vocabulary of controls, layout conventions, and concrete sizing (tab bars vs. sidebars, safe areas, the menu bar, SF Symbols, the Digital Crown, spatial depth). See `references/platforms.md` before designing or reviewing anything platform-specific — getting the *right* control for the platform matters as much as styling it correctly, and using the wrong numeric spec (a touch target, a bar height) is as much of a miss as using the wrong control.

## Icons, widgets, notifications, and sharing

Surfaces where a product shows up *outside* its own window, each with its own tooling and hard constraints: the Home Screen icon (now a layered source file with six required appearances, authored in Icon Composer), widgets (budgeted refresh, and Lock Screen widgets are read-only — not a design choice, a platform constraint), Live Activities/Dynamic Island (four distinct regions to design, an 8-hour-then-12-hour hard lifetime, a sandboxed data model), push notifications (rich media needs a Notification Service Extension, up to four actions), and the share sheet (use the system one — don't rebuild it as a custom modal). See `references/system-surfaces.md`, including the honest caveat that the icon pipeline is native-only and what to actually do for a PWA instead.

## Scroll storytelling for marketing pages

A different genre from the rest of this skill: apple.com's product pages tell a story as you scroll — a pinned, scroll-scrubbed frame sequence rather than a static document with some transitions on it — and narrative-paced copy reveals synced to specific beats in that sequence, not to raw scroll distance. Reach for `references/scroll-storytelling.md` for landing pages, product pages, and storefronts specifically (not app UI) when the goal is that same directed, cinematic feel. It covers how the actual technique works (canvas frame sequences, not video, and why), when native CSS scroll-driven animations are the right call versus when you actually need GSAP ScrollTrigger and a pinned section, and the restraint that keeps the effect from reading as a gimmick instead of a considered choice.

## Accessibility and multimodal feedback

Reduced motion doesn't mean "no feedback," Dynamic Type is a twelve-step scale that doesn't all move at the same rate, VoiceOver has specific rules for labels and grouping, and right-to-left layout mirrors some things and deliberately not others. See `references/accessibility-feedback.md` for all of it, plus the rules Apple uses for combining motion, sound, and haptics without them fighting each other. Treat this file as required reading, not optional polish — the Liquid Glass launch above is the concrete, recent proof that skipping it produces a real, shipped, widely-criticized product failure, not just a theoretical risk.

## Quick reference

| Need | Approach | Concrete starting point |
|---|---|---|
| Default UI spring | Critically damped, no bounce | damping ≈ 1.0, response ≈ 0.3–0.4s |
| Momentum/flick spring | Slight underdamped bounce | damping ≈ 0.8, response ≈ 0.3–0.4s |
| Handing off a drag's velocity | Feed release velocity into the settle animation | most spring libraries take a raw `velocity` param |
| Where a flick should land | Project the resting point from velocity, then snap to nearest target | exponential decay, not `v²/2a` |
| Interrupting a running animation | Read the *live* on-screen value, restart the spring from there | never restart from the logical/target value |
| Edge of a scrollable/draggable region | Resist progressively | rubber-band function, not a hard stop |
| Translucent chrome | Blur + tint layer, content scrolls underneath — controls layer *only*, never behind primary reading content | `backdrop-filter: blur() saturate()` |
| Nested corner radius | Inner radius = outer radius − padding, not a copy of the outer value | `border-radius: calc(var(--r-outer) - var(--pad))` |
| Minimum touch/tap target | 44×44pt on every interactive element, regardless of the visible glyph's size | pad the hit area, don't shrink the target |
| Icon weight | Match the adjacent text's weight | SF Symbols: 9 weights, ultralight→black |
| Large display type | Negative tracking, tight leading | e.g. `letter-spacing: -0.02em` |
| Small body type | Near-zero or slightly positive tracking, looser leading | scale with the user's text-size setting |
| Reduced motion | Cross-fade instead of slide/spring/parallax | `@media (prefers-reduced-motion: reduce)` |
| RTL layout | Mirror reading order, nav, icons with direction; do *not* mirror media controls, clocks, or numerals | see `references/accessibility-feedback.md` |
| Scroll-driven reveal (fade/slide/parallax) | Native CSS, no JS needed | `animation-timeline: view()` + `animation-range` |
| Scroll-driven pinned hero (product spins/assembles) | GSAP ScrollTrigger + canvas frame sequence | see `references/scroll-storytelling.md` |
| Lock Screen widget interactivity | Not available — read-only by platform constraint | `Button(intent:)` only works in Home Screen / StandBy sizes |
| Live Activity lifetime | Design for genuinely time-bound events only | 8hr active max, then 4hr more on Lock Screen, 12hr hard ceiling |

## When reviewing existing work

Walk through, in order: Does it respond instantly to touch/click? Can any in-flight animation be grabbed and reversed? Does motion carry the gesture's velocity, or does it feel like it "resets" into a new animation? Is translucency used with a real content layer underneath it and enough contrast to stay legible, or just as decoration over busy content (the exact failure Liquid Glass shipped with in 2025)? Do nested corners actually relate to their container, or do they just happen to look similar? Does type look considered at every size, or is it one style stretched across all of them? Would this still make sense to someone using Reduce Motion, Reduce Transparency, or the largest Dynamic Type size? That order roughly matches how much a violation breaks the "alive" feeling and how likely it is to make the interface genuinely unusable for someone — response, interruptibility, and legibility problems are the most serious.
