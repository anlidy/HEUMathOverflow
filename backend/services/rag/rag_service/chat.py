from __future__ import annotations

import json
from typing import Any

import httpx

from langchain_core.messages import AIMessage, BaseMessage, ToolMessage
from langchain_core.tools import BaseTool, tool
from pydantic import BaseModel, Field

from rag_service.config import ChatConfig, LLMConfig
from rag_service.context import ConversationContextManager
from rag_service.models import ChatMessage, ChatRequest, ChatResponse, SearchPostView, SearchRequest, SearchResponse, SummaryUpdate, ToolTrace
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
        self._traces: list[ToolTrace] = []
        self._tool = self._build_tool()

    @property
    def tool(self) -> BaseTool:
        return self._tool

    def reset(self) -> None:
        self._latest_response = None
        self._traces = []

    def latest_posts(self) -> list[SearchPostView]:
        if self._latest_response is None:
            return []
        return self._latest_response.posts

    def traces(self) -> list[ToolTrace]:
        return list(self._traces)

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
            content = self._build_trace_content(response)
            raw_result = response.model_dump(mode="json")
            self._traces.append(
                ToolTrace(
                    tool_name="search_knowledge_base",
                    tool_call_id="",
                    tool_args=request.model_dump(mode="json"),
                    tool_result=raw_result,
                    content=content,
                    status="ok",
                )
            )
            return json.dumps(response.model_dump(mode="json"), ensure_ascii=False)

        return search_knowledge_base

    def set_last_tool_call_id(self, tool_call_id: str) -> None:
        if not self._traces:
            return
        self._traces[-1].tool_call_id = tool_call_id

    def _build_trace_content(self, response: SearchResponse) -> str:
        chunk_ids: list[str] = []
        for post in response.posts:
            chunk_ids.extend(chunk.point_id for chunk in post.question_chunks)
            for answer in post.answers:
                chunk_ids.extend(chunk.point_id for chunk in answer.chunks)
        unique_ids = list(dict.fromkeys(chunk_ids))
        return f"检索到相关chunks {len(unique_ids)} 个,chunk ids:[{','.join(unique_ids)}]"


