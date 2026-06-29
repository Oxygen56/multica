# 第三方服务依赖风险评估报告

**文档编号**: STRAT-007
**评估日期**: 2026-06-29
**评估范围**: Auth0、Stripe、SendGrid、Mapbox、AWS S3、GitHub API

---

## 风险评估方法论

- **Likelihood（发生概率）**: 1-5 级（1=极低，5=极高）
- **Impact（影响程度）**: 1-5 级（1=可忽略，5=灾难性）
- **Risk Score** = Likelihood x Impact
- **风险等级**: 高风险（Score >= 15）、中风险（Score 9-14）、低风险（Score <= 8）
- 每个服务从以下维度评估：可用性、供应商锁定、定价变更、合规/数据隐私、迁移成本

---

## 汇总对比表

| 服务 | 功能 | 风险等级 | Risk Score (L×I) | 可用性风险 | 锁定风险 | 定价风险 | 合规风险 | 迁移成本 |
|------|------|---------|-------------------|-----------|---------|---------|---------|---------|
| Auth0 | 认证 | **高** | 16 (4×4) | 中 | 高 | 中 | 高 | 高 |
| Stripe | 支付 | **高** | 15 (3×5) | 低 | 高 | 中 | 高 | 高 |
| SendGrid | 邮件 | **中** | 9 (3×3) | 中 | 中 | 低 | 中 | 中 |
| Mapbox | 地图 | **低** | 6 (2×3) | 低 | 低 | 中 | 低 | 低 |
| AWS S3 | 存储 | **高** | 16 (4×4) | 低 | 高 | 中 | 中 | 高 |
| GitHub API | 代码集成 | **低** | 4 (2×2) | 低 | 低 | 低 | 低 | 低 |

---

## 1. Auth0（认证服务）

### 风险评估

| 维度 | 评分 | 说明 |
|------|------|------|
| Likelihood | 4 | 2023年曾发生多次区域性服务中断；Okta收购后产品路线不确定性增加 |
| Impact | 4 | 认证不可用 → 全部用户无法登录 → 业务完全停摆 |
| **综合风险** | **16 (4×4)** | **高风险** |

### 详细分析

**可用性风险**: Auth0 作为 SaaS 认证服务，其 SLA 为 99.9%（月），年化允许约 8.76 小时不可用。2023 年 Auth0/Okta 发生过几次影响广泛的中断事件，最长达数小时。认证是产品入口，任何中断都直接导致所有用户无法访问。

**供应商锁定风险**: **极高**。原因：
- 用户身份数据（用户名、密码哈希、社交登录绑定）完全存储在 Auth0 体系内
- 用户迁移不是简单的数据导出，涉及密码哈希算法兼容性、社交登录 token 迁移、用户通知和重新授权
- Auth0 的 Rules/Hooks/Actions 自定义逻辑使用了 Auth0 专有 API，不可直接迁移
- 迁移期间必然存在认证服务的并行运行期，工程复杂度极高

**定价变更风险**: 中。Okta 收购 Auth0 后，定价模式可能向企业级倾斜。Auth0 按 MAU（月活用户）计费，用户规模增长时成本非线性上升。Free tier 有 7,000 MAU 限制，超出后 B2C Essential 计划从 $240/月起跳，成本增长显著。

**合规/数据隐私风险**: **高**。认证服务处理用户 PII（邮箱、用户名、IP 地址、登录行为日志）。需确认：
- Auth0 的数据存储区域是否符合 GDPR 要求（当前支持 EU/US/AU 区域）
- 中国区业务需考虑数据本地化要求，Auth0 在中国大陆无节点
- SAML/OIDC 证书管理涉及密钥安全，证书过期未轮换会导致 SSO 全部失败

### 缓解策略

1. **架构层面——认证接口抽象层（IAM Gateway）**: 在应用与 Auth0 之间建立认证抽象层（如基于 OAuth2/OIDC 标准协议的统一认证网关），使后端不直接依赖 Auth0 SDK。这样切换 IdP 时只需修改网关配置，不改业务代码。这是最核心的防锁定措施。

