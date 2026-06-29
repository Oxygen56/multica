# CI/CD 工具选型对比报告：GitHub Actions vs GitLab CI vs Jenkins vs CircleCI vs Buildkite

**任务编号**: RESEARCH-015
**日期**: 2026-06-29
**目标团队**: 10人小型团队，代码托管于 GitHub
**自信度**: 0.90

---

## 一、概述

本报告对五款主流 CI/CD 工具——GitHub Actions、GitLab CI、Jenkins、CircleCI、Buildkite——进行系统性对比，覆盖价格、配置复杂度、生态系统、代码托管集成度、自托管能力五个核心维度。所有价格数据基于各工具截至2026年6月的公开定价页面，多源交叉验证。

**核心结论前置**：对于已使用 GitHub 托管代码的10人团队，GitHub Actions 是性价比最高、集成最自然的选择，推荐作为主力 CI/CD 工具。CircleCI 在测试密集型场景下可作为补充。除非团队有特殊的合规/安全需求要求代码不离开基础设施，否则 Jenkins 和 Buildkite 对当前团队规模而言运维成本过高。

---

## 二、团队背景与评估约束

| 维度 | 参数 |
|------|------|
| 团队规模 | 10人（含开发+DevOps） |
| 代码托管 | GitHub（已确定，不可变） |
| 月构建量估算 | 10人 x 4次构建/天 x 22工作日 = 约880次构建/月 |
| 单次构建时长估算 | 5-10分钟（中等复杂度项目，含lint/test/build） |
| 月构建总时长估算 | 约4,400-8,800分钟/月 |
| 运行环境偏好 | Linux 为主，偶需 macOS/Windows |
| 核心诉求 | 低维护成本、可预测开销、与 GitHub 深度集成 |

---

## 三、工具逐一分析

### 3.1 GitHub Actions

**定价（2026年6月生效）**

| 计划 | 月费 | 包含分钟数 | 存储 | 附加分钟单价 |
|------|------|-----------|------|-------------|
| Free | $0 | 2,000 分钟（仅私有仓库；公开仓库无限） | 500 MB | - |
| Team | $4/用户/月 | 3,000 分钟 | 2 GB | Linux 2核: $0.006/分钟 |
| Enterprise | $21/用户/月 | 50,000 分钟 | 50 GB | 同上 |

**自托管 Runner**：当前免费（2025年12月宣布的 $0.002/分钟收费计划已被无限期推迟，自托管 Runner 仍为免费使用）。这是关键优势——团队可以用自己的服务器跑 CI，零额外费用。

**2026年1月生效的降价**：托管 Runner 降价 20%-39%。例如 Linux 2核从 $0.008/分钟降至 $0.006/分钟，Linux 64核从 $0.256/分钟降至 $0.162/分钟。

**10人团队费用估算**（GitHub Team 计划，假设使用托管 Runner）：

- 平台费：10人 x $4/月 = $40/月
- 托管 Runner 费用：8,800分钟（高估算）x $0.006/分钟 = $52.80/月
- 减去包含的3,000分钟：5,800分钟 x $0.006 = $34.80/月
- **合计：约 $74.80/月**

如果使用自托管 Runner（在自己的服务器上跑），则仅需 $40/月的平台费。

**配置复杂度**：低。YAML 格式的工作流文件存放在 `.github/workflows/` 目录下。语法直观，学习曲线平缓。社区 Actions Marketplace 有 20,000+ 预构建 Actions，大部分常见操作（checkout、setup-node、docker build、deploy to cloud）直接用现有 Action 即可，无需自己写逻辑。

**生态系统**：业界最大。20,000+ Actions，覆盖几乎所有语言框架和云平台。GitHub 社区活跃，问题解答速度快。

**代码托管集成度**：原生集成，零配置。PR 状态检查、分支保护规则、Actions 日志、Secrets 管理、Environments 全部在同一个 GitHub UI 内完成，不需要任何外部系统或 Token 桥接。

