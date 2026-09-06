# syntax=docker/dockerfile:1

# =============================
# --- Global Build Arguments ---
# =============================
ARG GO_VERSION=1.25.4
ARG ALPINE_VERSION=3.22

# ==============================
# --- Dependencies Stage ---
# ==============================
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS dependencies

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod download && \
  go mod verify

# ==============================
# --- Builder Stage ---
# ==============================
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

ARG TARGETPLATFORM
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION=latest

WORKDIR /app

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static linking
# -ldflags options:
#   -w: omit DWARF symbol table
#   -s: omit symbol table and debug information
#   -X: set version variables for reproducibility
# -trimpath: remove all file system paths from the binary
# -a: force rebuilding of packages for reproducibility
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 \
  go build \
  -a \
  -trimpath \
  -ldflags="-w -s -X github.com/brpaz/sops-tui/cmd/main.Version=${VERSION} -X github.com/brpaz/sops-tui/cmd/main.BuildDate=${BUILD_DATE} -X github.com/brpaz/sops-tui/cmd/main.Commit=${VCS_REF}" \
  -o /build/app ./cmd/main.go

# ==============================
# --- Development Stage ---
# ==============================
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS development

WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go install github.com/go-delve/delve/cmd/dlv@latest

RUN addgroup -g ${GID} -S appgroup && \
  adduser -u ${UID} -S appuser -G appgroup

COPY . .

USER appuser

CMD ["go", "run", "./cmd/main.go"]

# ==============================
# --- Production Stage ---
# ==============================
FROM alpine:${ALPINE_VERSION} AS production

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION=latest

ARG GID=1000
ARG UID=1000

WORKDIR /app

RUN addgroup -g ${GID} -S appgroup && \
  adduser -u ${UID} -S appuser -G appgroup

RUN apk --no-cache add ca-certificates curl

COPY --from=builder /build/app /app

RUN chown appuser:appgroup /app && \
  chmod +x /app

USER appuser

LABEL org.opencontainers.image.title="" \
  org.opencontainers.image.version="${VERSION}" \
  org.opencontainers.image.created="${BUILD_DATE}" \
  org.opencontainers.image.revision="${VCS_REF}" \
  org.opencontainers.image.description="A k9s-style terminal UI for browsing, encrypting, and decrypting SOPS-protected secret files" \
  org.opencontainers.image.source="https://github.com/brpaz/sops-tui"

ENTRYPOINT ["/app"]