2. **用户数据定期导出备份**: 通过 Auth0 Management API 或 User Export 功能，定期（建议每周）将用户 profile 数据导出到自有存储（S3），确保始终拥有完整的用户数据副本。

3. **多 IdP 联邦设计**: 同时在 Auth0 中配置至少 2 个社交登录 provider（如 Google + GitHub），避免单一社交登录渠道中断导致部分用户无法登录。

4. **本地 Token 验证缓存**: JWT 验证引入本地公钥缓存（JWKS 缓存），即使 Auth0 的 JWKS endpoint 短暂不可达，已登录用户的 token 验证仍可继续（利用缓存公钥 + token 有效期窗口）。

5. **降级模式——本地账号兜底**: 除 Auth0 托管的登录页外，维护一套简化版本地登录表单作为紧急降级通道，通过 Auth0 Authentication API（非 Universal Login）调用，避免依赖 Auth0 托管页面。

6. **监控与告警**: 对 Auth0 的 `/jwks.json`、`/authorize`、`/oauth/token` 等关键端点做主动健康检查，异常时 2 分钟内告警。

### Fallback 方案（高风险必须）

**触发条件**: Auth0 服务不可用超过 15 分钟，或 Auth0 发布安全漏洞公告要求立即停止使用。

**Fallback 架构——自建 Ory 认证栈**:

| 阶段 | 操作 | 预计恢复时间 |
|------|------|-------------|
| 1. DNS 切换 | 将 auth.example.com 指向自建 Ory Kratos 实例 | 5 分钟（TTL 已预设为 60s） |
| 2. 数据导入 | 将预先导出的 Auth0 用户数据批量导入 Kratos 身份数据库 | 30 分钟（脚本已就绪） |
| 3. 密码重置 | 由于密码哈希算法差异，通知全部用户重置密码 | 批量邮件发送 |
| 4. 业务验证 | 验证登录、注册、社交登录、MFA 流程 | 15 分钟 |

**前置准备（当前就绪）**:
- Ory Kratos + Ory Hydra 已在 Kubernetes 集群中部署（standby 模式，资源占用极小）
- 用户数据导出脚本已通过 cron 每周执行
- DNS TTL 已设置为 60 秒
- 密码重置通知邮件模板已就绪
- Runbook 文档已编写（包含完整切换和回切流程）

**Fallback 局限性**:
- 用户需重置密码，体验有损
- 社交登录关联需用户重新绑定
- MFA 设备需用户重新注册
- 回切到 Auth0 的流程同样需要密码重置

---

## 2. Stripe（支付服务）

### 风险评估

| 维度 | 评分 | 说明 |
|------|------|------|
| Likelihood | 3 | Stripe 基础设施极其成熟，重大中断罕见，但 API 版本升级 / 合规政策变更是持续风险 |
| Impact | 5 | 支付不可用 → 零收入 + 订单丢失 + 用户信任受损 |
| **综合风险** | **15 (3×5)** | **高风险** |

### 详细分析

**可用性风险**: Stripe 的工程基础设施是行业标杆，SLA 达 99.99%+。历史上重大全站中断极为罕见（2023 年约 2 次、每次 < 30 分钟）。但 Stripe API 版本升级（每年数次）可能导致未及时迁移的集成出现部分功能故障。

**供应商锁定风险**: **高**。原因：
- PCI DSS 合规认证依赖 Stripe 的 tokenization（Stripe.js / Elements），更换支付提供商需要重新做 PCI 合规审计
- 已保存的支付方式（Saved Payment Methods / Setup Intents）存储在 Stripe，无法直接迁移到其他提供商
- Webhook 事件处理逻辑深度耦合 Stripe 事件模型（`payment_intent.succeeded`、`checkout.session.completed` 等），换提供商需重写全部 Webhook handler
- 订阅（Subscription）和发票（Invoice）逻辑与 Stripe Billing 深度绑定

**定价变更风险**: 中。Stripe 定价公开透明（2.9% + $0.30/笔），历史上未出现大幅涨价。但以下场景值得关注：
- 国际卡附加费（1.5%）随业务国际化自动触发
- 争议处理费（$15/次）在高争议率场景下会显著增加成本
- Stripe Billing 的订阅管理功能按 0.5%-0.8% 额外收费，随收入增长

