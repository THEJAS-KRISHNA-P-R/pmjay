# `frontend/app/case/[id]`

One file — `page.tsx` — and it's the most important file in the frontend: the page a family actually reads their answer on. Everything in `frontend/app/components/` exists to be composed here.

## The composition, top to bottom, and why each conditional is what it is

`CasePage` fetches once on mount (`useEffect`, keyed on `params.id`, with a `cancelled` flag so a fast route change can't set state on an unmounted component) and renders one of three states: loading, error, or the loaded case. The loaded case always renders `CareFirstBanner` and `TierPanel` first, unconditionally — then a sequence of *conditionally* rendered panels, each gated on the specific field actually being present in the response:

```tsx
<CareFirstBanner message={caseData.care_first_message} />     {/* always */}
<TierPanel outcome=... message=... citation=... />             {/* always */}

{outcome === "handoff" && handoff_summary && <HandoffPanel .../>}
<CaseDocumentPanel caseId={caseData.id} />                     {/* always */}
{action_steps?.length > 0 && <ActionSteps .../>}
{hospital_script && <CopyableTextBox title="Exact words to use at the desk" .../>}
{complaint_text && <CopyableTextBox title="Draft complaint, ready to review" helperText="..." .../>}
{evidence_prompt && <EvidenceForm .../>}
```

`CaseDocumentPanel` is the one panel here that's unconditional like the top two rather than field-gated like everything below it — every case gets a downloadable document regardless of outcome, so there's no backend field to check the presence of. It's placed after the handoff panel (so a family sees "here's your situation" fully before "here's the one thing to grab and go") and before the granular per-section content, on the reasoning that someone in a hurry should be able to find and use it without first reading every section on screen.

This ordering and gating is a direct, field-by-field mirror of what `backend/internal/response.Build` actually populates per outcome (see `backend/internal/response/README.md`'s table) — the frontend doesn't re-derive "should I show a complaint box for this outcome," it just checks whether `complaint_text` came back non-empty, because the backend already made that decision correctly and the frontend's job is to render what it was given, not second-guess it. **This is worth preserving if you touch this file**: adding a new frontend-side condition that tries to be smarter than "is this field present" (e.g. `outcome === "green" && complaint_text && ...`) would create a second place the green/amber/red/mixed/handoff logic has to be kept in sync, when one already exists and is already tested (`backend/internal/tiering`, `backend/internal/response`).

The two `CopyableTextBox` instances are the one place two uses of the same component are distinguished only by their props — `helperText` is what actually separates them (see `frontend/app/components/README.md` and that component's own test), not two different components.

## The three page states, and what's specific about each

- **Loading**: `role="status"` — a screen reader announces this without the user needing to hunt for a spinner. `loading` starts `true` and is only ever set `false` in `.finally()`, so there is no window where both loading text and case content could render simultaneously (this exact property — never showing both states at once — is one of the things `page.test.tsx`, described below, specifically checks).
- **Error**: `role="alert"`, and — this is the part worth understanding, not just noting — it includes its own PMJAY helpline (`tel:14555`) fallback, separate from `Header`'s own helpline link that's rendered above it on the same page. Two links to the same number on one page is intentional redundancy, not a mistake: a family whose case failed to load should be able to find the helpline number without having to scroll up past an error message to find `Header`'s copy. (This redundancy is exactly what made `page.test.tsx`'s error-state test need to scope its query with `within(alert)` rather than an unscoped `getByRole` — see that test's own comment for the specific ambiguity it caught.)
- **Loaded**: described above.

## `page.test.tsx`: the one genuine integration test in this frontend

Distinct from every other test in this codebase (`docs/TESTING.md` covers the full breakdown) — this is the only test that renders the *actual* `CasePage` component, composed exactly as production composes it, with only the network boundary mocked (`@/lib/api`'s `getCase`, `next/navigation`'s `useParams`). Every other frontend test exercises one component in isolation. This one exists specifically to catch the class of bug isolated tests structurally cannot: a prop passed under the wrong name, a conditional that hides a panel that should be showing, two components that individually work but were never actually wired together correctly. It covers:

- the loading → loaded transition never showing both states at once,
- the care-first message rendering for **every one of the five outcomes** — this page's own version of `backend/internal/response/builder_test.go`'s `TestBuild_CareFirstMessageIsAlwaysPresent_EveryOutcome`, checking the guarantee survives all the way to what a family's screen actually shows, not just what the backend's JSON says,
- `HandoffPanel` rendering for a handoff outcome and only a handoff outcome,
- the case document download link carrying the *correct* case ID through — `CaseDocumentPanel.test.tsx` can only confirm the component behaves correctly given some ID; this is the test that actually confirms `CasePage` passes it the right one,
- action steps / hospital script / complaint text each rendering only when the backend actually sent that specific field,
- the error state's helpline fallback.

**Why this isn't a real Playwright end-to-end test**: this sandbox's network allowlist doesn't reach a browser-binary CDN, so installing a real browser for Playwright wasn't achievable in the environment this was built in. Genuinely still missing relative to true e2e: real browser rendering (CSS layout, real click/focus/keyboard behavior), and testing against a live backend rather than a mocked API boundary. See `docs/TESTING.md` for the full, honest accounting of this gap — worth revisiting if this environment's constraints change, or from a developer's own machine.

## If you're extending this page

- **A new response field that needs its own panel**: add the conditional following the existing pattern (`{field && <Component .../>}`), add the component to `frontend/app/components/` with its own test, and extend `page.test.tsx` with a case that confirms it renders only when the field is present — matching how every existing conditional here is covered.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`. Keep it in sync with the code, in the same change.