**自托管能力**：支持自托管 Runner，可注册任意 Linux/macOS/Windows 机器作为 Runner。支持 Runner Groups 按仓库或组织隔离。2026年新增的 ARC（Actions Runner Controller）让 Kubernetes 上的 Runner 自动扩缩容更加成熟。

**优势总结**：
- 零额外集成成本（代码已在 GitHub）
- 20,000+ 市场 Actions，搭建流水线极快
- 自托管 Runner 免费
- YAML 语法简洁，团队成员学习成本低
- 降价后的托管 Runner 价格在行业中处于低位

**劣势总结**：
- Debug 体验一般（act 本地调试不完整体现 Actions 运行时环境）
- 复杂多仓库工作流编排能力有限
- 供应商锁定（迁移到其他 CI 工具需要重写所有 `.yml` 文件）
- macOS/Windows Runner 价格仍然偏高

---

### 3.2 GitLab CI

**定价（2026年）**

| 计划 | 价格 | 包含 CI/CD 分钟 | 存储 | 支持 |
|------|------|----------------|------|------|
| Free | $0 | 400 分钟/月 | 10 GiB | 社区 |
| Premium | $29/用户/月 | 10,000 分钟/月 | 500 GiB | Priority |
| Ultimate | $99/用户/月 | 50,000 分钟/月 | 500 GiB | 24/7 Priority |

**注意**：10,000 分钟在 Premium 计划中按组（group）共享，不是按用户。50人团队也只有10,000分钟总配额，这是经常被忽视的隐性限制。

额外分钟包：$10/1,000 分钟（一次性购买，不过期）。
自托管 Runner：免费且不限量（任何计划），可用 GitLab Runner（开源）注册任意机器。

**10人团队费用估算**（Premium 计划，SaaS）：

- 平台费：10人 x $29/月 = $290/月
- CI/CD 分钟：10,000分钟已含。估算月使用8,800分钟——刚好够用，但几乎没有余量。
- 如需额外分钟：1,000分钟包 x $10
- **合计：$290/月**（基础状况）。如果构建量增长，分钟数很快就需要额外购买。

如果使用自托管 GitLab（Self-Managed），Premium 降为约 $19/用户/月，总平台费约 $190/月，但需要自行维护 GitLab 实例。

**核心问题**：团队代码在 GitHub，不在 GitLab。使用 GitLab CI 意味着代码推送到 GitHub 后还需要同步或镜像到 GitLab——增加一个关键依赖和故障点。虽然有 GitLab CI/CD for GitHub 的集成方式，但本质上是为 GitHub 仓库挂一个 GitLab 的 CI 流水线，维护两套系统的账号、权限、Token，体验远不如原生。

**配置复杂度**：中等。`.gitlab-ci.yml` 语法比 GitHub Actions 丰富，支持 DAG（`needs:` 关键字）、父子流水线、下游触发、merge trains 等高级特性。但对10人团队日常使用而言，这些高级特性多数用不上，反而增加了配置文件的理解成本。Auto DevOps 可以自动生成流水线，但自动配置通常需要手动调整才能满足实际需求。

**生态系统**：强在 DevSecOps 一体化。GitLab 自带容器镜像仓库（Container Registry）、安全扫描（SAST/DAST/Secret Detection/Dependency Scanning）、合规仪表盘。但这些能力深度绑定在 GitLab 的 Ultimate 层级——SAST 和 DAST 需要 Ultimate（$99/用户/月），对10人团队来说成本高昂。

**代码托管集成度**：代码不在 GitLab 的情况下——差。需要维护 GitHub 到 GitLab 的镜像同步，增加延迟、故障点和维护负担。GitLab 作为 CI 工具独立使用时，失去了"单一平台"的核心价值主张。

**自托管能力**：强。GitLab Runner 开源，任何计划都可以注册自托管 Runner。Self-Managed GitLab 也有完善的企业部署方案。