**合规/数据隐私风险**: **高**。
- PCI DSS Level 1 合规——使用 Stripe.js/Elements 时合规责任较轻（SAQ A），但如果自建支付表单则升级为 SAQ D
- 支付数据涉及用户银行卡信息，GDPR 明确覆盖
- 跨境支付涉及不同司法管辖区的资金流转合规（如中国的外汇管制）
- Stripe 要求服务端不记录完整的卡号/CVC，代码中需确保日志不泄露 PAN

### 缓解策略

1. **支付接口抽象层**: 定义内部统一的支付接口（CreatePayment、Refund、CreateSubscription 等），Stripe 作为其实现之一。接口与实现分离，降低 SDK 耦合。

2. **幂等性设计**: 所有支付请求自带幂等 key（`Idempotency-Key`），确保 Stripe 超时或网络故障后重试不会重复扣款。这是 Stripe 官方强烈推荐的最佳实践。

3. **本地订单状态机**: 支付状态以本地数据库的订单状态机为准，不依赖 Stripe Webhook 作为唯一状态来源。Webhook 仅作为状态变更的辅助信号，关键状态通过 Stripe API 主动查询确认。

4. **Webhook 签名验证 + 重试队列**: 所有 Stripe Webhook 必须验证签名（`Stripe-Signature` header）。Webhook 处理失败的消息进入本地重试队列，避免丢失关键事件。

5. **定期 PCI 合规自查**: 每季度检查代码库确保无 PAN/CVC 泄露到日志、错误报告、数据库。使用自动化扫描工具（如 detect-secrets）检查。

6. **备用支付方式——加密货币支付**: 在支付接口抽象层中预集成一个轻量级加密支付通道（如通过 Coinbase Commerce），作为极端情况下的支付降级通道。此通道覆盖 Stripe 不支持地区的用户。

### Fallback 方案（高风险必须）

**触发条件**: Stripe 服务全站中断超过 30 分钟，或 Stripe 账户因合规原因被冻结。

**Fallback 架构——备用支付提供商切换**:

| 阶段 | 操作 | 预计恢复时间 |
|------|------|-------------|
| 1. 激活备用 provider | 配置开关将支付路由切换到 Adyen（已预集成） | 即时（feature flag） |
| 2. 前端切换 | 前端支付表单切换为 Adyen 组件 | 配置推送，CDN 缓存刷新后 5 分钟生效 |
| 3. 功能验证 | 端到端测试：下单 → 支付 → Webhook → 订单完成 | 15 分钟 |
| 4. 用户通知 | 已保存支付方式失效，通知用户重新绑卡 | 批量通知 |

**前置准备（当前就绪）**:
- Adyen 商户账号已开通（test mode），production 审批已预提交
- Payment Gateway 抽象层已实现 Stripe 和 Adyen 两个 adapter
- 前端支付组件支持动态 provider 切换（feature flag `payment_provider`）
- Webhook handler 已适配 Adyen 事件模型

**Fallback 局限性**:
- 已保存的信用卡 token 无法从 Stripe 迁移到 Adyen，用户需重新绑卡
- 当前订阅（Subscription）需人工迁移
- 结算账期和费率差异可能导致短期财务对账困难
- 跨境支付覆盖区域可能少于 Stripe

---

## 3. SendGrid（邮件服务）

### 风险评估

| 维度 | 评分 | 说明 |
|------|------|------|
| Likelihood | 3 | 邮件服务中断偶有发生，通常影响投递延迟而非完全不可用 |
| Impact | 3 | 验证邮件/通知延迟影响用户体验但不阻塞核心功能 |
| **综合风险** | **9 (3×3)** | **中风险** |

### 详细分析

**可用性风险**: SendGrid SLA 为 99.9%。实际中断多为区域性投递延迟而非全站不可用。但邮件投递的黑洞问题（邮件被接收方静默丢弃）比服务中断更难检测。SendGrid IP 池中个别 IP 被列入黑名单可能导致部分用户收不到邮件。

