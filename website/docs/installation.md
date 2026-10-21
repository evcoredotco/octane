---
sidebar_position: 3
---

# Installation

OCTANE is **pre-alpha**. Releases are published on
[GitHub Releases](https://github.com/evcoredotco/octane/releases) from
`v0.1.0` onwards, as binaries, `.deb` and `.rpm` packages, and a container
image. You can also install with `go install` or build from source.

## Install a release

```bash
# Go toolchain (Go 1.26 or newer)
go install github.com/evcoredotco/octane/cmd/octane@latest

# Container image (linux/amd64 and linux/arm64)
docker pull ghcr.io/evcoredotco/octane:v0
docker run --rm ghcr.io/evcoredotco/octane:v0 octane --version
```

For Debian/Ubuntu and Fedora/RHEL, download the `.deb` or `.rpm` file from
the release page and install it with `sudo apt install ./octane_*.deb` or
`sudo dnf install ./octane-*.rpm`. For other platforms, download the
archive for your operating system and architecture; it contains the binary,
man pages, and shell completions.

Verify a download against the signed checksums:

```bash
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp '^https://github.com/evcoredotco/octane/' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  checksums.txt
sha256sum --ignore-missing -c checksums.txt
```

## Requirements

- **Go 1.26 or newer** (`go version`)
- A POSIX-like shell for the examples (Linux, macOS, or WSL on Windows)
- Optionally **Docker**, to run the reference CSMS (CitrineOS)

## Build from source

```bash
git clone https://github.com/evcoredotco/octane
cd octane
go build ./cmd/octane          # produces ./octane in the repo root
```

Or via the project Makefile, which places the binary under `bin/`:

```bash
make build                     # produces ./bin/octane
```

Verify the build:

```bash
./octane --help
```

### Run without building

For one-off runs you can skip the binary entirely:

```bash
go run ./cmd/octane run --csms-endpoint ws://localhost:9210
```

## Shell completion

OCTANE generates completion scripts for bash, zsh, fish, and PowerShell.

```bash
# Load completion in the current session:
source <(octane completion bash)
source <(octane completion zsh)
octane completion fish | source
octane completion powershell | Out-String | Invoke-Expression
```

To install it permanently, write the script to your shell's completion
directory — for example:

```bash
# zsh
octane completion zsh > "${fpath[1]}/_octane"

# bash (Linux)
octane completion bash | sudo tee /etc/bash_completion.d/octane > /dev/null
```

Completion is both static (subcommands and flags) and dynamic (story file
paths and registered keywords).

## Man pages

OCTANE ships Unix-style manual pages generated from the command tree and
hand-written for the file formats:

```text
man 1 octane-run      # the run subcommand
man 5 octane.yml      # configuration file reference
man 5 octane.story    # story DSL reference
man 7 octane          # concepts overview
```

## Not yet available

- **Homebrew:** the release can update a tap at
  `evcoredotco/homebrew-octane`, but the tap does not exist yet.
- **APT and DNF repositories:** packages are attached to each release; there
  is no package repository to add.
- **Scoop:** not configured.

## Next steps

- **[Getting started](./getting-started.md)** — your first conformance run.
- **[CLI reference](./reference/cli.md)** — every subcommand and flag.
- **[Configuration](./reference/config-schema.md)** — `octane.yml` and
  environment variables.