**适用场景判断**：如果团队从一开始就用 GitLab 托管代码，GitLab CI 是一体化首选。但目前代码已在 GitHub，GitLab CI 的集成代价大于收益。

---

### 3.3 Jenkins

**定价**

Jenkins 本身是开源软件，永久免费。但"免费"两个字背后藏着巨大的隐性成本：

| 成本类别 | 估算 |
|---------|------|
| Jenkins 软件许可 | $0（开源） |
| 服务器/云主机 | $50-200/月（视配置） |
| 运维人力（安装/升级/插件管理/安全补丁） | 相当于0.25-0.5个全职DevOps的精力 |
| CloudBees 企业版（如需商业支持） | $30/用户/月起 = $300/月起（10人） |
| 年度总成本（自建） | 人力成本为主，估算 $15,000-30,000/年 |
| 年度总成本（CloudBees 托管） | ~$3,600-12,000+/年 |

**10人团队费用估算**（自建开源 Jenkins）：

- 云服务器（4核16GB，跑 Master + 2-3 Agent）：约 $100-150/月
- 运维时间：每月至少8-16小时用于插件更新、安全补丁、流水线维护
- **直接的 infra 费用：$100-150/月；隐性的运维人力成本远高于此**

**配置复杂度**：高。Jenkins 是五款工具中配置最复杂的。

- **传统方式**（Freestyle Job + UI 配置）：不可审计、不可版本控制、难以复制。任何有现代 DevOps 实践的团队都不应采用。
- **Pipeline as Code**（Jenkinsfile + Groovy）：支持声明式和脚本式两种风格。Groovy 语法对不熟悉 JVM 生态的开发者不友好。调试困难，每次修改需要提交+触发构建才能验证。
- **JCasC**（Jenkins Configuration as Code）：可以版本控制 Jenkins 系统配置，但初期搭建和调通 JCasC + 插件组合本身就是一项工程。
- **插件管理**：1,800+ 插件是 Jenkins 的武器库也是阿喀琉斯之踵。插件之间的兼容性矩阵、安全漏洞（Jenkins 插件是 CVE 的高频来源）、升级后破坏变更——每一项都需要持续投入精力管理。

**生态系统**：最大（历史意义上）。1,800+ 插件覆盖几乎所有需求。但插件质量参差不齐，维护状态不一，安全问题频发。Jenkins 的插件生态已经从"优势"逐渐变成"债务"——数量多但活跃维护的少，找到可靠的插件组合需要试错。

**代码托管集成度**：通过 GitHub/GitLab/Bitbucket 插件可实现 Webhook 触发、PR 状态更新、commit status 等。需要配置 Personal Access Token、Webhook、凭据管理。不是原生集成，初始配置约需1-2小时，后续维护较少。

**自托管能力**：完整（因为只能自托管）。Jenkins 本质上是 Java Web 应用，部署在任意服务器上。支持 Master-Agent 架构、Kubernetes 动态 Agent。但这也意味着——你必须自托管，没有 SaaS 版的 Jenkins 可用（CloudBees 的 SaaS 产品面向企业，个人和小团队不适用）。

**适用场景判断**：Jenkins 适合有专职 DevOps 团队、需要极端定制化流水线、或者有遗留系统必须兼容的场景。对于一个10人团队、代码在 GitHub 的现代开发团队——Jenkins 的运维负担远大于它带来的灵活性收益。**不推荐。**

---

### 3.4 CircleCI

**定价（2026年）**

| 计划 | 月费 | 包含额度 | 并发 | 用户限制 |
|------|------|---------|------|---------|
| Free | $0 | 30,000 credits/月 | 30 | 5 active users |
| Performance | $15起 | 30,000 free credits + 按需购买 | 80 | 5个免费 + 超出25,000 credits/用户 |
| Scale | 定制报价 | 定制 | 定制 | 定制 |

**Credit 消耗率**（Docker/Linux）：