**供应商锁定风险**: **中**。
- 邮件服务 API 标准化程度高（SMTP 协议 + REST API），切换提供商主要改 API key 和 endpoint
- 模板（Dynamic Templates）存在一定锁定——SendGrid 的模板语法（Handlebars）不完全与 Mailgun/Mandrill 兼容，需重写
- 发件域名验证（SPF/DKIM/DMARC）需在新 provider 的 DNS 记录中重新配置

**定价变更风险**: **低**。SendGrid 免费版 100封/天，Essentials 计划 $19.95/月起（50K封/月）。竞品（Mailgun、Amazon SES、Resend）价格接近且透明，SendGrid 大幅提价空间有限。

**合规/数据隐私风险**: **中**。
- 邮件内容可能包含用户 PII（用户名、邮箱、订单信息）
- 邮件开启追踪（Open Tracking）涉及用户行为追踪，GDPR 要求 opt-in
- 收件人邮箱地址存储在 SendGrid，属于跨境数据传输
- SendGrid（Twilio 旗下）的服务器位于美国，需评估 Schrems II 对 EU-US 数据传输的影响

### 缓解策略

1. **邮件发送抽象层**: 在应用内定义统一的 MailService 接口（Send、SendTemplate、SendBulk），SendGrid 作为其实现。接口参数使用内部模型，不暴露 SendGrid 专有概念。

2. **多 provider 路由**: 配置主备双 provider（SendGrid 主 + AWS SES 备），投递失败自动切换。关键事务邮件（密码重置、邮箱验证）双通道同时发送，取先到达者。

3. **邮件投递监控**: 实现投递率、打开率、点击率、退信率监控 dashboard。设置退信率阈值告警（> 5% 触发调查）。通过 Webhook 接收 bounce/complaint 事件，自动加入 suppression list。

4. **模板与内容分离**: 邮件模板使用自有的模板渲染引擎（如 MJML + Handlebars 在服务端渲染），不依赖 SendGrid Dynamic Templates。模板存储在自有代码仓库。

5. **SPF/DKIM/DMARC 预配置多个 provider**: DNS 中同时配置 SendGrid 和 SES 的 SPF/DKIM 记录，确保切换 provider 时邮件签名验证不受影响。

6. **GDPR 合规**: 关闭 SendGrid 的默认 Open Tracking（或改为 opt-in），邮件模板中的追踪像素仅在用户同意下启用。Suppression list 定期与自有 unsubscribe 数据库同步。

### Fallback 方案

**触发条件**: SendGrid 投递率下降到 < 80% 或 API 连续失败 10 分钟。

**Fallback——切换到 AWS SES**:
- 通过 feature flag 将邮件路由切换到 SES adapter
- DNS 记录已预配置，具备完整 SPF/DKIM
- 预计切换时间：< 5 分钟
- 用户无感知（发件人地址不变）

---

## 4. Mapbox（地图服务）

### 风险评估

| 维度 | 评分 | 说明 |
|------|------|------|
| Likelihood | 2 | Mapbox 服务稳定，中断罕见；地图加载为静态资源，CDN 缓存可缓冲 |
| Impact | 3 | 地图不可用影响产品体验但非致命——功能降级但不阻塞核心流程 |
| **综合风险** | **6 (2×3)** | **低风险** |

### 详细分析

**可用性风险**: Mapbox 在全球有完善的 CDN 分发网络，瓦片和样式服务极少全站中断。地图为前端静态资源加载，可被 CDN 和浏览器缓存缓冲。即使 Mapbox API 短时间不可用，已加载的地图瓦片仍可渲染。

**供应商锁定风险**: **低**。
- Mapbox 使用 Mapbox GL JS 专有库，但与 MapLibre GL JS（Mapbox GL JS v1 的 Apache 2.0 开源 fork）API 高度兼容
- 地图样式（Style JSON）是 Mapbox 专有格式，但可导出和迁移
- 切换主要涉及更换 JS 库 + 修改 tile server endpoint

**定价变更风险**: **中**。Mapbox 按请求量计费：
- 免费额度：50K 地图加载/月（mobile）或 50K Web 加载/月
- 超出后按百万次计费，高流量场景下成本增长显著
- 2021 年 Mapbox 曾调整定价结构（v1 → v2），部分用户成本上涨 3-10 倍
- 竞争者（MapTiler、Google Maps、自建方案）可作为定价谈判筹码

