# App Icons, Widgets & Live Activities

The rest of this skill covers the app you're inside of. This file covers the places your product shows up *outside* its own window: the Home Screen icon, a widget, a Live Activity/Dynamic Island, a push notification, and the system share sheet. Each has its own tooling, its own hard constraints, and its own specific ways to get it wrong.

## App icons: Icon Composer and the six appearances

Since iOS/iPadOS/macOS 26, an app icon is a **layered, resolution-independent source file**, not a flat exported PNG — because it has to render correctly in six distinct appearances (Default, Dark, Clear Light, Clear Dark, Tinted Light, Tinted Dark), and a single bitmap can't adapt across all of them. Apple's tool for this is **Icon Composer**, which ships with Xcode.

- **You author three modes; the system derives the other three.** Design for Default, Dark, and Mono (a grayscale/clear annotation) — Clear and Tinted variants, in both light and dark, get generated from that Mono annotation automatically. You're not hand-crafting six separate icons.
- **Structure: layers grouped into up to four depth groups.** Import vector (SVG) or raster (PNG) layers exported from Figma/Sketch/Illustrator/Photoshop at full canvas size, so their positioning carries over directly. Per layer, you control Liquid Glass on/off, opacity, blend mode, and fill; per group, blur, translucency, shadow type (neutral or chromatic), and specular highlights.
- **The canvas is 1024×1024 for iPhone/iPad/Mac, 1088×1088 for Watch.** Official templates for Figma, Sketch, Illustrator, and Photoshop are published at developer.apple.com/design/resources.
- **The four mistakes that show up first, in order of how often they bite:**
  1. Exporting one flattened image instead of separate layers — Icon Composer has nothing to add real depth *to*.
  2. Baking your own glossy highlights or gradients into the artwork — they fight the system's real-time lighting instead of working with it.
  3. Pushing elements into the corners — the automatic mask crops them; keep meaningful content away from the edges.
  4. Shipping without checking Clear and Tinted modes specifically — an icon that reads perfectly in Default can be nearly invisible in Clear, and this is the mode people actually miss in testing since it's not the default preview.
- **Design guidance that holds regardless of tooling:** keep shapes simple and graphic rather than finely detailed (detail gets lost at Home Screen size and under glass rendering), lean on one or two colors rather than a busy palette, skip text entirely (a long-standing rule, still true), and avoid photographic or screenshot-style icons.
- **As of iOS 27, Apple itself dialed back icon translucency** for sharper, more legible results — the same correction that hit Liquid Glass generally (see `materials-typography.md`) applied specifically to icons too. If an icon looks murky under glass, reducing translucency before redesigning the artwork is the first thing to try.
- **The honest caveat for your actual work:** this whole system is native-app tooling. If you're shipping a **PWA** (which is closer to what you'd actually do), the layered Icon Composer pipeline isn't available to you — as of this writing, PWA home-screen icons on iOS/iPadOS/macOS 26 render poorly in Clear and Tinted modes specifically, and there's no documented, reliable way to opt a web-app icon into the full six-appearance system. The pragmatic move for a PWA today is a single, high-contrast `apple-touch-icon` that holds up as a flat image, not an attempt to chase a system you can't actually plug into yet.

## Widgets: glanceable, budgeted, and mostly read-only

- **Home Screen families:** small, medium, large — and, starting with iOS 27 (autumn 2026), an extra-large 4×6 size. **Lock Screen families are different and separate:** `accessoryCircular`, `accessoryRectangular`, `accessoryInline`.
- **Lock Screen widgets are read-only.** `Button(intent:)` interactivity is only available in Home Screen and StandBy-sized widgets — not on the Lock Screen. That's a platform constraint, not a design choice you can override, so don't design a Lock Screen widget around an interaction it structurally can't have.
- **Refresh is budgeted, not live** — WidgetKit gives each widget roughly 40–70 timeline refreshes per day in production. A widget is the wrong tool for anything that needs to update in real time; that's what a Live Activity is for (see below), and the two are meant to complement each other, not duplicate each other. A single WidgetKit extension target commonly ships both: a weather app might have a Home Screen widget for today's forecast (scheduled refresh) and a Live Activity that only appears during an active severe-weather alert (real-time, temporary).
- **Since iOS 26, widgets pick up Liquid Glass** — transparent or tinted backgrounds that morph with whatever's behind them. Design the content to stay legible on its own, since a person can flip Reduce Transparency and flatten the effect at any time (see `accessibility-feedback.md`).
- **Glanceable first, interactive second.** A widget should communicate its core information the instant you look at it, the same way a watchOS complication does — interactivity (a quick toggle, a checkbox) is a bonus for the Home Screen/StandBy cases where it's available, not the reason the widget exists.

## Live Activities & Dynamic Island

Built with **ActivityKit**, rendered by the system across the Dynamic Island, the Lock Screen, and StandBy — you publish state changes, and iOS decides how and where to render them. You don't draw to the Island directly.

