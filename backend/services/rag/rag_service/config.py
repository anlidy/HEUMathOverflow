from __future__ import annotations

import os
from pathlib import Path

import yaml
from pydantic import BaseModel, Field


class ServerConfig(BaseModel):
    host: str = "0.0.0.0"
    port: int = 9091


class RabbitMQConfig(BaseModel):
    host: str
    port: int
    user: str
    password: str
    rag_worker_count: int = Field(default=1, ge=1)
    exchange: str = "app.events"
    exchange_type: str = "topic"
    queue_prefix: str = "rag.data.events"
    dead_letter_exchange: str = "app.events.dlx"
    dead_letter_queue: str = "rag.data.events.dlq"

    @property
    def url(self) -> str:
        return f"amqp://{self.user}:{self.password}@{self.host}:{self.port}/"


class QdrantConfig(BaseModel):
    host: str
    port: int
    collection: str
    dense_vector_name: str = "dense"
    sparse_vector_name: str = "sparse"
    enable_sparse: bool = False
    recreate_on_start: bool = False

    @property
    def url(self) -> str:
        return f"http://{self.host}:{self.port}"


class EmbeddingConfig(BaseModel):
    base_url: str
    api_key: str
    model: str


class SparseConfig(BaseModel):
    namespace_size: int = 1_000_003
    k1: float = 1.2
    b: float = 0.75
    avgdl: float = 120.0


class RerankConfig(BaseModel):
    base_url: str
    api_key: str
    model: str
    timeout_seconds: int = 30


class LLMConfig(BaseModel):
    base_url: str
    api_key: str
    model: str
    temperature: float = 0.0
    timeout_seconds: int = 60
    max_retries: int = 2


class ChunkingConfig(BaseModel):
    question_chunk_size: int = 900
    question_chunk_overlap: int = 120
    answer_chunk_size: int = 800
    answer_chunk_overlap: int = 120


class RetrievalConfig(BaseModel):
    global_chunk_limit: int = 36
    post_limit: int = 6
    answer_doc_limit: int = 2
    answer_chunk_limit: int = 4
    question_chunk_limit: int = 3
    neighbor_expand_limit: int = 1
    min_answer_chunks: int = 2
    min_question_chunks: int = 1


class ContextConfig(BaseModel):
    system_prompt: str = (
        "You are a MathOverflow AI assistant. Answer using the conversation context. "
        "Call tools only when they are needed to retrieve missing facts or forum knowledge."
    )
    max_history_messages: int = Field(default=12, ge=1)
    max_message_chars: int = Field(default=4_000, ge=1)


class ChatConfig(BaseModel):
    max_tool_round_trips: int = Field(default=3, ge=1)
    tool_post_limit: int = Field(default=3, ge=1)
    tool_global_chunk_limit: int = Field(default=12, ge=1)
    context: ContextConfig = Field(default_factory=ContextConfig)


class AppConfig(BaseModel):
    server: ServerConfig
    rabbitmq: RabbitMQConfig
    qdrant: QdrantConfig
    embedding: EmbeddingConfig
    sparse: SparseConfig = SparseConfig()
    rerank: RerankConfig
    llm: LLMConfig
    chunking: ChunkingConfig
    retrieval: RetrievalConfig
    chat: ChatConfig


def load_config(path: str | Path) -> AppConfig:
    with Path(path).open("r", encoding="utf-8") as file:
        data = yaml.safe_load(file) or {}
    llm_config = data.setdefault("llm", {})
    llm_config["base_url"] = os.getenv("RAG_LLM_BASE_URL", llm_config.get("base_url", ""))
    llm_config["api_key"] = os.getenv("RAG_LLM_API_KEY", llm_config.get("api_key", ""))
    return AppConfig.model_validate(data)
