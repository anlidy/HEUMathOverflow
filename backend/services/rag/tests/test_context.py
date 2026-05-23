from __future__ import annotations

import unittest

from rag_service.config import ContextConfig
from rag_service.context import ConversationContextManager
from rag_service.models import ChatMessage, ChatRequest


class ContextManagerTests(unittest.TestCase):
    def test_build_messages_applies_system_prompt_and_history_limit(self) -> None:
        manager = ConversationContextManager(
            ContextConfig(
                system_prompt="configured system",
                max_history_messages=2,
                max_message_chars=5,
            )
        )
        request = ChatRequest(
            messages=[
                ChatMessage(role="user", content="older user message"),
                ChatMessage(role="assistant", content="older assistant message"),
                ChatMessage(role="user", content="  newest user message  "),
            ]
        )

        messages = manager.build_messages(request)

        self.assertEqual(messages[0].content, "configured system")
        self.assertEqual(messages[1].content, "older")
        self.assertEqual(messages[2].content, "newes")

    def test_build_messages_includes_summary_and_tool_result(self) -> None:
        manager = ConversationContextManager(ContextConfig(system_prompt="configured system", max_history_messages=5, max_message_chars=100))
        request = ChatRequest(
            summary="existing summary",
            messages=[
                ChatMessage(role="assistant", type="tool_use", content="", tool_name="search_knowledge_base", tool_call_id="call-1", tool_args={"query": "compactness"}),
                ChatMessage(role="tool", type="tool_result", content="summary line", tool_call_id="call-1", tool_result={"query": "compactness"}),
                ChatMessage(role="user", content="next user message"),
            ],
        )

        messages = manager.build_messages(request)

        self.assertEqual(messages[1].content, "Conversation summary:\nexisting summary")
        self.assertEqual(getattr(messages[2], "tool_calls", [])[0]["name"], "search_knowledge_base")
        self.assertEqual(messages[3].tool_call_id, "call-1")


if __name__ == "__main__":
    unittest.main()
