
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

The gateway reads all application settings from `config.yaml`. To use another YAML file:

```bash
go run ./cmd/aegis -config path/to/config.yaml
```

See [config.yaml](./config.yaml) for a complete example. The YAML file is the sole source of application configuration; environment variables are not used.

Routes are matched by path-segment prefix. When multiple routes match, the longest prefix wins, regardless of YAML order. The matched prefix is preserved in the path sent upstream. Requests that match no route return `404 Not Found`.

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
- Health handlers are registered for GET; other methods follow the normal configured route matching rules.

The gateway handles `SIGINT` and `SIGTERM` by marking itself unready, stopping acceptance of new connections, and draining active requests for up to the configured shutdown timeout. If the deadline expires, remaining connections are closed.

Durations in `config.yaml` use Go duration syntax (for example, `15s` or `2m`) and must be greater than zero.

Read and write timeouts apply to the entire request/response, so tune them for the gateway's upload, response, and streaming workloads.

## Continuous Integration

GitHub Actions automatically runs formatting checks, tests, static analysis, and builds on pushes and pull requests targeting main.
