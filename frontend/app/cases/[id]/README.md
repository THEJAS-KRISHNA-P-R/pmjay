# `/cases/[id]` — Case Report & Action Workspace

## Purpose
The primary working surface for a specific billing denial evaluation. Delivers the 5-tier classification outcome, patient counter-script, CGRMS grievance draft, official PDF download, and evidence intake form.

## Architecture & Interaction Design
1. **Single-Page Flow with Sticky/Fixed TOC**:
   - Rather than hiding content behind tab widgets, the entire evaluation report is rendered sequentially on one scrollable page.
   - **Desktop**: Clean table of contents (`aside`) fixed at `top-[88px]` with zero initial scroll shift.
   - **Mobile**: Locked horizontal section pill bar fixed at `top-[60px]` with frosted blur background, completely flush under the header.
2. **Sections Included**:
   - **Overview**: Non-negotiable `CareFirstBanner`, `TierPanel` (Green/Amber/Red/Mixed/Handoff), and `DisclaimerNote`.
   - **Your Story**: Original verbatim patient narrative (`description`).
   - **Next Steps**: Step-by-step numbered instructions (`ActionSteps`) and point-of-denial verbal dialogue script (`CopyableTextBox`).
   - **Documents & Letters**: Official CGRMS complaint draft and one-click PDF generation button.
   - **Track Complaint**: Interactive local status tracker (`ComplaintStatusTracker`).
   - **Evidence & Log**: Time, staff name, and written denial notes log (`EvidenceForm`).
