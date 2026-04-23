# ADR 0001: Project governance baseline

## Context

The project needs a stable engineering baseline for repeatable builds, test execution, onboarding, and documentation synchronization.

## Decision

- Introduce unified command entry points in `Makefile`.
- Add bootstrap scripts for Linux/macOS and Windows.
- Add deployment + rollback documentation.
- Keep a canonical roadmap entry in [ROADMAP.md](../ROADMAP.md).
- Require Conventional Commits and documentation updates in `CONTRIBUTING.md`.

## Alternatives

- Keep ad-hoc scripts only: rejected due to inconsistent workflows.
- Keep process rules outside repository: rejected due to poor traceability.

## Consequences

- Faster onboarding and lower process ambiguity.
- Slight maintenance overhead for process docs and scripts.
