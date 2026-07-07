# loadbalancer

A layer-7 (HTTP) load balancer written in Go from scratch — a learning / resume project.
It acts as a reverse proxy that distributes incoming HTTP requests across a pool of backend
servers using **round robin**.

## Features (current)

- **HTTP reverse proxy** — forwards client requests to backends via `httputil.ReverseProxy`
- **Server pool** — manages multiple backends behind a single entrypoint
- **Round-robin balancing** — cycles requests evenly across healthy backends (atomic counter)
- **Active health checks** — a background goroutine periodically pings each backend's `/health`
  endpoint and flips its alive flag; dead backends are dropped from rotation automatically and
  rejoin when they recover

## Project layout

```
cmd/
  lb/main.go        # load balancer entrypoint (listens on :8080)
  backend/main.go   # test backend server; run several on different ports
internal/
  loadbalancer.go   # LoadBalancer: implements http.Handler (NewLoadBalancer, ServeHTTP, Start)
  backend.go        # Backend: parsed URL + ReverseProxy + atomic alive flag
  pool.go           # ServerPool: holds backends, Healthy(), NextRoundRobin()
  health.go         # HealthChecker: background ticker loop that updates alive flags
```

## Architecture

![Load balancer architecture](docs/architecture.png)

The **Pool** lives *inside* the load balancer — it's the LB's in-memory registry of backends,
where each `Backend` card points to a real server and carries an alive flag. Two independent
flows share that registry:

- **Client flow** (on demand): `Client → Proxy → pick an alive backend from the Pool → Server`.
  The proxy *reads* the alive flags to decide where to send traffic.
- **Health flow** (on a timer): the `HealthChecker` sends its own `GET /health` to each server
  every few seconds and *writes* the alive flags back onto the Pool's cards.

`LoadBalancer` implements `http.Handler`, and each `Backend` owns its own `ReverseProxy`.
Per request: **pick** a healthy backend from the pool → **forward** via that backend's proxy →
**stream** the response back to the client. If no backend is alive, the LB returns `503`.

## Requirements

- Go 1.22+

## Running it

The backend list is currently hardcoded in [`cmd/lb/main.go`](cmd/lb/main.go)
(`:9001`, `:9002`, `:9003`).

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

The load balancer now listens on `http://localhost:8080`.

## Testing it

Send requests and watch them rotate across backends. Each backend echoes its own port:

```bash
for i in $(seq 1 6); do curl -s localhost:8080/; done
```

Expected output (round robin cycling through the pool):

```
Hello from backend on port 9001 (path: /)
Hello from backend on port 9002 (path: /)
Hello from backend on port 9003 (path: /)
Hello from backend on port 9001 (path: /)
Hello from backend on port 9002 (path: /)
Hello from backend on port 9003 (path: /)
```

The load balancer also logs each request and the backend it chose:

```
[lb] GET / -> http://localhost:9001
[lb] GET / -> http://localhost:9002
[lb] GET / -> http://localhost:9003
```

Each backend exposes a `/health` endpoint (returns `200 OK`) that the upcoming health checker
will use.

## Roadmap

- [x] **Active health checks** — background goroutine that pings each backend and drops dead ones from rotation automatically
- [ ] Request logging as middleware (method, path, status, latency)
- [ ] YAML config to replace the hardcoded backend list
- [ ] `Strategy` interface + more algorithms (weighted, least-connections, IP hash)
- [ ] Metrics + `/stats` endpoint
- [ ] TLS termination
- [ ] Graceful shutdown
- [ ] Custom `Rewrite` / `SetXForwarded` (sets `X-Forwarded-Host` / `-Proto`, controls IP spoofing)
```
