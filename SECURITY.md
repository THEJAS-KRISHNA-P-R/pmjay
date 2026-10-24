# Security policy

## Reporting a vulnerability

Please do **not** open a public issue for a security problem, and never include
a case link, a patient's name, or medical details in any public issue.

Report privately through GitHub's **Security → Report a vulnerability** flow on
this repository, or contact the maintainer listed on the repository profile.
Include the affected route or file, a minimal reproduction, and what an
attacker could gain. Expect an acknowledgement within a few days; this is a
volunteer public-interest project, not a staffed service.

## Scope

In scope:

- the Go API (`backend/`) — auth bypass, injection, path traversal, request
  smuggling, rate-limit bypass, or a way to read another family's case;
- the Next.js app (`frontend/`) — XSS, open redirect, secret leakage, or a
  route that exposes a case to a crawler;
- the agent skill (`skills/`) — a leaked rate on an unverified record, a
  bundled secret or absolute local path, or packaging of frontend code.

Out of scope:

- missing security headers on a third-party host you control;
- rate limiting being disabled because `TRUST_PROXY_HEADERS` was set without a
  proxy in front (that is a deployment error, documented in
  `docs/DEPLOYMENT.md`);
- the fact that a case URL is a capability — this is deliberate and disclosed
  on `/privacy`. Treat a case link like a private document.

## Design notes worth knowing before testing

- No accounts, no passwords, no Aadhaar numbers. A case is reachable by its
  random UUID alone; the ID is the access key.
- Case IDs are redacted from server logs.
- `X-Forwarded-For` is only trusted when `TRUST_PROXY_HEADERS=true`.
- `/api/v1/health` is intentionally outside the rate limiters.
- Unverified packages cannot expose a rate through `GET /api/v1/packages`;
  `packageToSearchResult` strips it structurally.

## Supported versions

The `main` branch is the supported version. There are no long-lived release
branches.
