from langchain_core.embeddings import Embeddings
from langchain_core.runnables import RunnableLambda
from langchain_core.tools import tool
from langchain_core.vectorstores import VectorStoreRetriever
from langchain_qdrant import QdrantVectorStore
from qdrant_client import QdrantClient

from app.core.config import Settings


def build_qdrant_client(settings: Settings) -> QdrantClient:
    return QdrantClient(url=settings.qdrant_url, api_key=settings.qdrant_api_key or None)


def build_retriever(settings: Settings, embeddings: Embeddings) -> VectorStoreRetriever:
    vector_store = QdrantVectorStore(
        client=build_qdrant_client(settings),
        collection_name=settings.qdrant_collection,
        embedding=embeddings,
        content_key="page_content",
        metadata_payload_key="metadata",
    )
    return vector_store.as_retriever(search_kwargs={"k": settings.retrieval_top_k})


def build_retriever_as_tool(settings: Settings, embeddings: Embeddings):
    @tool
    async def search_forum_knowledge_base(query: str) -> str:
        """Search the forum knowledge base for posts, replies, summaries, and indexed documentation."""
        retriever = build_retriever(settings, embeddings)
        docs = await retriever.ainvoke(query)
        if not docs:
            return "No relevant documents found in the knowledge base."
        return "\n\n".join(doc.page_content for doc in docs)

    return search_forum_knowledge_base
