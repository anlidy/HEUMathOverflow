from __future__ import annotations

import hashlib
from collections.abc import Iterable
from typing import Any

from qdrant_client import QdrantClient, models

from rag_service.config import QdrantConfig
from rag_service.encoders import BM25SparseEncoder, QwenEmbeddingEncoder
from rag_service.models import ChunkRecord, SearchHit

DENSE_ANSWER_QUESTION_FALLBACK_THRESHOLD = 24


class QdrantIndex:
    def __init__(self, config: QdrantConfig, encoder: QwenEmbeddingEncoder, sparse_encoder: BM25SparseEncoder) -> None:
        self._config = config
        self._encoder = encoder
        self._sparse_encoder = sparse_encoder
        self._client = QdrantClient(url=config.url)

    @property
    def client(self) -> QdrantClient:
        return self._client

    def ensure_collection(self) -> None:
        exists = self._client.collection_exists(self._config.collection)
        if exists and self._config.recreate_on_start:
            self._client.delete_collection(self._config.collection)
            exists = False
        if not exists:
            create_kwargs = {
                "collection_name": self._config.collection,
                "vectors_config": {
                    self._config.dense_vector_name: models.VectorParams(
                        size=self._encoder.dimension,
                        distance=models.Distance.COSINE,
                    )
                },
            }
            if self._config.enable_sparse:
                create_kwargs["sparse_vectors_config"] = {
                    self._config.sparse_vector_name: models.SparseVectorParams(modifier=models.Modifier.IDF)
                }
            self._client.create_collection(**create_kwargs)
        self._ensure_payload_indexes()

    def replace_post(self, records: list[ChunkRecord]) -> None:
        if not records:
            return
        self.delete_post(records[0].post_id)

        dense_texts = [self._build_dense_index_text(record) for record in records]
        sparse_texts = [self._build_sparse_index_text(record) for record in records]
        dense_vectors = [self._encoder.embed(text) for text in dense_texts]
        sparse_vectors = [self._sparse_encoder.encode_document(text) for text in sparse_texts]
        points: list[models.PointStruct] = []
        for record, dense_vector, (indices, values) in zip(records, dense_vectors, sparse_vectors, strict=True):
            payload = self._build_payload(record)
            vector: dict[str, list[float] | models.SparseVector] = {
                self._config.dense_vector_name: dense_vector,
            }
            if self._config.enable_sparse and indices:
                vector[self._config.sparse_vector_name] = models.SparseVector(indices=indices, values=values)
            points.append(
                models.PointStruct(
                    id=self._point_id_value(record.point_id),
                    vector=vector,
                    payload=payload,
                )
            )
        self._client.upsert(collection_name=self._config.collection, wait=True, points=points)

    def delete_post(self, post_id: int) -> None:
        self._client.delete(
            collection_name=self._config.collection,
            wait=True,
            points_selector=models.FilterSelector(
                filter=models.Filter(
                    must=[
                        models.FieldCondition(
                            key="post_id",
                            match=models.MatchValue(value=post_id),
                        )
                    ]
                )
            ),
        )

    def hybrid_search(
        self,
        query_text: str,
        *,
        doc_types: list[str],
        limit: int,
        post_id: int | None = None,
        reply_ids: list[int] | None = None,
    ) -> list[SearchHit]:
        dense_vector = self._encoder.embed_query(query_text)
        sparse_indices, sparse_values = self._sparse_encoder.encode_query(query_text)
        query_filter = self._build_filter(doc_types=doc_types, post_id=post_id, reply_ids=reply_ids)
        sparse_query = None
        if self._config.enable_sparse and sparse_indices:
            sparse_query = models.SparseVector(indices=sparse_indices, values=sparse_values)

        prefetch: list[models.Prefetch] = [
            models.Prefetch(
                query=dense_vector,
                using=self._config.dense_vector_name,
                filter=query_filter,
                limit=limit,
            )
        ]
        if sparse_query is not None:
            prefetch.append(
                models.Prefetch(
                    query=sparse_query,
                    using=self._config.sparse_vector_name,
                    filter=query_filter,
                    limit=limit,
                )
            )

        response = self._client.query_points(
            collection_name=self._config.collection,
            prefetch=prefetch,
            query=models.FusionQuery(fusion=models.Fusion.RRF),
            with_payload=True,
            limit=limit,
        )
        points = getattr(response, "points", response)
        return [
            SearchHit(
                point_id=str(point.id),
                score=float(point.score),
                payload=dict(point.payload or {}),
            )
            for point in points
        ]

    def get_points(self, point_ids: Iterable[str]) -> dict[str, SearchHit]:
        ids = list(dict.fromkeys(point_ids))
        if not ids:
            return {}
        records = self._client.retrieve(
            collection_name=self._config.collection,
            ids=ids,
            with_payload=True,
        )
        return {
            str(record.id): SearchHit(
                point_id=str(record.id),
                score=0.0,
                payload=dict(record.payload or {}),
            )
            for record in records
        }

    def list_reply_chunks(self, post_id: int, reply_id: int) -> list[SearchHit]:
        records, _ = self._client.scroll(
            collection_name=self._config.collection,
            scroll_filter=self._build_filter(
                doc_types=["answer_chunk"],
                post_id=post_id,
                reply_ids=[reply_id],
            ),
            with_payload=True,
            limit=64,
        )
        hits = [
            SearchHit(
                point_id=str(record.id),
                score=0.0,
                payload=dict(record.payload or {}),
            )
            for record in records
        ]
        hits.sort(key=lambda hit: int(hit.payload.get("chunk_index", 0)))
        return hits

    def list_question_chunks(self, post_id: int) -> list[SearchHit]:
        records, _ = self._client.scroll(
            collection_name=self._config.collection,
            scroll_filter=self._build_filter(doc_types=["question_chunk"], post_id=post_id),
            with_payload=True,
            limit=64,
        )
        hits = [
            SearchHit(
                point_id=str(record.id),
                score=0.0,
                payload=dict(record.payload or {}),
            )
            for record in records
        ]
        hits.sort(key=lambda hit: int(hit.payload.get("chunk_index", 0)))
        return hits

    def _build_payload(self, record: ChunkRecord) -> dict[str, Any]:
        return {
            "doc_type": record.doc_type,
            "chunk_type": record.chunk_type,
            "post_id": record.post_id,
            "reply_id": record.reply_id,
            "chunk_index": record.chunk_index,
            "title": record.title,
            "tags": record.tags,
            "content": record.content,
            "answer_status": record.answer_status,
            "answer_rank": record.answer_rank,
            "prev_chunk_id": record.prev_chunk_id,
            "next_chunk_id": record.next_chunk_id,
            **record.metadata,
        }

    def _build_dense_index_text(self, record: ChunkRecord) -> str:
        lines = [f"Title: {record.title}"]
        if record.tags:
            lines.append(f"Tags: {', '.join(record.tags)}")
        if record.chunk_type == "question":
            lines.append("Question:")
            lines.append(record.content)
            return "\n".join(lines)
        if record.doc_type == "answer_doc":
            lines.append("Question Context:")
            lines.append(record.question_content)
            lines.append("Answer:")
            lines.append(record.content)
            return "\n".join(lines)
        if len(record.content) < DENSE_ANSWER_QUESTION_FALLBACK_THRESHOLD:
            lines.append("Question Context:")
            lines.append(record.question_content)
        lines.append("Answer:")
        lines.append(record.content)
        return "\n".join(lines)

    def _build_sparse_index_text(self, record: ChunkRecord) -> str:
        lines = [f"Title: {record.title}"]
        if record.tags:
            lines.append(f"Tags: {', '.join(record.tags)}")
        if record.chunk_type == "question":
            lines.append("Question:")
            lines.append(record.content)
        else:
            lines.append("Question Context:")
            lines.append(record.question_content)
            lines.append("Answer:")
            lines.append(record.content)
        return "\n".join(lines)

    def _build_filter(
        self,
        *,
        doc_types: list[str],
        post_id: int | None = None,
        reply_ids: list[int] | None = None,
    ) -> models.Filter:
        must: list[models.Condition] = [
            models.FieldCondition(
                key="doc_type",
                match=models.MatchAny(any=doc_types),
            )
        ]
        if post_id is not None:
            must.append(models.FieldCondition(key="post_id", match=models.MatchValue(value=post_id)))
        if reply_ids:
            must.append(models.FieldCondition(key="reply_id", match=models.MatchAny(any=reply_ids)))
        return models.Filter(must=must)

    def _ensure_payload_indexes(self) -> None:
        fields = {
            "doc_type": models.PayloadSchemaType.KEYWORD,
            "chunk_type": models.PayloadSchemaType.KEYWORD,
            "post_id": models.PayloadSchemaType.INTEGER,
            "reply_id": models.PayloadSchemaType.INTEGER,
            "answer_status": models.PayloadSchemaType.KEYWORD,
            "is_certified_post": models.PayloadSchemaType.BOOL,
        }
        for field, schema in fields.items():
            self._client.create_payload_index(
                collection_name=self._config.collection,
                field_name=field,
                field_schema=schema,
                wait=True,
            )

    def _point_id_value(self, point_id: str) -> int:
        digest = hashlib.blake2b(point_id.encode("utf-8"), digest_size=8).digest()
        return int.from_bytes(digest, "big", signed=False)
