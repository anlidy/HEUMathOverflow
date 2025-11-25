#### Go版本: 1.24.x

#### 安装依赖库: 
```bash
# 容器部署时由docker自动完成
 cd backend/
 go mod download
```

#### 项目结构:

`/bin` 部署用脚本

`/cmd `  各个go服务的启动入口(main.go)

`/docker` docker及k8s配置文件

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

本项目包含三个可运行的 Go 服务（user-service、forum-service、audit-service），并在 `docker/` 下提供了一个示例 `docker-compose.service.yml`，用于本地一键启动 Postgres、Redis、MinIO 以及三个服务。

快速开始（任意目录均可，相对位置调用脚本即可）：

```bash
# 脚本部署
# Windows把脚本换成run.cmd
# 在任何目录下可用
./bin/run up --build -d

# 查看docker容器运行状态
docker ps
```

服务端口（默认）：
- user-service: 8081
- forum-service: 8082
- audit-service: 8083

注意：源码中的配置文件已调整为在容器网络中使用服务名（如 `postgres`, `minio`, `redis`）作为 host；如果你需要在宿主机上运行服务并连接宿主机上的数据库，请把 `/cmd/**/*.yaml` 中的 host 改回 `localhost` 或使用适当的环境配置/挂载覆盖。

停止容器并清理：

```bash
# 停止服务容器, 自动删除容器和.env, -v 参数删除数据库挂载卷，--rmi local参数删除本地构建的image
./bin/run down
```

### kind 本地集群CI测试
1. `kind create cluster`创建集群
    kind为容器嵌套结构，使用容器来代替一个具核node，目前已知在linux上`kind create cluster`时kind会将proxy环境变量透传到node中，使得node中一切网络活动失效，影响集群运行，所以应当进入容器取消环境变量或直接关闭代理，使用TUN模式，在Windows+Docker desktop + WSL2平台可直接在cmd中使用集群创建指令，而不会有网络代理问题

