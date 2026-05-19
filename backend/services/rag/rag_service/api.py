from __future__ import annotations

from fastapi import FastAPI

from rag_service.chat import ChatService
from rag_service.models import ChatRequest, ChatResponse


def build_app(chat_service: ChatService) -> FastAPI:
    app = FastAPI(title="rag-service", version="0.1.0")

    @app.get("/healthz")
    async def healthz() -> dict[str, str]:
        return {"status": "ok"}

    @app.post("/chat", response_model=ChatResponse)
    async def chat(request: ChatRequest) -> ChatResponse:
        return chat_service.chat(request)

    return app
