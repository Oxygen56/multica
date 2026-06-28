# Multica 自部署指南 / Self-Hosting Deployment Guide

在自有服务器上部署 Multica 的完整指南。面向有 Docker 经验的运维人员和开发者。

## 目录

- [部署架构](#部署架构)
- [架构决策记录](#架构决策记录)
- [前置要求](#前置要求)
- [快速开始（推荐）](#快速开始推荐)
- [手动部署步骤](#手动部署步骤)
- [环境变量参考](#环境变量参考)
- [从源码构建](#从源码构建)
- [生产环境加固](#生产环境加固)
- [升级指南](#升级指南)
- [常见问题](#常见问题)

---

## 部署架构

```
┌──────────────────────────────────────────────────────┐
│                    用户机器                           │
│  ┌──────────────┐  ┌──────────────────────────────┐  │
│  │  multica CLI │  │  Agent Daemon (multica daemon)│  │
│  │  (管理工具)    │  │  - Claude Code / Codex / ...  │  │
│  └──────┬───────┘  └──────────────┬───────────────┘  │
└─────────┼──────────────────────────┼──────────────────┘
          │ HTTPS/WSS                │ WebSocket
          ▼                          ▼
┌──────────────────────────────────────────────────────┐
│                   服务器 (Docker)                      │
│                                                      │
│  ┌─────────────────┐  ┌────────────────────────────┐ │
│  │   Frontend       │  │   Backend (Go)             │ │
│  │   (Next.js 16)   │  │   - REST API               │ │
│  │   Port 3000      │  │   - WebSocket Server       │ │
│  │                  │  │   - Migration Runner       │ │
│  └────────┬─────────┘  │   Port 8080                │ │
│           │             └─────────────┬──────────────┘ │
│           │                           │                │
│           └───────────┬───────────────┘                │
│                       ▼                                │
│  ┌─────────────────────────────────────────────────┐  │
│  │   PostgreSQL 17 + pgvector                       │  │
│  │   Port 5432 (仅内部网络)                          │  │
│  │   数据卷: pgdata                                  │  │
│  └─────────────────────────────────────────────────┘  │
│                                                      │
│  ┌─────────────────────────────────────────────────┐  │
│  │   Redis（可选，水平扩展时需要）                      │  │
│  │   - 速率限制                                      │  │
│  │   - 实时消息 Pub/Sub                              │  │
│  └─────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

### 组件说明

| 组件 | 技术栈 | 职责 | 默认端口 |
|------|--------|------|----------|
| Backend | Go 1.26, Chi Router, pgx, Gorilla WebSocket | REST API、WebSocket、数据库迁移 | 8080 |
| Frontend | Next.js 16, React, TanStack Query | Web 管理界面 | 3000 |
| Database | PostgreSQL 17, pgvector | 主数据存储、向量检索 | 5432 |
| Redis | Redis 7 (可选) | 速率限制、跨节点 Pub/Sub | 6379 |

### 安全设计

所有服务默认绑定 `127.0.0.1`，不直接暴露到公网。生产环境需要在前面加一层反向代理（Caddy / nginx / Cloudflare Tunnel）来终止 TLS 并将流量转发到 `127.0.0.1:8080`（后端）和 `127.0.0.1:3000`（前端）。

> ⚠️ **不要**把服务绑定改为 `0.0.0.0`。Docker 默认绕过主机防火墙（UFW/iptables），改成 `0.0.0.0` 会让未认证端口暴露到公网。

---

## 架构决策记录

以下记录部署架构中的关键决策及其推理链。

### ADR-1：多容器部署（docker-compose）

**决策**：使用 docker-compose 管理三个独立容器（postgres + backend + frontend），而非单容器。

**推理链**：
- PostgreSQL 是独立的有状态服务，生命周期与无状态应用不同 → 必须独立容器
- Backend (Go) 和 Frontend (Next.js) 使用完全不同的运行时（Go vs Node.js），合并容器会引入不必要的依赖复杂度
- 独立容器支持独立扩缩容（虽然自部署场景通常单实例，但保留扩展可能）
- 独立健康检查和重启策略

**替代方案**：单容器（supervisord 管理多进程）。拒绝原因：违反单一职责、健康检查粒度粗、镜像膨胀。

**置信度**：高（已验证 — Dify、n8n、Plane 等同类平台均采用多容器模式）

### ADR-2：127.0.0.1 绑定

**决策**：所有服务端口绑定到 `127.0.0.1` 而非 `0.0.0.0`。

**推理链**：
- Docker 默认绕过 UFW/iptables 规则，`0.0.0.0` 映射会直接暴露端口到公网
- 默认 JWT_SECRET 和数据库密码需要用户主动修改 → 默认安全配置降低误操作风险
- 生产环境通过反向代理（nginx/Caddy）访问，反向代理通过 `127.0.0.1` 转发即可

**置信度**：高（Docker 网络安全最佳实践）

### ADR-3：Redis 可选

**决策**：Redis 作为可选依赖，不包含在基础 docker-compose 中。

**推理链**：
- 速率限制模块：REDIS_URL 未配置时降级为 no-op（允许所有请求）
- 实时消息：REDIS_URL 未配置时使用进程内 Pub/Sub（单节点模式，功能完整）
- 自部署用户多为单节点，不需要 Redis Pub/Sub
- 需要水平扩展的用户可自行添加 Redis 服务

**替代方案**：始终要求 Redis。拒绝原因：增加单节点部署的运维负担，与「快速开始」设计目标冲突。

**置信度**：高（已验证 — 后端启动日志明确区分 in-memory hub vs Redis relay 模式）

### ADR-4：多阶段构建 + 预构建镜像双通道

**决策**：提供两条镜像获取路径：（1）从 GHCR 拉取官方预构建镜像（默认）；（2）从源码本地构建（`make selfhost-build`）。

**推理链**：
- 大多数用户应使用官方发布的稳定版本 → 拉取 GHCR 镜像最快
- 开发者和需要自定义版本的用户需要从源码构建 → 提供 docker-compose 构建覆盖文件
- 两条路径通过同一个 docker-compose.selfhost.yml 管理，通过 `MULTICA_BACKEND_IMAGE` / `MULTICA_WEB_IMAGE` 环境变量切换

**置信度**：高（社区标准实践）

### ADR-5：Go 多阶段构建优化

**决策**：backend Dockerfile 使用 `golang:1.26-alpine` 构建 → `alpine:3.21` 运行。

**推理链**：
- 先复制 `go.mod` / `go.sum` 并执行 `go mod download`，利用 Docker 层缓存加速重复构建
- `CGO_ENABLED=0` 静态链接，无需 glibc
- `-ldflags "-s -w"` 去除调试信息，减小二进制体积
- 运行阶段只保留二进制 + 迁移文件，镜像体积小

**置信度**：高（Go 社区标准实践）

### ADR-6：自动数据库迁移

**决策**：通过 entrypoint.sh 在每次容器启动时自动执行 `migrate up`。

**推理链**：
- Multica 的 migrate 命令是幂等的（已执行的迁移会跳过）
- 避免了「先手动跑迁移再启动服务」的操作顺序错误
- 启动时间增加通常 < 1s（无新迁移时）

**替代方案**：独立 migration 容器或手动执行。拒绝原因：增加运维复杂度，容易忘记。

**置信度**：高（已验证 — 当前运行实例使用此方案）

---

## 前置要求

### 必需

- **Docker** ≥ 24.0（[安装指南](https://docs.docker.com/engine/install/)）
- **Docker Compose v2**（`docker compose` 命令可用，**不是**旧版 `docker-compose`）
  ```bash
  docker compose version  # 应显示 v2.x
  ```
- **1 GB+** 可用内存（PostgreSQL + Backend + Frontend 合计约 500 MB 基线）
- **2 GB+** 可用磁盘空间（不含数据增长）

### 可选

- **反向代理**（生产环境强烈建议）：Caddy / nginx / Traefik / Cloudflare Tunnel
- **SMTP 服务**（用于发送验证码邮件）：Resend（推荐）或自建 SMTP
- **Redis**（水平扩展时）：Redis 7+

---

## 快速开始（推荐）

一条命令完成安装：

```bash
curl -fsSL https://raw.githubusercontent.com/multica-ai/multica/main/scripts/install.sh | bash -s -- --with-server
```

脚本自动完成：
1. 检查 Docker 和 Docker Compose 是否已安装
2. 拉取 Multica 官方镜像（backend + frontend）
3. 生成 `.env` 配置文件（含随机 JWT_SECRET 和数据库密码）
4. 启动所有服务
5. 安装 `multica` CLI

完成后打开 http://localhost:3000。

---

## 手动部署步骤

### 步骤 1：获取代码和配置

```bash
git clone https://github.com/multica-ai/multica.git
cd multica
make selfhost
```

`make selfhost` 自动完成：
1. 从 `.env.example` 创建 `.env`（如不存在）
2. 生成随机 `JWT_SECRET` 和 `POSTGRES_PASSWORD`
3. 拉取官方 GHCR 镜像
4. 启动所有服务（postgres + backend + frontend）

如果 GHCR 镜像不可用（例如尚未发布），使用本地构建：

```bash
make selfhost-build
```

### 步骤 2：登录

打开 http://localhost:3000。

**推荐方式（生产环境）**：在 `.env` 中配置 `RESEND_API_KEY`，重启 backend 后可以使用真实邮箱验证码登录。详见[邮件配置](#smtp-邮件配置)。

**测试方式（无邮件服务）**：不配置 Resend，验证码会打印在 backend 容器日志中：
```bash
docker compose -f docker-compose.selfhost.yml logs backend | grep "Verification code"
```

**本地开发方式**：在 `.env` 中设置：
```
APP_ENV=development
MULTICA_DEV_VERIFICATION_CODE=888888
```
> ⚠️ 此方式仅用于本地私有测试，绝不要在公网可达实例上使用。

### 步骤 3：安装 CLI 和启动 Agent Daemon

每个需要运行 AI Agent 的团队成员：

```bash
# macOS / Linux
brew install multica-ai/tap/multica

# 配置并连接到自部署服务器
multica setup self-host
```

需要至少安装一个 AI Agent CLI：
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) (`claude` on PATH)
- [Codex](https://github.com/openai/codex) (`codex` on PATH)
- [GitHub Copilot CLI](https://docs.github.com/en/copilot) (`copilot` on PATH)
- [Gemini CLI](https://github.com/google-gemini/gemini-cli) (`gemini` on PATH)

### 步骤 4：手动 Docker Compose（备选）

如果不使用 Makefile：

```bash
# 1. 复制并编辑环境变量
cp .env.example .env
# 编辑 .env — 至少修改 JWT_SECRET 和 POSTGRES_PASSWORD

# 2. 启动（使用预构建镜像）
docker compose -f docker-compose.selfhost.yml up -d

# 或：从源码构建
docker compose -f docker-compose.selfhost.yml -f docker-compose.selfhost.build.yml up -d
```

---

## 环境变量参考

### 必需配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `JWT_SECRET` | JWT 签名密钥（至少 32 字符随机串） | `change-me-in-production` |
| `POSTGRES_PASSWORD` | 数据库密码 | `multica` |
| `DATABASE_URL` | PostgreSQL 连接字符串 | `postgres://multica:multica@localhost:5432/multica?sslmode=disable` |

### SMTP 邮件配置

| 变量 | 说明 | 必填 |
|------|------|------|
| `RESEND_API_KEY` | Resend API Key（推荐） | 否 |
| `RESEND_FROM_EMAIL` | 发件人地址 | 否 |
| `SMTP_HOST` | 自建 SMTP 服务器地址 | 否 |
| `SMTP_PORT` | SMTP 端口 | 否 |

### OAuth 登录（可选）

| 变量 | 说明 |
|------|------|
| `GOOGLE_CLIENT_ID` | Google OAuth Client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth Client Secret |
| `GOOGLE_REDIRECT_URI` | OAuth 回调地址 |

### 文件存储（可选）

默认存储在后端容器的本地文件系统。上传到 `backend_uploads` 卷。生产环境建议使用 S3：

| 变量 | 说明 |
|------|------|
| `S3_BUCKET` | S3 存储桶名称 |
| `S3_REGION` | S3 区域 |
| `AWS_ACCESS_KEY_ID` | AWS 访问密钥 |
| `AWS_SECRET_ACCESS_KEY` | AWS 秘密密钥 |
| `AWS_ENDPOINT_URL` | S3 兼容端点（MinIO 等） |

### 访问控制

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `ALLOW_SIGNUP` | 允许新用户注册 | `true` |
| `ALLOWED_EMAILS` | 白名单邮箱（逗号分隔） | 无限制 |
| `ALLOWED_EMAIL_DOMAINS` | 白名单邮箱域名 | 无限制 |
| `DISABLE_WORKSPACE_CREATION` | 禁止用户创建工作区 | 允许 |

### 服务端口

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `BACKEND_PORT` | Backend 主机端口 | `8080` |
| `FRONTEND_PORT` | Frontend 主机端口 | `3000` |

### 镜像通道

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MULTICA_IMAGE_TAG` | 镜像版本标签 | `latest` |
| `MULTICA_BACKEND_IMAGE` | Backend 镜像地址 | `ghcr.io/multica-ai/multica-backend` |
| `MULTICA_WEB_IMAGE` | Frontend 镜像地址 | `ghcr.io/multica-ai/multica-web` |

完整环境变量列表见 `.env.example`。

---

## 从源码构建

### 构建 Backend 镜像

```bash
docker build -t multica-backend:dev -f Dockerfile .
```

多阶段构建内容：
1. **构建阶段**（`golang:1.26-alpine`）：下载依赖、编译 server / multica CLI / migrate / backfill 等二进制
2. **运行阶段**（`alpine:3.21`）：仅保留二进制 + 数据库迁移文件 + entrypoint

### 构建 Frontend 镜像

```bash
docker build -t multica-web:dev -f Dockerfile.web .
```

多阶段构建内容：
1. **依赖阶段**（`node:22-alpine`）：pnpm install（利用缓存）
2. **构建阶段**（`node:22-alpine`）：Next.js standalone build
3. **运行阶段**（`node:22-alpine`）：仅保留 standalone 输出

### 一键构建 + 启动

```bash
make selfhost-build
```

使用构建覆盖文件 `docker-compose.selfhost.build.yml`，以本地构建镜像替代 GHCR 拉取。

---

## 生产环境加固

### 1. 反向代理 + TLS

以 Caddy 为例（`Caddyfile`）：

```
multica.example.com {
    reverse_proxy 127.0.0.1:3000
    reverse_proxy /api/* 127.0.0.1:8080
    reverse_proxy /ws 127.0.0.1:8080
}
```

### 2. 环境变量安全检查

```bash
# .env 文件权限
chmod 600 .env

# 必须修改的变量
JWT_SECRET=<openssl rand -hex 32 的结果>
POSTGRES_PASSWORD=<openssl rand -hex 24 的结果>
```

### 3. 数据库备份

```bash
# 通过 docker compose 执行 pg_dump
docker compose -f docker-compose.selfhost.yml exec postgres \
  pg_dump -U multica multica > backup_$(date +%Y%m%d).sql
```

建议配置定时备份（cron + 对象存储）。

### 4. 资源限制

在 `docker-compose.selfhost.yml` 中添加：

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 128M
```

### 5. 日志管理

在 `docker-compose.selfhost.yml` 中配置日志轮转：

```yaml
services:
  backend:
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "5"
```

---

## 升级指南

### 使用预构建镜像（推荐）

```bash
# 拉取新版本
docker compose -f docker-compose.selfhost.yml pull

# 重启服务（数据库迁移自动运行）
docker compose -f docker-compose.selfhost.yml up -d
```

### 从源码构建最新版

```bash
git pull origin main
make selfhost-stop
make selfhost-build
```

### 版本锁定

在 `.env` 中固定版本号：
```
MULTICA_IMAGE_TAG=v0.2.4
```

防止 `latest` 标签自动升级导致意外变更。

---

## 常见问题

### Q: 如何查看服务状态？

```bash
docker compose -f docker-compose.selfhost.yml ps
```

健康状态：`(healthy)` 表示服务正常。

### Q: Backend 启动后立即退出？

```bash
# 查看日志
docker compose -f docker-compose.selfhost.yml logs backend

# 常见原因：
# 1. 数据库连接失败 → 检查 DATABASE_URL 和 postgres 是否 healthy
# 2. 端口被占用 → 修改 BACKEND_PORT
# 3. 迁移失败 → 检查 postgres 中是否已有冲突数据
```

### Q: 前端页面显示「无法连接到服务器」？

检查反向代理是否正确转发 WebSocket 连接（`/ws` 路径需要 Upgrade 支持）。

### Q: 镜像拉取失败（GHCR 不可用）？

使用本地构建：
```bash
make selfhost-build
```

### Q: 如何从旧版本升级而不丢失数据？

数据库数据存储在 `pgdata` 卷中，重启不会丢失。升级只需拉取新镜像并重建容器：
```bash
docker compose -f docker-compose.selfhost.yml pull
docker compose -f docker-compose.selfhost.yml up -d
```

### Q: Redis 什么时候需要？

单节点部署（大多数自部署场景）不需要 Redis。以下情况需要：
- 多节点水平扩展（多个 backend 实例共享实时消息）
- 需要 IP 级别的速率限制（而非默认的 fail-open 模式）

添加 Redis 后在 `.env` 中设置 `REDIS_URL=redis://redis:6379/0`。

### Q: 如何配置 HTTPS？

不要在 Docker 内部配置 TLS。在前面加反向代理（Caddy / nginx / Cloudflare Tunnel），由反向代理终止 TLS，转发 HTTP 到容器的 `127.0.0.1` 端口。

### Q: 服务端口冲突怎么办？

在 `.env` 中修改：
```
BACKEND_PORT=8081
FRONTEND_PORT=3001
```

重启服务即可生效。

---

## 停止和清理

```bash
# 停止所有服务（保留数据）
make selfhost-stop
# 或
docker compose -f docker-compose.selfhost.yml down

# 停止并删除所有数据卷
docker compose -f docker-compose.selfhost.yml down -v
```

---

## 更多资源

- [完整自托管指南](SELF_HOSTING.md) — 详细步骤和 CLI 安装
- [高级配置](SELF_HOSTING_ADVANCED.md) — 邮件、OAuth、S3、飞书集成
- [AI Agent 配置](SELF_HOSTING_AI.md) — Agent Daemon 和 AI CLI 设置
- [贡献指南](CONTRIBUTING.md) — 参与开发