class OpenAICompatibleChatModel:
    def __init__(self, config: LLMConfig, tools: list[BaseTool] | None = None) -> None:
        self._config = config
        self._tools = list(tools or [])
        self._client = httpx.Client(timeout=config.timeout_seconds)

    def bind_tools(self, tools: list[BaseTool]) -> "OpenAICompatibleChatModel":
        return OpenAICompatibleChatModel(self._config, tools)

    def invoke(self, messages: list[Any]) -> AIMessage:
        payload: dict[str, Any] = {
            "model": self._config.model,
            "temperature": self._config.temperature,
            "stream": False,
            "messages": [self._serialize_message(message) for message in messages],
        }
        if self._tools:
            payload["tools"] = [self._serialize_tool(tool) for tool in self._tools]

        response_text = ""
        last_error: Exception | None = None
        for _ in range(max(self._config.max_retries, 0) + 1):
            try:
                response = self._client.post(
                    self._chat_completions_url(),
                    headers=self._build_headers(),
                    json=payload,
                )
                response.raise_for_status()
                response_text = response.text
                return self._decode_response(response)
            except (httpx.HTTPError, ValueError) as exc:
                last_error = exc
        raise RuntimeError(f"chat completion request failed: {last_error}; body={response_text[:500]}")

    def _build_headers(self) -> dict[str, str]:
        headers = {"Content-Type": "application/json"}
        if self._config.api_key:
            headers["Authorization"] = f"Bearer {self._config.api_key}"
        return headers

    def _chat_completions_url(self) -> str:
        base_url = self._config.base_url.rstrip("/")
        if base_url.endswith("/chat/completions"):
            return base_url
        return f"{base_url}/chat/completions"

    def _serialize_message(self, message: Any) -> dict[str, Any]:
        if isinstance(message, dict):
            return dict(message)
        if getattr(message, "type", None) == "system":
            return {"role": "system", "content": self._normalize_content(message.content)}
        if getattr(message, "type", None) == "human":
            return {"role": "user", "content": self._normalize_content(message.content)}
        if isinstance(message, ToolMessage):
            return {
                "role": "tool",
                "content": self._normalize_content(message.content),
                "tool_call_id": message.tool_call_id,
            }
        if isinstance(message, AIMessage):
            payload: dict[str, Any] = {
                "role": "assistant",
                "content": self._normalize_content(message.content),
            }
            tool_calls = getattr(message, "tool_calls", []) or []
            if tool_calls:
                payload["tool_calls"] = [self._serialize_tool_call(tool_call) for tool_call in tool_calls]
            return payload
        raise TypeError(f"unsupported message type: {type(message)!r}")

    def _serialize_tool(self, tool: BaseTool) -> dict[str, Any]:
        schema = tool.args_schema.model_json_schema() if tool.args_schema else {"type": "object", "properties": {}}
        return {
            "type": "function",
            "function": {
                "name": tool.name,
                "description": tool.description or "",
                "parameters": schema,
            },
        }

    def _serialize_tool_call(self, tool_call: dict[str, Any]) -> dict[str, Any]:
        return {
            "id": str(tool_call.get("id", "")),
            "type": tool_call.get("type", "function"),
            "function": {
                "name": str(tool_call.get("name", "")),
                "arguments": json.dumps(tool_call.get("args", {}), ensure_ascii=False),
            },
        }

    def _decode_response(self, response: httpx.Response) -> AIMessage:
        content_type = response.headers.get("content-type", "")
        body = response.text
        if "text/event-stream" in content_type or body.lstrip().startswith("data:"):
            return self._decode_sse_response(body)
        return self._decode_json_response(response.json())

    def _decode_json_response(self, payload: dict[str, Any]) -> AIMessage:
        message = ((payload.get("choices") or [{}])[0]).get("message") or {}
        ai_message = AIMessage(
            content=message.get("content") or "",
            tool_calls=self._parse_tool_calls(message.get("tool_calls") or []),
        )
        self._ensure_non_empty_completion(ai_message)
        return ai_message

    def _decode_sse_response(self, body: str) -> AIMessage:
        content_parts: list[str] = []
        tool_calls: dict[int, dict[str, Any]] = {}
        saw_chunks = False
        for raw_line in body.splitlines():
            line = raw_line.strip()
            if not line.startswith("data:"):
                continue
            data = line[5:].strip()
            if not data or data == "[DONE]":
                continue
            payload = json.loads(data)
            if payload.get("object") == "chat.completion":
                return self._decode_json_response(payload)
            if payload.get("object") != "chat.completion.chunk":
                continue
            saw_chunks = True
            for choice in payload.get("choices") or []:
                delta = choice.get("delta") or {}
                content = delta.get("content")
                if isinstance(content, str) and content:
                    content_parts.append(content)
                for tool_call in delta.get("tool_calls") or []:
                    index = int(tool_call.get("index", 0))
                    entry = tool_calls.setdefault(index, {"id": "", "name": "", "arguments": []})
                    if tool_call.get("id"):
                        entry["id"] = str(tool_call["id"])
                    function = tool_call.get("function") or {}
                    if function.get("name"):
                        entry["name"] = str(function["name"])
                    if function.get("arguments"):
                        entry["arguments"].append(str(function["arguments"]))
        if not saw_chunks:
            raise ValueError("invalid SSE chat completion response")
        ai_message = AIMessage(
            content="".join(content_parts),
            tool_calls=self._finalize_stream_tool_calls(tool_calls),
        )
        self._ensure_non_empty_completion(ai_message)
        return ai_message

    def _finalize_stream_tool_calls(self, tool_calls: dict[int, dict[str, Any]]) -> list[dict[str, Any]]:
        finalized: list[dict[str, Any]] = []
        for index in sorted(tool_calls):
            tool_call = tool_calls[index]
            arguments_text = "".join(tool_call["arguments"]).strip()
            args: dict[str, Any] = {}
            if arguments_text:
                parsed = json.loads(arguments_text)
                if not isinstance(parsed, dict):
                    raise ValueError("tool call arguments must decode to an object")
                args = parsed
            finalized.append(
                {
                    "name": tool_call["name"],
                    "args": args,
                    "id": tool_call["id"],
                    "type": "tool_call",
                }
            )
        return finalized

    def _parse_tool_calls(self, raw_tool_calls: list[dict[str, Any]]) -> list[dict[str, Any]]:
        parsed: list[dict[str, Any]] = []
        for tool_call in raw_tool_calls:
            function = tool_call.get("function") or {}
            arguments_text = function.get("arguments") or "{}"
            arguments = json.loads(arguments_text)
            if not isinstance(arguments, dict):
                raise ValueError("tool call arguments must decode to an object")
            parsed.append(
                {
                    "name": str(function.get("name", "")),
                    "args": arguments,
                    "id": str(tool_call.get("id", "")),
                    "type": str(tool_call.get("type", "tool_call")),
                }
            )
        return parsed

    def _normalize_content(self, content: Any) -> Any:
        if isinstance(content, list):
            return content
        return "" if content is None else str(content)

    def _ensure_non_empty_completion(self, message: AIMessage) -> None:
        content = self._normalize_content(message.content)
        has_content = bool(content if not isinstance(content, list) else len(content))
        has_tool_calls = bool(getattr(message, "tool_calls", []) or [])
        if not has_content and not has_tool_calls:
            raise ValueError("chat completion response contained neither content nor tool calls")


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
        effective_request, summary_update = self._prepare_request(request)
        messages = self._context_manager.build_messages(effective_request)
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
                    tool_traces=self._retrieval_tool.traces(),
                    summary_update=summary_update,
                )
            for tool_call in tool_calls:
                used_tools.append(str(tool_call.get("name", "")))
                result = self._retrieval_tool.tool.invoke(tool_call.get("args", {}))
                self._retrieval_tool.set_last_tool_call_id(str(tool_call["id"]))
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

    def _prepare_request(self, request: ChatRequest) -> tuple[ChatRequest, SummaryUpdate | None]:
        summary = request.summary or ""
        messages = list(request.messages)
        if not messages:
            return request, None
        total_chars = len(summary) + sum(self._message_cost(message) for message in messages)
        if total_chars <= self._config.compress_after_chars or len(messages) <= self._config.preserve_recent_messages:
            return request, None

        split_index = self._adjust_compression_boundary(messages)
        if split_index <= 0:
            return request, None
        compressible = messages[:split_index]
        if not compressible or compressible[-1].message_id is None:
            return request, None
        new_summary = self._summarize_messages(summary, compressible)
        summary_update = SummaryUpdate(summary=new_summary, up_to_message_id=int(compressible[-1].message_id))
        updated_request = request.model_copy(update={
            "summary": new_summary,
            "summary_up_to_message_id": summary_update.up_to_message_id,
            "messages": messages[split_index:],
        })
        return updated_request, summary_update

    def _adjust_compression_boundary(self, messages: list[ChatMessage]) -> int:
        split_index = max(len(messages) - self._config.preserve_recent_messages, 0)
        while split_index > 0:
            current = messages[split_index]
            previous = messages[split_index - 1]
            if current.role == "tool" or previous.role == "tool" or previous.type == "tool_use":
                split_index -= 1
                continue
            break
        return split_index

    def _message_cost(self, message: ChatMessage) -> int:
        payload = message.content
        if message.tool_result:
            payload += json.dumps(message.tool_result, ensure_ascii=False)
        elif message.tool_args:
            payload += json.dumps(message.tool_args, ensure_ascii=False)
        return len(payload)

    def _summarize_messages(self, existing_summary: str, messages: list[ChatMessage]) -> str:
        transcript = [f"Previous summary:\n{existing_summary}" if existing_summary else ""]
        for message in messages:
            if message.type == "tool_use":
                transcript.append(f"[assistant/tool_use] {message.tool_name}: {json.dumps(message.tool_args or {}, ensure_ascii=False)}")
            elif message.role == "tool":
                transcript.append(f"[tool/tool_result] {message.content}")
            else:
                transcript.append(f"[{message.role}/{message.type}] {message.content}")
        prompt = [
            {"role": "system", "content": "Summarize the conversation history for future context reuse. Preserve key facts, unresolved questions, and important tool findings."},
            {"role": "user", "content": "\n".join(line for line in transcript if line)},
        ]
        try:
            response = self._llm.invoke(prompt)
            summary = self._extract_text(response).strip()
            if summary:
                return summary
        except Exception:
            pass
        return "\n".join(line for line in transcript if line)[-4000:]


def build_chat_model(config: LLMConfig) -> OpenAICompatibleChatModel:
    return OpenAICompatibleChatModel(config)
