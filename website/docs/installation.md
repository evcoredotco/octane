---
sidebar_position: 3
---

# Installation

OCTANE is distributed as a Go CLI and as a GitHub Action wrapper around
the same engine.

During local development, build the CLI from the repository root:

```bash
make build
./bin/octane --help
```

Release packaging is handled by GoReleaser. For distribution details,
see `docs/distribution.md`.

