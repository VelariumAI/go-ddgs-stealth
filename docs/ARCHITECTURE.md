# Architecture

## Core Layers

- `Engine`: provider-chain orchestration for web search.
- `Client`: DDG-native search transport with anti-bot controls.
- `Fetcher` family:
  - `HTTPFetcher` reusing anti-bot transport/session/rate/circuit/solver stack.
  - `StealthyFetcher` (Rod browser runtime).
  - `DynamicFetcher` for interactive browser flows.
- `AdaptiveParser`: CSS selector self-healing with fingerprint persistence.
- `Spider`: concurrent crawler with checkpoint + JSONL output.
- `Service`: HTTP API for search and stealth operations.

## Anti-Bot Model

Shared anti-bot primitives are defined in `AntiBotConfig` and inherited by fetchers:

- UA rotation
- ChromeTLS profile
- session warmup/invalidation
- adaptive delay + jitter
- circuit breaker
- challenge solver chain
- proxy pool

## Observability

- Prometheus metrics for search and stealth runtime.
- OpenTelemetry span hooks (`http.stealth.fetch`, `http.stealth.crawl`) for service tracing.

## Runtime Entrypoints

- CLI: `cmd/goddgs`
- Service daemon: `cmd/goddgsd`
