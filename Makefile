# OCTANE — Makefile
#
# Contributors and CI use these targets by name. Keep the surface stable;
# any change requires updating AGENTS.md and CONTRIBUTING.md.

GO            ?= go
GOLANGCI_LINT ?= golangci-lint
# Keep in step with the versions pinned in .github/workflows/ci.yml.
GOLANGCI_LINT_VERSION ?= v2.15.0
STATICCHECK_VERSION   ?= 2026.2.2
DOCS_HOST     ?= 127.0.0.1
DOCS_PORT     ?= 3000

PKG    := ./...
BIN    := ./bin/octane
MODULE := github.com/evcoredotco/octane

.DEFAULT_GOAL := help

.PHONY: help
help: ## Display this help message, listing all available targets and their descriptions.
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1;34m${DOCKER_NAMESPACE}\033[0m\tCopyright (c) ${DATE} Alexis Sanchez\n \n\033[1;32mUsage:\033[0m\n  make \033[1;34m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[1;34m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1;33m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# ----------------------------------------------------------------------
# Format / lint / test
# ----------------------------------------------------------------------

##@ Format / lint / test
.PHONY: format
format: ## Format code with the formatters configured in golangci.yml (gci, gofumpt, golines).
	$(GOLANGCI_LINT) fmt --config golangci.yml

.PHONY: lint
lint: ## Run golangci-lint, go vet, staticcheck.
	$(GOLANGCI_LINT) run --config golangci.yml --timeout 5m
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

##@ Build
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

##@ Action image
.PHONY: action-image
action-image: ## Build the published octane-action Docker image locally.
	docker build -t octane-action:dev -f action/Dockerfile .

# ----------------------------------------------------------------------
# Spec helpers
# ----------------------------------------------------------------------

##@ Spec helpers
.PHONY: install-tools
install-tools: ## Install development tooling.
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	$(GO) install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
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

##@ Documentation: man pages, shell completions, web site
.PHONY: man
man: build ## Generate man pages (Section 1 from cobra, 5 and 7 from scdoc).
	bash ./scripts/gen-manpages.sh

.PHONY: completions
completions: build ## Generate shell completion scripts (bash, zsh).
	bash ./scripts/gen-completions.sh

.PHONY: docs-html
docs-html: ## Build the Docusaurus website into website/build.
	cd website && pnpm install && pnpm run build

.PHONY: docs-serve
docs-serve: ## Run the Docusaurus dev server on http://127.0.0.1:3000.
	cd website && pnpm install && pnpm run start --host $(DOCS_HOST) --port $(DOCS_PORT) --no-open

# ----------------------------------------------------------------------
# Packaging
# ----------------------------------------------------------------------

##@ Packaging
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
