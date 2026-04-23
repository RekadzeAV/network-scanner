# Contributing

## Branching and commits

- Use short-lived branches from `main`.
- Follow Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
- Reference a related issue/task in every PR description.

## Development workflow

1. Run project bootstrap:
   - Linux/macOS: `./scripts/bootstrap.sh`
   - Windows PowerShell: `.\scripts\bootstrap.ps1`
2. Implement changes.
3. Run checks:
   - `make test` (or `go test ./...`)
   - smoke checks from `scripts/`
4. Update documentation when behavior/API changes.

## Naming and style

- Keep package names short and lowercase.
- Use `gofmt` for all Go files.
- Keep new docs in `docs/` using Markdown.

## Pull requests

- Use repository PR template/checklist.
- Include test plan and rollback notes if release behavior changes.