| 资源等级 | vCPU/RAM | Credits/分钟 | 折合 USD/分钟 |
|---------|----------|-------------|--------------|
| Small | 1 CPU, 2GB | 5 | $0.003 |
| Medium | 2 CPU, 4GB | 10 | $0.006 |
| Medium+ | 3 CPU, 6GB | 15 | $0.009 |
| Large | 4 CPU, 8GB | 20 | $0.012 |
| macOS (Medium) | 4 CPU | 50 | $0.030 |

额外 Credits：$15/25,000 credits（约 $0.0006/credit）。

**10人团队费用估算**（Performance 计划）：

- 平台费：$15/月
- 10人（5人免费 + 5人额外）：5 x 25,000 credits = 125,000 credits 从用户费中已含（约 $15/用户/月 = $75/月）
- 实际构建消耗：880次构建/月 x 10分钟/次 x 10 credits/分钟（Medium）= 88,000 credits/月
- 包含 30,000 free + 用户费已覆盖剩余 58,000 → **已全额覆盖**
- **合计：约 $90/月**（$15 平台 + $75 用户费）

如果使用 Small（5 credits/分钟）则实际消耗仅 44,000 credits/月，总费用可压到 $30/月（仅需1个额外用户）。

**注意**：CircleCI 有一个 Startup Program，Pre-Seed 到 Series A 阶段的初创公司（成立5年内）可获得最高 $20,000 的免费 compute credits，最长24个月免费。如果团队符合条件，这是巨大的成本优势。

**配置复杂度**：中等偏低。YAML 格式（`.circleci/config.yml`），使用 Orbs（CircleCI 的可复用配置包，类似 GitHub Actions 的 Actions）可以快速搭建常见流水线。3,500+ Orbs 覆盖主流场景。CircleCI 的配置文件比 GitHub Actions 更结构化——使用 `executors`、`jobs`、`workflows` 三层结构，逻辑清晰。

**调试体验**是 CircleCI 的亮点：支持 SSH 进入失败的构建容器进行实时调试，这是 GitHub Actions 不原生支持的能力。

**生态系统**：3,500+ Orbs，规模小于 GitHub Actions 但质量相对可控。Orbs 需要经过 CircleCI 审核才能在 Registry 发布，质量门槛比 Actions Marketplace 高。

**代码托管集成度**：好（但不如 GitHub Actions）。支持 GitHub/Bitbucket。Webhook 触发、PR 状态更新、commit status 均自动配置。但 PR 上的 Checks 展示和 GitHub Actions 的原生体验相比略逊一筹。

**自托管能力**：支持 Self-Hosted Runner（CircleCI 称其为 "self-hosted machine runner"），但这是较新推出的功能，成熟度不如 GitHub Actions 的自托管 Runner。

**适用场景判断**：CircleCI 的核心优势在测试密集型场景——智能测试拆分（test splitting）、Docker Layer Caching、按资源等级精细控制、SSH 调试。如果团队的 CI 瓶颈在测试速度上，CircleCI 的并行化能力能让测试套件缩短 3 倍以上。但如果日常 CI 以 build/lint/deploy 为主，CircleCI 的速度优势体现不明显，而平台费+用户费的定价模型会让10人团队的总成本高于 GitHub Actions。

---

### 3.5 Buildkite

**定价（2026年）**

| 计划 | 价格 | 核心特性 |
|------|------|---------|
| Personal (Free) | $0 | 1用户, 3并发, 500分钟托管Agent/月 |
| Team | $15/用户/月 | 不限自托管Agent, 不限流水线 |
| Business/Pro | $30/用户/月 | SSO, 10并发自托管Agent, 2,000分钟托管Agent |
| Enterprise | 定制报价 | 无限制, SAML/SCIM, 审计日志, Premium SLA |

Buildkite 托管 Agent 按分钟计费（在计划费之外）：
- Small (2 vCPU, 4GB): $0.013/分钟
- Medium (4 vCPU, 16GB): $0.026/分钟
- Large (8 vCPU, 32GB): $0.052/分钟

