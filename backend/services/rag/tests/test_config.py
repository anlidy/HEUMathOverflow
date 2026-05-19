from __future__ import annotations

import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from rag_service.config import load_config


CONFIG_TEMPLATE = """
server:
  host: 0.0.0.0
  port: 9091
rabbitmq:
  host: rabbitmq
  port: 5672
  user: user
  password: pass
qdrant:
  host: qdrant
  port: 6333
  collection: forum_rag_chunks
embedding:
  base_url: http://rag-model-server:8000/v1
  api_key: embedding-key
  model: embed-model
rerank:
  base_url: http://rag-model-server:8000/v1
  api_key: rerank-key
  model: rerank-model
llm:
  base_url: https://yaml.example/v1
  api_key: yaml-key
  model: gpt-5.4
chunking:
  question_chunk_size: 900
  question_chunk_overlap: 120
  answer_chunk_size: 800
  answer_chunk_overlap: 120
retrieval:
  global_chunk_limit: 36
  post_limit: 6
  answer_doc_limit: 2
  answer_chunk_limit: 4
  question_chunk_limit: 3
  neighbor_expand_limit: 1
  min_answer_chunks: 2
  min_question_chunks: 1
chat:
  max_tool_round_trips: 3
  tool_post_limit: 3
  tool_global_chunk_limit: 12
"""


class ConfigEnvOverrideTests(unittest.TestCase):
    def test_llm_env_overrides_yaml_values(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            config_path = Path(temp_dir) / "config.yaml"
            config_path.write_text(CONFIG_TEMPLATE, encoding="utf-8")
            with patch.dict(
                os.environ,
                {
                    "RAG_LLM_BASE_URL": "https://env.example/v1",
                    "RAG_LLM_API_KEY": "env-key",
                },
                clear=False,
            ):
                config = load_config(config_path)

        self.assertEqual(config.llm.base_url, "https://env.example/v1")
        self.assertEqual(config.llm.api_key, "env-key")
        self.assertEqual(config.llm.model, "gpt-5.4")


if __name__ == "__main__":
    unittest.main()
