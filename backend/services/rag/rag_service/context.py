from __future__ import annotations

from langchain_core.messages import AIMessage, BaseMessage, HumanMessage, SystemMessage

from rag_service.config import ContextConfig
from rag_service.models import ChatMessage, ChatRequest


class ConversationContextManager:
    def __init__(self, config: ContextConfig) -> None:
        self._config = config

    def build_messages(self, request: ChatRequest) -> list[BaseMessage]:
        messages: list[BaseMessage] = [SystemMessage(content=self._config.system_prompt)]
        history = request.messages[-self._config.max_history_messages :]
        for item in history:
            content = self._trim_content(item.content)
            if not content:
                continue
            messages.append(self._to_langchain_message(ChatMessage(role=item.role, content=content)))
        return messages

    def _trim_content(self, content: str) -> str:
        return content.strip()[: self._config.max_message_chars]

    def _to_langchain_message(self, message: ChatMessage) -> BaseMessage:
        if message.role == "system":
            return SystemMessage(content=message.content)
        if message.role == "assistant":
            return AIMessage(content=message.content)
        return HumanMessage(content=message.content)
