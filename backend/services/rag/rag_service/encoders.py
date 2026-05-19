from __future__ import annotations

import hashlib
import logging
import re
import time
from collections import Counter

import httpx

from rag_service.config import EmbeddingConfig, RerankConfig, SparseConfig

LOGGER = logging.getLogger(__name__)
TOKEN_PATTERN = re.compile(r"[A-Za-z0-9_]+|[\u4e00-\u9fff]+")


class QwenEmbeddingEncoder:
    def __init__(self, config: EmbeddingConfig) -> None:
        self._base_url = config.base_url.rstrip("/")
        self._api_key = config.api_key
        self._model = config.model
        self._dimension: int | None = None

    @property
    def dimension(self) -> int:
        if self._dimension is None:
            self._dimension = len(self.embed_query("dimension probe"))
        return self._dimension

    def embed_documents(self, texts: list[str]) -> list[list[float]]:
        dense_vectors: list[list[float]] = []
        for text in texts:
            dense_vector = self.embed(text)
            dense_vectors.append(dense_vector)
        return dense_vectors

    def embed_query(self, text: str) -> list[float]:
        dense_vector = self.embed(text, prompt_name="query")
        return dense_vector

    def embed(self, text: str, prompt_name: str | None = None) -> list[float]:
        url = self._base_url.rstrip("/") + "/embeddings"
        payload = {
            "model": self._model,
            "input": text,
        }
        if prompt_name:
            payload["prompt_name"] = prompt_name
        headers = {
            "Authorization": f"Bearer {self._api_key}",
            "Content-Type": "application/json",
        }
        last_error: Exception | None = None
        for attempt in range(10):
            try:
                with httpx.Client(timeout=180) as client:
                    response = client.post(url, json=payload, headers=headers)
                    response.raise_for_status()
                break
            except httpx.HTTPError as exc:
                last_error = exc
                if attempt == 9:
                    raise
                time.sleep(3)
        else:  # pragma: no cover
            raise last_error or RuntimeError("embedding request failed")

        body = response.json()
        item = (body.get("data") or [{}])[0]
        dense = item.get("embedding") or item.get("dense_embedding") or []
        if self._dimension is None and dense:
            self._dimension = len(dense)
        return dense


class BM25SparseEncoder:
    def __init__(self, config: SparseConfig) -> None:
        self._namespace_size = config.namespace_size
        self._k1 = config.k1
        self._b = config.b
        self._avgdl = max(config.avgdl, 1.0)

    def encode_document(self, text: str) -> tuple[list[int], list[float]]:
        return self._encode(text, normalize_length=True)

    def encode_query(self, text: str) -> tuple[list[int], list[float]]:
        return self._encode(text, normalize_length=False)

    def _encode(self, text: str, *, normalize_length: bool) -> tuple[list[int], list[float]]:
        tokens = TOKEN_PATTERN.findall(text.lower())
        if not tokens:
            return [], []

        counts = Counter(tokens)
        doc_len = len(tokens)
        indices: list[int] = []
        values: list[float] = []
        length_norm = 1.0 - self._b + self._b * (doc_len / self._avgdl) if normalize_length else 1.0
        for token, tf in sorted(counts.items()):
            indices.append(self._hash_token(token))
            values.append((tf * (self._k1 + 1.0)) / (tf + self._k1 * length_norm))
        return indices, values

    def _hash_token(self, token: str) -> int:
        digest = hashlib.blake2b(token.encode("utf-8"), digest_size=8).digest()
        return int.from_bytes(digest, "big") % self._namespace_size


class Reranker:
    def __init__(self, config: RerankConfig) -> None:
        self._config = config

    def rerank(self, query: str, documents: list[str]) -> list[float]:
        if not documents:
            return []

        url = self._config.base_url.rstrip("/") + "/rerank"
        payload = {
            "model": self._config.model,
            "query": query,
            "documents": documents,
            "top_n": len(documents),
        }
        headers = {
            "Authorization": f"Bearer {self._config.api_key}",
            "Content-Type": "application/json",
        }
        try:
            with httpx.Client(timeout=self._config.timeout_seconds) as client:
                response = client.post(url, json=payload, headers=headers)
                response.raise_for_status()
            body = response.json()
            results = body.get("results") or body.get("data") or []
            if not results:
                return [0.0] * len(documents)
            scores = [0.0] * len(documents)
            for item in results:
                index = item.get("index")
                score = item.get("relevance_score", item.get("score", 0.0))
                if isinstance(index, int) and 0 <= index < len(scores):
                    scores[index] = float(score)
            return scores
        except Exception as exc:  # noqa: BLE001
            LOGGER.warning("rerank request failed, fallback to zero scores: %s", exc)
            return [0.0] * len(documents)
