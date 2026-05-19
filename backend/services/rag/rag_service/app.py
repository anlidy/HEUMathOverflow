from __future__ import annotations

import logging
from pathlib import Path

import uvicorn

from rag_service.api import build_app
from rag_service.chat import ChatService, RetrievalTool, build_chat_model
from rag_service.chunking import ChunkBuilder
from rag_service.context import ConversationContextManager
from rag_service.config import AppConfig, load_config
from rag_service.encoders import BM25SparseEncoder, QwenEmbeddingEncoder, Reranker
from rag_service.qdrant_store import QdrantIndex
from rag_service.consumer import RagEventConsumer
from rag_service.retrieval import RetrievalService


def create_runtime(config_path: str | Path = "config.yaml") -> tuple[AppConfig, RagEventConsumer, object]:
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(name)s %(message)s",
    )

    config = load_config(config_path)
    encoder = QwenEmbeddingEncoder(config.embedding)
    sparse_encoder = BM25SparseEncoder(config.sparse)
    reranker = Reranker(config.rerank)
    store = QdrantIndex(config.qdrant, encoder, sparse_encoder)
    store.ensure_collection()

    chunk_builder = ChunkBuilder(config.chunking)
    consumer = RagEventConsumer(config.rabbitmq, store, chunk_builder)
    retrieval = RetrievalService(store, reranker, config.retrieval)
    chat_model = build_chat_model(config.llm)
    context_manager = ConversationContextManager(config.chat.context)
    retrieval_tool = RetrievalTool(retrieval, config.chat)
    chat_service = ChatService(chat_model, context_manager, retrieval_tool, config.chat)
    app = build_app(chat_service)
    return config, consumer, app


def main() -> None:
    config, consumer, app = create_runtime()
    consumer.start()
    uvicorn.run(app, host=config.server.host, port=config.server.port)
