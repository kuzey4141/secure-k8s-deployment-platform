# Secure Kubernetes Deployment Platform

A Go-based internal platform for receiving deployment requests, validating them through security and quality gates, and preparing them for safe Kubernetes delivery.

This repository currently contains the first MVP step: a lightweight API gateway that accepts deployment requests, stores them in PostgreSQL, and exposes endpoints for listing and inspecting deployment history.

## Current MVP Scope

- `POST /api/deployments`
- `GET /api/deployments`
- `GET /api/deployments/:id`
- PostgreSQL schema migration
- Local PostgreSQL setup with Docker Compose

The next phases will add policy evaluation, service-to-service communication, eventing, Helm-based deployment, and observability.

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
policies/           # Reserved for upcoming Rego policy files
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

Each deployment is currently stored with a basic lifecycle status:

- `pending`

In the next phase, requests will be evaluated by policy rules and will return a more meaningful result such as:

- `accepted`
- `rejected`

## Roadmap

Planned next steps:

1. Add OPA/Rego-based policy checks to the create flow
2. Store policy violations for rejected deployments
3. Add a simple React dashboard
4. Introduce gRPC between services
5. Publish deployment events through Kafka
6. Add Helm-based Kubernetes deployment
7. Add Redis-backed status caching
8. Add Prometheus and Grafana observability

## Notes

- `.codex` is intentionally ignored and is not part of the repository
- The current implementation is focused on the backend MVP only
- Helm, Kafka, Redis, Terraform, and the dashboard are planned but not yet active
