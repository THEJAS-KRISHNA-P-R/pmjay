# Materials, Depth & Typography

## Materials and depth

Apple builds visual hierarchy with translucent "material" layers — blurred, tinted surfaces that sit above content — rather than flat opaque panels. Content keeps scrolling underneath; the chrome floats.

- **Build navigation bars, toolbars, and sheets as translucent layers**, not opaque strips. Content should be visible (blurred) moving underneath them.
- **Material weight signals hierarchy.** Heavier, darker materials separate large structural regions (a sidebar); lighter, airier materials draw attention to interactive elements (a floating action button). Avoid stacking two light translucent surfaces on top of each other — legibility falls apart fast.
- **Bigger surfaces should read as "thicker."** Scale blur radius and shadow depth up for large surfaces (sheets, panels) relative to small ones (chips, badges).
- **Use a scrim, not just translucency, for anything modal.** A task that blocks the rest of the UI should dim/push back what's behind it. A non-blocking, parallel panel (a popover, an inspector) can use translucency and an offset *without* a scrim, since the underlying flow isn't actually interrupted.
- **Vibrancy keeps text legible over a moving background.** Flat gray text on a blurred surface tends to wash out — bump contrast and weight slightly, and add a touch of letter-spacing, rather than relying on the same styling you'd use on a solid background.
- **Prefer soft scroll-edge fades over hard dividers.** Instead of a 1px border under a sticky header, fade a subtle blur/gradient where floating chrome overlaps scrolling content.
- **Animate materials as arriving, not just fading.** For a glass/blur surface, animating blur radius and scale together on enter/exit reads as a physical material appearing, rather than a flat opacity cross-fade.

```css
.toolbar {
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(20px) saturate(180%);
  border-top: 1px solid rgba(255, 255, 255, 0.4); /* a bright top edge reads as light catching glass */
}
```

## Liquid Glass: the current material system

Since iOS/iPadOS 26, macOS Tahoe 26, watchOS 26, and tvOS 26 (September 2025, refined in iOS 27 at WWDC 2026), the translucent layer above is a specific named material: **Liquid Glass**. It's worth understanding at the mechanical level, because the mechanics explain both why it looks the way it does and why it broke in the ways detailed below.

- **It's lensing, not blurring.** Liquid Glass bends and refracts the content behind it in real time (GPU-rendered, via Metal) rather than diffusing it the way a plain gaussian blur does. Specular highlights shift as the device moves. A `backdrop-filter: blur()` is a reasonable *approximation* on the web, but it's optically a different operation — don't expect it to read identically, and don't chase the exact look at the expense of performance.
- **It's cheap on-device because static backgrounds get cached** — the GPU only recomputes the refraction when what's behind the glass is actually moving. A naive web implementation that recalculates a heavy blur on every scroll frame will feel far more expensive than the native version looks; keep blur radius and layer count modest, and prefer `will-change`/compositing hints over recalculating filters per frame.
- **Native APIs, for context:** SwiftUI uses the `.glassEffect()` modifier inside a `GlassEffectContainer` (which auto-blends overlapping glass shapes and coordinates morphing between states); UIKit uses `UIGlassEffect` wrapped in a `UIVisualEffectView`, adopted explicitly rather than automatically.
- **Scope rule: it's for the controls/navigation layer, never for content.** Liquid Glass is meant to sit *above* content as a distinct functional layer (toolbars, tab bars, floating buttons, sheets) — not as a treatment for the content itself (lists, tables, media, the body of what someone is actually trying to read). The single biggest way teams misuse this material is applying it somewhere content should be opaque.

### Where Liquid Glass shipped wrong — and why it matters for you

This isn't a hypothetical warning; it's a documented, public case study, and it's the single best illustration of why "glass is a hierarchy tool, not a default" belongs at the top of this skill rather than buried as a caveat. When iOS 26 launched, Nielsen Norman Group published a named critique, and the specific failures are worth knowing individually because each one maps to a general rule you can check your own work against:

