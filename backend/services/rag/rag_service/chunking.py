from __future__ import annotations

from langchain_core.documents import Document
from langchain_text_splitters import RecursiveCharacterTextSplitter

from rag_service.config import ChunkingConfig
from rag_service.models import ChunkRecord, RagDataPayload


class ChunkBuilder:
    def __init__(self, config: ChunkingConfig) -> None:
        self._config = config
        separators = ["\n\n", "\n", "。", ". ", "!", "?", "；", ";", "，", ",", " "]
        self._question_splitter = RecursiveCharacterTextSplitter(
            chunk_size=config.question_chunk_size,
            chunk_overlap=config.question_chunk_overlap,
            separators=separators,
        )
        self._answer_splitter = RecursiveCharacterTextSplitter(
            chunk_size=config.answer_chunk_size,
            chunk_overlap=config.answer_chunk_overlap,
            separators=separators,
        )

    def build(self, payload: RagDataPayload) -> list[ChunkRecord]:
        question_chunks = self._build_question_chunks(payload)

        answer_records: list[ChunkRecord] = []
        for rank, answer in enumerate(payload.answers, start=1):
            answer_status = "selected"
            answer_records.extend(
                self._build_answer_records(
                    payload=payload,
                    answer_text=answer.content,
                    reply_id=answer.reply_id,
                    answer_rank=rank,
                    answer_status=answer_status,
                )
            )

        return question_chunks + answer_records

    def _build_question_chunks(self, payload: RagDataPayload) -> list[ChunkRecord]:
        chunks = self._split(self._question_splitter, payload.content)
        records: list[ChunkRecord] = []
        for index, chunk in enumerate(chunks):
            point_id = f"post:{payload.post_id}:question:{index}"
            records.append(
                ChunkRecord(
                    point_id=point_id,
                    doc_type="question_chunk",
                    chunk_type="question",
                    post_id=payload.post_id,
                    reply_id=None,
                    chunk_index=index,
                    title=payload.title,
                    tags=payload.tags,
                    content=chunk,
                    question_content=payload.content,
                    answer_status="",
                    answer_rank=0,
                    prev_chunk_id=f"post:{payload.post_id}:question:{index - 1}" if index > 0 else None,
                    next_chunk_id=f"post:{payload.post_id}:question:{index + 1}" if index + 1 < len(chunks) else None,
                    metadata={
                        "author_id": payload.author_id,
                        "is_certified_post": True,
                    },
                )
            )
        return records

    def _build_answer_records(
        self,
        payload: RagDataPayload,
        answer_text: str,
        reply_id: int,
        answer_rank: int,
        answer_status: str,
    ) -> list[ChunkRecord]:
        answer_chunks = self._split(self._answer_splitter, answer_text)
        records: list[ChunkRecord] = [
            ChunkRecord(
                point_id=f"post:{payload.post_id}:reply:{reply_id}:doc",
                doc_type="answer_doc",
                chunk_type="answer",
                post_id=payload.post_id,
                reply_id=reply_id,
                chunk_index=0,
                title=payload.title,
                tags=payload.tags,
                content=answer_text,
                question_content=payload.content,
                answer_status=answer_status,
                answer_rank=answer_rank,
                prev_chunk_id=None,
                next_chunk_id=None,
                metadata={
                    "author_id": payload.author_id,
                    "is_certified_post": True,
                },
            )
        ]

        for index, chunk in enumerate(answer_chunks):
            records.append(
                ChunkRecord(
                    point_id=f"post:{payload.post_id}:reply:{reply_id}:chunk:{index}",
                    doc_type="answer_chunk",
                    chunk_type="answer",
                    post_id=payload.post_id,
                    reply_id=reply_id,
                    chunk_index=index,
                    title=payload.title,
                    tags=payload.tags,
                    content=chunk,
                    question_content=payload.content,
                    answer_status=answer_status,
                    answer_rank=answer_rank,
                    prev_chunk_id=f"post:{payload.post_id}:reply:{reply_id}:chunk:{index - 1}" if index > 0 else None,
                    next_chunk_id=f"post:{payload.post_id}:reply:{reply_id}:chunk:{index + 1}" if index + 1 < len(answer_chunks) else None,
                    metadata={
                        "author_id": payload.author_id,
                        "is_certified_post": True,
                    },
                )
            )
        return records

    def _split(self, splitter: RecursiveCharacterTextSplitter, text: str) -> list[str]:
        source = [Document(page_content=text.strip())] if text.strip() else [Document(page_content="")]
        chunks = [doc.page_content.strip() for doc in splitter.split_documents(source)]
        return [chunk for chunk in chunks if chunk] or [text.strip() or ""]
