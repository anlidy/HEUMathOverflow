from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_prefix="RAG_", extra="ignore")

    app_name: str = "rag-service"
    app_env: str = "dev"
    app_port: int = 8090

    openai_api_key: str = ""
    openai_base_url: str = "https://api.openai.com/v1"
    openai_model: str = "gpt-4.1-mini"
    openai_embedding_model: str = "text-embedding-3-small"

    qdrant_url: str = "http://localhost:6333"
    qdrant_api_key: str = ""
    qdrant_collection: str = "forum_knowledge"

    redis_url: str = "redis://localhost:6379/0"
    postgres_dsn: str = "postgresql+asyncpg://postgres:postgres@localhost:5432/postgres"

    retrieval_top_k: int = 4
    chat_history_limit: int = 12
    cache_ttl_seconds: int = 300
    agent_verbose: bool = False


@lru_cache
def get_settings() -> Settings:
    return Settings()
