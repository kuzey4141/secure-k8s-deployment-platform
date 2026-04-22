# Secure Kubernetes Deployment Platform

Bu repo ilk olarak kucuk bir MVP ile baslar: deploy istegini alan, PostgreSQL'e kaydeden ve listeleyen bir Go API.

## Ilk Hedef

- `POST /api/deployments`
- `GET /api/deployments`
- `GET /api/deployments/:id`
- PostgreSQL migration
- Docker Compose ile lokal PostgreSQL

Policy check, gRPC, Kafka, Helm ve dashboard bir sonraki asamalarda eklenecek.

## Proje Yapisi

```text
backend/
  api-gateway/      # Ilk asamada calisan tek Go servisi
  migrations/       # PostgreSQL schema
policies/           # Rego policy dosyalari sonraki asama icin
docker-compose.yml  # Lokal PostgreSQL
```

## Calistirma

1. PostgreSQL'i kaldir:

```bash
docker compose up -d postgres
```

2. API'yi calistir:

```bash
go run ./backend/api-gateway/cmd/server
```

Varsayilan baglanti bilgileri:

```bash
export DATABASE_URL='postgres://secure:secure@localhost:5432/secure_deploy?sslmode=disable'
export HTTP_ADDR=':8080'
```

## Ornek Istek

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

Listeleme:

```bash
curl http://localhost:8080/api/deployments
```

## Sonraki Adim

Bir sonraki asamada `policies/deployment.rego` ekleyip create akisinda request'i accepted veya rejected yapacagiz.
