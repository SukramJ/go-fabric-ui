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

e2e: ## the end-to-end tests alone: a fabrictest site commissioned through the handlers, node and controller family
	CGO_ENABLED=0 $(GO) test -count=1 -run 'TestEndToEnd' -v ./fabricui/...

fmt: ## gofumpt -w .
	gofumpt -w .

fmt-check: ## fail when a file is not gofumpt-clean
	@diff=$$(gofumpt -l .); if [ -n "$$diff" ]; then echo "gofumpt would rewrite:"; echo "$$diff"; exit 1; fi

lint: fmt-check tidy-check ## gofumpt check + golangci-lint + tidy-check
	golangci-lint config verify
	golangci-lint run --timeout=5m ./...

tidy-check: ## verify go.mod/go.sum (root and tools/) are tidy and the module checksums verify
	$(GO) mod tidy -diff
	$(GO) mod verify
	cd tools && $(GO) mod tidy -diff && $(GO) mod verify

cover: ## coverage profile of every package
	CGO_ENABLED=0 $(GO) test -count=1 -timeout 300s -coverpkg=./... -coverprofile=coverage.out ./...

cover-check: cover ## fail when a package is below its floor (script/coverfloor)
	$(GO) run ./script/coverfloor coverage.out

pins-check: ## fail on a workflow `uses:` not pinned by a 40-hex commit SHA with a `# vX.Y.Z` comment
	@bad=$$(grep -nE '^[[:space:]-]*uses:' .github/workflows/*.yml | grep -vE 'uses: [^@[:space:]]+@[0-9a-f]{40} # v[0-9]+\.[0-9]+\.[0-9]+$$'); \
	if [ -n "$$bad" ]; then echo "pins-check: not pinned by SHA:"; echo "$$bad"; exit 1; fi; echo "pins-check: every action is pinned by SHA"

vuln: ## govulncheck ./...
	govulncheck ./...

# LICENSES_CHECKED are the dependencies whose licence go-licenses cannot
# classify and that were read by hand; every other unclassified licence
# fails the guard (`unknown` is disallowed). Per module the version, the
# licence and where it was read:
#   modernc.org/mathutil v1.7.1: BSD-3-Clause, read in the module's LICENSE
#     (module cache, modernc.org/mathutil@v1.7.1/LICENSE); reached through
#     go-fabric store -> modernc.org/sqlite -> modernc.org/libc.
LICENSES_CHECKED := modernc.org/mathutil

# GOROOT explicit: go-licenses tells the standard library by its GOROOT.
licenses: ## fail on a forbidden, restricted, reciprocal or unclassified Go dependency licence (go-licenses)
	GOROOT=$$($(GO) env GOROOT) go-licenses check ./... --disallowed_types=forbidden,restricted,reciprocal,unknown \
		$(addprefix --ignore ,$(LICENSES_CHECKED))

secrets: ## scan the tracked git history for committed secrets (gitleaks)
	$(GO) run github.com/zricethezav/gitleaks/v8@$(GITLEAKS_VERSION) git --no-banner --redact .

web-install: ## npm ci in ui/
	cd $(UI_DIR) && $(NPM) ci --no-audit --no-fund

npm-audit: ## fail on any npm advisory of a runtime (bundled) package (decision 54)
	cd $(UI_DIR) && $(NPM) audit --omit=dev --audit-level=low

npm-audit-all: ## report the npm advisories of the whole tree, development packages included; never fails
	-cd $(UI_DIR) && $(NPM) audit

web: ## build the standalone bundle into fabricui/dist
	cd $(UI_DIR) && $(NPM) run build

web-check: ## svelte-check + vitest (pages, kit, client against the fixtures, contract guard)
	cd $(UI_DIR) && $(NPM) run check && $(NPM) test

dist-check: ## rebuild the bundle into a temporary directory and fail on any byte of difference
	rm -rf $(CHECK_DIR)/dist && mkdir -p $(CHECK_DIR)
	cd $(UI_DIR) && FABRIC_UI_OUT_DIR=$(CHECK_DIR)/dist $(NPM) run build >/dev/null
	diff -r $(DIST_DIR) $(CHECK_DIR)/dist && echo "dist-check: fabricui/dist is the build of this tree"

# CSS_PACKAGES reach the bundle as generated CSS, which the vendor list of
# the JavaScript chunks does not see; notices-check requires each of them
# in THIRD-PARTY-NOTICES.md at its installed version, MIT, with its text.
CSS_PACKAGES := tailwindcss

notices-check: ## fail when a bundled npm package (JS chunks and CSS_PACKAGES) is missing from THIRD-PARTY-NOTICES.md
	rm -rf $(CHECK_DIR)/notices && mkdir -p $(CHECK_DIR)
	cd $(UI_DIR) && FABRIC_UI_OUT_DIR=$(CHECK_DIR)/notices FABRIC_UI_VENDOR_LIST=$(CHECK_DIR)/vendor.json $(NPM) run build >/dev/null
	node $(UI_DIR)/scripts/check-notices.mjs $(CHECK_DIR)/vendor.json THIRD-PARTY-NOTICES.md licenses/npm
	@for p in $(CSS_PACKAGES); do \
		v=$$(node -p "require('./$(UI_DIR)/node_modules/$$p/package.json').version"); \
		grep -qF "| \`$$p\` | $$v | MIT |" THIRD-PARTY-NOTICES.md || { echo "notices-check: $$p@$$v (CSS) is not in THIRD-PARTY-NOTICES.md"; exit 1; }; \
		test -f licenses/npm/$$p.txt || { echo "notices-check: no licence text at licenses/npm/$$p.txt"; exit 1; }; \
	done; echo "notices-check: CSS packages $(CSS_PACKAGES) listed with their licence texts"

contract-check: ## the contract guard (no removal against contract/previous) and fixture use on both sides
	cd $(UI_DIR) && $(NPM) run contract:check
	CGO_ENABLED=0 $(GO) test -count=1 -run 'TestEveryFixtureIsUsedByTheGoTests' ./fabricui/
	cd $(UI_DIR) && npx vitest run src/lib/api/fixtures.test.ts scripts/contract-guard.test.mjs

generate: ## regenerate the controller family's handlers (internal/api) from contract/openapi.yaml with the pinned oapi-codegen (tools/go.mod)
	script/gen-api.sh internal/api/api.gen.go

generate-check: ## fail when internal/api/api.gen.go is not the generation of contract/openapi.yaml
	mkdir -p $(CHECK_DIR)
	script/gen-api.sh $(CHECK_DIR)/api.gen.go
	diff -u internal/api/api.gen.go $(CHECK_DIR)/api.gen.go && echo "generate-check: internal/api is the generation of the contract"

gen-types: ## regenerate ui/src/lib/api/schema.gen.ts from contract/openapi.yaml
	cd $(UI_DIR) && $(NPM) run gen:types

# The browser smoke tests: the server is fabricui's TestServeE2EWeb, built
# once as a test binary; Chromium is installed into Playwright's cache.
E2E_BIN := $(CHECK_DIR)/fabricui-e2e.test

e2e-web: ## the browser smoke per page (Playwright, Chromium) against the Go e2e server
	mkdir -p $(CHECK_DIR)
	CGO_ENABLED=0 $(GO) test -c -o $(E2E_BIN) ./fabricui
	cd $(UI_DIR) && npx playwright install chromium
	cd $(UI_DIR) && FABRIC_UI_E2E_BIN=$(E2E_BIN) npx playwright test

version-check: ## internal/version, ui/package.json and the changelog's newest section agree
	@if [ -z "$(VERSION_GO)" ] || [ "$(VERSION_GO)" != "$(VERSION_NPM)" ] || [ "$(VERSION_GO)" != "$(VERSION_LOG)" ]; then \
		echo "version-check: internal/version=$(VERSION_GO) ui/package.json=$(VERSION_NPM) changelog=$(VERSION_LOG)"; exit 1; fi
	@echo "version-check: $(VERSION_GO)"

ci: build vet test lint generate-check web-check dist-check ## everything the pipeline's main jobs run

clean: ## remove local build artefacts (never the committed bundle)
	rm -rf coverage.out $(CHECK_DIR)

.PHONY: help setup build vet test race e2e e2e-web fmt fmt-check lint tidy-check pins-check cover cover-check vuln licenses secrets web-install npm-audit npm-audit-all web web-check dist-check notices-check contract-check generate generate-check gen-types version-check ci clean
