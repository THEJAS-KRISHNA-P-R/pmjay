# Scroll Storytelling (Apple Product-Page Style)

This is a different discipline from the rest of this skill: everything else here is about app chrome (controls, navigation, system UI). This file is about the scroll-driven narrative technique Apple uses on its own *marketing* pages — apple.com product pages, most famously — where the page tells a story as you scroll rather than being a static document with some transitions on it. It's the right reference when you're building a landing page, product page, or storefront that should carry that same sense of a directed, cinematic reveal.

## What's actually happening, demystified

The signature Apple hero effect — a product that rotates, unfolds, or assembles itself as you scroll — is not a video. It's a **canvas-rendered image sequence**: a "flip book" of anywhere from ~50 to a few hundred pre-rendered frames, where JavaScript maps scroll position to a frame index and draws that single frame to a `<canvas>` element.

- **Why not just use `<video>`?** Frame-accurate scrubbing — including scrubbing *backward* smoothly when someone scrolls up — is unreliable across browsers and codecs with a real video element. A canvas gives you exact, deterministic control over which frame is on screen at any scroll position, in either direction.
- **Why not just swap an `<img src>` per frame?** Decoding and repainting a freshly-loaded image on every scroll tick can visibly stutter. Drawing pre-loaded, already-decoded images to a canvas is smoother.
- **The section is pinned** (held in place, usually full-viewport) while the sequence plays through, then releases back into normal scroll once it's done — the scroll input is temporarily "spent" driving the animation instead of moving the page.
- **Copy is synced to specific points in the sequence**, not to raw scroll distance — a caption appears when the product reaches a specific rotation, not "40% of the way down the page," so the narrative beat and the visual beat land together.
- **There's a designed fallback.** Apple serves a static image instead of the full sequence on smaller screens, slower devices, and when the person has motion reduced — this is a deliberate, first-class design decision, not a degraded experience someone forgot to polish.

## Two tools, two different jobs — use both, not just one

**Native CSS scroll-driven animations** (`animation-timeline`) are the right default for the simpler 80% of scrollytelling: elements fading/sliding in as they enter the viewport, progress bars, parallax, staggered reveals. As of 2026 this has strong support in Chromium browsers and Safari, with Firefox catching up (verify current status — this shipped recently enough that it's still worth a quick check rather than assuming). It runs on the compositor thread, so it stays smooth even under main-thread load, and it needs zero JavaScript:

```css
.reveal {
  opacity: 0;
  transform: translateY(24px);
  animation-name: reveal-up;
  animation-timeline: view();          /* tracks this element's own passage through the viewport */
  animation-range: entry 0% entry 40%; /* only animate during entry, not the whole time it's visible */
  animation-fill-mode: both;           /* holds the 0%/100% state instead of snapping back at scroll-top */
}
@keyframes reveal-up {
  to { opacity: 1; transform: translateY(0); }
}
@supports not (animation-timeline: view()) {
  .reveal { opacity: 1; transform: none; } /* unsupported browsers get the resolved end-state, never a stuck mid-animation */
}
```

**GSAP ScrollTrigger driving a canvas** is still the right tool for the actual pinned, scrubbed, frame-sequence hero moment — native CSS scroll-driven animations don't yet cover *pinning* (freezing a section in place while its content changes) or this level of fine-grained scrub control. Reach for it deliberately for the one or two hero moments that need it, not as your default for every scroll effect on the page — see the restraint note below for why that distinction matters.

```js
gsap.registerPlugin(ScrollTrigger);

const frameCount = 148;
const frameURL = (i) => `/sequence/frame_${String(i).padStart(4, '0')}.jpg`;
const images = Array.from({ length: frameCount }, (_, i) => Object.assign(new Image(), { src: frameURL(i) }));
// show a preload/progress indicator while these load — a hero section that's blank
// or half-broken until every frame arrives is worse than not having the effect at all.

const canvas = document.querySelector('#hero canvas');
const ctx = canvas.getContext('2d');
const drawFrame = (i) => images[i]?.complete && ctx.drawImage(images[i], 0, 0, canvas.width, canvas.height);

if (matchMedia('(prefers-reduced-motion: reduce)').matches) {
  drawFrame(Math.floor(frameCount / 2)); // a single representative frame, no pin, no scrub — Apple's own fallback pattern
} else {
  const state = { frame: 0 };
  gsap.to(state, {
    frame: frameCount - 1,
    ease: 'none',
    scrollTrigger: { trigger: '#hero', pin: true, scrub: 0.5, start: 'top top', end: '+=3000' },
    onUpdate: () => drawFrame(Math.round(state.frame)),
  });
}
```

## What makes this feel considered rather than gimmicky

This is the same judgment call as everything else in this skill — a technique in service of the seven other principles in `design-foundations.md`, not a special exemption from them:

- **Reserve pinning/scroll-jacking for genuine hero moments, not the whole page.** A page where every section hijacks scroll input becomes exhausting and disorienting — it breaks the "where am I, where can I go" wayfinding baseline from `design-foundations.md`. Apple's own pages are mostly normal document scroll, with the frame-sequence treatment used sparingly, at the moments that earn it.
- **Never fully trap scroll input.** Someone should always be able to scroll past or skip through a pinned section faster than its "intended" pace if they choose to — don't force people to sit through the full duration of an animation to keep reading.
- **The reduced-motion fallback needs to be a real design decision, not a placeholder.** Landing on a single well-chosen static frame (as in the sketch above) respects the preference without just breaking the page. This is the same principle as the reduced-motion guidance in `accessibility-feedback.md`, applied to a marketing-page context instead of an app one.
- **Budget for the payload.** A 150-frame sequence is a lot of images — compress aggressively (modern formats, and dimensions matched to actual render size, not source resolution), and start loading a sequence shortly before someone scrolls into it rather than blocking the whole page's initial load on it.
- **Pace copy to narrative beats, not linear scroll distance.** A caption that appears exactly when the product reaches the rotation it's describing reads as intentional; one that's just evenly spaced down the page reads as a template.