**10人团队费用估算**（Team 计划 + 自建 Agent）：

- 平台费：10人 x $15/月 = $150/月
- Agent 服务器（自建，类似 Jenkins 场景）：约 $100-200/月
- **合计：约 $250-350/月**

如果使用 Buildkite 托管 Agent：
- 8,800分钟 x $0.013/分钟（Small）= $114.40/月
- **合计：平台 $150 + Agent $114.40 = $264.40/月**

**架构特点**：Buildkite 采用"混合架构"——控制平面（流水线管理、Web UI、调度）在 Buildkite 云端，执行平面（Agent）跑在客户自己的基础设施上。代码在构建过程中不经过 Buildkite 服务器。这是金融、医疗等合规敏感行业的常见选择，但10人小团队通常不需要这个级别的隔离。

**配置复杂度**：中等。使用 YAML 定义 pipeline steps，支持动态流水线（根据上一步输出动态生成后续步骤）。Buildkite 的流水线 DSL 比 GitHub Actions 和 CircleCI 更灵活，但学习和调试成本也更高。

**生态系统**：小。Buildkite 的 Plugin 生态远小于 GitHub Actions 和 CircleCI。很多功能需要自己写脚本封装。集成深度依赖团队自身能力。

**代码托管集成度**：中等。支持 GitHub/GitLab/Bitbucket。通过 Webhook 触发构建、更新 commit status。但与 GitHub 的集成不如 GitHub Actions 原生——需要配置 Webhook URL、API Token、更新 commit status 的权限。

**自托管能力**：核心卖点。Agent 始终跑在客户基础设施上，Buildkite 不接触代码。Agent 的部署、扩缩容、监控都由客户负责。对10人团队来说，管理 Agent 集群和 CI 平台自身的开销是一种不必要的负担。

**适用场景判断**：Buildkite 的典型客户是 Shopify、Airbnb、Lyft、Pinterest 级别的公司——他们有数百开发者、Monorepo、需要无限并发构建、并且对代码安全有极端要求。10人小团队用 Buildkite，等于花了高端工具的价钱却只用到了最基础的功能，性价比不高。

---

## 四、多维度对比总表

| 维度 | GitHub Actions | GitLab CI | Jenkins | CircleCI | Buildkite |
|------|:---:|:---:|:---:|:---:|:---:|
| **月费用（10人估算）** | $40-75 | $290 | $100-150 + 运维人力 | $30-90 | $150-350 |
| **费用可预测性** | 高（按分钟） | 中（分钟限制严格） | 低（运维不可预测） | 中（credits模型） | 高（按用户） |
| **配置复杂度** | 低 | 中 | 高 | 中偏低 | 中 |
| **YAML 语法** | 简洁直观 | 丰富但复杂 | Groovy（非YAML） | 结构清晰 | 灵活且强大 |
| **Marketplace/生态** | 20,000+ | CI/CD Catalog | 1,800+ Plugins | 3,500+ Orbs | 小 |
| **GitHub 集成度** | 原生（零配置） | 需镜像同步 | 需配置 Webhook | 好（自动） | 中（需手动配置） |
| **自托管 Runner/Agent** | 支持（当前免费） | 支持（完全免费） | 只能自托管 | 支持（较新） | 核心卖点 |
| **SaaS 可用性** | 是 | 是 | 否（开源） | 是 | 是（混合） |
| **本地调试** | act（有限） | gitlab-runner exec | 无 | SSH into build | bk local run |
| **并发限制** | 20（Free）/ 180（Team） | 取决于分钟配额 | 取决于服务器 | 30（Free）/ 80（Perf） | 3（Free）/ 10（Pro） |
| **学习曲线** | 1-3天 | 3-7天 | 7-30天 | 2-5天 | 3-7天 |
| **2026年市场占比** | ~60% | ~25%（GitLab用户内） | 下降中 | 下降中 | 细分领域强 |
| **目标用户** | GitHub用户 | GitLab用户 / DevSecOps | 传统企业 | 测试密集型团队 | 大型企业 / 合规敏感 |

