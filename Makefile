SHELL := /bin/sh

.PHONY: build test test-integration run deploy bootstrap lint smoke smoke-tools smoke-dtrack smoke-all p1-check p1-check-win p2-check p2-check-win p3-check p3-check-win stage2-p1-check stage2-p1-check-win stage2-p2-check stage2-p2-check-win stage2-p3-check stage2-p3-check-win ci-status ci-status-win ci-trigger ci-trigger-win p3-signoff p3-signoff-win p3-close-all p3-close-all-win

build:
	go build -o network-scanner ./cmd/network-scanner

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

lint:
	gofmt -w $$(rg --files -g '*.go')

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
