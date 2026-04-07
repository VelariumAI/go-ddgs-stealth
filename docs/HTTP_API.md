# HTTP API

`goddgsd` exposes a production HTTP surface.

## Endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /metrics`
- `POST /v1/search`
- `POST /v1/stealth/fetch`
- `POST /v1/stealth/crawl`

## /v1/search

Request:

```json
{
  "query": "golang",
  "max_results": 5,
  "region": "us-en"
}
```

## /v1/stealth/fetch

Request:

```json
{
  "url": "https://example.com",
  "method": "GET",
  "mode": "http",
  "human_like": true,
  "stealth_level": "strong"
}
```

Notes:

- `mode`: `http` or `stealth`.
- `stealth_level`: `basic|strong|aggressive`.

Response is `FetchResponse` JSON.

## /v1/stealth/crawl

Request:

```json
{
  "start_url": "https://example.com",
  "max_pages": 50,
  "concurrency": 4,
  "checkpoint_file": "/tmp/crawl.checkpoint.json",
  "streaming_jsonl": "/tmp/crawl.jsonl"
}
```

Response:

```json
{
  "started_at": "...",
  "ended_at": "...",
  "status": "ok"
}
```

## Error Payload

```json
{
  "error": "error message",
  "kind": "error_kind"
}
```
