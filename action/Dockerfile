# Multi-stage build for the published octane-action image.
#
# Stage 1 builds a static linux binary for the target platform.
# Stage 2 ships a minimal alpine runtime image carrying only the binary,
# CA certificates, and bash (required by entrypoint.sh).

# ----- builder ---------------------------------------------------------

FROM golang:1.23-alpine@sha256:383395b794dffa5b53012a212365d40c8e37109a626ca30d6151c8348d380b5f AS builder

# BuildKit injects TARGETARCH and TARGETOS when --platform is used.
# Declare them here so the shell expansion below resolves correctly.
ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

# Bring in source.
COPY . .

# Static binary for the target platform.
RUN apk add --no-cache git \
 && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags='-s -w' -o bin/octane ./cmd/octane

# ----- runtime ---------------------------------------------------------

# Alpine provides a minimal shell environment needed by entrypoint.sh.
# SHA pinned for reproducible builds. Refresh with:
#   docker pull alpine:3.21 && docker inspect alpine:3.21 --format '{{index .RepoDigests 0}}'
FROM alpine:3.21@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d

RUN apk add --no-cache bash ca-certificates

LABEL org.opencontainers.image.source="https://github.com/octane-project/octane"
LABEL org.opencontainers.image.title="octane-action"
LABEL org.opencontainers.image.description="OCPP Conformance Testing & Network Evaluation"
LABEL org.opencontainers.image.licenses="Apache-2.0"

COPY --from=builder /src/bin/octane /usr/local/bin/octane

# Entrypoint script translates INPUT_* env vars to CLI flags.
COPY action/entrypoint.sh /entrypoint.sh

# Action workspace will be mounted at /github/workspace by the runner.
WORKDIR /github/workspace

ENTRYPOINT ["/entrypoint.sh"]
CMD []
