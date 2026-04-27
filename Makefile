# OCTANE — Makefile
#
# All agents reference these targets by name. Keep the surface stable;
# any change requires updating .claude/agents/*.md and AGENTS.md.
#
# STATUS: This Makefile's Go-related targets (format, lint, test,
# build, package) are intentionally preserved as the contract for
# when implementation begins. They will fail cleanly today because
# pkg/ is empty and go.mod does not exist; this is expected. The
# specs (specs/001-bootstrap-engine, specs/002-story-framework) and
# ADRs (notably 0007, 0015, 0016, 0017) are the design source of
# truth for the implementation that will populate the surface.
#
# Until Go code lands:
#   make spec-check    works (validates the specs)
#   make man           works (generates man pages from scdoc)
#   make completions   works (generates shell completions; depends on the binary)
#   make docs-html     works (Docusaurus build)
#   make format/lint/test/build  fail cleanly until pkg/ exists

GO            ?= go
GOLANGCI_LINT ?= golangci-lint
GOFUMPT       ?= gofumpt
GOLINES       ?= golines
GCI           ?= gci

PKG    := ./...
BIN    := ./bin/octane
MODULE := github.com/evcoreco/octane

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} \
		/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ----------------------------------------------------------------------
# Format / lint / test
# ----------------------------------------------------------------------

.PHONY: format
format: ## Run gofmt, gofumpt, golines, gci.
	$(GO) fmt $(PKG)
	$(GOFUMPT) -l -w .
	$(GOLINES) -w -m 80 .
	$(GCI) write \
		--skip-generated \
		--section standard \
		--section "prefix($(MODULE))" \
		--section default \
		.

.PHONY: lint
lint: ## Run golangci-lint, go vet, staticcheck.
	$(GOLANGCI_LINT) run --timeout 5m
	$(GO) vet $(PKG)

.PHONY: test
test: ## Run unit tests with -race.
	$(GO) test -race -timeout 120s $(PKG)

.PHONY: test-reference
test-reference: ## Run the full suite against the pinned CitrineOS.
	@echo "Spinning CitrineOS at pinned commit..."
	cd test/reference && docker compose up -d --wait
	$(GO) test -race -tags=reference -timeout 600s ./test/integration/...
	cd test/reference && docker compose down -v

.PHONY: fuzz
fuzz: ## Run fuzz targets for 30 seconds each.
	$(GO) test -fuzz=. -fuzztime=30s $(PKG) || true

# ----------------------------------------------------------------------
# Build
# ----------------------------------------------------------------------

.PHONY: build
build: ## Build the octane CLI.
	mkdir -p bin
	$(GO) build -o $(BIN) ./cmd/octane

.PHONY: build-static
build-static: ## Build a static linux/amd64 binary for the Action image.
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		$(GO) build -trimpath -ldflags='-s -w' -o $(BIN) ./cmd/octane

# ----------------------------------------------------------------------
# Action image
# ----------------------------------------------------------------------

.PHONY: action-image
action-image: ## Build the published octane-action Docker image locally.
	docker build -t octane-action:dev -f action/Dockerfile .

# ----------------------------------------------------------------------
# Spec helpers
# ----------------------------------------------------------------------

.PHONY: spec-check
spec-check: ## Validate every spec under specs/.
	@for d in specs/*/; do \
		.specify/scripts/bash/check-spec.sh "$$d" || exit 1; \
	done

.PHONY: install-tools
install-tools: ## Install development tooling.
	$(GO) install mvdan.cc/gofumpt@latest
	$(GO) install github.com/segmentio/golines@latest
	$(GO) install github.com/daixiang0/gci@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/goreleaser/goreleaser/v2@latest

.PHONY: dev-setup
dev-setup: install-tools ## One-time dev environment setup (runs install-tools + configures GOPRIVATE).
	$(GO) env -w GOPRIVATE=github.com/evcoreco/*
	@echo "GOPRIVATE set to github.com/evcoreco/* (required for ocpp16types — ADR 0020)"
	@echo "Run 'go env GOPRIVATE' to verify."

# Run this once github.com/evcoreco/ocpp16types publishes its first tagged release.
# Replace vX.Y.Z with the actual release tag.
.PHONY: pin-ocpp16types
pin-ocpp16types: ## Pin github.com/evcoreco/ocpp16types to its latest release (ADR 0020).
	@echo "Checking GOPRIVATE..."
	@go env GOPRIVATE | grep -q "evcoreco" || (echo "ERROR: run 'make dev-setup' first to configure GOPRIVATE" && exit 1)
	$(GO) get github.com/evcoreco/ocpp16types@latest
	$(GO) mod tidy
	@echo "ocpp16types pinned. Commit go.mod and go.sum."

# ----------------------------------------------------------------------
# Documentation: man pages, shell completions, web site
# ----------------------------------------------------------------------

.PHONY: man
man: build ## Generate man pages (Section 1 from cobra, 5 and 7 from scdoc).
	bash ./scripts/gen-manpages.sh

.PHONY: completions
completions: build ## Generate shell completion scripts (bash, zsh).
	bash ./scripts/gen-completions.sh

.PHONY: docs-html
docs-html: ## Build the Docusaurus website into website/build.
	cd website && npm ci && npm run build

.PHONY: docs-serve
docs-serve: ## Run the Docusaurus dev server on http://localhost:3000.
	cd website && npm ci && npm run start

# ----------------------------------------------------------------------
# Packaging
# ----------------------------------------------------------------------

.PHONY: package
package: build man completions ## Snapshot release: binaries + .deb + .rpm + Homebrew formula.
	goreleaser release --snapshot --clean

.PHONY: package-deb
package-deb: build man completions ## Build .deb only via nfpm.
	nfpm package --packager deb --config packaging/nfpm.yaml --target dist/

.PHONY: package-rpm
package-rpm: build man completions ## Build .rpm only via nfpm.
	nfpm package --packager rpm --config packaging/nfpm.yaml --target dist/

.PHONY: clean
clean: ## Remove build artefacts.
	rm -rf bin/ build/ dist/ website/build/ website/.docusaurus/
