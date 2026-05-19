# rag-service

`rag-service` consumes forum `rag-events`, builds dense+sparse Qdrant indexes, and exposes an LLM chat API with internal retrieval tool calling.

## Flow

1. Consume `rag.data.created`, `rag.data.updated`, `rag.data.deleted` from RabbitMQ.
2. Split post question and selected answers into chunks.
3. Generate dense vectors with local `Qwen3-Embedding-4B` and BM25-style sparse vectors locally.
4. Upsert points into Qdrant.
5. Let the chat model call the internal retrieval tool when it needs indexed forum knowledge.

## Model endpoints

The service expects OpenAI-compatible local model endpoints:

- Embedding: `/v1/embeddings`
- Rerank: `/v1/rerank`
- Chat LLM: `/v1/chat/completions` compatible endpoint

Configure embedding/rerank in `config.yaml`.
Configure chat LLM `base_url` and `api_key` through Docker environment variables:

- `RAG_LLM_BASE_URL`
- `RAG_LLM_API_KEY`

In docker-compose, these are sourced from `backend/deploy/docker/.env`.

## API

- `GET /healthz`
- `POST /chat`

Example request:

```json
{
  "messages": [
    {
      "role": "user",
      "content": "how to prove a continuous function on [a,b] is uniformly continuous"
    }
  ]
}
```
