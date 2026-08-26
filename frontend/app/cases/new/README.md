# `/cases/new` — New Case Evaluation

## Purpose
Dedicated standalone page for starting a new case evaluation, embedded within the `AppShell` product chrome.

## Architecture
1. **2-Column Balanced Workspace**:
   - Left Column (7 cols): Full `IntakeForm` with multiline text area, language selector, and submit action.
   - Right Column (5 cols): NHA Patient Protection summary card, key guidelines, and emergency helpline shortcuts (`14555`, `15100`, `112`).
2. **Submission Flow**:
   - Submits narrative to `POST /api/v1/cases`.
   - On success, saves minted UUIDv4 to local `caseHistory` and immediately redirects to `/cases/[id]`.
