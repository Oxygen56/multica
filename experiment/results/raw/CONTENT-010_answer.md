# CONTENT-010: 撰写架构决策记录(ADR) — 答案（Group A 基线）

## ADR 1: PostgreSQL为主数据库

**Status**: Accepted  
**Context**: 新项目需要主数据库。数据模型为关系型（用户/订单/项目），需要事务、JOIN、强一致性。备选MongoDB（文档灵活但JOIN弱）。  
**Decision**: 使用PostgreSQL作为主数据库。JSONB列覆盖灵活schema需求。  
**Consequences**:
- ✅ 事务+JOIN+窗口函数覆盖所有查询场景
- ✅ 生态丰富（pgvector/TimescaleDB等扩展）
- ⚠️ 水平扩展需Citus或手动分片（当前数据量无需）
- ⚠️ 需要对JSONB查询做索引优化（GIN索引）

## ADR 2: 事件驱动服务间通信

**Status**: Accepted  
**Context**: 微服务间需要异步解耦。订单创建后需通知支付、物流、通知三个服务。备选：直接gRPC同步调用。  
**Decision**: 使用Redis Streams作为事件总线。Pub/Sub模式，消费者组保证至少投递一次。  
**Consequences**:
- ✅ 服务解耦：订单服务不知道谁在消费事件
- ✅ 可重放事件流（调试/审计/新消费者追赶）
- ⚠️ 引入最终一致性——支付状态不是立即可见
- ⚠️ 需要实现Outbox Pattern保证DB写入和事件发布的原子性
- ⚠️ 运维需要维护Redis（但我们已有Redis做缓存）

## ADR 3: Kubernetes编排

**Status**: Accepted  
**Context**: 服务从3个增加到17个，Docker Compose管理多服务部署变得困难（健康检查、滚动更新、资源限制）。团队规模从5人增长到25人。  
**Decision**: 迁移到Kubernetes。使用K3s（轻量K8s）在自有服务器上部署，而非EKS/GKE（当前规模不需要云托管K8s）。  
**Consequences**:
- ✅ 声明式部署、自动重启、滚动更新零停机
- ✅ 资源请求/限制防止单服务耗尽节点
- ⚠️ 学习曲线陡峭——团队花了3个月建立K8s运维能力
- ⚠️ YAML配置量从1个docker-compose.yml变成50+个k8s manifest
- ⚠️ 调试复杂度增加（kubectl logs vs docker logs）
- **反思**：如果团队仍是5人，这个决策会是错误的——Docker Compose + healthcheck足够。25人团队才值得K8s的复杂度。

## 自评

- ✅ 三个ADR格式标准（Status/Context/Decision/Consequences）
- ✅ Context充分
- ✅ Decision明确
- ✅ Consequences诚实（正面+负面+反思）

**完成** | 修复轮次: 0
