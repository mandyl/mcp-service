# mcp-service

MCP 业务服务，提供基于 Elasticsearch 的数据查询 API。

## 功能

- `POST /api/v1/search`：全文搜索，支持分页与字段过滤
- `GET /api/v1/indices`：查询可用 ES 索引列表
- `GET /health`：健康检查（含 ES 连通性）

## 快速开始

### 本地运行

```bash
export PORT=8080
export ES_ADDR=http://localhost:9200   # 或使用 ES_HOST + ES_PORT
go run ./cmd/server
```

### Docker 构建

```bash
docker build -t mandyl/mcp-service:latest .
```

### Kubernetes 部署

```bash
kubectl create namespace backend
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
```

## 环境变量

| 变量        | 默认值                                               | 说明                    |
|------------|------------------------------------------------------|------------------------|
| `PORT`     | `8080`                                               | 服务监听端口             |
| `ES_ADDR`  | `http://elasticsearch.backend.svc.cluster.local:9200` | ES 完整地址（优先级高）  |
| `ES_HOST`  | `elasticsearch.backend.svc.cluster.local`            | ES 主机名               |
| `ES_PORT`  | `9200`                                               | ES 端口                 |

> **注意**：若设置了 `ES_ADDR`，则 `ES_HOST` / `ES_PORT` 被忽略。

## API

### POST /api/v1/search

```json
// 请求
{
  "index": "logs-2026-03",
  "query": "error",
  "from": 0,
  "size": 10,
  "filters": { "field": "level", "value": "ERROR" }
}

// 响应（成功）
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 42,
    "hits": [{ "_id": "xxx", "_score": 1.5, "_source": {} }],
    "from": 0,
    "size": 10
  }
}
```

### GET /api/v1/indices

```json
{ "code": 0, "message": "success", "data": { "indices": ["logs-2026-03"] } }
```

### GET /health

```json
{ "status": "ok", "es_connected": true, "timestamp": "2026-03-06T12:00:00Z" }
```

## 错误码

| 错误码  | 说明               |
|--------|--------------------|
| 0      | 成功               |
| 40001  | 索引不存在          |
| 40002  | 请求参数错误        |
| 40003  | 查询语法错误        |
| 50001  | Elasticsearch 连接失败 |
| 50002  | 服务内部错误        |
