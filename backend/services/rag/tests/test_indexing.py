from __future__ import annotations

import unittest

from rag_service.qdrant_store import DENSE_ANSWER_QUESTION_FALLBACK_THRESHOLD, QdrantIndex
from rag_service.models import ChunkRecord


def make_record(*, doc_type: str, chunk_type: str, content: str, question_content: str = "Question body", title: str = "Title", tags: list[str] | None = None) -> ChunkRecord:
    return ChunkRecord(
        point_id="p1",
        doc_type=doc_type,
        chunk_type=chunk_type,
        post_id=1,
        reply_id=2 if chunk_type == "answer" else None,
        chunk_index=0,
        title=title,
        tags=tags or ["tag1", "tag2"],
        content=content,
        question_content=question_content,
        answer_status="selected" if chunk_type == "answer" else "",
        answer_rank=1 if chunk_type == "answer" else 0,
        prev_chunk_id=None,
        next_chunk_id=None,
        metadata={"author_id": 1, "is_certified_post": True},
    )


class IndexTextTests(unittest.TestCase):
    def setUp(self) -> None:
        self.store = object.__new__(QdrantIndex)

    def test_question_chunk_dense_and_sparse_match_question_content(self) -> None:
        record = make_record(doc_type="question_chunk", chunk_type="question", content="Question chunk text")
        dense_text = self.store._build_dense_index_text(record)
        sparse_text = self.store._build_sparse_index_text(record)
        self.assertIn("Question:\nQuestion chunk text", dense_text)
        self.assertEqual(dense_text, sparse_text)

    def test_answer_doc_dense_and_sparse_keep_full_question(self) -> None:
        record = make_record(doc_type="answer_doc", chunk_type="answer", content="Full answer text")
        dense_text = self.store._build_dense_index_text(record)
        sparse_text = self.store._build_sparse_index_text(record)
        self.assertIn("Question Context:\nQuestion body", dense_text)
        self.assertIn("Question Context:\nQuestion body", sparse_text)

    def test_short_answer_chunk_dense_falls_back_to_question(self) -> None:
        short_content = "x" * (DENSE_ANSWER_QUESTION_FALLBACK_THRESHOLD - 1)
        record = make_record(doc_type="answer_chunk", chunk_type="answer", content=short_content)
        dense_text = self.store._build_dense_index_text(record)
        self.assertIn("Question Context:\nQuestion body", dense_text)

    def test_long_answer_chunk_dense_omits_question_but_sparse_keeps_it(self) -> None:
        long_content = "x" * DENSE_ANSWER_QUESTION_FALLBACK_THRESHOLD
        record = make_record(doc_type="answer_chunk", chunk_type="answer", content=long_content)
        dense_text = self.store._build_dense_index_text(record)
        sparse_text = self.store._build_sparse_index_text(record)
        self.assertNotIn("Question Context:\nQuestion body", dense_text)
        self.assertIn("Question Context:\nQuestion body", sparse_text)
        self.assertIn(f"Answer:\n{long_content}", dense_text)
        self.assertIn(f"Answer:\n{long_content}", sparse_text)


if __name__ == "__main__":
    unittest.main()
