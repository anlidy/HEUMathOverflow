from __future__ import annotations

import json
import unittest

from fastapi.testclient import TestClient
from langchain_core.messages import AIMessage

from rag_service.api import build_app
from rag_service.chat import ChatService, OpenAICompatibleChatModel, RetrievalTool
from rag_service.config import ChatConfig, ContextConfig, LLMConfig
from rag_service.context import ConversationContextManager
from rag_service.models import ChatMessage, ChatRequest, ChatResponse, SearchPostView, SearchRequest, SearchResponse


class FakeRetrievalService:
    def search(self, request: SearchRequest) -> SearchResponse:
        return SearchResponse(
            query=request.query,
            posts=[
                SearchPostView(
                    post_id=1,
                    title="Uniform continuity on compact intervals",
                    tags=["analysis"],
                    score=1.0,
                    question_chunks=[],
                    answers=[],
                )
            ],
        )


class FakeBoundLLM:
    def __init__(self, responses: list[AIMessage]) -> None:
        self._responses = responses
        self._index = 0

    def invoke(self, _messages: list[object]) -> AIMessage:
        response = self._responses[self._index]
        self._index += 1
        return response


class FakeLLM:
    def __init__(self, responses: list[AIMessage]) -> None:
        self._responses = responses

    def bind_tools(self, _tools: list[object]) -> FakeBoundLLM:
        return FakeBoundLLM(self._responses)

    def invoke(self, _messages: list[object]) -> AIMessage:
        return self._responses.pop(0)


class FakeChatService:
    def chat(self, _request: ChatRequest) -> ChatResponse:
        return ChatResponse(answer="ok")


class FakeHTTPResponse:
    def __init__(self, text: str, *, content_type: str = "application/json") -> None:
        self.text = text
        self.headers = {"content-type": content_type}

    def raise_for_status(self) -> None:
        return None

    def json(self) -> dict[str, object]:
        return json.loads(self.text)


class FakeHTTPClient:
    def __init__(self, responses: list[FakeHTTPResponse]) -> None:
        self._responses = responses
        self.requests: list[dict[str, object]] = []

    def post(self, url: str, *, headers: dict[str, str], json: dict[str, object]) -> FakeHTTPResponse:
        self.requests.append({"url": url, "headers": headers, "json": json})
        return self._responses.pop(0)


class ChatServiceTests(unittest.TestCase):
    def test_chat_service_runs_retrieval_tool_loop(self) -> None:
        llm = FakeLLM(
            [
                AIMessage(
                    content="",
                    tool_calls=[
                        {
                            "name": "search_knowledge_base",
                            "args": {"query": "uniform continuity"},
                            "id": "call-1",
                        }
                    ],
                ),
                AIMessage(content="Use Heine-Cantor on [a,b]."),
            ]
        )
        service = ChatService(
            llm=llm,
            context_manager=ConversationContextManager(ContextConfig(system_prompt="system")),
            retrieval_tool=RetrievalTool(FakeRetrievalService(), ChatConfig()),
            config=ChatConfig(),
        )

        response = service.chat(ChatRequest(messages=[ChatMessage(role="user", content="How?")]))

        self.assertEqual(response.answer, "Use Heine-Cantor on [a,b].")
        self.assertEqual(response.used_tools, ["search_knowledge_base"])
        self.assertEqual(len(response.retrieved_posts), 1)
        self.assertEqual(len(response.tool_traces), 1)

    def test_chat_service_returns_summary_update_when_threshold_exceeded(self) -> None:
        llm = FakeLLM([
            AIMessage(content="compressed summary"),
            AIMessage(content="final answer"),
        ])
        service = ChatService(
            llm=llm,
            context_manager=ConversationContextManager(ContextConfig(system_prompt="system", max_history_messages=20)),
            retrieval_tool=RetrievalTool(FakeRetrievalService(), ChatConfig(compress_after_chars=20, preserve_recent_messages=1)),
            config=ChatConfig(compress_after_chars=20, preserve_recent_messages=1),
        )

        response = service.chat(
            ChatRequest(
                messages=[
                    ChatMessage(message_id=1, role="user", content="first message"),
                    ChatMessage(message_id=2, role="assistant", content="second message"),
                    ChatMessage(message_id=3, role="user", content="third message"),
                ]
            )
        )

        self.assertEqual(response.answer, "final answer")
        self.assertIsNotNone(response.summary_update)
        self.assertEqual(response.summary_update.summary, "compressed summary")
        self.assertEqual(response.summary_update.up_to_message_id, 2)

    def test_api_exposes_chat_and_not_search(self) -> None:
        client = TestClient(build_app(FakeChatService()))

        chat_response = client.post("/chat", json={"messages": [{"role": "user", "content": "hello"}]})
        search_response = client.post("/search", json={"query": "hello"})

        self.assertEqual(chat_response.status_code, 200)
        self.assertEqual(chat_response.json()["answer"], "ok")
        self.assertEqual(search_response.status_code, 404)

    def test_openai_compatible_model_parses_json_tool_calls(self) -> None:
        model = OpenAICompatibleChatModel(LLMConfig(base_url="https://example.com/v1", api_key="key", model="gpt-test", max_retries=0))
        model._client = FakeHTTPClient(
            [
                FakeHTTPResponse(
                    json.dumps(
                        {
                            "choices": [
                                {
                                    "message": {
                                        "role": "assistant",
                                        "content": "",
                                        "tool_calls": [
                                            {
                                                "id": "call-1",
                                                "type": "function",
                                                "function": {
                                                    "name": "search_knowledge_base",
                                                    "arguments": '{"query":"compactness"}',
                                                },
                                            }
                                        ],
                                    }
                                }
                            ]
                        }
                    )
                )
            ]
        )

        response = model.invoke([{"role": "user", "content": "hello"}])

        self.assertEqual(response.tool_calls[0]["name"], "search_knowledge_base")
        self.assertEqual(response.tool_calls[0]["args"], {"query": "compactness"})

    def test_openai_compatible_model_parses_sse_tool_calls(self) -> None:
        body = "\n".join(
            [
                'data: {"id":"1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"call-1","type":"function","function":{"name":"search_knowledge_base","arguments":"{\\"query\\":\\"compact"}}]}}]}',
                'data: {"id":"1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"ness\\"}"}}]}}]}',
                'data: {"id":"1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Found it."}}]}',
                'data: [DONE]',
            ]
        )
        model = OpenAICompatibleChatModel(LLMConfig(base_url="https://example.com/v1", api_key="key", model="gpt-test", max_retries=0))
        model._client = FakeHTTPClient([FakeHTTPResponse(body, content_type="text/event-stream")])

        response = model.invoke([{"role": "user", "content": "hello"}])

        self.assertEqual(response.content, "Found it.")
        self.assertEqual(response.tool_calls[0]["name"], "search_knowledge_base")
        self.assertEqual(response.tool_calls[0]["args"], {"query": "compactness"})

    def test_openai_compatible_model_rejects_empty_sse_completion(self) -> None:
        body = "\n".join(
            [
                'data: {"id":"1","object":"chat.completion.chunk","choices":[],"usage":{"total_tokens":1}}',
                'data: [DONE]',
            ]
        )
        model = OpenAICompatibleChatModel(
            LLMConfig(base_url="https://example.com/v1", api_key="key", model="gpt-test", max_retries=0)
        )
        model._client = FakeHTTPClient([FakeHTTPResponse(body, content_type="text/event-stream")])

        with self.assertRaisesRegex(RuntimeError, "neither content nor tool calls"):
            model.invoke([{"role": "user", "content": "hello"}])


if __name__ == "__main__":
    unittest.main()
