# mcp-service

MCP Service — provides an HTTP REST API for querying a local Elasticsearch instance.
Sits behind the [Higress](https://higress.ai) API Gateway; all requests have already
been authenticated by `ext-auth-service` before reaching this service.

## Architecture

```
Client
  │  Authorization: Bearer <token>
  ▼
Higress Gateway
  │  ext-auth plugin validates token
  ▼
mcp-service  ──▶  Elasticsearch
```

## Environment Variables

| Variable      | Default       | Description                           |
|---------------|---------------|---------------------------------------|
| `PORT`        | `8080`        | HTTP port                             |
| `HOST`        | `0.0.0.0`     | Bind address                          |
| `ES_HOST`     | `localhost`   | Elasticsearch hostname                |
| `ES_PORT`     | `9200`        | Elasticsearch port                    |
| `ES_USERNAME` | *(none)*      | Elasticsearch username (optional)     |
| `ES_PASSWORD` | *(none)*      | Elasticsearch password (optional)     |
| `LOG_LEVEL`   | `info`        | Pino log level                        |

## Local Development

Prerequisites: a running Elasticsearch instance (Docker is easiest).

```bash
# Start Elasticsearch locally
docker run -d --name es -p 9200:9200 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  docker.elastic.co/elasticsearch/elasticsearch:8.12.2

# Install dependencies
npm install

# Run in dev mode
npm run dev
```

### Example Requests

```bash
# Search
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"index": "my-index", "query": "hello world", "size": 5}'

# List indices
curl http://localhost:8080/api/v1/indices

# Health check
curl http://localhost:8080/health
```

## Docker Build

```bash
docker build -t mandyl/mcp-service:latest .

docker run -p 8080:8080 \
  -e ES_HOST=host.docker.internal \
  mandyl/mcp-service:latest
```

## Kubernetes Deployment

```bash
# Ensure namespace exists
kubectl create namespace backend --dry-run=client -o yaml | kubectl apply -f -

# Deploy ConfigMap, Deployment, and Service
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml

# Verify
kubectl get pods -n backend -l app=mcp-service
kubectl logs -n backend -l app=mcp-service
```

## API Reference

### `POST /api/v1/search`

Full-text search against an Elasticsearch index.

**Request body:**

```json
{
  "index": "my-index",
  "query": "search keywords",
  "from": 0,
  "size": 10,
  "filters": {
    "field": "status",
    "value": "active"
  }
}
```

**Success response (`code: 0`):**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 42,
    "hits": [
      { "_id": "doc1", "_score": 1.5, "_source": { "title": "..." } }
    ],
    "from": 0,
    "size": 10
  }
}
```

**Error codes:**

| Code    | Meaning                          |
|---------|----------------------------------|
| `0`     | Success                          |
| `40001` | Index not found                  |
| `40002` | Invalid request parameters       |
| `50001` | Elasticsearch connection failed  |
| `50002` | Internal server error            |

### `GET /api/v1/indices`

Returns all non-system Elasticsearch indices.

### `GET /health`

Liveness/readiness probe. Includes ES connectivity status.
