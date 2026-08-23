# Deployment

Two real paths are documented here: **colocated** (the backend rides along on infrastructure already paid for elsewhere, frontend on Vercel's free tier — the cheapest realistic option, and the one `ARCHITECTURE.md`'s cost table assumes) and **fully standalone** (no existing infrastructure assumed, everything on free tiers). Both get you to $0 marginal cost outside of LLM inference. Concrete commands throughout, not just prose.

## Topology, either way

```
Browser
  │
  ├──────────────► Next.js frontend (Vercel, its own domain)
  │                        │
  │                        │  fetch(NEXT_PUBLIC_API_BASE_URL + "/v1/...")
  │                        ▼
  └──────────────► Caddy (:443, TLS auto-issued) ──► Go binary (127.0.0.1:8080)
                    on the backend host
```

The frontend and backend are **different origins** in both paths below (Vercel's domain vs. wherever the backend is reachable), so `NEXT_PUBLIC_API_BASE_URL` must be set at frontend build time to the backend's full public URL, and the backend's `ALLOWED_ORIGINS` must list the frontend's Vercel URL. `lib/api.ts` defaults to same-origin `/api` only if you instead choose to serve both under one domain through Caddy — see the aside at the bottom for that variant.

**One operational requirement either way, worth being explicit about**: the Go process should never be reachable directly from the public internet — only Caddy should be. `internal/api/middleware.go`'s `clientIP()` trusts the *last* `X-Forwarded-For` value for per-IP rate limiting — the entry Caddy itself appends, not whatever a client sent — which is what makes it safe against a caller simply forging that header when Caddy is the only path in. That safety is contingent on the topology, not something the code alone can guarantee: if the Go port is ever also reachable directly, a caller hitting it straight can set a single `X-Forwarded-For` value with no comma, which becomes indistinguishable from a proxy-appended one. Bind the Go process to `127.0.0.1` at the OS/firewall level (or a cloud security group rule restricting the port to localhost) so Caddy is the only thing anyone outside the host can reach on 80/443.

---

## Path 1: Colocated (recommended if you already run a VM)

This assumes what `ARCHITECTURE.md`'s cost table assumes: an existing always-on instance (a t4g.micro, ARM64/Graviton, already paid for as part of a separate project) with headroom for one more 5.7 MB static binary. Marginal cost: **$0**.

### 1. Cross-compile for the target architecture

Already verified working during the build (see `../HANDOVER.md`):

```bash
cd backend
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o pmjay-advocate-server ./cmd/server
# 5.7 MB, stripped, no C toolchain needed even when cross-compiling from a different arch
```

(Swap `GOARCH=arm64` for `amd64` if the target instance is x86_64 instead of Graviton.)

### 2. Ship the binary and the seed data directory

```bash
scp pmjay-advocate-server your-user@your-host:/opt/pmjay-advocate/
ssh your-user@your-host 'mkdir -p /opt/pmjay-advocate/data'
```

The HBP dataset is `go:embed`-ed into the binary itself (see `internal/hbp`), so nothing else needs to ship — `/opt/pmjay-advocate/data` is only for `FileStore`'s `cases.json`, created automatically on first write.

### 3. Run it as a systemd service

`/etc/systemd/system/pmjay-advocate.service` on the host:

```ini
[Unit]
Description=PMJAY Point-of-Denial Advocate backend
After=network.target

[Service]
Type=simple
User=pmjay
WorkingDirectory=/opt/pmjay-advocate
EnvironmentFile=/opt/pmjay-advocate/.env
ExecStart=/opt/pmjay-advocate/pmjay-advocate-server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

`/opt/pmjay-advocate/.env` holds the real environment values — copy `backend/.env.example`, fill in the API key for whichever `LLM_PROVIDER` you're using (`ANTHROPIC_API_KEY`, `GROQ_API_KEY`, or `GEMINI_API_KEY` — see `internal/config/README.md` for the full provider-selection story), and set `ALLOWED_ORIGINS` to your actual Vercel URL. Then:

```bash
sudo useradd -r -s /usr/sbin/nologin pmjay   # once
sudo systemctl daemon-reload
sudo systemctl enable --now pmjay-advocate
sudo systemctl status pmjay-advocate         # confirm it's up
journalctl -u pmjay-advocate -f              # structured JSON logs, tail them live
```

`PORT` in `.env` should stay `8080` (or whatever's free) and bind only to loopback in practice — Go's `http.Server` binds all interfaces on `:PORT` by default, so the host's firewall (or cloud security group) is what actually needs to restrict external access to that port, per the note above.

### 4. Caddy: reverse proxy `/api` to the Go binary

If a Caddy instance is already running on the host for another project (e.g. `yourfee.in`), add a new site block for this project's own domain/subdomain rather than trying to path-merge into the existing one — cleanest separation, and each project's `Caddyfile` block stays independently editable:

```caddyfile
api.pmjay-advocate.example.com {
	reverse_proxy 127.0.0.1:8080

	header {
		X-Content-Type-Options nosniff
		X-Frame-Options DENY
	}
}
```

```bash
sudo systemctl reload caddy   # picks up the new block, auto-issues TLS via Let's Encrypt
curl https://api.pmjay-advocate.example.com/api/v1/health   # should return the health JSON
```

### 5. Frontend: Vercel free tier

```bash
cd frontend
npm install -g vercel   # if not already
vercel link
vercel env add NEXT_PUBLIC_API_BASE_URL production
# paste: https://api.pmjay-advocate.example.com/api
vercel --prod
```

Then set `ALLOWED_ORIGINS` back on the backend (in `/opt/pmjay-advocate/.env`) to the exact Vercel production URL Vercel gives you, and `systemctl restart pmjay-advocate` to pick it up. CORS will reject the frontend's requests until this matches exactly — `corsMiddleware` does an exact string match against the `Origin` header, no wildcards, no subdomain matching.

### Redeploying after a change

```bash
cd backend
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o pmjay-advocate-server ./cmd/server
scp pmjay-advocate-server your-user@your-host:/opt/pmjay-advocate/
ssh your-user@your-host 'sudo systemctl restart pmjay-advocate'
```

`cases.json` on disk is untouched by a binary swap — `FileStore` just reloads it on the next start.

---

## Path 2: Fully standalone (no existing infrastructure assumed)

Same $0 target, using free tiers end to end. Backend on Fly.io (Oracle Cloud's Always Free ARM tier is an equally valid swap if preferred — same binary, same Dockerfile).

### Backend on Fly.io

`backend/Dockerfile` (see that file) already builds the same zero-dependency static binary in a distroless final stage, so Fly.io's Docker-based deploy works with no extra setup:

```bash
cd backend
fly launch --no-deploy   # generates fly.toml, pick a region, decline a Postgres add-on (not needed)
fly secrets set ANTHROPIC_API_KEY=sk-ant-...
# Using Groq or Gemini instead? Set LLM_PROVIDER and the matching key —
# see internal/config/README.md — instead of (or alongside) the above.
# Example: fly secrets set LLM_PROVIDER=groq GROQ_API_KEY=gsk_...
fly secrets set ALLOWED_ORIGINS=https://your-frontend.vercel.app
fly volumes create pmjay_data --size 1   # 1GB is wildly generous for a single JSON file
```

In the generated `fly.toml`, mount the volume so `cases.json` survives redeploys, and point `DATA_FILE_PATH` at it:

```toml
[mounts]
  source = "pmjay_data"
  destination = "/data"

[env]
  DATA_FILE_PATH = "/data/cases.json"
  PORT = "8080"
```

```bash
fly deploy
fly status   # confirm it's healthy; note the public https://<app>.fly.dev URL
```

Fly's free allowance covers this comfortably — the workload is one small request at a time, not sustained traffic.

### Frontend: identical Vercel steps as Path 1

Same as above, just point `NEXT_PUBLIC_API_BASE_URL` at the `fly.dev` URL instead of a custom domain.

---

## Aside: same-origin deployment (skip CORS entirely)

If cross-origin CORS configuration is something you'd rather avoid altogether, both frontend and backend can sit behind the *same* Caddy instance on one domain, with `lib/api.ts`'s default `/api` base URL working with zero frontend env configuration:

```caddyfile
pmjay-advocate.example.com {
	handle /api/* {
		reverse_proxy 127.0.0.1:8080
	}
	handle {
		reverse_proxy 127.0.0.1:3000   # `next start`, also run as a systemd service
	}
}
```

This trades away Vercel's free CDN/edge network for one less moving part (no CORS config, no `NEXT_PUBLIC_API_BASE_URL` to keep in sync) — a reasonable trade for a hackathon-stage deploy on a host you already control, less so if you want Vercel's edge caching for the frontend specifically.

---

## Backing up `cases.json`

There is no database here — `FileStore` is one JSON file (see `internal/store/filestore.go`), which is the right tradeoff for this system's actual scale (see `../ARCHITECTURE.md`, "reducing bills"), but it means the *entire* record of every family's case lives in one file with no replication of its own. Losing it isn't losing generic data — it's losing the one place a family's evidence log and draft complaint were being kept for them. This needs an actual answer, not just an assumption that the VM disk is fine.

**Path 1 (colocated):** the file is `/opt/pmjay-advocate/data/cases.json`. A daily cron job covers both failure modes that matter — a bad deploy or bug corrupting the file (local rotating snapshots), and losing the disk entirely (an off-box copy):

```bash
# /opt/pmjay-advocate/backup-cases.sh — run daily via cron
#!/usr/bin/env bash
set -euo pipefail
SRC=/opt/pmjay-advocate/data/cases.json
DEST_DIR=/opt/pmjay-advocate/backups
mkdir -p "$DEST_DIR"
[ -f "$SRC" ] || exit 0   # nothing written yet — fine, not an error
cp "$SRC" "$DEST_DIR/cases-$(date +%F).json"
find "$DEST_DIR" -name 'cases-*.json' -mtime +30 -delete   # keep 30 days locally

# Off-box copy — this is the line that actually protects against losing
# the VM/disk, not just the local rotation above. Any destination works;
# scp to a second cheap host is the zero-new-infrastructure option:
scp -q "$SRC" your-user@your-second-host:/opt/pmjay-advocate/backups/cases-latest.json
```

```bash
chmod +x /opt/pmjay-advocate/backup-cases.sh
sudo crontab -u pmjay -e
# add: 0 3 * * *  /opt/pmjay-advocate/backup-cases.sh
```

Restoring is just copying a snapshot back to `data/cases.json` and restarting the service — `FileStore` reads the whole file on startup, no migration step.

**Path 2 (Fly.io):** the volume Fly mounts for `DATA_FILE_PATH` already gets Fly's own automatic daily volume snapshots — nothing to set up. Confirm they're on and see how far back they go: `fly volumes list`, then `fly volumes snapshots list <volume-id>`. Restoring is `fly volumes create --snapshot-id <snapshot-id>` (see Fly's own volume-snapshot docs for the exact current flags — this is exactly the kind of platform-specific detail worth checking against Fly's docs at deploy time rather than trusting a snapshot of it here).

## Environment variables

See `backend/.env.example` for the full list with defaults. The settings that must be changed from their defaults for any real deployment: the API key for whichever `LLM_PROVIDER` is active (`ANTHROPIC_API_KEY` by default — empty by default, the server starts without it but every intake request fails clearly until it's set; `GROQ_API_KEY`/`GEMINI_API_KEY` if `LLM_PROVIDER` is set to `groq`/`gemini` instead) and `ALLOWED_ORIGINS` (defaults to `http://localhost:3000`, i.e. local dev only).

## Verifying a deployment

```bash
curl https://<backend-url>/api/v1/health
# {"status":"ok","packages_loaded":"315","exclusions_loaded":"4"}
```

A `200` with those three keys means: the binary is running, the embedded dataset loaded, and the process is reachable through whatever proxy sits in front of it. It does **not** confirm the active provider's API key (`ANTHROPIC_API_KEY`, `GROQ_API_KEY`, or `GEMINI_API_KEY`, whichever `LLM_PROVIDER` selects) is valid — that's only exercised by an actual `POST /api/v1/cases` call, which costs a real (tiny, or on Groq/Gemini's free tier, zero) API charge, so it's not something to hit repeatedly just to check config.
