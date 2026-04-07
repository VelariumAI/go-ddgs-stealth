# Changelog

## Unreleased

## v0.3.0 - 2026-04-07

- Added unified fetcher architecture with shared anti-bot inheritance:
  - `Fetcher` / `StealthFetcher` interfaces
  - `HTTPFetcher`, `StealthyFetcher` (Rod), and `DynamicFetcher`
- Added adaptive parsing subsystem:
  - `AdaptiveParser` with selector self-healing
  - selector fingerprint persistence (`Save` / `Load`)
- Added crawler runtime:
  - concurrent `Spider` with per-domain pacing
  - checkpoint + JSONL streaming
  - pause/resume support
  - policy controls (`MaxDepth`, `AllowHosts`, `DenyHosts`)
- Added session orchestration utilities:
  - `FetcherSessionPool`
  - TTL-based `SessionHub`
- Added stealth HTTP APIs:
  - `POST /v1/stealth/fetch`
  - `POST /v1/stealth/crawl`
- Added stealth endpoint guards:
  - optional bearer token auth (`GODDGS_API_TOKEN`)
  - per-IP rate limiting (`GODDGS_STEALTH_RATE_PER_MIN`)
- Added observability extensions:
  - Prometheus stealth metrics
  - OpenTelemetry tracing hooks for stealth HTTP lifecycle
- Expanded CLI:
  - `stealth-fetch`, `stealth-crawl`, and `repl`
  - enhanced doctor diagnostics for stealth configuration paths
- Added benchmark scaffold under `bench/`.
- Improved release pipeline to publish `SHA256SUMS.txt` with release artifacts.
- Strengthened docs and roadmap for `go-ddgs-stealth` identity and release scope.

## v0.2.1 - 2026-04-07

- Completed post-rename path consistency updates for `go-ddgs-stealth`.
- Updated README/docs badges, links, imports, and release helper scripts to new repo/module path.

## v0.2.0 - 2026-04-07

- Renamed project repository to `go-ddgs-stealth`.
- Updated Go module path to `github.com/velariumai/go-ddgs-stealth` (breaking import path change).
- Updated documentation, badges, scripts, and examples for the new naming.

## v0.1.3 - 2026-04-07

- Added transport-level decompression for `gzip`, `br`, and `zstd`.
- Fixed gzip decompression close-path bug and hardened `Content-Encoding` parsing.
- Updated Chrome header profile (`Accept-Encoding` with `zstd`, `Priority` hints).
- Corrected `Sec-Fetch-Site` for script requests to `same-site`.
- Added decompression regression tests (including noop/uncompressed paths).

## v0.1.2 - 2026-04-07

- Re-added Go Reference badge in README.
- Applied `gofmt -s` formatting fixes to solver sources.

## v0.1.1 - 2026-04-07

- Aligned release/tag state with reconciled remote `main`.
- Revamped documentation for consistency with implemented solver and anti-bot capabilities.
- Added comprehensive docs index and architecture/configuration/anti-bot guides.

## v0.1.0 - 2026-04-07

- Added DDG-first resilient search client with typed block detection.
- Added provider failover engine with adapters for Brave, Tavily, and SerpAPI.
- Added `goddgs` CLI and `goddgsd` HTTP service runtimes.
- Added structured event hooks and Prometheus observability.
- Added anti-bot resilience hardening (fresh VQD retry, solver retry budget fix, circuit breaker fail-fast).
- Added OSS governance/release scaffolding and CI quality gates.
- Enforced total test coverage gate at `>=85.0%`.
