from __future__ import annotations

from collections import defaultdict
from dataclasses import dataclass, field

from rag_service.config import RetrievalConfig
from rag_service.encoders import Reranker
from rag_service.models import AnswerView, ChunkView, SearchHit, SearchPostView, SearchRequest, SearchResponse
from rag_service.qdrant_store import QdrantIndex


@dataclass(slots=True)
class PostGroup:
    post_id: int
    title: str
    tags: list[str]
    question_hits: list[SearchHit] = field(default_factory=list)
    answer_hits: list[SearchHit] = field(default_factory=list)


class RetrievalService:
    def __init__(self, store: QdrantIndex, reranker: Reranker, config: RetrievalConfig) -> None:
        self._store = store
        self._reranker = reranker
        self._config = config

    def search(self, request: SearchRequest) -> SearchResponse:
        global_limit = request.global_chunk_limit or self._config.global_chunk_limit
        post_limit = request.post_limit or self._config.post_limit

        initial_hits = self._store.hybrid_search(
            request.query,
            doc_types=["question_chunk", "answer_chunk"],
            limit=global_limit,
        )
        groups = self._group_hits(initial_hits)
        for group in groups.values():
            if not group.answer_hits and group.question_hits:
                group.answer_hits.extend(self._supplement_answers(request.query, group))
            if not group.question_hits and group.answer_hits:
                group.question_hits.extend(self._supplement_questions(request.query, group))

        ranked_groups = sorted(
            groups.values(),
            key=lambda group: self._score_group(request.query, group),
            reverse=True,
        )[:post_limit]

        return SearchResponse(
            query=request.query,
            posts=[self._build_post_view(request.query, group) for group in ranked_groups],
        )

    def _group_hits(self, hits: list[SearchHit]) -> dict[int, PostGroup]:
        groups: dict[int, PostGroup] = {}
        for hit in hits:
            post_id = int(hit.payload["post_id"])
            group = groups.setdefault(
                post_id,
                PostGroup(
                    post_id=post_id,
                    title=str(hit.payload.get("title", "")),
                    tags=list(hit.payload.get("tags", [])),
                ),
            )
            if hit.payload.get("chunk_type") == "question":
                group.question_hits.append(hit)
            else:
                group.answer_hits.append(hit)
        return groups

    def _supplement_answers(self, query: str, group: PostGroup) -> list[SearchHit]:
        question_context = "\n".join(hit.payload.get("content", "") for hit in group.question_hits[:2])
        search_text = self._build_answer_search_text(query, group.title, question_context)
        answer_docs = self._store.hybrid_search(
            search_text,
            doc_types=["answer_doc"],
            limit=self._config.answer_doc_limit,
            post_id=group.post_id,
        )
        top_reply_ids: list[int] = []
        for hit in answer_docs:
            reply_id = hit.payload.get("reply_id")
            if isinstance(reply_id, int) and reply_id not in top_reply_ids:
                top_reply_ids.append(reply_id)
        if not top_reply_ids:
            return []

        answer_chunks = self._store.hybrid_search(
            search_text,
            doc_types=["answer_chunk"],
            limit=self._config.answer_chunk_limit,
            post_id=group.post_id,
            reply_ids=top_reply_ids,
        )
        return self._expand_answer_neighbors(answer_chunks)

    def _supplement_questions(self, query: str, group: PostGroup) -> list[SearchHit]:
        answer_context = "\n".join(hit.payload.get("content", "") for hit in group.answer_hits[:2])
        search_text = f"User query: {query}\nMatched answer context: {answer_context}"
        question_hits = self._store.hybrid_search(
            search_text,
            doc_types=["question_chunk"],
            limit=self._config.question_chunk_limit,
            post_id=group.post_id,
        )
        if question_hits:
            return self._expand_question_neighbors(group.post_id, question_hits)
        return []

    def _build_answer_search_text(self, query: str, title: str, question_context: str) -> str:
        return "\n".join(
            [
                f"User query: {query}",
                f"Post title: {title}",
                "Matched question context:",
                question_context,
            ]
        )

    def _expand_answer_neighbors(self, hits: list[SearchHit]) -> list[SearchHit]:
        if not hits:
            return []
        by_reply: dict[int, list[SearchHit]] = defaultdict(list)
        for hit in hits:
            reply_id = hit.payload.get("reply_id")
            if isinstance(reply_id, int):
                by_reply[reply_id].append(hit)

        expanded: dict[str, SearchHit] = {hit.point_id: hit for hit in hits}
        for reply_id, reply_hits in by_reply.items():
            best_hit = max(reply_hits, key=lambda item: item.score)
            post_id = int(best_hit.payload["post_id"])
            reply_chunks = self._store.list_reply_chunks(post_id, reply_id)
            for neighbor in self._window_chunks(reply_chunks, best_hit, self._config.neighbor_expand_limit):
                neighbor.score = best_hit.score
                expanded[neighbor.point_id] = neighbor
            if len(expanded) < self._config.min_answer_chunks:
                for neighbor in reply_chunks:
                    if neighbor.point_id in expanded:
                        continue
                    neighbor.score = best_hit.score
                    expanded[neighbor.point_id] = neighbor
                    if len(expanded) >= self._config.min_answer_chunks:
                        break

        expanded_hits = list(expanded.values())
        expanded_hits.sort(key=lambda item: item.score, reverse=True)
        return expanded_hits

    def _expand_question_neighbors(self, post_id: int, hits: list[SearchHit]) -> list[SearchHit]:
        if not hits:
            return []
        all_chunks = self._store.list_question_chunks(post_id)
        expanded: dict[str, SearchHit] = {hit.point_id: hit for hit in hits}
        for hit in hits:
            for neighbor in self._window_chunks(all_chunks, hit, self._config.neighbor_expand_limit):
                neighbor.score = hit.score
                expanded[neighbor.point_id] = neighbor
        result = list(expanded.values())
        result.sort(key=lambda item: item.score, reverse=True)
        return result

    def _window_chunks(self, chunks: list[SearchHit], center_hit: SearchHit, radius: int) -> list[SearchHit]:
        center_index = int(center_hit.payload.get("chunk_index", 0))
        return [
            chunk
            for chunk in chunks
            if abs(int(chunk.payload.get("chunk_index", 0)) - center_index) <= radius
        ]

    def _score_group(self, query: str, group: PostGroup) -> float:
        best_question = max((hit.score for hit in group.question_hits), default=0.0)
        best_answer = max((hit.score for hit in group.answer_hits), default=0.0)
        both_bonus = 0.2 if group.question_hits and group.answer_hits else 0.0
        return best_question + best_answer + both_bonus

    def _build_post_view(self, query: str, group: PostGroup) -> SearchPostView:
        question_hits = self._dedupe_hits(group.question_hits)
        answer_hits = self._dedupe_hits(group.answer_hits)
        question_chunks = [self._to_chunk_view(hit) for hit in question_hits[: self._config.question_chunk_limit]]

        answer_groups: dict[int, list[SearchHit]] = defaultdict(list)
        for hit in answer_hits:
            reply_id = hit.payload.get("reply_id")
            if isinstance(reply_id, int):
                answer_groups[reply_id].append(hit)

        answer_docs: list[tuple[AnswerView, str]] = []
        for reply_id, reply_hits in answer_groups.items():
            reply_hits.sort(key=lambda item: (item.score, -int(item.payload.get("chunk_index", 0))), reverse=True)
            chunks = [self._to_chunk_view(hit) for hit in reply_hits[: self._config.answer_chunk_limit]]
            answer = AnswerView(
                reply_id=reply_id,
                answer_status=str(reply_hits[0].payload.get("answer_status", "")),
                answer_rank=int(reply_hits[0].payload.get("answer_rank", 0)),
                score=max(hit.score for hit in reply_hits),
                chunks=chunks,
            )
            answer_text = "\n".join(chunk.content for chunk in chunks)
            answer_docs.append((answer, answer_text))

        rerank_scores = self._reranker.rerank(query, [doc for _, doc in answer_docs])
        reranked: list[AnswerView] = []
        for index, (answer, _) in enumerate(answer_docs):
            answer.score += rerank_scores[index]
            reranked.append(answer)
        reranked.sort(key=lambda item: (item.score, -item.answer_rank), reverse=True)

        return SearchPostView(
            post_id=group.post_id,
            title=group.title,
            tags=group.tags,
            score=self._score_group(query, group),
            question_chunks=question_chunks,
            answers=reranked[: self._config.answer_doc_limit],
        )

    def _to_chunk_view(self, hit: SearchHit) -> ChunkView:
        return ChunkView(
            point_id=hit.point_id,
            chunk_index=int(hit.payload.get("chunk_index", 0)),
            chunk_type=str(hit.payload.get("chunk_type", "question")),
            content=str(hit.payload.get("content", "")),
            score=float(hit.score),
        )

    def _dedupe_hits(self, hits: list[SearchHit]) -> list[SearchHit]:
        deduped: dict[str, SearchHit] = {}
        for hit in sorted(hits, key=lambda item: item.score, reverse=True):
            deduped.setdefault(hit.point_id, hit)
        return list(deduped.values())
