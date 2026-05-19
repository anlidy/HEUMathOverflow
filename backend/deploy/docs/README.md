#### Go版本: 1.24.x

#### 安装依赖库: 
```bash
# 容器部署时由docker自动完成
 cd backend/
 go mod download
```

#### 项目结构:

`/deploy/bin` 部署用脚本

`/services` 各个微服务代码与入口(main.go)

`/deploy/docker` docker及k8s配置文件

`/common` 公共库

`/api` proto与生成代码

- `/common` 各服务共享的一些模块
- `/common/config` 共享的配置文件
- `/common/db` 共享的数据库实例化函数
- `/common/middleware` 共享的中间件函数
- `/common/utils` 共享的工具函数

`/services/user/internal` 用户服务

- `/handler` 服务控制层, 负责解析请求, 调用service层的方法进行处理, 返回请求
- `/service` 服务层, 负责完成业务逻辑, 调用resository层的方法来控制数据库, 返回处理结果
- `/repository` 数据层, 封装了数据库操作逻辑, 对service层暴露接口
- `/server` 路由与服务启动相关代码
- `/model` 数据模型, 包括用户表结构, 请求/响应体结构等 



### Docker 部署

本项目包含多个可运行服务，并在 `deploy/docker/` 下提供了 `docker-compose.service.yml`，用于本地一键启动依赖和服务。

快速开始（任意目录均可，相对位置调用脚本即可）：

```bash
# 脚本部署
# Windows把脚本换成run.cmd
# 在任何目录下可用
./deploy/bin/run up --build -d

# 查看docker容器运行状态
docker ps
```

服务端口（默认）：
- user-service: 8081
- forum-service: 8082
- audit-service: 8083

注意：源码中的配置文件已调整为在容器网络中使用服务名（如 `postgres`, `minio`, `redis`）作为 host；如果你需要在宿主机上运行服务并连接宿主机上的数据库，请把 `/services/*/config/*.yaml` 中的 host 改回 `localhost` 或使用适当的环境配置/挂载覆盖。

停止容器并清理：

```bash
# 停止服务容器, 自动删除容器和.env, -v 参数删除数据库挂载卷，--rmi local参数删除本地构建的image
./deploy/bin/run down
```

### kind 本地集群CI测试
1. `kind create cluster`创建集群
    kind为容器嵌套结构，使用容器来代替一个具核node，目前已知在linux上`kind create cluster`时kind会将proxy环境变量透传到node中，使得node中一切网络活动失效，影响集群运行，所以应当进入容器取消环境变量或直接关闭代理，使用TUN模式，在Windows+Docker desktop + WSL2平台可直接在cmd中使用集群创建指令，而不会有网络代理问题

2. `docker compose -f ./backend/deploy/docker/docker-compose.service.yml build`生成容器
3. docker tag 添加tag，如`docker tag kind-user-service:latest user-service:latest`
4. kind load到node中，如`kind load user-service:latest`
5. `kubectl apply -f ./backend/deploy/docker/cluster/k8s.config`将各配置文件部署即可