**合规/数据隐私风险**: **低**。
- 地图展示不涉及用户 PII
- 地理位置数据（用户标记的地点）在服务端处理，Mapbox 仅接收瓦片请求
- 注意：Mapbox 的 Telemetry SDK 默认收集使用数据，需在产品中禁用（`mapboxgl.accessToken` 后调用 `mapboxgl.setRTLTextPlugin` 等相关设置）

### 缓解策略

1. **库层面切换到 MapLibre GL JS**: 从 Mapbox GL JS v2（专有协议）迁移到 MapLibre GL JS（BSD 协议开源 fork），API 兼容度 > 95%。这是最有效的防锁定措施——MapLibre 不绑定任何 tile server 提供商。

2. **地图瓦片来源多样化**: 配置瓦片服务为可替换式——通过环境变量指定 tile server URL。除 Mapbox 外，预配置一个免费 tile provider（如 OpenStreetMap 标准瓦片、Protomaps 自部署方案）。

3. **静态瓦片预缓存**: 对核心地理区域（如用户活跃城市）预缓存地图瓦片到自有 CDN/S3，极端情况下可离线提供基础地图。

4. **地图功能降级设计**: UI 中地图区域为可降级组件——当地图加载失败时，显示静态占位图 + 列表/文本视图替代交互式地图，保证功能可用的最低体验。

5. **Telemetry 关闭**: 明确配置 MapLibre/Mapbox 关闭所有遥测数据收集，避免合规隐患。

### Fallback 方案

*低风险服务不需要完整 multi-provider fallback，采用降级方案即可。*

**降级触发**: Mapbox tile 请求失败率 > 50%
**降级操作**:
- 前端自动切换瓦片源到 OpenStreetMap 免费瓦片（`https://tile.openstreetmap.org/{z}/{x}/{y}.png`）
- 功能标记 `map_use_fallback_tiles = true`
- 切换时间：即时（前端逻辑已内置）
- 用户影响：地图视觉效果下降（无自定义样式），但地理位置信息完整可用

---

## 5. AWS S3（对象存储）

### 风险评估

| 维度 | 评分 | 说明 |
|------|------|------|
| Likelihood | 4 | S3 极其可靠（99.999999999% 持久性），但 AWS 账户级风险（欠费/合规封禁/配置错误）高 |
| Impact | 4 | 存储不可用 → 用户上传/下载失败 → 产品核心功能受损 |
| **综合风险** | **16 (4×4)** | **高风险** |

### 详细分析

**可用性风险**: S3 作为 AWS 最核心的基础设施服务，具有极高的可用性（SLA 99.9%）和持久性（11个9）。但以下场景值得关注：
- **单区域中断**：如果所有数据存储在单个 region，该 region 中断会导致全部不可用
- **配置错误**：人为误改 bucket policy、lifecycle rule、IAM role，可能导致数据无法访问甚至被删除
- **账户级风险**：AWS 账户欠费、违规封禁、IAM key 泄露导致的恶意操作

**供应商锁定风险**: **高**。
- S3 API 已成为行业事实标准，但 S3 特有的功能（版本控制、生命周期策略、事件通知、跨区域复制、Intelligent Tiering）在其竞品（Cloudflare R2、Google Cloud Storage）中的支持程度和 API 行为有差异
- 大规模数据迁移（TB 级以上）成本高、耗时长
- 项目代码中如果深度使用 AWS SDK 的 S3 高级功能（如 presigned URL、multipart upload 的特定行为），切换成本高

**定价变更风险**: **中**。AWS 历史上持续降价，但存在隐性成本：
- 数据传出（egress）费用高，流量增大时成本增长显著
- API 请求费用在小文件高并发场景下累计可观
- 跨 region 复制费用

**合规/数据隐私风险**: **中**。
- 需确保 S3 bucket 所在 region 符合数据本地化要求
- 默认加密（SSE-S3 / SSE-KMS）需开启
- Bucket 必须默认阻止公开访问（Block Public Access），避免数据泄露事故
- 访问日志需开启用于审计
- 敏感数据（用户上传的身份证明等）需要额外的字段级加密

