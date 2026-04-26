# Multi-stage build for the published octane-action image.
#
# Stage 1 builds a static linux binary using the same Makefile target as
# the local dev workflow. Stage 2 ships a minimal runtime image
# (distroless) carrying only the binary and CA certificates.

# ----- builder ---------------------------------------------------------

FROM golang:1.23-alpine AS builder

WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

# Bring in source.
COPY . .

# Static binary; the make target enforces the same flags used in CI.
RUN apk add --no-cache make git \
 && make build-static

# ----- runtime ---------------------------------------------------------

FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.source="https://github.com/octane-project/octane"
LABEL org.opencontainers.image.title="octane-action"
LABEL org.opencontainers.image.description="OCPP Conformance Testing & Network Evaluation"
LABEL org.opencontainers.image.licenses="Apache-2.0"

COPY --from=builder /src/bin/octane /usr/local/bin/octane

# Action workspace will be mounted at /github/workspace by the runner.
WORKDIR /github/workspace

ENTRYPOINT ["/usr/local/bin/octane"]
