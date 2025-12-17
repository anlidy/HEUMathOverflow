# HEU Math Overflow - 公共数学智慧答疑与社区认证系统

一个融合 AI 大模型实时应答能力、学生社区协作学习以及教师专业权威引导的智慧学习生态系统。

## 项目简介

本项目旨在为公共数学课程提供一个新型的在线答疑平台，实现：
- 学生提问与互动讨论
- AI 实时自动解答
- 教师认证优质答案
- 积累高质量问答数据

## 快速导航

### 📚 文档
- [需求文档](./需求文档.md) - 完整的项目需求和功能说明
- [MVP文档](./MVP文档.md) - 最小可行产品规划
- [数据库 ER 图（中文）](./数据库ER图.md) - 完整的数据库架构和实体关系图
- [Database ER Diagram (English)](./Database-ER-Diagram.md) - Complete database schema and ER diagram
- [前后端接口字段差异分析](./前后端接口字段差异分析.md)
- [后端待补充接口字段](./后端待补充接口字段.md)

### 💻 技术栈

**后端**
- 语言：Go 1.24+
- 框架：Gin + GORM
- 数据库：PostgreSQL 14+
- 缓存：Redis 7+
- 消息队列：RabbitMQ
- 搜索引擎：Elasticsearch
- 对象存储：MinIO
- API 文档：参见 [后端文档](./backend/)

**前端**
- 框架：Vue.js
- 详见 [前端文档](./frontend/)

### 🗄️ 数据库架构

项目采用微服务架构，使用三个独立的 PostgreSQL 数据库：

1. **user_db** - 用户服务数据库
   - User 表：用户信息（ID、用户名、邮箱、角色等）

2. **forum_db** - 论坛服务数据库
   - Post 表：帖子/问题
   - Reply 表：回复/答案
   - PostLike、PostStar 表：点赞和收藏
   - ReplyLike 表：回复点赞

3. **audit_db** - 审核服务数据库（预留）

**查看完整 ER 图：**
- [中文版 ER 图](./数据库ER图.md) - 包含详细的表结构、字段说明、关系约束
- [English ER Diagram](./Database-ER-Diagram.md) - Complete table structures, field descriptions, and relationships

## 核心功能

### 🎓 用户角色
- **学生**：发帖提问、回复讨论、点赞评论
- **助教/教师**：拥有学生权限，可认证答案、管理版块、审核内容
- **AI 助教**：系统账户，自动发布 AI 生成的答案

### 📝 论坛社区
- 富文本编辑，支持 LaTeX 数学公式渲染
- 标签分类（微积分、线性代数等）
- 帖子状态筛选（未回答、已回答、已认证）
- 点赞、收藏、评论功能

### 🤖 AI 集成
- 实时调用 AI 大模型（豆包、Kimi、智谱清言等）
- 自动生成答案并发布
- 支持上传教材、辅导书作为知识库

### ✅ 内容审核
- AI 自动审核不当内容
- 教师人工审核管理界面
- 违规内容自动拦截

### 🏆 教师认证
- 教师可认证优质答案
- 自动收集认证的问答对
- 支持导出数据用于模型训练

## 快速开始

### 后端部署

详细的部署说明请参考 [后端 README](./backend/README.md)

**使用 Docker Compose（推荐）：**

```bash
cd backend

# Windows
python bin\run.py up --build -d

# Linux/Mac
./bin/run up --build -d
```

**服务端口：**
- user-service: 8081
- forum-service: 8082
- audit-service: 8083
- MinIO Console: 9001

### 前端部署

详细说明请参考 [前端 README](./frontend/README.md)

## 项目结构

```
HEUMathOverflow/
├── backend/                # 后端服务
│   ├── cmd/               # 服务入口点
│   │   ├── user-service/
│   │   ├── forum-service/
│   │   └── audit-service/
│   ├── internal/          # 内部实现
│   │   ├── common/        # 共享模块
│   │   ├── user-service/
│   │   ├── forum-service/
│   │   └── audit-service/
│   └── docker/            # Docker 配置
├── frontend/              # 前端应用
├── 数据库ER图.md          # 数据库架构文档（中文）
├── Database-ER-Diagram.md # Database schema doc (English)
└── 需求文档.md            # 需求文档
```

## 开发团队

公共数学教研室

## 许可证

[待定]

---

**文档版本**: 1.0  
**最后更新**: 2025-12-17
