# Traefik 集成部署指南

## 架构

```
浏览器
  ↓ (localhost:80)
┌─────────────────────────┐
│   Traefik 3.6.4         │
│   反向代理/负载均衡      │
├─────────────────────────┤
├→ /              → Nginx (前端静态文件)
├→ /api/v1/user   → user-service:8081
├→ /api/v1/forum  → forum-service:8082
├→ /api/v1/audit  → audit-service:8083
└→ /dashboard     → Traefik Dashboard (8080)
```

## 文件说明

- `docker-compose.service.yml` - 已更新，集成了 Traefik
- `nginx-frontend.conf` - Nginx 前端配置（SPA 回退）

## 前置准备

### 1. 构建前端

```bash
cd frontend
npm install
npm run build
```

这会生成 `frontend/dist` 目录。

### 2. 检查目录结构

```
backend/docker/
├── docker-compose.yml
├── docker-compose.service.yml
├── nginx-frontend.conf              # ← 新增
├── .env
└── ...
```

## 启动服务

```bash
cd backend

# 启动所有服务（基础设施 + 应用 + Traefik）
./bin/run up --build -d

# 或手动启动
cd docker
docker compose -p backend -f docker-compose.yml up -d
docker compose -p backend -f docker-compose.service.yml up --build -d
```

## 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| **前端应用** | http://localhost | 从 Nginx 提供的静态文件 |
| **用户服务 API** | http://localhost/api/v1/user | Traefik 转发 |
| **论坛服务 API** | http://localhost/api/v1/forum | Traefik 转发 |
| **审核服务 API** | http://localhost/api/v1/audit | Traefik 转发 |
| **Traefik 仪表板** | http://localhost:8080/dashboard | 监控路由和服务 |

## 工作流程

### 开发阶段

如果需要修改前端：

```bash
# 更新前端代码后重新构建
cd frontend
npm run build

# 重启 frontend 容器
docker compose -p backend restart frontend
```

### 负载均衡示例

如果要运行多个相同服务的实例（负载均衡）：

```yaml
# 在 docker-compose.service.yml 中添加
user-service-2:
  build:
    context: ..
    dockerfile: docker/Dockerfile
    target: production
    args:
      SERVICE: user
      EXPOSE_PORT: 8081
  restart: unless-stopped
  environment:
    GIN_MODE: release
  expose:
    - "8081"
  labels:
    - "traefik.enable=true"
    - "traefik.http.routers.user-service.rule=PathPrefix(`/api/v1/user`)"
    - "traefik.http.services.user-service.loadbalancer.server.port=8081"
    - "traefik.http.services.user-service.loadbalancer.server=user-service-2:8081"
  networks:
    - default
```

Traefik 会自动在两个实例间进行轮询负载均衡。

## Traefik 配置详解

### Router 规则

```yaml
labels:
  - "traefik.http.routers.user-service.rule=PathPrefix(`/api/v1/user`)"
```

- `PathPrefix` - 前缀匹配
- `Path` - 精确路径
- `Host` - 域名匹配
- 支持组合：`PathPrefix(`/api`) && Host(`api.example.com`)`

### 服务发现

```yaml
labels:
  - "traefik.http.services.user-service.loadbalancer.server.port=8081"
```

自动从 Docker 容器标签发现服务。

## 常用命令

```bash
# 查看 Traefik 日志
docker compose -p backend logs traefik

# 查看前端容器日志
docker compose -p backend logs frontend

# 重启 Traefik
docker compose -p backend restart traefik

# 进入前端容器
docker exec -it backend-frontend-1 sh

# 检查 Traefik 路由配置
curl http://localhost:8080/api/entrypoints
curl http://localhost:8080/api/routers
curl http://localhost:8080/api/services
```

## 生产环保议

1. **HTTPS 支持**：使用 Let's Encrypt 自动化证书
   ```yaml
   traefik:
     command:
       - "--entrypoints.websecure.address=:443"
       - "--certificatesresolvers.letsencrypt.acme.httpchallenge=true"
       - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
       - "--certificatesresolvers.letsencrypt.acme.email=your@email.com"
       - "--certificatesresolvers.letsencrypt.acme.storage=acme.json"
   ```

2. **身份验证**：添加中间件
   ```yaml
   labels:
     - "traefik.http.middlewares.auth.basicauth.users=admin:$$2y$$05$$..."
     - "traefik.http.routers.user-service.middlewares=auth"
   ```

3. **速率限制**：防止 DDoS
   ```yaml
   labels:
     - "traefik.http.middlewares.ratelimit.ratelimit.average=100"
     - "traefik.http.middlewares.ratelimit.ratelimit.burst=200"
   ```

## 故障排查

### 问题：无法访问前端

```bash
# 检查 Nginx 容器状态
docker ps | grep frontend

# 查看 Nginx 日志
docker compose -p backend logs frontend

# 检查 Nginx 配置是否正确
docker exec backend-frontend-1 nginx -t
```

### 问题：API 路由无效

```bash
# 查看 Traefik 路由配置
docker compose -p backend logs traefik

# 检查服务是否被正确发现
curl http://localhost:8080/api/services
```

### 问题：静态文件 404

- 确保 `frontend/dist` 存在且有 `index.html`
- 检查 volume 挂载：`docker inspect backend-frontend-1 | grep -A 5 Mounts`
- 验证 Nginx 配置中的 root 路径正确

## 与原方案的对比

| 特性 | Vite 代理 | Traefik |
|------|---------|---------|
| 开发效率 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| 生产就绪 | ⭐ | ⭐⭐⭐⭐⭐ |
| 负载均衡 | ❌ | ✅ |
| HTTPS/TLS | ❌ | ✅ |
| 监控仪表板 | ❌ | ✅ |
| 中间件支持 | 部分 | ✅ |
| 资源消耗 | 低 | 中等 |

## 切换回 Vite 开发

如果想恢复使用 Vite 开发模式（不用 Traefik）：

```bash
# 停止 Traefik 相关的容器
docker compose -p backend -f docker-compose.service.yml down

# 改用 watch.yml（开发热重载）
docker compose -p backend -f docker-compose.watch.yml up --build

# 前端仍使用 Vite 开发服务器
cd frontend
npm run dev
```
