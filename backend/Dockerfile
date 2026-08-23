# PMJAY Point-of-Denial Advocate — backend
#
# Multi-stage build. The backend has zero third-party Go dependencies
# (see ../ARCHITECTURE.md) and compiles to a single static binary with
# CGO_ENABLED=0, so the final stage can be `scratch` — no shell, no
# package manager, no libc, nothing an attacker could use even if they
# got code execution. The only reason this isn't literally FROM scratch
# is that the HTTPS call to api.anthropic.com needs a CA certificate
# bundle, which scratch doesn't have; distroless/static ships exactly
# that and nothing else.

FROM golang:1.22 AS builder

WORKDIR /src

# A separate `go.mod`-only COPY + `go mod download` layer would go here
# in a normal Go project, to cache dependency downloads across builds
# that only change source, not deps. Skipped entirely, on purpose:
# there is no go.sum and nothing to download — one COPY is all this
# zero-dependency module needs.
COPY . .

# CGO_ENABLED=0 for a fully static binary (no libc dependency at
# runtime, which is what makes the distroless final stage possible).
# -s -w strips debug symbols — smaller image, and there is no
# in-container debugging story for a binary this size anyway.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/server ./server

# FileStore's default DATA_FILE_PATH is ./data/cases.json (relative to
# WORKDIR, so /app/data/cases.json here) — declared as a volume mount
# point since distroless has no `mkdir` to create it at runtime if it's
# missing. DATA_FILE_PATH is what actually controls where the app
# writes, though, not this directive — if your platform mounts a
# volume somewhere else (docs/DEPLOYMENT.md's Fly.io example mounts at
# /data), just set DATA_FILE_PATH to match wherever that is, so data
# survives redeploys either way.
VOLUME ["/app/data"]

# distroless/static's "nonroot" variant already runs as uid 65532, not
# root — no USER directive needed, unlike a typical Debian/Alpine base.

EXPOSE 8080

ENTRYPOINT ["/app/server"]
