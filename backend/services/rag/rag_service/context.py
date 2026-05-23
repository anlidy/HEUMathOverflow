from __future__ import annotations

import json

from langchain_core.messages import AIMessage, BaseMessage, HumanMessage, SystemMessage, ToolMessage

from rag_service.config import ContextConfig
from rag_service.models import ChatMessage, ChatRequest


class ConversationContextManager:
    def __init__(self, config: ContextConfig) -> None:
        self._config = config

    def build_messages(self, request: ChatRequest) -> list[BaseMessage]:
        messages: list[BaseMessage] = [SystemMessage(content=self._config.system_prompt)]
        if request.summary:
            messages.append(SystemMessage(content=f"Conversation summary:\n{self._trim_content(request.summary)}"))
        history = request.messages[-self._config.max_history_messages :]
        for item in history:
            content = self._trim_content(item.content)
            if not content and item.type == "text":
                continue
            messages.append(self._to_langchain_message(item.model_copy(update={"content": content})))
        return messages

    def _trim_content(self, content: str) -> str:
        return content.strip()[: self._config.max_message_chars]

    def _to_langchain_message(self, message: ChatMessage) -> BaseMessage:
        if message.role == "system":
            return SystemMessage(content=message.content)
        if message.role == "tool":
            tool_content = message.content
            if message.tool_result:
                tool_content = json.dumps(message.tool_result, ensure_ascii=False)
            return ToolMessage(content=tool_content, tool_call_id=message.tool_call_id or "")
        if message.role == "assistant":
            if message.type == "tool_use" and message.tool_name and message.tool_call_id:
                return AIMessage(
                    content=message.content,
                    tool_calls=[
                        {
                            "name": message.tool_name,
                            "args": message.tool_args or {},
                            "id": message.tool_call_id,
                            "type": "tool_call",
                        }
                    ],
                )
            return AIMessage(content=message.content)
        return HumanMessage(content=message.content)
