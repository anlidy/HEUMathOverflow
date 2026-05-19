from __future__ import annotations

import json
import logging
import math
import os
import re
from functools import lru_cache
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse

from sentence_transformers import CrossEncoder, SentenceTransformer

LOGGER = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s %(message)s")

EMBEDDING_MODEL = os.environ.get("EMBEDDING_MODEL_PATH", "/app/models/Qwen3-Embedding-4B")
RERANK_MODEL = os.environ.get("RERANK_MODEL_PATH", "/app/models/Qwen3-Reranker-4B")
EMBEDDING_DIMENSION = 2560
TOKEN_PATTERN = re.compile(r"[A-Za-z0-9_]+|[\u4e00-\u9fff]+")


@lru_cache(maxsize=1)
def get_embedder() -> SentenceTransformer | None:
    try:
        return SentenceTransformer(EMBEDDING_MODEL, trust_remote_code=True)
    except Exception as exc:  # noqa: BLE001
        LOGGER.warning("failed to load embedding model, fallback active: %s", exc)
        return None


@lru_cache(maxsize=1)
def get_reranker() -> CrossEncoder | None:
    try:
        return CrossEncoder(RERANK_MODEL, trust_remote_code=True)
    except Exception as exc:  # noqa: BLE001
        LOGGER.warning("failed to load reranker model, fallback active: %s", exc)
        return None


def _normalize_inputs(value: str | list[str]) -> list[str]:
    return [value] if isinstance(value, str) else value


def _fallback_embedding(text: str) -> list[float]:
    vector = [0.0] * EMBEDDING_DIMENSION
    tokens = TOKEN_PATTERN.findall(text.lower())
    if not tokens:
        return vector
    for token in tokens:
        index = hash(token) % EMBEDDING_DIMENSION
        vector[index] += 1.0
    norm = math.sqrt(sum(value * value for value in vector))
    if norm > 0:
        vector = [value / norm for value in vector]
    return vector


def _fallback_rerank_score(query: str, document: str) -> float:
    query_tokens = set(TOKEN_PATTERN.findall(query.lower()))
    doc_tokens = TOKEN_PATTERN.findall(document.lower())
    if not query_tokens or not doc_tokens:
        return 0.0
    overlap = sum(1 for token in doc_tokens if token in query_tokens)
    return overlap / len(query_tokens)


def _json_response(handler: BaseHTTPRequestHandler, status: int, payload: dict) -> None:
    data = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    handler.send_response(status)
    handler.send_header("Content-Type", "application/json")
    handler.send_header("Content-Length", str(len(data)))
    handler.end_headers()
    handler.wfile.write(data)


class RagModelHandler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:  # noqa: N802
        if urlparse(self.path).path == "/healthz":
            _json_response(self, 200, {"status": "ok"})
            return
        _json_response(self, 404, {"error": "not found"})

    def do_POST(self) -> None:  # noqa: N802
        path = urlparse(self.path).path
        content_length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(content_length) if content_length > 0 else b"{}"
        try:
            body = json.loads(raw)
        except Exception:
            _json_response(self, 400, {"error": "invalid json"})
            return

        if path == "/v1/embeddings":
            self._handle_embeddings(body)
            return
        if path == "/v1/rerank":
            self._handle_rerank(body)
            return
        _json_response(self, 404, {"error": "not found"})

    def _handle_embeddings(self, body: dict) -> None:
        texts = _normalize_inputs(body.get("input", ""))
        prompt_name = body.get("prompt_name")
        model = get_embedder()
        if model is not None:
            encode_kwargs = {
                "normalize_embeddings": True,
                "convert_to_numpy": True,
            }
            if isinstance(prompt_name, str) and prompt_name:
                encode_kwargs["prompt_name"] = prompt_name
            dense_vectors = model.encode(texts, **encode_kwargs).tolist()
        else:
            dense_vectors = [_fallback_embedding(text) for text in texts]
        data = [
            {
                "index": index,
                "embedding": vector,
            }
            for index, vector in enumerate(dense_vectors)
        ]
        _json_response(self, 200, {"object": "list", "model": body.get("model", EMBEDDING_MODEL), "data": data})

    def _handle_rerank(self, body: dict) -> None:
        query = str(body.get("query", ""))
        documents = list(body.get("documents", []))
        top_n = body.get("top_n")
        reranker = get_reranker()
        if reranker is not None:
            pairs = [[query, doc] for doc in documents]
            scores = reranker.predict(pairs).tolist()
        else:
            scores = [_fallback_rerank_score(query, doc) for doc in documents]
        results = [
            {"index": index, "relevance_score": float(score)}
            for index, score in enumerate(scores)
        ]
        results.sort(key=lambda item: item["relevance_score"], reverse=True)
        if isinstance(top_n, int):
            results = results[:top_n]
        _json_response(self, 200, {"object": "list", "model": body.get("model", RERANK_MODEL), "results": results})

    def log_message(self, format: str, *args: object) -> None:  # noqa: A003
        LOGGER.info("%s - %s", self.address_string(), format % args)


def main() -> None:
    server = ThreadingHTTPServer(("0.0.0.0", 8000), RagModelHandler)
    LOGGER.info("rag-model-server listening on 0.0.0.0:8000")
    server.serve_forever()


if __name__ == "__main__":
    main()
