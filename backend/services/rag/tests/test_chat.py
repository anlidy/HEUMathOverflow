from __future__ import annotations

import unittest

from fastapi.testclient import TestClient
from langchain_core.messages import AIMessage

from rag_service.api import build_app
from rag_service.chat import ChatService, RetrievalTool
from rag_service.config import ChatConfig, ContextConfig
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


class FakeChatService:
    def chat(self, _request: ChatRequest) -> ChatResponse:
        return ChatResponse(answer="ok")


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

    def test_api_exposes_chat_and_not_search(self) -> None:
        client = TestClient(build_app(FakeChatService()))

        chat_response = client.post("/chat", json={"messages": [{"role": "user", "content": "hello"}]})
        search_response = client.post("/search", json={"query": "hello"})

        self.assertEqual(chat_response.status_code, 200)
        self.assertEqual(chat_response.json()["answer"], "ok")
        self.assertEqual(search_response.status_code, 404)


if __name__ == "__main__":
    unittest.main()
