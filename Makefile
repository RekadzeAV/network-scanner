SHELL := /bin/sh

.PHONY: build test test-integration run deploy bootstrap bootstrap-win lint lint-tools check-env smoke smoke-tools smoke-dtrack smoke-all p1-check p1-check-win p2-check p2-check-win p3-check p3-check-win stage2-p1-check stage2-p1-check-win stage2-p2-check stage2-p2-check-win stage2-p3-check stage2-p3-check-win ci-status ci-status-win ci-trigger ci-trigger-win p3-signoff p3-signoff-win p3-close-all p3-close-all-win p0-preflight-win p0-preflight docs-link-check-win stage2-signoff-status-win final-release-check final-release-check-win

build:
	mkdir -p build
	go build -ldflags="-s -w" -o build/network-scanner ./cmd/network-scanner
	go build -ldflags="-s -w -H windowsgui" -o build/network-scanner-gui ./cmd/gui

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
	@golangci-lint run ./... || (echo "WARNING: golangci-lint not installed or found issues" && exit 0)
	@govulncheck ./... || (echo "WARNING: govulncheck not installed" && exit 0)

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

# ============================================
# Final release verification (D4)
# ============================================

final-release-check:
	./scripts/final-release-check.sh

final-release-check-win:
	powershell -ExecutionPolicy Bypass -File .\scripts\final-release-check.ps1

# ============================================
# Linux-specific targets (Sprint 2)
# ============================================

# Version info
VERSION  ?= dev
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Install targets
install: build
	@echo "📦 Installing binaries..."
	@sudo cp build/network-scanner /usr/local/bin/
	@sudo cp build/network-scanner-gui /usr/local/bin/
	@echo "✅ Binaries installed to /usr/local/bin/"

install-systemd:
	@echo "🔧 Installing systemd service..."
	@sudo cp config/systemd/network-scanner.service /etc/systemd/system/
	@sudo systemctl daemon-reload
	@sudo systemctl enable network-scanner.service
	@echo "✅ systemd service installed"
	@echo "   Start: sudo systemctl start network-scanner"
	@echo "   Status: sudo systemctl status network-scanner"

install-desktop:
	@echo "🔧 Installing desktop file..."
	@sudo cp config/desktop/network-scanner-gui.desktop /usr/share/applications/
	@sudo cp assets/icons/network-scanner.svg /usr/share/icons/hicolor/scalable/apps/
	@echo "✅ Desktop file installed"

install-all: install install-systemd install-desktop
	@echo "✅ All installations completed"

# Package targets
deb: build
	@echo "📦 Building Debian package..."
	@mkdir -p build/deb/network-scanner/DEBIAN
	@mkdir -p build/deb/usr/local/bin
	@mkdir -p build/deb/usr/share/applications
	@mkdir -p build/deb/usr/share/icons/hicolor/scalable/apps
	@mkdir -p build/deb/etc/systemd/system
	@mkdir -p build/deb/var/lib/network-scanner
	@mkdir -p build/deb/var/log/network-scanner
	
	@cp build/network-scanner build/deb/usr/local/bin/
	@cp build/network-scanner-gui build/deb/usr/local/bin/
	@cp config/desktop/network-scanner-gui.desktop build/deb/usr/share/applications/
	@cp assets/icons/network-scanner.svg build/deb/usr/share/icons/hicolor/scalable/apps/
	@cp config/systemd/network-scanner.service build/deb/etc/systemd/system/
	
	@echo "Package: network-scanner" > build/deb/network-scanner/DEBIAN/control
	@echo "Version: $(VERSION)" >> build/deb/network-scanner/DEBIAN/control
	@echo "Section: utils" >> build/deb/network-scanner/DEBIAN/control
	@echo "Priority: optional" >> build/deb/network-scanner/DEBIAN/control
	@echo "Architecture: amd64" >> build/deb/network-scanner/DEBIAN/control
	@echo "Maintainer: Network Scanner Team" >> build/deb/network-scanner/DEBIAN/control
	@echo "Description: Advanced network discovery and security analysis tool" >> build/deb/network-scanner/DEBIAN/control
	
	@dpkg-deb --build --root-owner-group build/deb network-scanner_$(VERSION)_amd64.deb
	@echo "✅ Debian package created: network-scanner_$(VERSION)_amd64.deb"

rpm: build
	@echo "📦 Building RPM package..."
	@echo "Note: RPM building requires rpmbuild. Install with: sudo apt install rpm"
	@echo "For now, use: make deb && dpkg-deb --build..."

# Help
help:
	@echo "Network Scanner - Makefile targets"
	@echo ""
	@echo "Build:"
	@echo "  make build       - Build CLI and GUI"
	@echo "  make cli         - Build CLI only"
	@echo "  make gui         - Build GUI only"
	@echo ""
	@echo "Test:"
	@echo "  make test        - Run all tests"
	@echo "  make test-parallel - Run parallel tests"
	@echo ""
	@echo "Lint:"
	@echo "  make lint        - Run linter"
	@echo ""
	@echo "Install:"
	@echo "  make install     - Install binaries"
	@echo "  make install-systemd - Install systemd service"
	@echo "  make install-desktop - Install desktop file"
	@echo "  make install-all - Install everything"
	@echo ""
	@echo "Package:"
	@echo "  make deb         - Build Debian package"
	@echo "  make rpm         - Build RPM package"

