#### Go版本: 1.23.5

#### 安装依赖库: 
```bash
 cd backend/
 go mod download
```

#### 项目结构:

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

本项目包含三个可运行的 Go 服务（user-service、forum-service、audit-service），并在 `docker/` 下提供了一个示例 `docker-compose.yml`，用于本地一键启动 Postgres、Redis、MinIO 以及三个服务。

快速开始（在 `backend` 目录下执行）：

```bash
# 复制示例环境变量文件并按需修改
cp docker/.env.example docker/.env

# 在 docker 目录下启动（compose 文件位于 docker/docker-compose.yml）
cd docker
docker compose up --build -d
```

服务端口（默认）：
- user-service: 8081
- forum-service: 8082
- audit-service: 8083

注意：源码中的配置文件已调整为在容器网络中使用服务名（如 `postgres`, `minio`, `redis`）作为 host；如果你需要在宿主机上运行服务并连接宿主机上的数据库，请把 `internal/common/config/*.yaml` 中的 host 改回 `localhost` 或使用适当的环境配置/挂载覆盖。

停止容器并清理：

```bash
docker compose down -v
```
