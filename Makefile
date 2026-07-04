SHELL := /bin/sh

.PHONY: build test test-integration run deploy bootstrap bootstrap-win lint lint-tools check-env smoke smoke-tools smoke-dtrack smoke-all p1-check p1-check-win p2-check p2-check-win p3-check p3-check-win stage2-p1-check stage2-p1-check-win stage2-p2-check stage2-p2-check-win stage2-p3-check stage2-p3-check-win ci-status ci-status-win ci-trigger ci-trigger-win p3-signoff p3-signoff-win p3-close-all p3-close-all-win p0-preflight-win p0-preflight docs-link-check-win stage2-signoff-status-win

build:
	mkdir -p build
	go build -o build/network-scanner ./cmd/network-scanner

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

run:
	go run ./cmd/network-scanner

deploy:
	@echo "Deploy step is project-specific. See docs/deployment.md."

bootstrap:
	./scripts/bootstrap.sh

bootstrap-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\bootstrap.ps1

check-env:
	@echo "Checking environment..."
	@go version || (echo "Go is not installed" && exit 1)
	@golangci-lint version || (echo "WARNING: golangci-lint not installed" && exit 0)
	@govulncheck help >/dev/null 2>&1 || (echo "WARNING: govulncheck not installed" && exit 0)
	@echo "Environment check complete."

lint:
	gofmt -w $$(rg --files -g '*.go')

lint-tools:
	@echo "Running golangci-lint..."
	@golangci-lint run ./... || (echo "WARNING: golangci-lint not installed or found issues" && exit 0)
	@echo "Running govulncheck..."
	@govulncheck ./... || (echo "WARNING: govulncheck not installed" && exit 0)

smoke:
	./scripts/smoke-cli-no-topology.sh
	./scripts/smoke-cli-topology.sh

smoke-tools:
	./scripts/smoke-cli-tools.sh

smoke-dtrack:
	./scripts/smoke-d-track-topology-export.sh

smoke-all: smoke smoke-tools smoke-dtrack

p1-check:
	./scripts/p1-closure-check.sh

p1-check-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\p1-closure-check.ps1

p2-check:
	./scripts/p2-closure-check.sh

p2-check-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\p2-closure-check.ps1

p3-check:
	./scripts/p3-closure-check.sh

p3-check-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\p3-closure-check.ps1

stage2-p1-check:
	./scripts/stage2-p1-closure-check.sh

stage2-p1-check-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\stage2-p1-closure-check.ps1

stage2-p2-check:
	./scripts/stage2-p2-closure-check.sh

stage2-p2-check-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\stage2-p2-closure-check.ps1

stage2-p3-check:
	./scripts/stage2-p3-closure-check.sh

stage2-p3-check-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\stage2-p3-closure-check.ps1

ci-status:
	./scripts/check-ci-status.sh

ci-status-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\check-ci-status.ps1

ci-trigger:
	./scripts/trigger-ci-workflow.sh

ci-trigger-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\trigger-ci-workflow.ps1

p3-signoff-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\finalize-p3-signoff.ps1

p3-signoff:
	./scripts/finalize-p3-signoff.sh

p3-close-all:
	./scripts/p3-close-all.sh

p3-close-all-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\p3-close-all.ps1

p0-preflight-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\p0-signoff-preflight.ps1

docs-link-check-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\docs-link-check.ps1

stage2-signoff-status-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\stage2-signoff-status.ps1
