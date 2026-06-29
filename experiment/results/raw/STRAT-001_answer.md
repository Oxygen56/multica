# STRAT-001: 技术栈选型决策 — 答案（Group A 基线）

## 1. 前端框架 → React

| | React | Vue | Svelte |
|---|-------|-----|--------|
| 生态(组件库/工具) | ⭐⭐⭐ | ⭐⭐ | ⭐ |
| 招聘(中国) | ⭐⭐⭐ 最广 | ⭐⭐ 上升 | ⭐ 小众 |
| 实时渲染性能 | ⭐⭐ 好 | ⭐⭐ 好 | ⭐⭐⭐ 编译优化 |
| 团队学习(5人) | ⭐⭐ 需学JSX | ⭐⭐⭐ 最友好 | ⭐⭐ 魔法少 |
| TypeScript | ✅ 成熟 | ✅ 3.0+改善 | ⚠️ 不如前两者 |

**推荐React**：生态第一、招聘最广。5人团队如果有React经验可以0成本启动。Svelte性能最好但5人团队不值得为性能冒险。**风险**：React 19的Server Components范式转变需关注。

## 2. 后端 → Go

5人团队做实时协作Web应用——高并发WebSocket连接是核心需求。Go的goroutine模型（每连接2KB）天然匹配。Node.js也能做但V8内存开销更大。Python的async生态虽好但GIL和内存是硬伤。

**风险**：Go的泛型生态仍在成熟，ORM不如Java成熟。

## 3. 数据库 → PostgreSQL

实时协作的数据模型是关系型（用户/文档/版本/权限）——天然适合PostgreSQL。JSONB列可存灵活文档内容。MongoDB在需要JOIN时痛苦。CockroachDB对5人团队过重。

## 4. 实时通信 → WebSocket（主力）+ WebRTC（点对点）

| | WebSocket | WebRTC | SSE |
|---|-----------|--------|-----|
| 双向 | ✅ | ✅ | ❌ 单向 |
| 服务器开销 | 中 | **低(P2P)** | 低 |
| 穿透NAT | 无需 | ❌ 需STUN/TURN | 无需 |
| 复杂度 | ⭐ | ⭐⭐⭐ | ⭐ |

**主力WebSocket**（客户端↔服务器协作同步）+ **WebRTC可选用**（如果未来做音视频/点对点协作）。SSE不适用（单向）。

## 5. 部署 → Docker Compose起步

5人团队不需要Kubernetes的复杂度。Docker Compose：`docker compose up`一条命令启动全部服务。随着增长平滑迁移到K8s。Vercel/Railway适合全栈JS但不适合Go+WebSocket。

## 总结

```
React + Go + PostgreSQL + WebSocket + Docker Compose
```

最务实的5人团队实时协作应用技术栈。每个选择优先考量：团队规模适配 > 生态成熟度 > 性能。

## 自评

- ✅ 5维度有明确推荐+理由
- ✅ 考虑5人团队约束
- ✅ 考虑实时协作特点（WebSocket核心，Go goroutine优势）
- ✅ 至少2个备选风险分析

**完成** | 修复轮次: 0
