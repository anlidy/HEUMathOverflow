from fastapi import APIRouter, Depends, Request

from app.schemas.chat import ChatRequest, ChatResponse
from app.services.rag_service import RagService

router = APIRouter()


def get_rag_service(request: Request) -> RagService:
    return request.app.state.container.rag_service


@router.post("/ask", response_model=ChatResponse)
async def ask_question(payload: ChatRequest, service: RagService = Depends(get_rag_service)) -> ChatResponse:
    return await service.ask(payload)
