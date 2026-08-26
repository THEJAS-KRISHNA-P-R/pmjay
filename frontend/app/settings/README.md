# `/settings` — User & Accessibility Settings

## Purpose
Manages local client preferences, accessibility settings, and browser privacy controls.

## Key Features
1. **Local Contact Info & Preferred Language**:
   - Optional fields saved exclusively in browser `localStorage` (`lib/localPrefs.ts`).
   - Zero centralized database retention.
2. **Accessible Large Text Mode**:
   - Pre-paint DOM attribute toggle (`data-text-scale="large"` on `html`) preventing layout shift or flash of unscaled text.
3. **Data Purge & Privacy Controls**:
   - One-click deletion of saved profile data.
   - One-click purge of all locally stored case history and notes.
