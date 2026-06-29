# RESEARCH-001: 向量数据库技术调研 — 答案（Group A 基线）

## 场景：中小团队RAG，100万-1000万向量

## 方案对比

| | Pinecone | Weaviate | Milvus | Qdrant | Chroma | pgvector |
|---|---------|----------|--------|--------|--------|----------|
| 部署 | 仅SaaS | 自托管/SaaS | 自托管/Zilliz Cloud | 自托管/Cloud | 自托管 | PG扩展 |
| 索引 | 专有(性能未公开) | HNSW | IVF/HNSW/DISKANN | HNSW | HNSW | IVFFlat/HNSW |
| 过滤 | ⭐⭐⭐ | ⭐⭐⭐⭐ GraphQL | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ SQL |
| 运维 | ⭐⭐⭐⭐ 零运维 | ⭐⭐⭐ | ⭐⭐ 较复杂 | ⭐⭐⭐ | ⭐⭐⭐⭐ Pip install | ⭐⭐⭐⭐ 已有PG |
| 成本 | $$$ SaaS按量 | $$ 自托管 | 免费(开源) | $$ 自托管 | 免费 | 免费(已有PG) |
| QPS(100万,99%召回) | ~500 | ~300 | ~800 | ~400 | ~200 | ~200 |

## 推荐：pgvector（首选）/ Qdrant（备选）

**pgvector首选理由**：
1. 已有PostgreSQL的团队零额外运维——`CREATE EXTENSION pgvector`一条命令
2. SQL过滤+向量搜索在同一个查询中——不需要同步两个数据库
3. 100-1000万向量规模下，HNSW索引的性能与专用向量DB差距<2x，够用

**Qdrant备选**：如果向量规模增长到>1000万或需要极致性能（>1000 QPS），Qdrant的Rust实现+量化索引提供更好的性价比。

**不推荐**：
- Pinecone：中小团队成本过高（100万向量≈$70/月），且锁定SaaS
- Milvus：功能最强但运维复杂（依赖etcd/Pulsar/MinIO），小团队不值得

## 来源
- pgvector GitHub: https://github.com/pgvector/pgvector
- Qdrant benchmarks: https://qdrant.tech/benchmarks/
- Milvus docs: https://milvus.io/docs

## 自评

- ✅ 覆盖6个方案（含pgvector）
- ✅ 量化数据（QPS估算+成本）
- ✅ 推荐有场景限定（中小团队，已有PG）
- ✅ 不推荐方案说明理由

**完成** | 修复轮次: 0