| What shipped wrong | The general rule it violated |
|---|---|
| Messages let people set a photo as a text-thread background; message text became hard to read against busy areas of the photo | Never place primary reading text directly on unpredictable, high-frequency-detail content. Glass/translucency needs a legibility floor regardless of what's behind it. |
| A floating search button, visually isolated from the tab bar, blended into busy backgrounds and read as ambiguous (icon? action button?) | An element's *material* shouldn't be the only thing distinguishing it from decoration — shape, position, and consistent placement still have to carry meaning on their own. |
| Tab bars and controls shrank and became more crowded to accommodate the glass treatment, pushing some targets below the 44×44pt minimum | A material change is never sufficient justification for shrinking a touch target. Legibility and hit-target size are non-negotiable inputs, not variables the aesthetic gets to trade away. |
| Buttons shimmered, tab bars rippled, and some elements animated without a clear triggering action; several users reported motion discomfort | Motion needs a cause (see `motion-and-gestures.md`'s "causality" principle) — ambient/decorative motion on system chrome that people stare at all day is a cost, not a delight. |
| No system-wide way to reduce (rather than fully disable via Reduce Transparency) the glass effect, so many people enabled accessibility settings as a workaround for a design problem, not an accessibility need | Provide a middle ground on intensity for any effect that trades legibility for aesthetics. A binary on/off is not the same as user control. |

Apple's own fix, shipped in iOS 27: a system-wide **intensity slider** giving graduated control over glass opacity (rather than the binary Reduce Transparency toggle), plus reintegrating search back into the tab bar instead of isolating it. Nine months elapsed between the criticized launch and the correction.

**What to actually do with this, on native or web:**
- Test every glass/translucent surface against the *busiest* realistic background you'll ship with (a photo, video, or saturated color) — not a clean placeholder gray. If it's illegible there, it's illegible in production.
- Never let a material choice shrink a touch target or push text weight/size down to compensate visually. Fix the material's opacity or backing tint before you touch the content.
- If you're offering a translucent theme as a product-wide aesthetic choice, offer an intensity control, or at minimum respect `prefers-reduced-transparency` — see `accessibility-feedback.md`.
- Motion on chrome that people see on every screen (not just once, in a modal) should be tied to an actual state change, not ambient.

## Concentricity: corner radius as a relationship, not a value

Liquid Glass introduced (and named) a layout discipline worth using even outside a glass context: a nested element's corner curve should share its container's curve *center*, not just look like a similar rounded shape. Apple calls this **concentricity**.

- **The formula:** `inner radius = outer radius − padding`. If a card has a 20px radius and 12px of padding, a button flush against that padding should have an 8px radius — not a value you picked because it looked fine in isolation. Copying the outer radius onto the inner element (a very common default) is the mistake this corrects.
- **Capsule (fully-rounded/pill) shapes satisfy this automatically**, since their radius is defined as half their height — that's part of why capsule buttons and search fields feel so at-home nested inside other rounded containers without extra math.
- **It's a relationship to an actual nested element, not a rule to apply uniformly everywhere.** Applying one large radius to all four corners of a container that only has child elements along one edge (say, a toolbar at the top but nothing at the bottom) can look worse, not better — critics of macOS Tahoe's window corners have made exactly this point about corners with no adjacent element to be concentric *with*. Use it where there's something to relate to; don't cargo-cult the aesthetic onto edges with nothing nested in them.
- **SwiftUI:** `RoundedRectangle(cornerRadius: .containerConcentric, style: .continuous)` calculates this automatically relative to the nearest container.

```css
:root { --r-outer: 20px; --pad: 12px; }
.card { border-radius: var(--r-outer); padding: var(--pad); }
.card > .button { border-radius: calc(var(--r-outer) - var(--pad)); } /* not var(--r-outer) again */
```

## SF Symbols

SF Symbols are Apple's icon system, optically matched to the San Francisco font so icons and adjacent text scale and weight together. Even when you can't use the actual symbol set (its usage terms restrict it to Apple-platform contexts), the discipline it encodes is worth replicating with whatever icon set you do use.

- **Nine weights** (ultralight → black), matching SF Pro's text weights one-to-one — pick the icon weight to match the weight of the text it sits next to, not a weight you happened to like.
- **Three scales** (small / medium / large) relative to the current text size.
- **Four rendering modes:** monochrome (single flat color — the default), hierarchical (Apple auto-applies a small set of opacity levels to a symbol's own layers to suggest depth), palette (you assign specific colors to specific layers yourself), and multicolor (the symbol's inherent, baked-in coloring, for symbols where that's meaningful). Use **hierarchical** for depth/hierarchy within a symbol; use **variable color** (a separate mechanism — a fillable subset of a symbol's layers) to communicate *change or magnitude* (signal strength, a battery level), not depth. Conflating the two reads as a bug ("why is this icon's opacity flickering") rather than a signal.
- **Avoid ultralight/thin/light weights** at small sizes or for anything that needs to hold up in bright light or for people with low vision — this is the same legibility logic as the type-weight guidance below, applied to icons.
- **On the web:** pick one icon family and hold it to the same discipline — one visual system, a small number of weights used deliberately, sized and colored to match adjacent text rather than left at whatever default the library ships.

## Color: semantic roles, not hex values

Apple's system colors (`systemBlue`, `label`, `secondaryLabel`, `tertiaryLabel`, `systemBackground`, and the rest) are **adaptive roles, not fixed values** — they're documented as colors that automatically remap for light mode, dark mode, and Increased Contrast, and Apple does not publish a guaranteed hex value for them. The widely-cited `#007AFF` for `systemBlue` is a community-measured approximation, not an Apple spec, and it can and does drift slightly between OS versions. Two implications:

- **On native platforms, reference the semantic API** (`Color(.label)`, `Color(.systemBackground)`) rather than hardcoding a hex value you measured off a screenshot — you'll silently drift out of sync with the system as it updates.
- **On the web, don't try to pixel-match a moving target.** Instead, define your own small set of role-based custom properties (`--color-label`, `--color-secondary-label`, `--color-background`, `--color-background-secondary`) and independently tune each for both light and dark — treat the widely-cited hex values as a reasonable *starting point* for a first pass, not ground truth to defend.

Two specific, non-obvious behaviors worth designing around:

- **Tint colors get lighter in dark mode and darker in light mode** — a tint tuned to look right in one mode is not guaranteed to have sufficient contrast in the other. Check both independently; don't assume symmetry.
- **Dark mode is not an inversion — it needs *more* separation between background levels, not less.** A common mistake is picking dark grays that sit close together (they look "correctly dark" in isolation but muddy together on screen). Spread them out:

```css
/* Light: levels can sit close together */
--bg-primary-light: #FFFFFF;
--bg-secondary-light: #F5F5F7;

/* Dark: levels need more separation, not less */
--bg-primary-dark: #1C1C1E;
--bg-secondary-dark: #2C2C2E;
```

Opaque (non-translucent) separators are used specifically where transparency would create visual artifacts — most notably intersecting grid lines, where overlapping semi-transparent strokes produce moiré-like optical noise. Reach for a flat, opaque hairline there instead of continuing the translucency pattern.

## Typography

Apple's system typefaces reshape themselves at different sizes rather than using one fixed style scaled up or down — the same discipline is worth applying anywhere type spans a wide size range.

- **Know the family, not just "the Apple font."** SF Pro is the system sans-serif (iOS, macOS, tvOS); SF Compact is its rounded counterpart for watchOS; SF Mono is for code and tabular data; New York is the companion serif for editorial content; SF Arabic and SF Hebrew are localized variants for RTL scripts. Reaching for the right one (SF Mono for a table of numbers, New York for a long-form article) does real work.
- **Avoid light font weights.** Apple's own guidance is explicit here: prefer Regular, Medium, Semibold, or Bold over Ultralight, Thin, or Light — light weights lose legibility at small sizes, in bright ambient light, and for people with visual impairments.
- **Letter-spacing (tracking) should vary by size, never be one fixed value.** Large display type needs *negative* tracking — at big sizes, letters read as too far apart by default. Small text needs slightly *positive* tracking for legibility. A single fixed `letter-spacing` value will be wrong at one end of your type scale or the other.
- **Line-height (leading) moves the opposite direction from tracking.** Tight leading for large headlines, looser leading for small body copy. Loosen it further for scripts with tall ascenders/descenders, or for dense, information-heavy UI where you need more visual separation between lines.
- **Build hierarchy from weight + size + leading together**, not size alone — a heavier weight adds visual presence without taking up more horizontal space, which is often the better lever.
- **Respect the user's text-size preference (Dynamic Type on Apple platforms).** Lay out with relative units (`rem`/`em`) rather than fixed pixel spacing, so the layout scales gracefully instead of breaking when text grows. Dynamic Type's actual scale, its non-uniform per-style scaling behavior, and the concrete bugs it exposes live in `accessibility-feedback.md` — read that before assuming "I used `rem`" is the whole job.
- **Default to the platform's system font.** It already ships with size-specific optical adjustments, tracking tables, and legibility tuning that took Apple's type team a long time to get right — only reach for a custom typeface when you have a specific reason to.

```css
:root { font: 100%/1.5 system-ui, sans-serif; } /* comfortable default leading for body text */

.display {
  font-size: clamp(2rem, 5vw, 4rem);
  line-height: 1.05;        /* tight leading at large sizes */
  letter-spacing: -0.02em;  /* negative tracking as size grows */
  font-optical-sizing: auto;
}
```
