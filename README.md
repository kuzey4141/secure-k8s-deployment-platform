# Secure Kubernetes Deployment Platform

A Go-based internal platform for receiving deployment requests, validating them through security and quality gates, and preparing them for safe Kubernetes delivery.

This repository currently contains the first MVP step: a lightweight API gateway that accepts deployment requests, stores them in PostgreSQL, and exposes endpoints for listing and inspecting deployment history.

## Current MVP Scope

- `POST /api/deployments`
- `GET /api/deployments`
- `GET /api/deployments/:id`
- In-process OPA/Rego policy evaluation from `policies/`
- Shared control metadata from `policies/controls.json`
- Rejected deployment violations stored in PostgreSQL
- PostgreSQL schema migration
- Local PostgreSQL setup with Docker Compose

The next phases will add richer policy reporting, service-to-service communication, eventing, Helm-based deployment, and observability.

## Why This Project?

The goal of this project is not to repeat a cluster security scanning workflow. Instead, it focuses on the secure deployment path itself:

- Receive deployment requests through a backend API
- Validate requests before deployment
- Build a foundation for policy-as-code with OPA/Rego
- Prepare the platform for future gRPC, Kafka, Helm, Redis, and observability integrations

This makes the project a good portfolio piece for backend engineering, platform engineering, Kubernetes, and cloud security.

## Project Structure

```text
backend/
  api-gateway/      # The only active Go service in the current MVP
  migrations/       # PostgreSQL schema files
policies/           # Rego controls plus shared JSON control metadata
docker-compose.yml  # Local PostgreSQL setup
```

## Getting Started

1. Start PostgreSQL:

```bash
docker compose up -d postgres
```

2. Run the API:

```bash
go run ./backend/api-gateway/cmd/server
```

Default environment values:

```bash
export DATABASE_URL='postgres://secure:secure@localhost:5432/secure_deploy?sslmode=disable'
export HTTP_ADDR=':8080'
export POLICY_PATH='policies'
```

## API Example

```bash
curl -X POST http://localhost:8080/api/deployments \
  -H 'Content-Type: application/json' \
  -d '{
    "app_name": "payment-service",
    "image": "bugra/payment-service:v1.2.0",
    "namespace": "production",
    "replicas": 2,
    "cpu_limit": "500m",
    "memory_limit": "512Mi",
    "privileged": false
  }'
```

List deployments:

```bash
curl http://localhost:8080/api/deployments
```

Get a single deployment by ID:

```bash
curl http://localhost:8080/api/deployments/<deployment-id>
```

## Current Response Model

Each deployment request is now evaluated against Rego control files before it is stored. Shared metadata such as `severity` and `message` comes from `policies/controls.json`.

The resulting lifecycle status is one of:

- `accepted`
- `rejected`

Rejected requests also create `policy_violations` rows linked to the deployment record.

When a deployment has stored violations, `GET /api/deployments/:id` now includes a `violations` array in the response.

## Roadmap

Planned next steps:

1. Add a simple React dashboard
2. Introduce gRPC between services
3. Publish deployment events through Kafka
4. Add Helm-based Kubernetes deployment
5. Add Redis-backed status caching
6. Add Prometheus and Grafana observability

## Notes

- `.codex` is intentionally ignored and is not part of the repository
- The current implementation is focused on the backend MVP only
- Helm, Kafka, Redis, Terraform, and the dashboard are planned but not yet active
