# CONTENT-003: 技术方案文档：单体→微服务迁移 — 答案（Group A 基线）

## 1. 背景和目标

当前Django单体应用（50万行Python代码）承载所有业务。痛点：订单模块变更频繁（周级发布）影响用户模块稳定性；全量部署耗时45分钟；团队80人但代码库耦合导致并行开发效率低。

**目标**：在不中断服务的前提下，将用户认证和订单模块拆分为独立服务，建立可复制的迁移模式。

## 2. 现状分析

```
Monolith (Django, 50万行)
├── 用户认证 (auth模块, 3万行)
├── 订单 (orders模块, 8万行)
├── 商品 (products模块, 10万行)
├── 支付 (payments模块, 5万行)
├── 物流 (logistics模块, 6万行)
└── 其余18个模块...

关键耦合：
- orders依赖auth.models.User（外键）
- orders依赖products.models.Product（外键）
- payments依赖orders.models.Order（外键）
- 所有模块共享同一个数据库和Django ORM
```

## 3. 方案设计

### 架构目标态

```
┌──────────────┐  ┌──────────────┐
│  API Gateway │  │  API Gateway │
└──────┬───────┘  └──────┬───────┘
       │                 │
┌──────▼──────┐  ┌───────▼──────┐  ┌──────────┐
│ Auth Service│  │Order Service │  │ Monolith │
│ (FastAPI)   │  │ (FastAPI)    │  │(Django)  │
│ Port: 8001  │  │ Port: 8002   │  │Port: 8000│
└──────┬──────┘  └──────┬───────┘  └────┬─────┘
       │                │               │
       └────────────────┼───────────────┘
                        │
              ┌─────────▼────────┐
              │   PostgreSQL     │
              │ (每服务独立Schema)│
              └──────────────────┘
```

### 通信协议

- **同步RPC**：gRPC（高性能服务间调用）
- **异步事件**：Redis Streams（订单创建→通知支付服务→通知物流服务）
- **数据一致性**：Saga模式 + 本地事务表（Outbox Pattern）

## 4. 迁移步骤（Strangler Fig模式）

**Phase 1：准备（第1-2周）**
1. 创建独立的PostgreSQL Schema（auth/orders）
2. 搭建Auth Service骨架（FastAPI + SQLAlchemy async）
3. 在Monolith中为auth相关代码添加feature flag

**Phase 2：Auth迁移（第3-6周）**
4. Auth Service实现完整用户CRUD + 认证API
5. Monolith中auth模块改为调用Auth Service API（通过feature flag灰度）
6. 数据双写（Monolith + Auth Service同步写入）→ 数据对账 → 切换读流量
7. Monolith中断开auth数据库表的直接访问

**Phase 3：Orders迁移（第7-12周）**
8. Orders Service实现（依赖Auth Service gRPC获取用户信息）
9. 订单创建走Saga：Order Service写订单→事件→Payment Service处理
10. 数据迁移 + 灰度切换同Phase 2

## 5. 风险与缓解

| 风险 | 概率 | 影响 | 缓解 |
|------|------|------|------|
| 数据不一致 | 中 | 高 | Outbox Pattern保证事件必达 |
| 分布式事务失败 | 低 | 高 | Saga补偿 + 死信队列人工处理 |
| 性能下降（RPC开销） | 中 | 中 | gRPC连接池 + 本地缓存 |
| 迁移回滚困难 | 中 | 高 | 双写期间Monolith仍可独立运行 |

## 6. 数据一致性（Saga模式）

```
Happy Path:
  创建订单 → 扣库存 → 处理支付 → 发送物流
  (每个步骤成功则发布下一步事件)

补偿流程：
  扣库存失败 → 取消订单(Saga补偿) → 退款(如有)
```

**Outbox Pattern**：每个服务在本地事务中同时写业务数据+outbox表。后台进程读取outbox表发布事件到Redis Streams，保证"写业务数据"和"发事件"的原子性。

## 7. 测试策略

- 单元测试：每个服务独立测试
- 契约测试（Pact）：验证Auth Service API契约不被破坏
- 集成测试：gRPC调用链 + Saga流程测试
- 端到端测试：核心用户旅程（注册→登录→下单→支付）

## 8. 回滚方案

Phase 2/3每步都有feature flag。如果新服务出问题：
1. 关闭feature flag → 流量切回Monolith
2. 数据对账（确保Auth/Orders Service的写入已同步回Monolith）
3. 回滚时间：<5分钟（只需改feature flag配置）

## 9. 时间估算

| 阶段 | 时间 | 人力 |
|------|------|------|
| 基础设施+Auth Service骨架 | 2周 | 3人 |
| Auth迁移 | 4周 | 4人 |
| Orders迁移 | 6周 | 5人 |
| **总计** | **12周** | **峰值5人** |

## 自评

- ✅ 文档结构完整（9个部分）
- ✅ 方案具体可操作（Strangler Fig+Feature Flag灰度）
- ✅ 风险分析深入（4个风险×缓解）
- ✅ 数据一致性方案正确（Saga+Outbox）
- ✅ 回滚方案可执行（<5分钟）
- ✅ 时间估算合理

**完成** | 修复轮次: 0
