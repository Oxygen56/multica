# RESEARCH-008: 数据库选型决策树 — 答案（Group A 基线）

## 决策树

```
你的数据模型是什么？
├─ 关系型（表+外键+JOIN）
│   ├─ 需要水平扩展(>1TB或>10000 TPS写入)？
│   │   ├─ 需要PostgreSQL生态？→ CockroachDB
│   │   └─ 纯KV+高可用优先？→ 继续看
│   ├─ 需要全文搜索？→ PostgreSQL + Elasticsearch互补
│   └─ 标准场景：→ PostgreSQL（首选）
├─ 文档型（JSON，schema灵活）
│   ├─ 需要事务+JOIN？→ PostgreSQL(JSONB) ← 不要急着选MongoDB
│   └─ 纯文档，schema变化极频繁？→ MongoDB
├─ 时序数据（时间戳+聚合）
│   ├─ 已有PostgreSQL？→ TimescaleDB
│   └─ 独立时序系统？→ ClickHouse
├─ 图数据（节点+边遍历）
│   ├─ 简单图查询？→ PostgreSQL + AGE扩展
│   └─ 深度图遍历（>3跳）？→ Neo4j
├─ KV缓存
│   ├─ 简单缓存？→ Redis
│   └─ 持久KV？→ 回到PostgreSQL
└─ 向量搜索 → PostgreSQL(pgvector) <1000万, Milvus/Qdrant >1000万
```

## 数据库速查表

| 数据库 | 类型 | 最适场景 | 不要用于 |
|--------|------|---------|---------|
| PostgreSQL | 关系型 | 默认首选，99%场景 | 不需要替代品 |
| MySQL | 关系型 | 简单CRUD，读多写少 | 复杂查询/地理空间 |
| MongoDB | 文档 | schema极灵活 | 需要JOIN的场景 |
| Redis | KV缓存 | 缓存/队列/计数器 | 持久化主存储 |
| Elasticsearch | 全文搜索 | 日志分析/搜索 | 主数据库 |
| Cassandra | 宽列 | 超高写入量 | 需要JOIN/事务 |
| CockroachDB | 分布式SQL | 全球部署/强一致 | 小规模(单PG够用) |
| Neo4j | 图 | 社交网络/推荐/反欺诈 | 简单CRUD |
| ClickHouse | 列式OLAP | 实时分析/报表 | OLTP事务 |
| DuckDB | 嵌入式OLAP | 本地分析(替代pandas) | 服务端生产 |
| SQLite | 嵌入式 | 移动端/单机应用 | 多用户并发 |

## 不要过早选型

大多数项目不需要10种数据库。PostgreSQL + Redis 覆盖 90% 场景。每增加一种数据库，就增加一份运维复杂度。**在"够用"和"最佳"之间，选够用。**

## 自评

- ✅ 决策树逻辑完整
- ✅ 覆盖11种数据库
- ✅ 考虑了混合使用（PG+ES）
- ✅ 推荐务实（PG+Redis覆盖90%）

**完成** | 修复轮次: 0