---

## 五、10人团队费用对比（具体数字）

假设场景：月构建 880 次，平均每次 10 分钟 Medium Linux，使用各平台的推荐计划。

| 工具 | 计划 | 月平台费 | 月计算费 | 其他费用 | **月总计** | **年总计** |
|------|------|---------|---------|---------|----------|----------|
| GitHub Actions | Team | $40 (10x$4) | $35 (托管Runner) | $0 | **$75** | **$900** |
| GitHub Actions (自托管) | Team | $40 (10x$4) | $0 (自托管Runner) | ~$100 (服务器) | **$140** | **$1,680** |
| GitLab CI | Premium SaaS | $290 (10x$29) | $0 (10,000含) | $0 | **$290** | **$3,480** |
| Jenkins (自建) | 开源 | $0 | $0 | $150 (服务器) + 运维人力 | **$150+** | **$1,800+** |
| CircleCI | Performance | $15 | $0 (credits够用) | $75 (额外5用户) | **$90** | **$1,080** |
| CircleCI (优化) | Performance | $15 | $0 | $15 (1额外用户) | **$30** | **$360** |
| Buildkite | Team | $150 (10x$15) | ~$114 (托管Agent) | $0 | **$264** | **$3,168** |
| Buildkite (自建Agent) | Team | $150 (10x$15) | $0 | ~$150 (Agent服务器) | **$300** | **$3,600** |

**关键观察**：

1. **GitHub Actions** 在年费上是最低的之一（$900/年），且集成度最高。如果使用自托管 Runner，稍贵但获得完全的计算控制权。
2. **CircleCI** 在精细优化后（Small 资源 + 控制用户数），可以达到极低的年费（$360-1,080/年），前提是构建需求对 Small 资源类足够。
3. **GitLab CI** 对10人团队偏贵（$3,480/年），且代码不在 GitLab 时需要额外维护代码同步。
4. **Jenkins** 的直接费用不高，但运维人力是最大的隐性成本——这是10人小团队最稀缺的资源。
5. **Buildkite** 整体最贵，不是为小团队设计的工具。

---

## 六、推荐方案

### 首选：GitHub Actions（自信度 0.90）

**推荐理由**：

1. **集成零成本**：代码已在 GitHub。Actions 与 PR、Issues、分支保护、Secrets、Environments 的原生集成不需要任何额外配置或 Token 桥接。团队成员不需要学习新平台——流水线状态就在他们每天用的 PR 页面上。

2. **价格合理且可预测**：Team 计划 $4/人/月 + 托管 Runner 按分钟计费，月开销 $75 左右。2026年降价后，Linux 2核仅 $0.006/分钟，是目前主流 CI 工具中最低的单价之一。

3. **自托管 Runner 免费且成熟**：如果未来团队需要使用自己的服务器跑 CI（降低成本或满足合规需求），自托管 Runner 完全免费（$0.002/分钟的收费计划已无限期推迟），且有成熟的 Kubernetes Runner Controller (ARC) 支撑自动扩缩容。

4. **生态最大**：20,000+ Actions 意味着10人团队的常见需求（部署到 AWS/阿里云/Vercel、构建 Docker、运行测试、Slack 通知）直接使用现成 Action 即可，不需要自己写任何脚本。大幅缩短流水线搭建时间。

5. **学习成本最低**：YAML 配置简洁直观，开发者上手1-2天即可写出生产级流水线。这对没有专职 DevOps 的10人团队至关重要——每个人都能维护 CI，不依赖某一个人。

6. **可扩展性好**：未来团队扩编、项目增多、构建复杂度提升时，Actions 的 Matrix Builds、Reusable Workflows、Environments、OIDC 等特性可以无缝应对，无需迁移工具。

**需要关注的风险**：

