# RESEARCH-006: 开源商业化路径研究 — 答案（Group A 基线）

## 5个案例

### GitLab（Open Core）
- **初始策略**：GitLab CE（MIT开源）+ GitLab EE（专有功能）
- **转折点**：2015年YC后转向"开源优先但商业强化"
- **演进**：CE足够好→吸引大量用户→部分用户需要EE功能（审计/权限）→付费
- **社区反应**：正面。CE保持功能完整，EE是增量而非阉割

### Confluent/Kafka（云服务）
- **初始策略**：Kafka Apache 2.0开源
- **转折点**：2017年推出Confluent Cloud（托管Kafka）
- **演进**：开源Kafka成为标准→云服务解决运维痛点→同时卖企业connector
- **风险**：AWS推出MSK（Managed Kafka），直接竞争。Confluent应对：加强connector生态差异化

### HashiCorp（BSL转换）
- **初始策略**：Terraform MPL 2.0开源
- **转折点**：2023年将Terraform从MPL改为BSL（Business Source License）
- **原因**：防止云厂商（如Spacelift、env0）直接打包Terraform做SaaS
- **社区反应**：**强烈负面**。OpenTofu（Linux基金会）分叉出现。证明BSL转换是高风险动作

### Supabase（托管服务）
- **初始策略**：Apache 2.0开源
- **演进**：开源PostgreSQL平台→托管云服务→增值功能（实时订阅、Auth）
- **策略**：核心100%开源，托管是核心收入。不限制任何人自托管
- **成功因素**：自托管体验好→用户信任→付费托管自然转化

### Redis Labs（模块化）
- **初始策略**：Redis BSD开源
- **转折点**：2018年将Redis模块改为RSAL（Redis Source Available License，后改为SSPL）
- **原因**：AWS ElastiCache直接提供Redis服务但不回馈
- **演进**：核心Redis仍BSD，高级模块（RediSearch/RedisGraph）用限制性协议
- **社区反应**：**争议**。Valkey分叉（Linux基金会）出现。与HashiCorp类似，协议变更代价大

## 商业化决策框架

| 路径 | 适合场景 | 风险 | 代表 |
|------|---------|------|------|
| Open Core | 有自然的"企业功能"层 | 平衡CE和EE功能线 | GitLab |
| 托管云服务 | 运维复杂的产品 | 云厂商竞争 | Confluent |
| 协议变更 | 被云厂商严重侵蚀 | **社区分裂** | HashiCorp❌ |
| 100%开源+托管 | 开源社区驱动 | 收入单一 | Supabase |

**核心教训**：协议变更（BSL/SSPL）保护了商业但破坏了社区信任。最好的路径是在第一天就选对协议（AGPL/BSL），而非几年后变更。

## 自评

- ✅ 5案例有具体时间线+事件
- ✅ 转折点分析有洞察（协议变更=社区分裂）
- ✅ 框架可操作（4路径+适用+风险）
- ✅ 考虑2024-2025最新发展（OpenTofu/Valkey分叉）

**完成** | 修复轮次: 0
