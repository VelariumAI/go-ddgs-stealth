# go-ddgs-stealth
[![CI](https://github.com/velariumai/go-ddgs-stealth/actions/workflows/ci.yml/badge.svg)](https://github.com/velariumai/go-ddgs-stealth/actions/workflows/ci.yml)
[![Release](https://github.com/velariumai/go-ddgs-stealth/actions/workflows/release.yml/badge.svg)](https://github.com/velariumai/go-ddgs-stealth/actions/workflows/release.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/velariumai/go-ddgs-stealth.svg)](https://pkg.go.dev/github.com/velariumai/go-ddgs-stealth)

`go-ddgs-stealth` is a Go-native search + stealth fetching toolkit.

## What It Includes

- DDG-first search engine with provider failover (`ddg`, `brave`, `tavily`, `serpapi`).
- Unified fetcher interfaces:
  - `HTTPFetcher` (anti-bot transport primitives)
  - `StealthyFetcher` (Rod browser-backed)
  - `DynamicFetcher` (interactive flow wrapper)
- Adaptive parser with selector self-healing and persistence.
- Spider/crawler with concurrency, per-domain pacing, JSONL streaming, and checkpoint resume.
- Session pooling helpers for multi-fetcher orchestration.
- Prometheus metrics and OpenTelemetry span hooks.
- CLI + HTTP service runtime.

## Install

```bash
go get github.com/velariumai/go-ddgs-stealth
```

## Quick Start

```go
cfg := goddgs.LoadConfigFromEnv()
engine, err := goddgs.NewDefaultEngineFromConfig(cfg)
if err != nil {
    panic(err)
}
resp, err := engine.Search(context.Background(), goddgs.SearchRequest{Query: "golang", MaxResults: 5})
if err != nil {
    panic(err)
}
fmt.Println(resp.Provider, len(resp.Results))
```

## CLI

```bash
go run ./cmd/goddgs search --q "golang" --json
go run ./cmd/goddgs stealth-fetch --url https://example.com --mode http
go run ./cmd/goddgs stealth-crawl --url https://example.com --max 20 --out /tmp/crawl.jsonl
go run ./cmd/goddgs doctor
```

## HTTP Service

```bash
go run ./cmd/goddgsd
```

Endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /metrics`
- `POST /v1/search`
- `POST /v1/stealth/fetch`
- `POST /v1/stealth/crawl`

## Documentation

- [docs/README.md](docs/README.md)
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- [docs/API_REFERENCE.md](docs/API_REFERENCE.md)
- [docs/HTTP_API.md](docs/HTTP_API.md)
- [docs/CLI.md](docs/CLI.md)
- [ROADMAP.md](ROADMAP.md)

## Development

```bash
make fmt
make vet
go test ./...
./scripts/check_coverage.sh 85.0
```
