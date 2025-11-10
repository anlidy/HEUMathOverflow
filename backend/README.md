#### Go版本: 1.24.x

#### 安装依赖库: 
```bash
 cd backend/
 go mod download
```

#### 项目结构:

`/bin` 部署用脚本

`/cmd `  各个go服务的启动入口(main.go)

`/docker` docker配置文件

`/internal` 程序的内部实现

- `/common` 各服务共享的一些模块
- `/common/config` 共享的配置文件
- `/common/db` 共享的数据库实例化函数
- `/common/middleware` 共享的中间件函数
- `/common/utils` 共享的工具函数

`/internal/user-service` 用户服务

- `/controller` 服务控制层, 负责解析请求, 调用service层的方法进行处理, 返回请求
- `/service` 服务层, 负责完成业务逻辑, 调用resository层的方法来控制数据库, 返回处理结果
- `/repository` 数据层, 封装了数据库操作逻辑, 对service层暴露接口
- `/router` 路由层, 负责定义服务的请求路径和处理函数, 调用中间件
- `/model` 数据模型, 包括用户表结构, 请求/响应体结构等 



### Docker 部署

本项目包含三个可运行的 Go 服务（user-service、forum-service、audit-service），并在 `docker/` 下提供了 Docker Compose 配置文件，用于一键启动 Postgres、Redis、MinIO、MongoDB 以及三个服务。

#### 前置要求

- Docker Desktop（Windows/Mac）或 Docker Engine（Linux）
- Docker Compose v2（通常包含在 Docker Desktop 中）

#### 快速开始

**方式一：使用部署脚本（推荐）**

**Windows:**

**方式A：使用批处理脚本**
```cmd
# 进入 backend 目录
cd backend

# 启动所有服务（构建镜像并后台运行）
bin\run.cmd up --build -d

# 查看容器状态
docker ps

# 查看服务日志
docker compose -p backend logs -f user-service

# 停止并清理
bin\run.cmd down
```

**方式B：使用 Python 脚本（推荐，更可靠）**
```cmd
# 进入 backend 目录
cd backend

# 启动所有服务（构建镜像并后台运行）
python bin\run.py up --build -d

# 查看容器状态
docker ps

# 停止并清理
python bin\run.py down
```

**注意**：如果批处理脚本 `run.cmd` 出现问题，建议使用 Python 脚本 `run.py`。

**Linux/Mac:**
```bash
# 进入 backend 目录
cd backend

# 启动所有服务（构建镜像并后台运行）
./bin/run up --build -d

# 查看容器状态
docker ps

# 查看服务日志
docker compose -p backend logs -f user-service

# 停止并清理
./bin/run down
```

**方式二：直接使用 Docker Compose**

```bash
# 进入 docker 目录
cd backend/docker

# 1. 检查并创建 .env 文件（如果不存在，需要手动创建或从 .env.example 复制）
# Windows: copy .env.example .env
# Linux/Mac: cp .env.example .env

# 2. 启动基础设施服务（PostgreSQL, Redis, MinIO, MongoDB）
docker compose -p backend -f docker-compose.yml up -d

# 3. 启动应用服务（user-service, forum-service, audit-service）
docker compose -p backend -f docker-compose.service.yml up --build -d

# 4. 查看所有容器状态
docker ps

# 5. 停止所有服务
docker compose -p backend -f docker-compose.service.yml down
docker compose -p backend -f docker-compose.yml down
```

#### 服务端口

- **user-service**: 8081
- **forum-service**: 8082
- **audit-service**: 8083
- **MinIO Console**: 9001（管理界面）
- **PostgreSQL**: 5432（默认不暴露，仅在容器网络内访问）
- **Redis**: 6379（默认不暴露，仅在容器网络内访问）
- **MongoDB**: 27017（默认不暴露，仅在容器网络内访问）

#### 常用命令

```bash
# 查看运行中的容器
docker ps

# 查看所有容器（包括已停止的）
docker ps -a

# 查看服务日志
docker compose -p backend logs [service-name]

# 实时查看日志
docker compose -p backend logs -f [service-name]

# 重启特定服务
docker compose -p backend restart user-service

# 进入容器
docker exec -it backend-user-service-1 sh

# 查看网络
docker network ls
docker network inspect backend_default

# 清理未使用的镜像和容器
docker system prune -a
```

#### 环境配置

部署脚本会自动检查 `backend/docker/.env` 文件是否存在，如果不存在会尝试从 `.env.example` 复制。请确保 `.env` 文件中配置了正确的环境变量：

- `POSTGRES_USER` - PostgreSQL 用户名
- `POSTGRES_PASSWORD` - PostgreSQL 密码
- `POSTGRES_DB` - PostgreSQL 数据库名
- `MINIO_ROOT_USER` - MinIO 根用户
- `MINIO_ROOT_PASSWORD` - MinIO 根密码
- `MONGO_USER` - MongoDB 用户名
- `MONGO_PASSWORD` - MongoDB 密码

#### 注意事项

1. **网络配置**：所有服务在 `backend_default` Docker 网络中运行，使用服务名（如 `postgres`, `minio`, `redis`）作为主机名进行通信
2. **数据持久化**：数据存储在 `backend/docker/data/` 目录下，包括：
   - PostgreSQL 数据：`data/Postgre/`
   - Redis 数据：`data/Redis/`
   - MinIO 数据：`data/Minio/`
   - MongoDB 数据：`data/Mongo/`
3. **配置文件**：服务配置文件位于 `cmd/*/` 目录下的 yaml 文件，已配置为使用容器网络中的服务名
4. **宿主机运行**：如果需要在宿主机上运行服务并连接容器中的数据库，需要修改配置文件中的 host 为 `localhost` 并暴露相应端口

#### 故障排查

```bash
# 检查容器日志
docker compose -p backend logs user-service

# 检查容器状态
docker compose -p backend ps

# 检查网络连接
docker network inspect backend_default

# 重启所有服务
docker compose -p backend restart

# 完全清理并重新部署
bin\run.cmd down
bin\run.cmd up --build -d
```
