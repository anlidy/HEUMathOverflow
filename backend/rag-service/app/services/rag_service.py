from hashlib import sha256

from langchain_core.messages import AIMessage, HumanMessage
from redis.asyncio import Redis

from app.core.config import Settings
from app.repositories.chat_session_repository import ChatSessionRepository
from app.schemas.chat import ChatRequest, ChatResponse


class RagService:
    def __init__(
        self,
        settings: Settings,
        redis_client: Redis,
        chat_repository: ChatSessionRepository,
        agent_executor,
    ):
        self._settings = settings
        self._redis = redis_client
        self._chat_repository = chat_repository
        self._agent_executor = agent_executor

    async def ask(self, payload: ChatRequest) -> ChatResponse:
        cache_key = self._build_cache_key(payload.session_id, payload.question)
        cached_answer = await self._redis.get(cache_key)
        if cached_answer:
            return ChatResponse(session_id=payload.session_id, answer=cached_answer, cached=True)

        history = await self._chat_repository.list_recent_messages(
            session_id=payload.session_id,
            limit=self._settings.chat_history_limit,
        )
        chat_history = [self._to_message(item["role"], item["content"]) for item in history]

        result = await self._agent_executor.ainvoke(
            {
                "input": payload.question,
                "chat_history": chat_history,
            }
        )
        answer = result["output"]

        await self._chat_repository.save_message(payload.session_id, "user", payload.question)
        await self._chat_repository.save_message(payload.session_id, "assistant", answer)
        await self._redis.set(cache_key, answer, ex=self._settings.cache_ttl_seconds)

        return ChatResponse(session_id=payload.session_id, answer=answer, cached=False)

    @staticmethod
    def _build_cache_key(session_id: str, question: str) -> str:
        digest = sha256(question.encode("utf-8")).hexdigest()
        return f"rag:answer:{session_id}:{digest}"

    @staticmethod
    def _to_message(role: str, content: str):
        if role == "assistant":
            return AIMessage(content=content)
        return HumanMessage(content=content)
