from langchain_core.tools import tool

from app.repositories.chat_session_repository import ChatSessionRepository


def build_tools(chat_repository: ChatSessionRepository):
    @tool
    async def get_recent_conversation(session_id: str, limit: int = 6) -> str:
        """Return recent conversation context for the current chat session."""

        rows = await chat_repository.list_recent_messages(session_id=session_id, limit=limit)
        if not rows:
            return "No prior conversation found for this session."
        return "\n".join(f"{row['role']}: {row['content']}" for row in rows)

    return [get_recent_conversation]
