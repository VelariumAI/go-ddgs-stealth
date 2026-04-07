# go-ddgs-stealth Roadmap

## Current Focus (v0.3.x)

1. Unified fetcher stack (`HTTPFetcher`, `StealthyFetcher`, `DynamicFetcher`) with shared anti-bot controls.
2. Production HTTP API for search and stealth endpoints.
3. Adaptive parsing and resilient spider runtime with checkpoint/resume.
4. End-to-end observability: Prometheus + OpenTelemetry spans.
5. Strong CI quality bars (tests, race, static analysis, vuln checks, cross-platform builds).

## Near-Term Milestones

- Improve dynamic browser controls (selector waits, click/script steps, screenshots).
- Expand adaptive selector persistence and conflict resolution.
- Add richer crawl policies (depth limits, allow/deny patterns, robots options).
- Add strict API auth/rate-limit middleware for hosted deployments.

## Release Discipline

- Semantic versioning with clear changelog entries.
- Breaking API changes only in major versions.
- Every feature requires tests + docs updates + release notes.
