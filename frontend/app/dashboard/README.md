# `/dashboard` — Case Dashboard

## Purpose
The personalized management hub for a family's ongoing cases. Displays evaluated cases, complaint tracking states, and quick-start actions for new disputes.

## Key Architecture & Design Details
1. **Local-First Zero Account Storage**:
   - Because PMJAY Advocate avoids centralized accounts for privacy, the dashboard queries `lib/caseHistory.ts` to retrieve cases created by or viewed on this specific browser device.
   - Case data is populated asynchronously via direct API fetches or cached metadata.
2. **Needs Attention & Status Tracker**:
   - Cases with pending evidence or unresolved grievances are prioritized.
   - Families can toggle complaint submission status (Draft, Submitted to CGRMS, Under Review, Resolved) locally.
3. **AppShell Layout**:
   - Rendered within `AppShell` with fixed navigation, responsive mobile bottom tabs, and clean 2-column widescreen balancing.
