import os
from flask import Flask, request, jsonify
from elasticsearch import Elasticsearch

app = Flask(__name__)


def get_es_client():
    # Bug fix 3: 支持 ES_ADDR 完整地址（优先级高于 ES_HOST/ES_PORT）
    es_addr = os.environ.get("ES_ADDR", "").strip()
    if es_addr:
        return Elasticsearch([es_addr])
    es_host = os.environ.get("ES_HOST", "localhost")
    es_port = int(os.environ.get("ES_PORT", 9200))
    return Elasticsearch([f"http://{es_host}:{es_port}"])


# 延迟初始化，避免启动时 ES 尚未就绪导致连接失败
es_client = get_es_client()


@app.route("/health", methods=['GET'])
def health_check():
    """Health check endpoint for Kubernetes probes."""
    try:
        es_connected = es_client.ping()
    except Exception:
        es_connected = False
    return jsonify({"status": "ok", "es_connected": es_connected})


@app.route("/api/v1/debug/headers", methods=['GET'])
def debug_headers():
    """Returns all request headers for debugging."""
    return jsonify({"headers": dict(request.headers)})


@app.route("/api/v1/indices", methods=['GET'])
def get_indices():
    """Lists available Elasticsearch indices."""
    try:
        indices = [index['index'] for index in es_client.cat.indices(format="json")]
        return jsonify({"code": 0, "message": "success", "data": {"indices": indices}})
    except Exception as e:
        return jsonify({"code": 50001, "message": f"Failed to get indices: {str(e)}", "data": None}), 500


@app.route("/api/v1/search", methods=['POST'])
def search():
    """Performs a search query in Elasticsearch."""
    req_data = request.get_json()

    if not req_data or "index" not in req_data or "query" not in req_data:
        return jsonify({"code": 40002, "message": "index and query are required fields", "data": None}), 400

    index = req_data["index"]
    query_text = req_data["query"]
    size = int(req_data.get("size", 10))
    from_ = int(req_data.get("from", 0))

    if size > 100:
        return jsonify({"code": 40002, "message": "size cannot be greater than 100", "data": None}), 400

    es_query = {
        "query": {
            "multi_match": {
                "query": query_text,
                "fields": ["*"]
            }
        }
    }

    # Bug fix 2: 实现 filters 参数（按字段值过滤）
    filters = req_data.get("filters")
    if filters and isinstance(filters, dict):
        field = filters.get("field")
        value = filters.get("value")
        if field and value is not None:
            es_query = {
                "query": {
                    "bool": {
                        "must": es_query["query"],
                        "filter": [{"term": {field: value}}]
                    }
                }
            }

    try:
        if not es_client.indices.exists(index=index):
            return jsonify({"code": 40001, "message": f"index '{index}' not found", "data": None}), 404

        res = es_client.search(index=index, body=es_query, size=size, from_=from_)

        hits = [
            {"_id": hit['_id'], "_score": hit['_score'], "_source": hit['_source']}
            for hit in res["hits"]["hits"]
        ]

        response_data = {
            "total": res["hits"]["total"]["value"],
            "hits": hits,
            "from": from_,
            # Bug fix 4: 返回请求的 size，而非实际命中数
            "size": size,
        }
        return jsonify({"code": 0, "message": "success", "data": response_data})

    except Exception as e:
        return jsonify({"code": 50002, "message": f"ES search request failed: {str(e)}", "data": None}), 500


if __name__ == '__main__':
    app.run(host="0.0.0.0", port=int(os.environ.get("PORT", 8080)))
