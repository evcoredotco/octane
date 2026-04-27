# Distribution

OCTANE is published as a multi-architecture container image to the GitHub
Container Registry (GHCR).

## Image coordinates

```
ghcr.io/octane-project/octane:<version>
ghcr.io/octane-project/octane:latest
```

Use a versioned tag (e.g. `v0.1.0`) in any automated pipeline. The `latest`
tag tracks the most recent stable release and is suitable only for ad-hoc
exploration on a developer workstation.

## Supported platforms

Each image manifest covers two target platforms:

- `linux/amd64` — standard x86-64 runners (GitHub-hosted, most cloud VMs)
- `linux/arm64` — Apple Silicon, AWS Graviton, and equivalent ARM64 hosts

Docker and compatible runtimes select the correct variant automatically via
the manifest list.

## Image contents

The runtime layer is Alpine-based and contains only what OCTANE needs at
execution time:

- the `octane` binary compiled for the target platform
- `bash` (required by the entrypoint)
- CA certificates (`ca-certificates` package) so TLS connections to a CSMS
  succeed out of the box

The image entrypoint is `entrypoint.sh`, which maps well-known environment
variables (e.g. `OCTANE_CACHE_DIR`, `OCTANE_TARGET`) to their corresponding
CLI flags before invoking the `octane` binary. Consumers may also override
the entrypoint and invoke the binary directly.

## Using the image in GitLab CI

Pin to a semver tag rather than `latest` so that a registry push cannot
silently change the binary your pipeline runs:

```yaml
variables:
  OCTANE_IMAGE: "ghcr.io/octane-project/octane:v0.1.0"
  OCTANE_CACHE_DIR: "${CI_PROJECT_DIR}/.octane-cache"

conformance:
  stage: conformance
  image: ${OCTANE_IMAGE}
  cache:
    key: octane-${CI_COMMIT_REF_SLUG}
    paths:
      - .octane-cache/
  script:
    - octane run scenarios/ --fail-on major
```

`OCTANE_CACHE_DIR` is read by the entrypoint to set the cache root; the
GitLab cache block then persists that directory between pipeline runs.

## Using the image directly from the CLI

To bypass the entrypoint and invoke the binary directly:

```
docker run --rm \
  --entrypoint /usr/local/bin/octane \
  ghcr.io/octane-project/octane:v0.1.0 \
  run scenarios/
```

Mount a local directory with `-v $(pwd)/scenarios:/scenarios` to pass
scenario files into the container.

## Release cadence

Every push of a `vMAJOR.MINOR.PATCH` tag triggers the `release.yml`
workflow, which:

1. runs `goreleaser` to produce platform-specific binaries and a GitHub
   Release;
2. calls `docker/build-push-action` in multi-arch mode to build and push the
   image manifest for both `linux/amd64` and `linux/arm64`;
3. tags the image with the full semver (`vX.Y.Z`), the minor series
   (`vX.Y`), and updates `latest`.

Pre-release tags (`v0.1.0-rc.1`) publish a versioned image but do not move
the `latest` tag.

## Pulling a pinned digest

For fully reproducible CI where even a tag re-push must not change the image,
pull by digest:

```
docker pull ghcr.io/octane-project/octane@sha256:<digest>
```

Digests are stable across architectures for a given manifest list. Record the
digest in your pipeline configuration alongside the tag for human readability.

## Caching note

OCTANE caches parsed scenario artefacts under `OCTANE_CACHE_DIR` to speed up
repeated runs.

- **GitHub Actions**: pair the image with `actions/cache@v4`, using a cache
  key that includes the scenario file hashes and the OCTANE version. Point
  the `path` input at the value of `OCTANE_CACHE_DIR`.
- **GitLab CI**: use GitLab's built-in `cache:` block (see example above)
  with `paths` set to `.octane-cache/` (or the value of `OCTANE_CACHE_DIR`).

A partial cache hit on a key prefix is better than no cache; prefix the key
with a stable component (e.g. the minor version) and append a content hash
so that scenario edits produce a cold start rather than stale entries.
