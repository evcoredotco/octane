---
sidebar_position: 3
---

# Installation

OCTANE is **pre-alpha**. No packages are published yet — the supported
installation method today is **building from source** with the standard
Go toolchain. The release machinery for binaries and OS packages is in
place and will be enabled once the first version is tagged.

## Requirements

- **Go 1.26 or newer** (`go version`)
- A POSIX-like shell for the examples (Linux, macOS, or WSL on Windows)
- Optionally **Docker**, to run the reference CSMS (CitrineOS)

## Build from source

```bash
git clone https://github.com/evcoreco/octane
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

## Planned distribution channels

These channels are wired into the release pipeline (`goreleaser` plus
`packaging/nfpm.yaml`) but are **not yet published**. They will become
available with the first tagged release.

| Channel | Planned command |
|---|---|
| Debian / Ubuntu | `apt install octane` |
| Fedora / RHEL | `dnf install octane` |
| macOS (Homebrew) | `brew install evcoreco/octane/octane` |
| Windows (Scoop) | `scoop install octane` |
| Docker | `docker pull ghcr.io/evcoreco/octane` |
| Direct download | static binaries, signed with cosign, SBOM-attested |

:::note
Until then, build from source as shown above. The
[GitHub Action](./operations/ci-integration.md) references the container
image `ghcr.io/evcoreco/octane:v0`, which is published by the release
workflow on each version tag.
:::

## Next steps

- **[Getting started](./getting-started.md)** — your first conformance run.
- **[CLI reference](./reference/cli.md)** — every subcommand and flag.
- **[Configuration](./reference/config-schema.md)** — `octane.yml` and
  environment variables.
