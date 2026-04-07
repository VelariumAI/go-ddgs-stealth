# CLI Reference

Binary: `goddgs` (or `go run ./cmd/goddgs`).

## Commands

- `providers`
- `search --q <query> [--max N] [--region us-en] [--json]`
- `doctor`
- `stealth-fetch --url <url> [--mode http|stealth] [--level basic|strong|aggressive] [--human] [--json]`
- `stealth-crawl --url <url> [--max N] [--concurrency N] [--out file.jsonl]`
- `repl`

## Exit Codes

- `0`: success
- `2`: invalid input / usage / classified no-results-blocked
- `3`: runtime probe/fetch/crawl failed
- `4`: initialization/provider error
