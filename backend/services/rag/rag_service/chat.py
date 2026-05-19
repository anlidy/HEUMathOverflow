from __future__ import annotations

import json
from typing import Any

from langchain_core.messages import AIMessage, BaseMessage, ToolMessage
from langchain_core.tools import BaseTool, tool
from langchain_openai import ChatOpenAI
from pydantic import BaseModel, Field

from rag_service.config import ChatConfig, LLMConfig
from rag_service.context import ConversationContextManager
from rag_service.models import ChatRequest, ChatResponse, SearchPostView, SearchRequest, SearchResponse
from rag_service.retrieval import RetrievalService


class RetrievalToolInput(BaseModel):
    query: str = Field(description="Question to search in the indexed MathOverflow content.")
    post_limit: int | None = Field(default=None, ge=1, description="Maximum number of posts to return.")
    global_chunk_limit: int | None = Field(default=None, ge=1, description="Maximum number of chunks to inspect.")


class RetrievalTool:
    def __init__(self, retrieval: RetrievalService, config: ChatConfig) -> None:
        self._retrieval = retrieval
        self._config = config
        self._latest_response: SearchResponse | None = None
        self._tool = self._build_tool()

    @property
    def tool(self) -> BaseTool:
        return self._tool

    def reset(self) -> None:
        self._latest_response = None

    def latest_posts(self) -> list[SearchPostView]:
        if self._latest_response is None:
            return []
        return self._latest_response.posts

    def _build_tool(self) -> BaseTool:
        @tool("search_knowledge_base", args_schema=RetrievalToolInput)
        def search_knowledge_base(query: str, post_limit: int | None = None, global_chunk_limit: int | None = None) -> str:
            """Search indexed MathOverflow posts and answers for relevant context."""
            request = SearchRequest(
                query=query,
                post_limit=post_limit or self._config.tool_post_limit,
                global_chunk_limit=global_chunk_limit or self._config.tool_global_chunk_limit,
            )
            response = self._retrieval.search(request)
            self._latest_response = response
            return json.dumps(response.model_dump(mode="json"), ensure_ascii=False)

        return search_knowledge_base


class ChatService:
    def __init__(
        self,
        llm: Any,
        context_manager: ConversationContextManager,
        retrieval_tool: RetrievalTool,
        config: ChatConfig,
    ) -> None:
        self._llm = llm
        self._context_manager = context_manager
        self._retrieval_tool = retrieval_tool
        self._config = config

    def chat(self, request: ChatRequest) -> ChatResponse:
        self._retrieval_tool.reset()
        messages = self._context_manager.build_messages(request)
        llm_with_tools = self._llm.bind_tools([self._retrieval_tool.tool])
        used_tools: list[str] = []

        for _ in range(self._config.max_tool_round_trips + 1):
            ai_message = llm_with_tools.invoke(messages)
            messages.append(ai_message)
            tool_calls = getattr(ai_message, "tool_calls", []) or []
            if not tool_calls:
                return ChatResponse(
                    answer=self._extract_text(ai_message),
                    used_tools=used_tools,
                    retrieved_posts=self._retrieval_tool.latest_posts(),
                )
            for tool_call in tool_calls:
                used_tools.append(str(tool_call.get("name", "")))
                result = self._retrieval_tool.tool.invoke(tool_call.get("args", {}))
                messages.append(ToolMessage(content=str(result), tool_call_id=str(tool_call["id"])))

        raise RuntimeError("chat tool-calling exceeded configured round-trip limit")

    def _extract_text(self, message: BaseMessage) -> str:
        content = getattr(message, "content", "")
        if isinstance(content, str):
            return content
        if isinstance(content, list):
            texts: list[str] = []
            for item in content:
                if isinstance(item, str):
                    texts.append(item)
                    continue
                if isinstance(item, dict) and item.get("type") == "text":
                    texts.append(str(item.get("text", "")))
            return "\n".join(texts).strip()
        return str(content)


def build_chat_model(config: LLMConfig) -> ChatOpenAI:
    return ChatOpenAI(
        model=config.model,
        api_key=config.api_key,
        base_url=config.base_url,
        temperature=config.temperature,
        timeout=config.timeout_seconds,
        max_retries=config.max_retries,
    )