### 缓解策略

1. **S3 API 兼容的存储抽象层**: 使用 S3 兼容协议作为存储接口标准。这意味着任何兼容 S3 API 的存储服务（AWS S3、Cloudflare R2、MinIO、Ceph）都可以作为后端。代码中使用 AWS SDK 的 S3 客户端，但仅调用标准 S3 API（PutObject、GetObject、DeleteObject、ListObjects、HeadObject）、避免使用 S3 专有高级功能。

2. **多 Region 跨区域复制（CRR）**: 配置 S3 Cross-Region Replication，将数据实时复制到另一个 AWS region 的备用 bucket。主 region 不可用时，DNS/应用配置切换到备用 region。

3. **定期数据备份到异云**: 使用 rclone 或自定义脚本，每周将 S3 数据全量/增量同步到 Cloudflare R2（S3 兼容，egress 免费）。这同时解决了供应商锁定和定价（egress 成本）问题。

4. **Presigned URL 本地生成**: 需要外部访问文件时生成 presigned URL——它基于 HMAC 签名，不需要与 S3 交互即可生成。生成的 URL 时效控制在合理范围（默认 1 小时内有效）。

5. **IAM 最小权限 + 防护栏**: 存储访问使用 IAM role（非 IAM user 的 AK/SK），权限按最小原则分配。配置 SCP（Service Control Policy）防止删除 bucket 或关闭 versioning。

6. **Infrastructure as Code + 变更审批**: S3 bucket 配置通过 Terraform 管理，变更走 PR review。关键的 bucket policy 和 lifecycle 变更需要 additional approval。

7. **Bucket 默认安全基线**: 所有 bucket 默认开启——Block Public Access、Default Encryption、Versioning、Access Logging、Object Lock（防止删除）。

### Fallback 方案（高风险必须）

**触发条件**: 主 region S3 不可用持续 10 分钟以上，或 AWS 账户被封禁。

**Fallback——切换到备用存储**:

| 阶段 | 操作 | 预计恢复时间 |
|------|------|-------------|
| 1. DNS/配置切换 | 将存储 endpoint 从主 S3 bucket 切换到备用 bucket（CRR 副本） | 5 分钟 |
| 2. 应用重启/热加载 | 应用重新加载存储配置 | 3 分钟 |
| 3. 功能验证 | 验证上传、下载、列表、删除操作 | 10 分钟 |
| 4. 增量数据追平 | 切换期间产生的数据通过 rclone 增量同步补全 | 视数据量 |

**三层 Fallback 架构**:

```
Layer 1: AWS S3 主区域 (ap-southeast-1)  ← 主存储
Layer 2: AWS S3 备用区域 (ap-southeast-3)  ← CRR 实时复制，5 分钟切换
Layer 3: Cloudflare R2                 ← 每周增量备份 + 实时双写关键文件
```

**前置准备（当前就绪）**:
- S3 CRR 已配置，主备 region 数据延迟 < 1 秒
- R2 备份同步脚本每周执行
- 存储 endpoint 通过环境变量 `STORAGE_ENDPOINT` 配置，支持热更新
- Terraform 管理所有 bucket 配置

**Fallback 局限性**:
- CRR 副本在极端情况下可能有秒级延迟，切换瞬间可能丢失最新写入
- Presigned URL 需按 bucket 生成，切换后旧 URL 失效
- R2 作为第三层在某些边缘功能上不完全兼容 S3（如 Object Lock 的 legal hold 模式）

---

## 6. GitHub API（代码集成）

### 风险评估

| 维度 | 评分 | 说明 |
|------|------|------|
| Likelihood | 2 | GitHub 服务极其稳定，API 中断频率低；即使中断，核心产品功能不受影响 |
| Impact | 2 | GitHub 集成是辅助功能（代码仓库浏览/Webhook），不阻塞核心业务流程 |
| **综合风险** | **4 (2×2)** | **低风险** |

### 详细分析