**Four regions to design, each with its own discipline:**

- **Compact** (the default, everyday state — small leading/trailing content around the camera cutout): this is what people see almost all the time an activity is running, so keep its layout *consistent* update to update. An activity that visually reshuffles itself every update reads as broken, not dynamic.
- **Minimal** (shown when more than one Live Activity is competing for space, so yours gets squeezed to the smallest possible footprint): design it as a single icon or a tiny number that still makes sense with zero surrounding context. If it can't stand alone like that, it will look broken the moment a second activity shows up — which will happen in real usage, not just as an edge case.
- **Expanded** (on long-press, or shown automatically for a few seconds after a significant update): treat this as a rich tooltip, not a mini-app — a small number of high-confidence, immediate actions (rate, confirm, mute, check in), never a multi-step flow. Anything more belongs in the actual app.
- **Lock Screen** (a distinct fourth layout, not just a bigger version of expanded): roughly a 360pt-wide banner, variable height up to about 160pt. This is where most people actually interact with a Live Activity, and it can carry more information density than the other three — secondary data, color-coded status, controls for common actions.

**Concrete constraints worth knowing before you design around them:**

- The Dynamic Island's corners are rounded at 44 points, deliberately matching the TrueDepth camera housing's own rounding — the same concentricity logic from `materials-typography.md`, applied to an actual physical component. Keep content in compact and expanded states properly aligned around that cutout rather than centered as if it weren't there.
- A Live Activity can update for up to **8 hours** before the system automatically ends it and removes it from the Dynamic Island; it can then persist on the Lock Screen for up to **4 more hours** (12 hours total, maximum) before being cleared. Design for genuinely time-bound events — a delivery, a ride, a timer, a live score — not open-ended state that needs to persist indefinitely.
- Each Live Activity runs sandboxed: **no direct network access, no direct location updates.** State arrives via push notification or from the host app. Plan your "live" data source around that constraint from the start, not after the UI is built.
- People can disable Live Activities per app in Settings — design the fallback (a normal notification) as a real path, not an afterthought.
- Not every iPhone has the physical Dynamic Island hardware; Live Activities still need to work correctly via the Lock Screen and StandBy on models that don't. Never assume Island hardware is present.
- As of iOS 27, **test compact and minimal layouts in landscape specifically** — this wasn't a consideration before and is a new, real gap if skipped.

## Push notifications: rich content and actions

A plain push notification (title, body, sound, badge) needs no extra work. Rich content — an attached image, video, audio, GIF, or a custom expanded UI — needs two separate, optional app extensions:

- **Notification Service Extension (NSE)** — intercepts a notification *before* it displays and enriches it: downloads and attaches media, adjusts text. Requires `"mutable-content": 1` in the payload's `aps` dictionary; the extension won't fire without it, and a silent notification (sound/badge only, no alert) can't be modified at all — plan the payload with this in mind, not after.
- **Notification Content Extension (NCE)** — provides a fully custom UI shown when the person expands the notification (a carousel across multiple images, for instance, rather than a single static one).
- **Actions** are declared via a `category` identifier in the payload, matched against up to **four** `UNNotificationAction` entries registered in the extension's `Info.plist`. Keep actions to genuinely one-tap operations (confirm, dismiss, reply, mark as read) — anything requiring a follow-up screen belongs in the app, the same discipline as Live Activity actions above.
- Two content shapes cover most real use cases: a single attached image/media alongside the message (good for a promotion, an update, an event reminder), or a swipeable multi-image carousel (good for showcasing a few related items in one notification rather than sending several).

## The share sheet: use the system one, don't rebuild it

Apple's own guidance here is unambiguous and worth stating plainly, because it's a genuinely common mistake: **when people tap Share, they expect the system share sheet — don't substitute a custom in-app share modal that looks similar but isn't the one they already know.** The system-provided composition view exists specifically so sharing feels consistent across every app; reach for a fully custom interface only when you have a real reason to, not as a default styling choice.

- **Triggering the share sheet from your own app** is the common case and is simple: hand `UIActivityViewController` (or the SwiftUI `ShareLink`) the content to share. On the web, the direct equivalent is the **Web Share API** (`navigator.share()`) — same principle applies: prefer the platform's native share UI over a custom "share to X/Y/Z" modal you built yourself, on both native and web.
- **Building a Share Extension** (so your app becomes a destination inside *other* apps' share sheets) is a separate, more involved undertaking — a distinct extension target, native-only, with no web equivalent.
- **Your own app-specific actions appear first**, ahead of general system-wide actions (Add to Files, AirPlay, etc.) — but people can reorder and prune the list themselves, so never assume your action holds a fixed position.
- Keep the initial preview/compose surface small and simple. It's tempting to make the compose view tall and rich, but extra height risks landing behind the keyboard once someone starts typing, forcing a scroll — a small, efficient preview with room to grow beats a large one that immediately fights the keyboard.
