# Motion & Gestures

The mental model behind Apple's motion design (rooted in Apple's "Designing Fluid Interfaces" WWDC talk): an interface feels physical, not operated, when it behaves like an object in the real world — it responds the instant you touch it, moves continuously with you, keeps going when you let go, resists at its limits, and can be caught and redirected mid-motion. Everything below is a technique in service of that.

## 1. Kill latency first

Nothing else on this page matters if there's lag between input and response — the feeling of "directness" collapses the moment a delay is noticeable.

- Show feedback on **press/pointer-down**, not on release. A button that only highlights on `click` feels unresponsive even if the actual action is fast.
- Audit anything sitting on the input path that isn't strictly necessary: debounces, artificial delays, transition-end waits.
- During a drag, slider, or any continuous gesture, update the UI in lockstep with the pointer for the *entire* gesture — don't wait until release to animate.

```css
/* Feedback fires on press, not on release */
.button:active {
  transform: scale(0.97);
  transition: transform 100ms ease-out;
}
```

## 2. Direct manipulation

When something is dragged, it needs to stick to the pointer exactly — including the offset from wherever the user actually grabbed it. Re-centering the element under the cursor on grab is an immediate tell that the interface isn't really "physical."

```js
el.addEventListener('pointerdown', (e) => {
  el.setPointerCapture(e.pointerId); // keep tracking even if the pointer leaves the element
  const grabOffsetY = e.clientY - el.getBoundingClientRect().top;
  // record position + timestamp on each pointermove so you can derive velocity later
});
```

Keep a short rolling history of recent positions and timestamps (not just the latest point) — you'll need it to compute velocity at release.

## 3. Interruptibility — the load-bearing principle

Any animation the user can plausibly want to interrupt must be interruptible *at any frame*, and reversible without waiting for it to finish. A sheet that's closing should follow the finger back open the instant it's grabbed again — it shouldn't have to finish closing first.

- Never block input while a transition is running.
- **Always start the next animation from the current, on-screen (presentation) value** — never from the target/logical value. Reading the live transform and animating from there is what makes a mid-flight grab look continuous instead of jumpy.
- Prefer spring-driven motion over CSS `@keyframes`/fixed transitions for anything gesture-driven — springs naturally animate from wherever they currently are, which is exactly the property interruption needs.
- When a gesture reverses direction, the velocity should carry through into the new animation rather than being discarded — a hard velocity reset reads as a "wall." Use a spring implementation that re-targets from the current velocity rather than swapping in a brand-new animation from rest.
- Split 2D drags into independent X and Y springs. A single spring driving a 2D distance will desync if X and Y have different velocities at handoff.

## 4. Prefer springs to fixed-duration animation

A scripted, fixed-duration animation can't respond to new input mid-flight — you can only replace it. A spring can just be re-targeted, and the motion stays continuous. Use springs for anything the user directly touches; fixed-duration easing is fine for things that are purely programmatic (a toast entrance no one is dragging, for example).

Apple's spring model uses two parameters that map more directly to how a designer thinks than raw physics does:

- **Damping ratio** — how much the motion overshoots before settling. `1.0` is *critically damped*: it reaches the target with no bounce. Below `1.0`, it overshoots and oscillates before settling; lower values bounce more.
- **Response** — roughly, how quickly the spring reacts to a new target, in seconds. This is *not* a fixed duration — a spring doesn't have one; how long it takes to visually settle emerges from damping + response together.

**Rules of thumb:**

- Default most interface motion to **damping ≈ 1.0** (no overshoot) — it reads as controlled and doesn't call attention to itself.
- Only introduce bounce (**damping ≈ 0.7–0.85**) when the motion is a direct continuation of a gesture that had real momentum — a flick, a thrown card, a release. Bounce on something that just faded or slid in programmatically (a menu opening, say) reads as noise, not physicality.
- A useful starting point for response is **0.3–0.4s** for most UI-scale motion; tighten it for small, snappy controls, loosen it slightly for large surfaces like sheets.

