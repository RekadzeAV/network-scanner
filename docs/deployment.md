# Deployment

## Overview

Deployment is environment-specific and should be automated through CI/CD workflows.

Recommended entry points:

- `make build`
- `make test`
- `make deploy`

## Rollback

If a release must be reverted:

1. Identify the last stable tag (`git tag --sort=-v:refname`).
2. Redeploy artifacts built from the stable tag.
3. Validate via smoke checks:
   - `./scripts/integration-check.sh` (Linux/macOS)
   - `.\scripts\integration-check.ps1` (Windows)
4. Record the rollback reason and impact in release notes/changelog.
