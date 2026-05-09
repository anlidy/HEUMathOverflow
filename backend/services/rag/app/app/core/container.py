from langchain.agents import AgentExecutor, create_tool_calling_agent
from langchain.prompts import ChatPromptTemplate, MessagesPlaceholder

from app.core.config import Settings
from app.integrations.openai_client import build_chat_model, build_embeddings
from app.integrations.postgres import build_engine, build_session_factory
from app.integrations.qdrant_store import build_retriever_as_tool
from app.integrations.redis_client import build_redis_client
from app.repositories.chat_session_repository import ChatSessionRepository
from app.services.rag_service import RagService
from app.tools.forum_tools import build_tools


class AppContainer:
    def __init__(self, settings: Settings):
        self.settings = settings
        self.redis = build_redis_client(settings)
        self.engine = build_engine(settings)
        self.session_factory = build_session_factory(self.engine)
        self.chat_repository = ChatSessionRepository(self.session_factory)

        self.llm = build_chat_model(settings)
        self.embeddings = build_embeddings(settings)

        knowledge_tool = build_retriever_as_tool(settings, self.embeddings)
        self.tools = [knowledge_tool, *build_tools(self.chat_repository)]
        self.agent_executor = self._build_agent_executor()
        self.rag_service = RagService(
            settings=settings,
            redis_client=self.redis,
            chat_repository=self.chat_repository,
            agent_executor=self.agent_executor,
        )

    def _build_agent_executor(self) -> AgentExecutor:
        prompt = ChatPromptTemplate.from_messages(
            [
                (
                    "system",
                    "You are the MathOverflow forum knowledge assistant. Use retrieval before answering specific knowledge questions. Use tools when external context is needed. If you are unsure, state the uncertainty clearly.",
                ),
                MessagesPlaceholder("chat_history", optional=True),
                ("human", "{input}"),
                MessagesPlaceholder("agent_scratchpad"),
            ]
        )
        agent = create_tool_calling_agent(self.llm, self.tools, prompt)
        return AgentExecutor(agent=agent, tools=self.tools, verbose=self.settings.agent_verbose)

    async def startup(self) -> None:
        await self.chat_repository.init_schema()

    async def shutdown(self) -> None:
        await self.redis.close()
        await self.engine.dispose()
