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


if __name__ == "__main__":
    unittest.main()