```js
import { animate } from 'motion';

// Default: no overshoot
animate(el, { y: 0 }, { type: 'spring', bounce: 0, duration: 0.4 });

// Momentum-driven: a little bounce, because a flick preceded it
animate(el, { y: target }, { type: 'spring', bounce: 0.2, duration: 0.4 });
```

## 5. Hand off velocity at the seam between drag and animation

The moment a gesture ends, the follow-up animation should continue at the exact velocity the finger was moving — otherwise there's a visible "seam" between dragging and animating, which is usually the single biggest tell that an interaction *isn't* Apple-quality.

Most spring libraries accept raw release velocity directly. If an API wants a *relative* velocity instead, normalize by the remaining distance:

```
relativeVelocity = releaseVelocity / (targetValue - currentValue)
```

## 6. Project momentum instead of snapping to the nearest point

Don't decide the resting position purely from where the gesture ended — factor in how fast it was moving, the same way scroll deceleration does. This is what makes a flick feel like it actually "throws" something rather than just letting go of it.

```js
// decelerationRate ~0.998 feels like normal scrolling; ~0.99 feels snappier
function project(initialVelocityPxPerSec, decelerationRate = 0.998) {
  return (initialVelocityPxPerSec / 1000) * decelerationRate / (1 - decelerationRate);
}

const projectedRestingPoint = currentPosition + project(releaseVelocity);
const target = nearestSnapPoint(projectedRestingPoint);
animateSpringTo(target, { velocity: releaseVelocity }); // then hand off velocity per §5
```

Note this is an exponential-decay projection, not the constant-deceleration `v²/(2a)` formula — the exponential form is what actually matches how scrolling and flick gestures feel.

## 7. Keep motion paths spatially consistent

- **Symmetric enter/exit.** If a panel enters from the right, it should exit to the right. Entering from one direction and exiting through another reads as disconnected.
- **Anchor to the trigger.** A menu, popover, or sheet should visually originate from whatever opened it — set its transform-origin at the trigger rather than the screen or element center, so the spatial relationship is obvious.
- **Mirror the easing on round-trip transitions**, so the return path doesn't feel like a different animation from the outbound one.

## 8. Let intermediate motion hint at the destination

People predict an end state from a trajectory, not just the final frame. Motion that visibly moves *toward* where it's going (say, a control that grows in the direction it's about to end up) reads as more intentional than motion that just linearly interpolates between two states.

## 9. Rubber-banding at boundaries

A hard stop at a scroll or drag limit reads as "broken" or "frozen." Progressive resistance — the element still follows the gesture, just less and less — reads as "there's nothing further here" instead.

```js
// The further past the boundary, the less the element follows
function rubberband(overshoot, dimension, constant = 0.55) {
  return (overshoot * dimension * constant) / (dimension + constant * Math.abs(overshoot));
}
```

## 10. Gesture recognition details

- **Tap:** highlight on touch-down, commit on touch-up. Give it a little hit-padding (~10px) and let the user cancel by dragging off the target before release.
- **Drag/swipe:** require a small movement threshold (~10px) before committing to a direction — this avoids misfiring on what was meant to be a tap.
- Where possible, track multiple plausible gestures in parallel from the first movement, and only cancel the ones that didn't match once intent is clear — recognizers that only fire on a *final* state (a generic `swipeleft` event, say) throw away the continuous feedback you need along the way.
- Minimize gesture-disambiguation delays. Double-tap detection, for instance, unavoidably delays single-tap response — only pay that cost where a real double-tap gesture exists.

## 11. Frame-level smoothness

- Keep per-frame positional change small enough to avoid visible "strobing" during fast motion.
- For very fast movement, a touch of motion blur/stretch reads as smoother and faster than a hard, sharp streak.
- Animate only compositor-friendly properties (`transform`, `opacity`) and drive them off `requestAnimationFrame` (the web equivalent of `CADisplayLink`) rather than timers.
