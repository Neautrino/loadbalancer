# loadbalancer

A layer-7 (HTTP) load balancer written in Go from scratch — a learning / resume project.
It reverse-proxies incoming HTTP requests across a pool of backend servers, with **pluggable
balancing algorithms**, **active + passive health checking**, and **per-backend circuit breakers**.

## Features

- **HTTP reverse proxy** — forwards client requests to backends via `httputil.ReverseProxy`
- **Server pool** — manages multiple backends behind a single entrypoint
- **5 pluggable algorithms** behind a `Strategy` interface — swap with one line:
  round robin, least connections, weighted round robin, IP hash, and consistent hashing
- **Active health checks** — a background goroutine pings each backend's `/health` and drops
  dead backends from rotation automatically (they rejoin on recovery)
- **Passive health via circuit breakers** — each backend has its own breaker
  (`closed → open → half-open`) that trips after repeated failures and self-probes for recovery
- **Structured request logging** — `slog`-based middleware capturing method, path, status, latency

## Architecture

![Load balancer architecture](docs/architecture.png)

A request flows: **Client → LoggingMiddleware → Strategy → ReverseProxy → Server**.
The `Strategy` only ever picks from `Pool.Healthy()`, and a backend is healthy only when
**both** gates agree:

```
Healthy() = alive (set by HealthChecker)  AND  breaker.Allow() (per-backend CircuitBreaker)
```

Three independent flows coordinate through shared per-backend state, never calling each other directly:

- **Request flow** (on demand): pick a healthy backend → forward → stream the response back.
- **Health flow** (background timer): `HealthChecker` GETs `/health` on each server → sets `alive`.
- **Breaker flow** (on real outcomes): the proxy's `ModifyResponse`/`ErrorHandler` hooks call
  `RecordSuccess`/`RecordFailure`, tripping the circuit after N failures.

## Balancing algorithms

| Algorithm | How it picks | Best for |
|---|---|---|
| **Round robin** | cycles through backends in order (atomic counter) | even load, simplest |
| **Least connections** | fewest in-flight requests (atomic per-backend counter) | uneven request durations |
| **Weighted round robin** | smooth WRR proportional to per-backend weight | heterogeneous backends |
| **IP hash** | `hash(client IP) % n` — sticky sessions | simple client affinity |
| **Consistent hash** | hash ring + virtual nodes — sticky with minimal remapping | affinity that survives backend changes |

The algorithm is selected by the strategy injected in [`cmd/lb/main.go`](cmd/lb/main.go)
(config-driven selection is on the roadmap).

## Project layout

```
cmd/
  lb/main.go            # LB entrypoint — wires pool + strategy, starts the server
  backend/main.go       # test backend server (run several on different ports)
internal/
  loadbalancer.go       # LoadBalancer: http.Handler (pick → forward), Start()
  middleware.go         # LoggingMiddleware + statusRecorder (captures status + latency)
  health.go             # HealthChecker: background ticker pinging /health
  pool/
    pool.go             # ServerPool: Healthy(), Backends()
    backend.go          # Backend: URL, ReverseProxy (+ hooks), alive, activeConns, weight, breaker
    circuitbreaker.go   # CircuitBreaker: closed/open/half-open state machine
  algorithms/
    strategy.go         # Strategy interface (Next)
    roundrobin.go       # round robin
    leastconn.go        # least active connections
    weighted.go         # smooth weighted round robin
    iphash.go           # IP hash (sticky)
    consistenthash.go   # consistent hashing (sticky, minimal remap on failover)
```

## Requirements

- Go 1.22+

## Running it

The backend list and algorithm are currently set in [`cmd/lb/main.go`](cmd/lb/main.go).

**1. Start the backend servers** (three terminals, or background them with `&`):

```bash
go run ./cmd/backend -port=9001
go run ./cmd/backend -port=9002
go run ./cmd/backend -port=9003
```

**2. Start the load balancer:**

```bash
go run ./cmd/lb
```

The load balancer listens on `http://localhost:8080`.

## Testing it

Each backend echoes its own port, so you can watch requests distribute:

```bash
for i in $(seq 1 6); do curl -s localhost:8080/; done
```

Round robin cycles through the pool:

```
Hello from backend on port 9001 (path: /)
Hello from backend on port 9002 (path: /)
Hello from backend on port 9003 (path: /)
...
```

**Least connections / weighted** only differ visibly under concurrent load — use the `/slow`
endpoint and background the requests:

```bash
for i in $(seq 1 6); do curl -s localhost:8080/slow & done; wait
```

**Consistent hashing** stickiness (same client → same backend) can be tested with spoofed IPs:

```bash
for ip in 1.1.1.1 2.2.2.2 3.3.3.3; do curl -s -H "X-Forwarded-For: $ip" localhost:8080/; done
```

**Health / circuit breaker**: kill a backend and watch it drop from rotation (active check), or
watch a backend's circuit trip to `open` after repeated failures in the logs.

## Roadmap

- [x] Reverse proxy + server pool
- [x] `Strategy` interface + 5 algorithms (round robin, least-conn, weighted, IP hash, consistent hash)
- [x] Active health checks
- [x] Passive health checks via per-backend circuit breakers
- [x] Structured request logging middleware (`slog`)
- [ ] **YAML config** — replace the hardcoded backends/algorithm/weights (next)
- [ ] Metrics + `/stats` endpoint
- [ ] TLS termination
- [ ] Graceful shutdown
- [ ] Custom `Rewrite` / `SetXForwarded` (sets `X-Forwarded-Host` / `-Proto`, controls IP spoofing)
```