**可用性风险**: GitHub 拥有世界级的可用性记录。API 中断通常与大规模事件相关（如 DDoS），但频率低、恢复快。增量操作（如 commit 同步）天然支持重试。GitHub 的状态页面（githubstatus.com）提供透明的实时更新。

**供应商锁定风险**: **低**。
- GitHub REST API 和 GraphQL API 是 GitHub 专有，但核心集成（webhook、OAuth、git clone）是标准化协议
- 如果要迁移到 GitLab/Bitbucket，OAuth App 和 Webhook 需重新注册，但功能等价
- Webhook 事件格式（push、pull_request、issues）各平台不完全一致，但都基于标准 HTTP POST+JSON

**定价变更风险**: **低**。GitHub 免费版已包含 API 访问，API rate limit（5000 req/h 认证用户）在大多数场景下足够。即使未来对 API 调用收费，对产品的影响可控。

**合规/数据隐私风险**: **低**。
- GitHub 应用的 OAuth scope 应限制在最小必要范围（通常 `repo` 或 `public_repo`）
- 用户授权 token 不存储在服务端日志中
- GitHub 不存储用户代码文件（仅通过 API 临时读取元数据），降低了数据泄露面

### 缓解策略

1. **API 调用缓存**: 对 GitHub API 的只读请求（仓库信息、文件内容、commit 历史）做本地缓存，设置合理 TTL，减少 API 调用次数和失败影响。

2. **指数退避重试**: 所有 GitHub API 请求包装重试逻辑——429（Rate Limit）使用 `Retry-After` header，5xx 错误使用指数退避（1s → 2s → 4s → 8s），最多重试 3 次。

3. **Personal Access Token 轮换**: GitHub 集成使用的 PAT 每 90 天轮换一次（或使用 GitHub App 的 installation token，自动过期）。

4. **Webhook 签名验证 + 幂等处理**: 所有 GitHub webhook 验证 HMAC-SHA256 签名。通过 `X-GitHub-Delivery` header 做幂等去重，避免重复事件处理。

5. **功能降级**: GitHub 集成为非核心功能，API 不可用时功能模块静默降级——显示缓存数据或 "暂时不可用" 状态，不影响产品其他功能。

### Fallback 方案

*低风险服务，采用降级方案即可。*

**降级触发**: GitHub API 连续失败 5 分钟。
**降级操作**:
- 展示本地缓存的代码仓库元数据
- 代码 Webhook 自动同步暂停，显示 "同步延迟" 状态
- 所有非 GitHub 功能正常运行
- 恢复后触发增量同步补全缺失数据

---

## 附录 A：风险矩阵可视化

```
Impact
  5 |                    | Stripe(15)           |
    |                    |                      |
  4 |                    | Auth0(16) AWS S3(16) |
    |                    |                      |
  3 | Mapbox(6)          | SendGrid(9)          |
    |                    |                      |
  2 | GitHub API(4)      |                      |
    |                    |                      |
  1 |                    |                      |
    +--------------------------------------------
      1         2         3         4         5
                       Likelihood
```

- 绿色区域（Score <= 8）：低风险，常规缓解 + 降级方案
- 黄色区域（Score 9-14）：中风险，缓解策略全量执行
- 红色区域（Score >= 15）：高风险，必须有完整 fallback 方案

---

## 附录 B：行动优先级

| 优先级 | 行动项 | 目标服务 | 类型 |
|--------|--------|---------|------|
| P0 | 建立认证抽象层（IAM Gateway） | Auth0 | 架构改造 |
| P0 | 配置 S3 跨区域复制 + 双写 | AWS S3 | 基础设施 |
| P1 | 支付接口抽象层 + Adyen adapter | Stripe | 架构改造 |
| P1 | 部署 Ory Kratos standby 栈 | Auth0 | Fallback 准备 |
| P1 | S3 → R2 定期备份 | AWS S3 | 数据备份 |
| P2 | 切换到 MapLibre GL JS | Mapbox | 库替换 |
| P2 | 邮件多 provider 路由 | SendGrid | 功能增强 |
| P3 | GitHub API 缓存层 | GitHub API | 性能优化 |

---

<!-- answer_complete -->
