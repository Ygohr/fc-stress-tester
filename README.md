# Stress Tester - Go + Cobra

A CLI stress testing tool for HTTP services, written in idiomatic Go.

---

## Project Overview

`stress-tester` performs concurrent HTTP GET requests against a target URL and produces a detailed execution report. It is built to be simple, testable, and production-ready without unnecessary complexity.

---

## Architecture Decisions

### Dependency Direction

```
CLI (cmd/)
  ↓
Load Test Service (internal/service/stress_test/)
  ↓
HTTPClient Interface (internal/contract/)
  ↓
HTTPClient Implementation (infrastructure/http/)
```

Business logic never imports `net/http` directly — it depends only on the `contract.HTTPClient` interface. This enables full unit testing via mocks.

### No repository or use-case layers

The domain is simple: one service, one operation. Extra layers would add indirection without value.

### Strategy Pattern for HTTP Client

The `contract.HTTPClient` interface allows swapping implementations (mock, real, custom transport) without touching business logic.

---

## Folder Structure

```
.
├── cmd/
│   ├── main/ 
│   |   └── main.go      # Entry point
│   ├── root.go          # Cobra root command
│   └── run.go           # Load test command + flag validation + report printing
│
├── internal/
│   ├── contract/
│   │   └── http_client.go      # HTTPClient interface
│   │
│   ├── service/loadtest/
│   │   ├── model.go            # Config, Report, result types
│   │   ├── service.go          # Orchestrates the load test
│   │   ├── worker_pool.go      # Concurrent worker pool
│   │   ├── report.go           # Report aggregation
│   │   ├── service_test.go     # Unit tests for service
│   │   ├── worker_pool_test.go # Unit tests for worker pool
│   │   └── integration_test.go # Integration tests (real httptest server)
│   │
│   └── testutil/
│       ├── mock_http_client.go # testify mock
│       └── test_server.go      # httptest helper
│
├── infrastructure/http/
│   ├── client.go        # Concrete HTTP client (30s timeout)
│   └── client_test.go   # HTTP client tests
│
├── Dockerfile
└── README.md
```

---

## Worker Pool Design

```
main goroutine
  │
  ├── fills jobs channel (buffered, size = totalRequests)
  ├── closes jobs channel
  │
  ├── spawns N worker goroutines (N = concurrency)
  │     each reads from jobs channel until drained
  │     each writes result to results channel
  │
  └── waiter goroutine: wg.Wait() → closes results channel
        main goroutine drains results channel
```

- No goroutine leaks: workers stop when `jobs` is drained; the waiter closes `results` after all workers finish.
- Context cancellation: workers check `ctx.Done()` before each request and record a context error instead of making a network call.
- Exact request count guaranteed: jobs channel is pre-filled with exactly `totalRequests` items.

---

## How to Run Locally

**Prerequisites:** Go 1.22+

```bash
go mod download
go run ./cmd/main/main.go --url=http://example.com --requests=100 --concurrency=10
```

---

## How to Execute Tests

```bash
go test ./... -v -race
```

## Docker

### Build the image

```bash
docker build -t stress-tester .
```

### Run the container

```bash
docker run --rm stress-tester \
  --url=http://example.com \
  --requests=1000 \
  --concurrency=10
```

### Run with Docker Compose

```bash
docker compose up --build
```

With custom parameters:

```bash
# Linux / macOS
URL=https://example.com REQUESTS=1000 CONCURRENCY=20 docker compose up --build

# Windows PowerShell
$env:URL="https://example.com"; $env:REQUESTS=1000; $env:CONCURRENCY=20; docker compose up --build

# Windows CMD
set URL=https://example.com && set REQUESTS=1000 && set CONCURRENCY=20 && docker compose up --build
```
---