- **供应商锁定**：`.github/workflows/*.yml` 是 GitHub Actions 专有格式，迁移到其他 CI 工具需要重写。但考虑到团队已在 GitHub 生态中，这一锁定的实际影响有限。
- **调试体验**：本地调试依赖 `act`（社区工具），不能完全复现 Actions 运行时。对于复杂流水线，调试可能需要在 GitHub 上反复提交触发，效率不如 CircleCI 的 SSH 调试。
- **未来定价不确定性**：自托管 Runner 的收费计划虽然被推迟，但 GitHub 明确表示这是"推迟而非取消"。未来可能以不同形式重新收费。建议保持关注。

### 补充方案：CircleCI（用于测试密集型模块）

如果团队中有特定的测试密集型项目（大型测试套件，运行时间 >15分钟），可以考虑在 CircleCI 上为该特定仓库运行测试流水线。CircleCI 的 Test Splitting + Docker Layer Caching 能将大型测试套件的运行时间缩短 3 倍。此时 CircleCI 作为 GitHub Actions 的补充而非替代。

---

## 七、不推荐的选项及原因

| 工具 | 不推荐原因 |
|------|-----------|
| **GitLab CI** | 代码已在 GitHub，使用 GitLab CI 需要维护代码镜像同步，增加故障点和维护成本。$29/用户/月对10人团队来说也偏贵。 |
| **Jenkins** | 运维负担与10人团队的规模严重不匹配。插件管理、安全补丁、JCasC 搭建需要持续的 DevOps 精力投入，而这些精力应该花在产品开发上。 |
| **Buildkite** | 为大型企业和合规场景设计。$15/用户/月的平台费 + 自建 Agent 的运维成本，性价比对所有方面都不如 GitHub Actions。10人团队不需要 Buildkite 的混合架构和无限并发能力。 |

---

## 八、实施建议

1. **立即启动**：在现有 GitHub 仓库中创建 `.github/workflows/ci.yml`，配置 lint → test → build 基础流水线。预计1小时内可完成。
2. **规划自托管 Runner**：如果月构建时长超过 3,000 分钟，部署自托管 Runner 到团队现有服务器（或一台 $20/月的云主机），将计算费归零。推荐使用 ARC（Actions Runner Controller）管理。
3. **关注定价变化**：GitHub 自托管 Runner 的收费在"推迟"状态，未来可能以按作业次数（而非按分钟）的方式回归。每月检查一次 GitHub Blog 和 community discussions。
4. **在特定场景测试 CircleCI**：如果测试套件运行超过15分钟，在 CircleCI Free 计划上做一个对比测试（同样的测试在 CircleCI vs GitHub Actions），用数据决定是否需要双工具策略。

---

## 九、参考来源

- GitHub Actions Pricing: https://docs.github.com/en/billing/managing-billing-for-github-actions/about-billing-for-github-actions
- GitHub Actions 2026 Price Changes: https://github.blog/changelog/2025-12-16-coming-soon-simpler-pricing-and-a-better-experience-for-github-actions/
- GitHub Self-Hosted Runner Pricing Postponed: https://github.com/orgs/community/discussions/182186
- CircleCI Pricing: https://circleci.com/pricing/
- CircleCI Plans Overview: https://circleci.com/docs/guides/plans-pricing/plan-overview/
- Buildkite Pricing: https://buildkite.com/pricing/
- GitLab Pricing: https://about.gitlab.com/pricing/
- CloudBees (Managed Jenkins) Pricing: https://www.cloudbees.com/products/saas-platform/pricing
- CI/CD Tools Comparison 2026: https://www.deployhq.com/blog/best-ci-cd-software-top-10-tools-to-know-in-2025
- CI/CD Systems Deep Dive 2026: https://www.youngju.dev/blog/culture/2026-05-16-cicd-systems-github-actions-buildkite-circleci-gitlab-jenkins-argo-tekton-earthly-dagger-2026-deep-dive.en
- CI/CD Cost Optimization 2026: https://sesamedisk.com/ci-cd-pipelines-2026-comparison/

---

<!-- answer_complete -->
