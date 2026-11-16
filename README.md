
# Aegis API Gateway

A high-performance API gateway written in Go.

## Prerequisites

- Go (project-supported version)
- Git

## Getting Started

Clone the repository:

```bash
git clone https://github.com/HitanshuDev/aegis-api-gateway.git
cd aegis-api-gateway
```

## Development

Run the gateway:

```bash
make run
```

Run tests:

```bash
make test
```

Run static analysis:

```bash
make vet
```

Build:

```bash
make build
```

## Health Checks and Shutdown

- `GET /health/live` returns `200 OK` while the process is running.
- `GET /health/ready` returns `200 OK` while the gateway is accepting traffic and `503 Service Unavailable` after shutdown begins.
- Other methods to these health endpoints return `405 Method Not Allowed`.

The gateway handles `SIGINT` and `SIGTERM` by marking itself unready, stopping acceptance of new connections, and draining active requests for up to the configured shutdown timeout. If the deadline expires, remaining connections are closed.

Durations use Go duration syntax (for example, `15s` or `2m`). The following environment variables are configurable and must be greater than zero:

| Variable | Default |
| --- | --- |
| `AEGIS_READ_TIMEOUT` | `15s` |
| `AEGIS_READ_HEADER_TIMEOUT` | `5s` |
| `AEGIS_WRITE_TIMEOUT` | `30s` |
| `AEGIS_IDLE_TIMEOUT` | `60s` |
| `AEGIS_SHUTDOWN_TIMEOUT` | `30s` |

Read and write timeouts apply to the entire request/response, so tune them for the gateway's upload, response, and streaming workloads.

## Continuous Integration

GitHub Actions automatically runs formatting checks, tests, static analysis, and builds on pushes and pull requests targeting main.
