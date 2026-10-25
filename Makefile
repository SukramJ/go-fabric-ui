# go-fabric-ui — make targets. `make help` lists them.
GO  ?= go
NPM ?= npm

UI_DIR   := ui
DIST_DIR := fabricui/dist
# dist-check and notices-check build here, never into the committed bundle.
CHECK_DIR := $(or $(TMPDIR),/tmp)/go-fabric-ui-check

FUZZTIME ?= 10000x

GOFUMPT_VERSION       ?= v0.12.0
GOLANGCI_LINT_VERSION ?= v2.14.0
GOVULNCHECK_VERSION   ?= v1.8.0
GOLICENSES_VERSION    ?= v1.6.0
GITLEAKS_VERSION      ?= v8.30.1

# The one release version: internal/version, ui/package.json and the
# changelog's newest section must agree (version-check).
VERSION_GO   = $(shell sed -n 's/^const Version = "\(.*\)"$$/\1/p' internal/version/version.go)
VERSION_NPM  = $(shell node -p 'require("./$(UI_DIR)/package.json").version' 2>/dev/null)
VERSION_LOG  = $(shell sed -n 's/^## \[\([0-9][^]]*\)\].*/\1/p' changelog.md | head -1)

.DEFAULT_GOAL := help

help: ## show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

setup: ## install the pinned gofumpt, golangci-lint, govulncheck and go-licenses
	$(GO) install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	$(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	$(GO) install github.com/google/go-licenses@$(GOLICENSES_VERSION)

build: ## go build ./...
	CGO_ENABLED=0 $(GO) build ./...

vet: ## go vet ./...
	$(GO) vet ./...

test: ## go test -count=1 ./... (handlers, fixtures, import guard, end-to-end on fabrictest)
	CGO_ENABLED=0 $(GO) test -count=1 -timeout 300s ./...

race: ## go test -race (CGO_ENABLED=1 for this run only)
	CGO_ENABLED=1 $(GO) test -race -count=1 -timeout 600s ./...

e2e: ## the end-to-end test alone: a fabrictest site commissioned through the handler
	CGO_ENABLED=0 $(GO) test -count=1 -run 'TestEndToEnd' -v ./fabricui/

fmt: ## gofumpt -w .
	gofumpt -w .

fmt-check: ## fail when a file is not gofumpt-clean
	@diff=$$(gofumpt -l .); if [ -n "$$diff" ]; then echo "gofumpt would rewrite:"; echo "$$diff"; exit 1; fi

lint: fmt-check ## gofumpt check + golangci-lint + go mod tidy -diff + go mod verify
	golangci-lint config verify
	golangci-lint run --timeout=5m ./...
	$(GO) mod tidy -diff
	$(GO) mod verify

tidy-check: ## verify go.mod/go.sum are tidy and the module checksums verify
	$(GO) mod tidy -diff
	$(GO) mod verify

cover: ## coverage profile of every package
	CGO_ENABLED=0 $(GO) test -count=1 -timeout 300s -coverpkg=./... -coverprofile=coverage.out ./...

cover-check: cover ## fail when a package is below its floor (script/coverfloor)
	$(GO) run ./script/coverfloor coverage.out

vuln: ## govulncheck ./...
	govulncheck ./...

# GOROOT explicit: go-licenses tells the standard library by its GOROOT.
licenses: ## fail on a forbidden, restricted or reciprocal Go dependency licence (go-licenses)
	GOROOT=$$($(GO) env GOROOT) go-licenses check ./... --disallowed_types=forbidden,restricted,reciprocal

secrets: ## scan the tracked git history for committed secrets (gitleaks)
	$(GO) run github.com/zricethezav/gitleaks/v8@$(GITLEAKS_VERSION) git --no-banner --redact .

web-install: ## npm ci in ui/
	cd $(UI_DIR) && $(NPM) ci --no-audit --no-fund

web: ## build the standalone bundle into fabricui/dist
	cd $(UI_DIR) && $(NPM) run build

web-check: ## svelte-check + vitest (pages, kit, client against the fixtures, contract guard)
	cd $(UI_DIR) && $(NPM) run check && $(NPM) test

dist-check: ## rebuild the bundle into a temporary directory and fail on any byte of difference
	rm -rf $(CHECK_DIR)/dist && mkdir -p $(CHECK_DIR)
	cd $(UI_DIR) && FABRIC_UI_OUT_DIR=$(CHECK_DIR)/dist $(NPM) run build >/dev/null
	diff -r $(DIST_DIR) $(CHECK_DIR)/dist && echo "dist-check: fabricui/dist is the build of this tree"

notices-check: ## fail when a bundled npm package is missing from THIRD-PARTY-NOTICES.md
	rm -rf $(CHECK_DIR)/notices && mkdir -p $(CHECK_DIR)
	cd $(UI_DIR) && FABRIC_UI_OUT_DIR=$(CHECK_DIR)/notices FABRIC_UI_VENDOR_LIST=$(CHECK_DIR)/vendor.json $(NPM) run build >/dev/null
	node $(UI_DIR)/scripts/check-notices.mjs $(CHECK_DIR)/vendor.json THIRD-PARTY-NOTICES.md licenses/npm

contract-check: ## the contract guard (no removal against contract/previous) and fixture use on both sides
	cd $(UI_DIR) && $(NPM) run contract:check
	CGO_ENABLED=0 $(GO) test -count=1 -run 'TestEveryFixtureIsUsedByTheGoTests' ./fabricui/
	cd $(UI_DIR) && npx vitest run src/lib/api/fixtures.test.ts scripts/contract-guard.test.mjs

gen-types: ## regenerate ui/src/lib/api/schema.gen.ts from contract/openapi.yaml
	cd $(UI_DIR) && $(NPM) run gen:types

version-check: ## internal/version, ui/package.json and the changelog's newest section agree
	@if [ -z "$(VERSION_GO)" ] || [ "$(VERSION_GO)" != "$(VERSION_NPM)" ] || [ "$(VERSION_GO)" != "$(VERSION_LOG)" ]; then \
		echo "version-check: internal/version=$(VERSION_GO) ui/package.json=$(VERSION_NPM) changelog=$(VERSION_LOG)"; exit 1; fi
	@echo "version-check: $(VERSION_GO)"

ci: build vet test lint web-check dist-check ## everything the pipeline's main jobs run

clean: ## remove local build artefacts (never the committed bundle)
	rm -rf coverage.out $(CHECK_DIR)

.PHONY: help setup build vet test race e2e fmt fmt-check lint tidy-check cover cover-check vuln licenses secrets web-install web web-check dist-check notices-check contract-check gen-types version-check ci clean
