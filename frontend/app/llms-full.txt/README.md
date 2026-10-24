# `frontend/app/llms-full.txt`

The longer companion to `/llms.txt`: system context for an agent that is
about to reason about this project rather than just link to it. Also a route
handler, also built from `lib/site.ts`, so its URLs match the deployment.

The difference is scope, not tone. `/llms.txt` is an index; this file
explains the decision layer (Green/Amber/Red/Mixed/Handoff), what the tool
cannot do, the data provenance, the retention window, and the fact that a
missing package means "not in this index", never "not covered".

## Keep it honest

This file is read by systems that will summarise it to a person in distress.
Anything added here should survive the question "is this true on a fresh
deployment with no API key?" — which is why the availability caveat and the
helpline numbers are stated plainly rather than in a footer.
