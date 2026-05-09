from collections.abc import Sequence

from sqlalchemy import text
from sqlalchemy.ext.asyncio import async_sessionmaker


class ChatSessionRepository:
    def __init__(self, session_factory: async_sessionmaker):
        self._session_factory = session_factory

    async def init_schema(self) -> None:
        query = text(
            """
            CREATE TABLE IF NOT EXISTS rag_chat_messages (
                id BIGSERIAL PRIMARY KEY,
                session_id VARCHAR(128) NOT NULL,
                role VARCHAR(32) NOT NULL,
                content TEXT NOT NULL,
                created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
            );
            CREATE INDEX IF NOT EXISTS idx_rag_chat_messages_session_created_at
            ON rag_chat_messages (session_id, created_at DESC);
            """
        )
        async with self._session_factory() as session:
            await session.execute(query)
            await session.commit()

    async def save_message(self, session_id: str, role: str, content: str) -> None:
        query = text(
            """
            INSERT INTO rag_chat_messages (session_id, role, content)
            VALUES (:session_id, :role, :content)
            """
        )
        async with self._session_factory() as session:
            await session.execute(query, {"session_id": session_id, "role": role, "content": content})
            await session.commit()

    async def list_recent_messages(self, session_id: str, limit: int) -> Sequence[dict]:
        query = text(
            """
            SELECT role, content, created_at
            FROM rag_chat_messages
            WHERE session_id = :session_id
            ORDER BY created_at DESC
            LIMIT :limit
            """
        )
        async with self._session_factory() as session:
            result = await session.execute(query, {"session_id": session_id, "limit": limit})
            rows = result.mappings().all()
        return list(reversed(rows))
