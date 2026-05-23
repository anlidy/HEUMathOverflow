from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime
from typing import Any, Literal

from pydantic import BaseModel, Field, model_validator


class PostAnswer(BaseModel):
    reply_id: int = Field(alias="reply_id")
    replier_id: int = Field(alias="replier_id")
    content: str


class RagDataPayload(BaseModel):
    post_id: int = Field(alias="post_id")
    author_id: int = Field(alias="author_id")
    title: str
    content: str
    tags: list[str] = Field(default_factory=list)
    answers: list[PostAnswer] = Field(default_factory=list)


class RagEvent(BaseModel):
    type: Literal["rag.data.created", "rag.data.updated", "rag.data.deleted"]
    payload: RagDataPayload
    created_at: datetime = Field(alias="created_at")


class SearchRequest(BaseModel):
    query: str
    post_limit: int | None = None
    global_chunk_limit: int | None = None


class ChunkView(BaseModel):
    point_id: str
    chunk_index: int
    chunk_type: Literal["question", "answer"]
    content: str
    score: float


class AnswerView(BaseModel):
    reply_id: int
    answer_status: str
    answer_rank: int
    score: float
    chunks: list[ChunkView]


class SearchPostView(BaseModel):
    post_id: int
    title: str
    tags: list[str]
    score: float
    question_chunks: list[ChunkView]
    answers: list[AnswerView]


class SearchResponse(BaseModel):
    query: str
    posts: list[SearchPostView]


class ChatMessage(BaseModel):
    message_id: int | None = None
    role: Literal["system", "user", "assistant", "tool"]
    type: Literal["text", "tool_use", "tool_result", "event"] = "text"
    content: str
    tool_name: str | None = None
    tool_call_id: str | None = None
    tool_args: dict[str, Any] | None = None
    tool_result: dict[str, Any] | None = None


class ChatRequest(BaseModel):
    session_id: str | None = None
    summary: str | None = None
    summary_up_to_message_id: int | None = None
    messages: list[ChatMessage]

    @model_validator(mode="after")
    def validate_messages(self) -> "ChatRequest":
        if not self.messages:
            raise ValueError("messages must not be empty")
        if not any(message.role == "user" for message in self.messages):
            raise ValueError("messages must include at least one user message")
        return self


class ToolTrace(BaseModel):
    tool_name: str
    tool_call_id: str
    tool_args: dict[str, Any] = Field(default_factory=dict)
    tool_result: dict[str, Any] = Field(default_factory=dict)
    content: str
    status: Literal["ok", "error"] = "ok"


class SummaryUpdate(BaseModel):
    summary: str
    up_to_message_id: int


class ChatResponse(BaseModel):
    answer: str
    used_tools: list[str] = Field(default_factory=list)
    retrieved_posts: list[SearchPostView] = Field(default_factory=list)
    tool_traces: list[ToolTrace] = Field(default_factory=list)
    summary_update: SummaryUpdate | None = None


@dataclass(slots=True)
class ChunkRecord:
    point_id: str
    doc_type: str
    chunk_type: str
    post_id: int
    reply_id: int | None
    chunk_index: int
    title: str
    tags: list[str]
    content: str
    question_content: str
    answer_status: str
    answer_rank: int
    prev_chunk_id: str | None
    next_chunk_id: str | None
    metadata: dict[str, Any]


@dataclass(slots=True)
class SearchHit:
    point_id: str
    score: float
    payload: dict[str, Any]
