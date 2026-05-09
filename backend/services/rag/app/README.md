# services/rag

论坛站内知识问答服务的最小 Python 骨架，当前采用：

- `FastAPI` 作为 HTTP 入口
- `LangChain` 作为编排层
- `OpenAI API` 作为模型与向量化接口
- `Qdrant` 作为向量库
- `Redis` 作为缓存
- `PostgreSQL` 作为会话/持久化存储

## 目录结构

```text
services/rag/
  app/
    api/routes/           HTTP 路由
    core/                 配置、日志、容器
    integrations/         OpenAI / Qdrant / Redis / PostgreSQL 客户端
    repositories/         数据访问层
    schemas/              请求响应模型
    services/             RAG 业务编排
    tools/                LangChain tools
    main.py               服务入口
  .env.example
  pyproject.toml
```

## 快速启动

```bash
cd backend/services/rag/app
python -m venv .venv
source .venv/bin/activate
pip install -e .
cp .env.example .env
uvicorn app.main:app --reload --port 8090
```

## 当前已提供

- `/healthz` 健康检查
- `/api/v1/chat/ask` 问答接口骨架
- OpenAI tool calling agent 骨架
- Qdrant 检索工具接入位
- PostgreSQL 聊天记录表初始化与读写
- Redis 问答结果缓存

## 后续建议

1. 接入论坛帖子、回复、标签等正式知识入库流程。
2. 增加文档切片、embedding、upsert 到 Qdrant 的离线任务。
3. 为 tool 增加真实的帖子详情、用户画像、热帖查询能力。
4. 引入鉴权、监控、限流与评测集。
