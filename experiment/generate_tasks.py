#!/usr/bin/env python3
"""Generate 100 benchmark tasks across 6 cognitive domains for the AGI architecture experiment."""

import json

tasks = []

# ===== CODE DOMAIN (17 tasks) =====
code = [
    ("CODE-001", 2, "修复并发竞态条件",
     "以下Go代码在高并发场景下偶发panic。分析代码，定位根因，提供修复方案并附修改后的代码。\n\n"
     "type Cache struct { mu sync.RWMutex; items map[string]*Item }\n"
     "func (c *Cache) Get(key string) *Item {\n"
     "    c.mu.RLock(); defer c.mu.RUnlock()\n"
     "    item := c.items[key]\n"
     "    if item != nil && item.expired() { c.items[key] = nil; return nil }\n"
     "    return item\n}\n\n"
     "已知：expired()方法没有副作用，items的写操作在另一个goroutine中进行。请给出完整分析。",
     ["正确识别竞态条件的根因","给出至少两种可行修复方案并比较优劣","修复后的代码不再有竞态条件","解释为什么原始代码在高并发下会panic"],
     {"correctness":0.5,"completeness":0.3,"quality":0.2}, 4, ["RESEARCH"], "code_with_explanation"),

    ("CODE-002", 3, "设计可扩展的RBAC权限系统",
     "为一个SaaS平台设计基于RBAC的权限系统。要求支持：\n"
     "1. 租户隔离（多租户，每个租户可有自己的角色定义）\n"
     "2. 细粒度权限（资源级别，如只能编辑自己创建的文档）\n"
     "3. 权限继承（角色可继承其他角色的权限）\n"
     "4. 临时权限授予（有时效性的权限提升）\n\n"
     "请给出：数据模型设计、API接口设计、关键代码片段。解释设计中的关键权衡。",
     ["数据模型覆盖所有四种需求","API设计RESTful且完整","代码片段可编译","明确说明至少2个设计权衡及选择理由"],
     {"correctness":0.4,"completeness":0.3,"quality":0.3}, 5, ["DESIGN","STRAT"], "design_doc_with_code"),

    ("CODE-003", 4, "实现分布式ID生成器",
     "设计并实现一个分布式唯一ID生成器，满足：\n"
     "1. 全局唯一（跨数据中心）\n2. 趋势递增（利于数据库索引）\n"
     "3. 高性能（单机>10万QPS）\n4. 不依赖外部服务也能工作\n"
     "5. 支持从ID中解析出生成时间戳和机器标识\n\n"
     "实现语言不限，需给出完整代码和设计文档。",
     ["ID结构定义清晰，含时间戳+机器标识+序列号","给出冲突分析（时钟回拨如何处理）","代码可编译运行","性能分析支持10万QPS的结论"],
     {"correctness":0.5,"completeness":0.25,"quality":0.25}, 6, ["MATH","DESIGN"], "code_with_design_doc"),

    ("CODE-004", 1, "SQL查询优化",
     "以下SQL查询在百万级数据量下执行超过10秒。分析慢查询原因，给出优化方案（包括但不限于索引优化、查询重写），并提供优化后的SQL。\n\n"
     "SELECT u.name, COUNT(o.id) as order_count, SUM(o.amount) as total_amount\n"
     "FROM users u LEFT JOIN orders o ON u.id = o.user_id\n"
     "WHERE o.created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)\n"
     "  AND o.status IN ('paid','shipped','delivered')\n"
     "GROUP BY u.id ORDER BY total_amount DESC LIMIT 50;\n\n"
     "已知：users表10万行，orders表500万行。",
     ["正确分析慢查询根因","给出索引优化方案（含具体DDL）","重写后的SQL正确且语义等价","说明优化前后的预期性能差异"],
     {"correctness":0.5,"completeness":0.3,"quality":0.2}, 3, ["RESEARCH"], "sql_with_explanation"),

    ("CODE-005", 3, "重构上帝类",
     "以下是一个上帝类（Java），包含太多职责。分析问题，给出重构方案，输出重构后的代码结构。\n\n"
     "public class OrderService {\n"
     "  public Order createOrder(Cart cart, User user) { /* 验证库存、计算价格、创建订单、发送邮件 */ }\n"
     "  public void processPayment(Order order, PaymentInfo info) { /* 支付处理、更新状态、通知仓库 */ }\n"
     "  public void shipOrder(Order order, Address addr) { /* 选择物流、生成运单、更新库存 */ }\n"
     "  public Report generateSalesReport(DateRange range) { /* 查询订单、聚合数据、格式化 */ }\n"
     "  public void sendPromotion(User user, Promotion promo) { /* 过滤用户、生成内容、发送 */ }\n"
     "}\n\n"
     "重构目标：单一职责、开闭原则、可测试性。给出重构后的类图/代码结构。",
     ["正确识别至少5个独立职责","重构方案遵循SOLID原则","给出至少3个新类的完整代码骨架","说明重构对可测试性的提升"],
     {"correctness":0.4,"completeness":0.3,"quality":0.3}, 4, ["DESIGN"], "code_with_design_explanation"),

    ("CODE-006", 5, "设计实时协作编辑系统的CRDT",
     "设计一个基于CRDT的实时协作文本编辑器核心数据结构。要求：\n"
     "1. 支持多用户并发插入/删除字符\n2. 最终一致性保证（无需OT）\n"
     "3. 支持光标位置同步\n4. 给出合并算法伪代码\n"
     "5. 分析两个用户同时在同位置插入不同字符的行为\n\n"
     "给出核心数据结构和关键算法。",
     ["CRDT数据结构定义完整","插入和删除操作定义清晰","合并算法无冲突","正确分析边界场景","给出至少一个已知CRDT方案的对比"],
     {"correctness":0.5,"completeness":0.3,"quality":0.2}, 7, ["MATH","RESEARCH"], "design_doc_with_pseudocode"),

    ("CODE-007", 4, "实现无锁并发队列",
     "设计并实现一个无锁（lock-free）多生产者-多消费者队列。要求：\n"
     "1. 使用CAS操作而非互斥锁\n2. 支持有界和无界两种模式\n"
     "3. 给出ABA问题的分析和解决方案\n"
     "4. 给出与std::mutex+condition_variable版本的性能对比分析\n"
     "实现语言C++或Rust，需完整可编译代码。",
     ["队列接口定义清晰","CAS操作使用正确","ABA问题的分析和解决方案正确","性能对比有数据支撑"],
     {"correctness":0.5,"completeness":0.3,"quality":0.2}, 6, ["MATH"], "code_with_analysis"),

    ("CODE-008", 2, "编写可测试的HTTP客户端封装",
     "为一个微服务系统编写HTTP客户端封装层。要求：\n"
     "1. 支持重试策略（指数退避）\n2. 支持断路器模式\n"
     "3. 支持请求/响应拦截器链\n4. 完全可测试（依赖注入，mock友好）\n\n"
     "给出接口设计+实现代码+单元测试用例设计。语言Python或Go。",
     ["重试和断路器逻辑正确","拦截器链设计合理","依赖注入使得所有依赖可mock","单元测试用例设计覆盖关键路径"],
     {"correctness":0.4,"completeness":0.3,"quality":0.3}, 4, ["DESIGN"], "code_with_tests"),

    ("CODE-009", 3, "实现简单正则表达式引擎",
     "实现一个简单的正则表达式引擎，至少支持：\n"
     "1. 字面量匹配\n2. . (匹配任意单字符)\n3. * (零个或多个)\n4. + (一个或多个)\n5. ? (零个或一个)\n6. ^ 和 $ (行首行尾锚点)\n\n"
     "使用NFA/DFA方法，给出设计和完整实现代码。语言不限。",
     ["正则引擎可编译运行","至少5个测试用例通过（含边界情况）","解释NFA→DFA转换过程","给出时间/空间复杂度分析"],
     {"correctness":0.5,"completeness":0.3,"quality":0.2}, 5, ["MATH"], "code_with_explanation"),

    ("CODE-010", 5, "设计分布式事务方案",
     "一个电商系统需要确保订单创建和库存扣减的原子性，但订单服务和库存服务使用不同数据库。\n"
     "设计分布式事务方案，考虑：\n1. 最终一致性 vs 强一致性\n2. 补偿事务（Saga模式）\n"
     "3. 幂等性保证\n4. 故障恢复（任意步骤失败的回滚策略）\n5. 并发冲突处理\n\n"
     "给出：架构图（文字描述）、关键代码伪代码、状态机定义。",
     ["明确选择一致性模型并给出理由","Saga编排/编排模式定义清晰","幂等性方案具体可实施","故障恢复状态机完整","考虑至少2个边界情况"],
     {"correctness":0.5,"completeness":0.3,"quality":0.2}, 7, ["STRAT","DESIGN"], "architecture_doc"),

    ("CODE-011", 1, "编写安全的用户输入验证函数",
     "为一个Web应用编写用户输入验证模块。需处理：\n"
     "1. XSS注入防护\n2. SQL注入防护（输入层面）\n"
     "3. 邮箱/手机号/URL格式验证\n4. 长度和字符集限制\n\n"
     "给出完整实现代码（Python或JavaScript），包含至少10个测试用例。",
     ["XSS和SQL注入向量被正确防护","格式验证正则正确","边界情况处理（空输入、超长输入、Unicode）","测试用例覆盖正常和异常路径"],
     {"correctness":0.5,"completeness":0.25,"quality":0.25}, 3, ["RESEARCH"], "code_with_tests"),

    ("CODE-012", 4, "设计API限流系统",
     "设计一个API限流（Rate Limiting）中间件。要求：\n"
     "1. 支持多种限流策略：固定窗口、滑动窗口、令牌桶、漏桶\n"
     "2. 支持多维度限流：按用户/按IP/按API端点/按租户\n"
     "3. 分布式场景下的一致性保证\n"
     "4. 给出Redis Lua脚本或等效的原子操作实现\n\n"
     "给出设计文档+核心代码实现。",
     ["四种限流策略都有实现或伪代码","多维度限流设计合理","分布式一致性方案可行","Redis脚本或等效实现正确"],
     {"correctness":0.5,"completeness":0.3,"quality":0.2}, 6, ["DESIGN","MATH"], "design_doc_with_code"),

    ("CODE-013", 3, "排查内存泄漏",
     "以下Python服务在运行48小时后内存从200MB增长到2GB。分析可能的内存泄漏原因，"
     "给出排查方法和工具，提供代码审查检查清单。\n\n"
     "已知：服务使用asyncio，大量使用闭包和装饰器，有自定义缓存，使用了第三方ML库。",
     ["识别至少4种可能的内存泄漏来源","给出具体的排查命令/工具","提供代码审查检查清单","对每种泄漏给出修复建议"],
     {"correctness":0.4,"completeness":0.3,"quality":0.3}, 4, ["RESEARCH"], "analysis_with_checklist"),

    ("CODE-014", 5, "实现LSM-Tree存储引擎",
     "设计并实现一个简化的LSM-Tree存储引擎。要求：\n"
     "1. MemTable（内存中的有序结构）\n2. SSTable（磁盘上的不可变文件）\n"
     "3. Compaction策略（至少Leveled Compaction）\n"
     "4. Bloom Filter加速点查询\n5. Write-Ahead Log保证持久性\n\n"
     "给出核心数据结构和关键算法实现。语言不限。",
     ["MemTable和SSTable结构定义清晰","Compaction策略逻辑正确","Bloom Filter实现正确","WAL机制完整","至少5个测试用例"],
     {"correctness":0.5,"completeness":0.3,"quality":0.2}, 8, ["MATH","DESIGN"], "code_with_design_doc"),

    ("CODE-015", 2, "设计数据库Migration方案",
     "为一个生产环境的PostgreSQL数据库设计零停机Schema变更方案。\n"
     "场景：需要给一个大表（2亿行）添加一个NOT NULL列（带默认值），同时保证服务不中断。\n\n"
     "给出：详细步骤、每个步骤的风险评估、回滚方案、代码/脚本。",
     ["步骤设计覆盖expand-contract模式","风险评估和回滚方案完整","考虑了锁和复制延迟问题","给出的脚本可直接执行"],
     {"correctness":0.5,"completeness":0.3,"quality":0.2}, 4, ["STRAT","RESEARCH"], "procedure_with_scripts"),

    ("CODE-016", 3, "实现Event Sourcing模式",
     "为一个任务管理系统实现Event Sourcing模式。要求：\n"
     "1. 所有状态变更以事件形式存储\n2. 支持事件回放重建当前状态\n"
     "3. 支持快照优化回放性能\n4. 给出CQRS读模型投影\n\n"
     "以Issue状态变更（todo→in_progress→in_review→done）为例，给出完整代码实现。",
     ["事件类型定义完整","事件存储和发布机制正确","回放逻辑可正确重建状态","快照策略合理","CQRS读模型投影正确"],
     {"correctness":0.5,"completeness":0.3,"quality":0.2}, 5, ["DESIGN"], "code_with_design"),

    ("CODE-017", 4, "实现共识算法Raft的领导者选举",
     "实现Raft共识算法中的领导者选举部分。要求：\n"
     "1. 三种节点状态：Follower、Candidate、Leader\n"
     "2. 随机选举超时\n3. 任期(term)管理\n4. 投票逻辑\n"
     "5. 处理网络分区场景（给出测试用例）\n\n"
     "给出完整代码实现（Go或Python），包含模拟网络层的测试。",
     ["状态转换逻辑与Raft论文一致","选举超时随机化正确","任期比较逻辑正确","包含至少3个测试场景（正常选举、网络分区、分割投票）"],
     {"correctness":0.5,"completeness":0.3,"quality":0.2}, 6, ["MATH"], "code_with_tests"),
]

# ===== STRAT DOMAIN (17 tasks) =====
strat = [
    ("STRAT-001", 3, "技术栈选型决策",
     "一个创业团队（5名开发者）要构建一个实时协作的Web应用（类似Figma的简化版）。\n"
     "需要做以下技术选型决策：\n"
     "1. 前端框架：React vs Vue vs Svelte\n"
     "2. 后端：Node.js vs Go vs Python(FastAPI)\n"
     "3. 数据库：PostgreSQL vs MongoDB vs CockroachDB\n"
     "4. 实时通信：WebSocket vs WebRTC vs Server-Sent Events\n"
     "5. 部署：Kubernetes vs Docker Compose vs Vercel/Railway\n\n"
     "对每个维度给出推荐方案+理由+风险分析。考虑团队规模和产品特性。",
     ["每个维度都有明确的推荐和理由","考虑了团队规模（5人）的约束","考虑了实时协作的技术特点","至少分析了2个备选方案的风险"],
     {"correctness":0.3,"reasoning_quality":0.4,"practicality":0.3}, 5, ["RESEARCH","CODE"], "decision_memo"),

    ("STRAT-002", 4, "微服务拆分边界决策",
     "一个电商Monolith（订单、商品、用户、支付、物流、营销6个模块）正在考虑微服务化。\n"
     "当前痛点：订单模块变更频繁影响其他模块，部署耦合严重。\n\n"
     "请给出：\n"
     "1. 微服务拆分方案（哪些模块先拆、拆成几个服务）\n"
     "2. 拆分顺序和优先级\n"
     "3. 数据一致性策略\n"
     "4. 不要拆分的情况/理由\n"
     "5. 与「继续优化Monolith」方案的对比分析",
     ["拆分方案有清晰的边界定义","拆分顺序有逻辑支撑","数据一致性策略具体","客观评估了不拆分的选项","给出至少3个拆分决策原则"],
     {"correctness":0.3,"reasoning_quality":0.4,"practicality":0.3}, 6, ["CODE","DESIGN"], "analysis_report"),

    ("STRAT-003", 2, "开源协议选择",
     "你要把一个内部工具开源。该工具是一个CLI开发效率工具，你希望：\n"
     "1. 最大化社区采用\n2. 防止云厂商打包成SaaS服务与你竞争\n"
     "3. 允许企业用户内部使用但不允许修改后闭源分发\n"
     "4. 个人开发者和小公司免费使用\n\n"
     "给出开源协议建议（从MIT/Apache2/GPLv3/AGPLv3/BSL/Elastic License中选择），"
     "解释每种协议的适用性和不适用原因。",
     ["给出明确的主推荐协议","分析了所有6种协议的适用性","考虑了防止云厂商竞争的需求","解释了推荐协议的权衡"],
     {"correctness":0.4,"reasoning_quality":0.4,"practicality":0.2}, 4, ["RESEARCH"], "decision_memo"),

    ("STRAT-004", 5, "AGI架构设计决策",
     "你在设计一个AGI agent运行平台。需要在以下维度做决策：\n"
     "1. Agent间通信：消息队列 vs 共享内存 vs 事件总线\n"
     "2. 记忆系统：向量数据库 vs 图谱数据库 vs 文件系统\n"
     "3. 任务分解：静态DAG vs LLM动态分解 vs 混合\n"
     "4. 安全边界：沙箱隔离 vs 权限系统 vs 审计日志\n"
     "5. 扩展策略：插件系统 vs 标准化API vs Agent即代码\n\n"
     "对于每个维度，给出你的推荐方案、设计原则和关键权衡。"
     "考虑长期演进（5-10年），不限于当前技术限制。",
     ["每个维度有明确推荐和设计原则","考虑了长期演进（5-10年）","每个决策有清晰的trade-off分析","总体架构一致（各决策相互协调）"],
     {"correctness":0.3,"reasoning_quality":0.5,"vision":0.2}, 7, ["RESEARCH","CODE","DESIGN"], "architecture_decision_record"),

    ("STRAT-005", 3, "技术债务管理策略",
     "一个运行3年的SaaS产品累积了大量技术债务：\n"
     "- 30%代码无测试覆盖\n- 有3个已EOL的依赖\n"
     "- 数据库Schema有5个历史遗留的冗余字段\n"
     "- 日志系统混乱（3种不同格式）\n"
     "- CI/CD pipeline平均耗时45分钟\n\n"
     "你的团队有8名开发者，每两周一个sprint。产品团队同时要求新功能开发不减速。\n"
     "设计一个技术债务偿还策略，给出优先级排序和具体执行计划。",
     ["优先级排序有清晰标准","执行计划具体到可操作层面","平衡了新功能开发和还债","给出了度量改进的指标"],
     {"correctness":0.3,"reasoning_quality":0.4,"practicality":0.3}, 5, ["RESEARCH"], "strategy_doc"),

    ("STRAT-006", 4, "数据迁移策略",
     "你需要将100TB的生产数据从一个云提供商迁移到另一个，要求：\n"
     "1. 停机时间<5分钟\n2. 数据完整性100%\n3. 迁移过程中服务可读\n"
     "4. 可回滚到源端\n\n"
     "数据包括：PostgreSQL(2TB)、Elasticsearch(5TB)、S3对象存储(93TB)。\n"
     "设计完整迁移方案，包括时间线估算。",
     ["迁移步骤完整且顺序合理","停机时间分析符合要求","回滚方案具体可操作","数据完整性验证方案存在","时间估算有依据"],
     {"correctness":0.4,"reasoning_quality":0.3,"practicality":0.3}, 6, ["CODE","RESEARCH"], "migration_plan"),

    ("STRAT-007", 2, "第三方服务依赖风险评估",
     "你的产品依赖以下第三方服务：\n"
     "1. Auth0（认证）2. Stripe（支付）3. SendGrid（邮件）\n"
     "4. Mapbox（地图）5. AWS S3（存储）6. GitHub API（代码集成）\n\n"
     "请评估每个依赖的风险等级（高/中/低），给出风险缓解策略，"
     "并为高风险依赖设计fallback方案。",
     ["每个依赖有风险等级评估和理由","高风险依赖有具体fallback方案","考虑了供应商锁定和定价变更风险","考虑了合规/数据隐私风险"],
     {"correctness":0.3,"reasoning_quality":0.4,"practicality":0.3}, 4, ["RESEARCH"], "risk_assessment"),

    ("STRAT-008", 5, "构建vs购买vs开源决策框架",
     "设计一个通用决策框架，用于决定一个软件组件应该自建（Build）、购买（Buy）还是使用开源（Open Source）。\n"
     "框架应包含：\n1. 评估维度（至少8个）\n2. 每个维度的权重建议\n"
     "3. 评分方法\n4. 边界条件（什么情况下自动排除某个选项）\n\n"
     "用以下两个案例验证你的框架：\nA. CI/CD平台（团队50人）\nB. 日志分析系统（团队5人，预算有限）",
     ["评估维度至少8个且相互独立","评分方法可操作","边界条件定义清晰","两个案例的评估结果合理","框架本身简单易用"],
     {"correctness":0.3,"reasoning_quality":0.4,"practicality":0.3}, 6, ["DESIGN","RESEARCH"], "decision_framework"),

    ("STRAT-009", 3, "灰度发布策略",
     "你负责一个日活1000万用户的社交App后端。需要上线一个重构的推荐算法（完全替换旧算法）。\n"
     "设计灰度发布策略，考虑：\n"
     "1. 用户分桶策略\n2. 回滚条件和自动化\n3. 监控指标和告警阈值\n"
     "4. A/B实验设计（确保统计显著性）\n5. 逐步放量的时间线\n\n"
     "特别注意：新旧算法的结果不能混合，同一用户必须在同一个桶中。",
     ["分桶策略保证用户一致性","回滚条件自动化且具体","监控指标覆盖业务和技术维度","A/B设计有样本量计算","时间线合理"],
     {"correctness":0.4,"reasoning_quality":0.3,"practicality":0.3}, 5, ["MATH","CODE"], "rollout_plan"),

    ("STRAT-010", 4, "On-Call事件响应流程设计",
     "为一个25人的工程团队设计On-Call事件响应流程。背景：\n"
     "- 服务可用性目标：99.9%\n- 平均每月发生5次P1/P2事件\n"
     "- 团队分布在2个时区\n"
     "- 现有监控覆盖：基础设施metrics、APM、日志聚合\n\n"
     "设计内容：\n1. 事件分级标准（P1-P4）\n"
     "2. 值班轮换制度\n3. 升级路径\n4. 事后复盘流程\n5. 衡量指标",
     ["分级标准明确可操作","轮换制度考虑时区差异","升级路径覆盖所有严重级别","复盘流程有模板","指标可衡量且不鼓励错误行为"],
     {"correctness":0.3,"reasoning_quality":0.4,"practicality":0.3}, 5, ["DESIGN"], "process_doc"),

    ("STRAT-011", 2, "API版本管理策略",
     "你的平台对外提供REST API，已有50个第三方集成在使用v1版本。现在需要做破坏性变更。\n"
     "设计API版本管理策略，包括：\n"
     "1. 版本标识方案（URL路径 vs Header vs 媒体类型）\n"
     "2. 废弃(v1)的时间线和通知机制\n"
     "3. 向后兼容的边界定义\n"
     "4. API变更的分类（哪些算破坏性，哪些不算）",
     ["版本标识方案有理由支撑","废弃时间线对第三方合理","破坏性变更定义清晰","考虑了SemVer和API版本的关系"],
     {"correctness":0.3,"reasoning_quality":0.4,"practicality":0.3}, 4, ["CODE","DESIGN"], "strategy_doc"),

    ("STRAT-012", 5, "平台战略：开放vs封闭",
     "你在运营一个AI Agent平台（类似Multica）。需要决定平台开放策略：\n"
     "1. 完全开放API → 吸引开发者但可能被竞品利用\n"
     "2. 半开放（认证开发者）→ 生态可控但增长慢\n"
     "3. 封闭生态（官方Agent only）→ 质量可控但天花板低\n\n"
     "做全面分析，给出分阶段建议。考虑：开发者生态、商业模型、技术壁垒、数据飞轮。",
     ["三种模式的优劣分析完整","分阶段建议具体","考虑商业模型可持续性","分析了数据飞轮效应","给出可逆决策和不可逆决策的区分"],
     {"correctness":0.2,"reasoning_quality":0.5,"vision":0.3}, 6, ["RESEARCH","DESIGN"], "strategy_paper"),

    ("STRAT-013", 3, "安全事件响应计划",
     "为你的SaaS产品设计安全事件响应计划（SIRP）。场景：发现了生产数据库被未授权访问的迹象。\n"
     "设计：\n1. 检测到确认的时间线（前15分钟做什么）\n"
     "2. 遏制措施\n3. 取证流程\n4. 通知义务（用户、监管）\n5. 恢复和加固",
     ["时间线覆盖T0到T+72h","遏制措施具体可立即执行","取证流程不破坏证据","考虑了GDPR/CCPA等通知义务","恢复后加固措施具体"],
     {"correctness":0.3,"reasoning_quality":0.4,"practicality":0.3}, 5, ["RESEARCH"], "incident_response_plan"),

    ("STRAT-014", 4, "多产品线资源分配",
     "你的公司有3个产品线：\n"
     "A. 成熟产品（占收入70%，增长5%/年）\n"
     "B. 成长产品（占收入25%，增长40%/年）\n"
     "C. 探索产品（占收入5%，增长100%/年，但绝对值小）\n\n"
     "你有50名工程师。设计资源分配方案和决策理由。"
     "需考虑2年后的产品组合目标。",
     ["分配方案有数字支撑","考虑了长期战略而不仅是短期收入","三个产品线有不同的人力策略","考虑了人员流动和知识传承"],
     {"correctness":0.2,"reasoning_quality":0.5,"practicality":0.3}, 5, ["MATH"], "resource_allocation_plan"),

    ("STRAT-015", 2, "什么时候该重写而不是重构",
     "设计一个决策框架，帮助技术负责人判断一个系统是应该重写（Rewrite）还是重构（Refactor）。\n"
     "必须包含：\n1. 评估维度\n2. 定性/定量指标\n3. 决策树\n"
     "4. 至少3个已知案例（如Netscape重写失败、Basecamp重构成功等）",
     ["评估维度覆盖技术/业务/人员","决策树逻辑清晰","案例分析与框架一致","考虑了渐进重写（Strangler Fig）作为中间选项"],
     {"correctness":0.3,"reasoning_quality":0.4,"practicality":0.3}, 4, ["RESEARCH","CODE"], "decision_framework"),

    ("STRAT-016", 5, "技术组织架构设计",
     "为一个150人的技术组织设计团队拓扑。业务场景：\n"
     "- 3条产品线（B2B SaaS、B2C App、API平台）\n"
     "- 共享基础设施（认证、支付、通知、日志）\n"
     "- 需要保持快速迭代（周级发布）同时保证平台稳定性\n\n"
     "设计团队结构、职责划分、协作接口。参考Team Topologies方法论。",
     ["团队拓扑符合Team Topologies原则","共享基础设施的治理模式明确","协作接口定义清晰","考虑了Conway's Law的影响","150人的规模假设合理"],
     {"correctness":0.3,"reasoning_quality":0.4,"practicality":0.3}, 6, ["DESIGN"], "org_design_doc"),

    ("STRAT-017", 3, "定价模型设计",
     "为你的AI Agent平台设计定价模型。产品特点：\n"
     "- 按Agent运行时间计费（类似计算资源）\n"
     "- 客户从个人开发者到企业团队\n"
     "- 竞争对手按subscription收费\n\n"
     "设计：\n1. 定价维度（什么该收费什么不该）\n"
     "2. 分层方案\n3. Free tier设计（吸引用户vs防止滥用）\n4. 企业定价策略",
     ["定价维度逻辑清晰","分层方案覆盖目标客户段","Free tier有防滥用机制","考虑了land-and-expand策略","与竞品的差异分析合理"],
     {"correctness":0.3,"reasoning_quality":0.4,"practicality":0.3}, 5, ["RESEARCH","MATH"], "pricing_strategy"),
]

# ===== RESEARCH DOMAIN (17 tasks) =====
research = [
    ("RESEARCH-001", 3, "向量数据库技术调研",
     "调研当前主流的向量数据库方案，用于RAG（检索增强生成）场景。\n"
     "调研对象至少包含：Pinecone、Weaviate、Milvus、Qdrant、Chroma、pgvector。\n"
     "评估维度：性能（QPS/延迟）、扩展性、运维复杂度、成本、生态集成、过滤能力。\n"
     "给出推荐方案（针对中小团队，数据量100万-1000万向量）和选型理由。",
     ["覆盖至少5个方案","每个方案有量化数据或合理估算","推荐有明确场景限定","说明了不做推荐的方案的缺点","至少引用3个来源"],
     {"correctness":0.4,"completeness":0.3,"source_quality":0.3}, 5, ["CODE"], "research_report"),

    ("RESEARCH-002", 4, "LLM推理优化技术综述",
     "调研当前LLM推理优化的主流技术，包括但不限于：\n"
     "1. 量化（GPTQ、AWQ、GGUF）\n2. 推测解码（Speculative Decoding）\n"
     "3. Flash Attention / Paged Attention\n4. 批处理策略（Continuous Batching）\n"
     "5. KV Cache优化\n6. 模型蒸馏\n\n"
     "对每项技术：原理简述、性能提升幅度（有数据）、适用场景、成熟度评估。"
     "给出实际部署建议（中小规模，单GPU/多GPU推理）。",
     ["覆盖全部6类技术","每项技术有性能数据","成熟度评估合理","部署建议有场景针对性","引用最新论文/博客（2024-2025）"],
     {"correctness":0.4,"completeness":0.3,"source_quality":0.3}, 6, ["CODE","MATH"], "technical_survey"),

    ("RESEARCH-003", 2, "竞品功能矩阵分析",
     "你的产品是AI Agent运行平台。分析以下竞品的功能矩阵：\n"
     "1. LangChain/LangGraph\n2. AutoGPT\n3. CrewAI\n4. Microsoft AutoGen\n5. Anthropic的Agent SDK\n\n"
     "维度：Agent定义方式、任务分解、记忆系统、人机协作、多Agent协调、可观测性、部署方式。\n"
     "输出功能矩阵表格+差异化机会分析。",
     ["功能矩阵覆盖所有维度和竞品","差异化机会不是泛泛而谈","对每个竞品的优劣势有具体判断","至少发现3个可行的差异化方向"],
     {"correctness":0.4,"completeness":0.3,"insight_quality":0.3}, 4, ["STRAT"], "competitive_analysis"),

    ("RESEARCH-004", 5, "AGI评估体系调研",
     "调研当前AGI/LLM评估体系的最新进展。覆盖：\n"
     "1. 学术基准：MMLU、BIG-bench、ARC、HellaSwag等（说明各测什么、局限）\n"
     "2. 实际能力评估：SWE-bench、AgentBench、WebArena\n"
     "3. 人类评估方法：Chatbot Arena、比较判断\n"
     "4. 评估的元问题：基准泄漏、Goodhart's Law、评估与能力的gap\n\n"
     "重点关注：如何评估Agent系统的cognitive能力（不仅是模型能力），"
     "这对DeepSeek的AGI核心业务管培生面试可能有参考价值。",
     ["覆盖全部4个层面","每个基准有局限分析","Agent评估vs模型评估的区分清晰","对元问题（基准泄漏等）有深入讨论","结论部分提炼出面试可用的见解"],
     {"correctness":0.3,"completeness":0.3,"insight_quality":0.4}, 7, ["STRAT","MATH"], "research_report"),

    ("RESEARCH-005", 1, "技术文档调研：gRPC vs REST vs GraphQL",
     "为一个内部微服务通信场景（高吞吐、低延迟、强类型）调研gRPC、REST、GraphQL的适用性。\n"
     "输出：对比表格（性能、工具链、学习曲线、类型安全、调试体验、生态）、推荐方案。\n"
     "至少引用3个生产环境案例。",
     ["对比维度完整","推荐方案有场景针对性","引用了真实生产案例","考虑了中国互联网公司的主流选择"],
     {"correctness":0.4,"completeness":0.3,"source_quality":0.3}, 3, ["CODE"], "comparison_report"),

    ("RESEARCH-006", 3, "开源项目的商业化路径研究",
     "研究5个成功的开源商业化案例：\n"
     "1. GitLab（open core）\n2. Confluent/Kafka（云服务）\n"
     "3. HashiCorp（BSL转换）\n4. Supabase（托管服务）\n5. Redis/Redis Labs（模块化）\n\n"
     "分析：每个案例的初始开源策略、转折点、商业模型演进、社区反应。\n"
     "提炼可复用的商业化路径决策框架。",
     ["5个案例均有具体时间线和事件","转折点分析有洞察","决策框架可操作","考虑了2024-2025年的最新发展"],
     {"correctness":0.3,"completeness":0.3,"insight_quality":0.4}, 5, ["STRAT"], "case_study_report"),

    ("RESEARCH-007", 4, "实时协作底层技术调研",
     "调研实现多人实时协作（类似Google Docs/Figma）的底层技术方案。\n"
     "覆盖：\n1. OT（Operational Transformation）的历史和现状\n"
     "2. CRDT的主流实现（Yjs、Automerge、Loro）\n"
     "3. 同步协议（WebSocket、WebRTC data channel）\n"
     "4. 冲突解决策略\n5. 离线支持和同步\n\n"
     "给出技术选型建议（Web应用，目标用户10万+）。",
     ["OT和CRDT的对比客观","至少3个CRDT库的对比","考虑了Web和移动端的差异","选型建议考虑了团队技术栈"],
     {"correctness":0.4,"completeness":0.3,"source_quality":0.3}, 5, ["CODE"], "technical_survey"),

    ("RESEARCH-008", 2, "数据库选型决策树",
     "研究并输出一个数据库选型决策树（或决策矩阵）。覆盖至少10种数据库：\n"
     "PostgreSQL、MySQL、MongoDB、Redis、Elasticsearch、Cassandra、CockroachDB、"
     "Neo4j、ClickHouse、DuckDB、SQLite。\n\n"
     "输入条件：数据模型（关系/文档/图/时序/向量）、读写比例、一致性要求、"
     "扩展需求、运维能力。输出：推荐数据库+理由。",
     ["决策树逻辑完整","覆盖所有10+数据库","考虑了混合使用场景","推荐不教条（承认PostgreSQL可以覆盖很多场景）"],
     {"correctness":0.4,"completeness":0.3,"practicality":0.3}, 4, ["CODE","STRAT"], "decision_tree"),

    ("RESEARCH-009", 5, "DeepSeek技术栈深度调研",
     "深度调研DeepSeek的技术栈、架构决策和AGI战略。为面试准备。\n"
     "调研内容：\n"
     "1. DeepSeek的模型架构演进（V1→V2→V3，MoE设计）\n"
     "2. 训练基础设施和创新（通信优化、成本控制）\n"
     "3. 产品战略：开源vs API服务，目标用户群\n"
     "4. 组织和文化：团队结构、研究文化（参考公开访谈）\n"
     "5. 与OpenAI/Anthropic/Google的差异化定位\n"
     "6. 挑战和风险\n\n"
     "输出：结构化调研报告 + 面试中可引用的关键见解 + 可以提问的好问题。",
     ["架构演进分析准确","数据有来源支撑","差异化分析不夸大不贬低","见解可在面试中自然引用","问题展示深度思考"],
     {"correctness":0.3,"completeness":0.3,"insight_quality":0.4}, 7, ["STRAT"], "company_deep_dive"),

    ("RESEARCH-010", 3, "WebAssembly应用场景调研",
     "调研WebAssembly（Wasm）在浏览器外（non-browser）的应用场景和生态状态。\n"
     "聚焦:\n1. Serverless/FaaS中的Wasm\n2. 插件系统（如Envoy、Zed编辑器）\n"
     "3. 边缘计算\n4. 区块链智能合约\n\n"
     "评估Wasm在服务端的成熟度和未来2年的发展趋势。",
     ["4个场景都有具体案例分析","成熟度评估有依据","趋势预测有理有据","讨论了Wasm vs Docker的差异化定位"],
     {"correctness":0.4,"completeness":0.3,"source_quality":0.3}, 4, ["CODE","STRAT"], "technology_landscape"),

    ("RESEARCH-011", 1, "Python异步编程最佳实践",
     "调研Python异步编程（asyncio）的最佳实践和常见陷阱。\n"
     "覆盖：\n1. async/await使用模式\n2. 事件循环管理\n3. 并发vs并行的选择\n"
     "4. 第三方库兼容性（requests→httpx，同步ORM→异步ORM）\n"
     "5. 调试和性能分析工具\n\n"
     "输出：最佳实践指南 + 反模式清单。",
     ["覆盖5个方面","最佳实践有代码示例","反模式有具体例子说明危害","工具推荐有版本依据"],
     {"correctness":0.4,"completeness":0.3,"practicality":0.3}, 3, ["CODE"], "best_practices_guide"),

    ("RESEARCH-012", 4, "Agent记忆系统技术调研",
     "调研AI Agent的记忆系统实现方案。覆盖：\n"
     "1. 工作记忆（上下文窗口管理、注意力机制）\n"
     "2. 短期记忆（会话内历史管理、摘要压缩）\n"
     "3. 长期记忆（向量存储、知识图谱、 episodic memory）\n"
     "4. 混合方案（MemGPT、Letta、Mem0等）\n"
     "5. 评估方法：如何衡量记忆系统的效果\n\n"
     "重点关注实际可部署的方案（而非学术原型），给出现有开源方案的对比。",
     ["覆盖全部5类记忆","开源方案对比有量化数据","评估方法有可操作性","部署建议务实"],
     {"correctness":0.4,"completeness":0.3,"source_quality":0.3}, 5, ["CODE","STRAT"], "technical_survey"),

    ("RESEARCH-013", 3, "技术雷达：2025年值得关注的工具",
     "调研2025年值得关注的新兴开发工具和平台。按以下分类：\n"
     "1. AI辅助编程（超Codex/Copilot的下一代工具？）\n"
     "2. 基础设施（新的IaaS/PaaS范式）\n"
     "3. 数据库和数据工具\n"
     "4. 前端和全栈框架\n"
     "5. 可观测性和DevOps\n\n"
     "每个分类推荐1-3个工具，说明为什么值得关注、成熟度、采用风险。",
     ["每个分类有具体工具推荐","推荐有理有据（不是广告）","成熟度评估诚实","采用风险评估务实"],
     {"correctness":0.3,"completeness":0.3,"insight_quality":0.4}, 4, ["STRAT"], "tech_radar"),

    ("RESEARCH-014", 5, "AGI安全与对齐研究现状",
     "调研AGI安全（AI Safety）和对齐（Alignment）的当前研究状态。覆盖：\n"
     "1. 主要研究机构和团队（Anthropic、DeepMind、OpenAI、Conjecture、ARC等）\n"
     "2. 核心技术路线：RLHF、Constitutional AI、RRHF、Debate、Iterated Amplification\n"
     "3. 评估方法：红队测试、自动化对齐评估\n"
     "4. 治理和监管：EU AI Act、美国行政令、中国AI治理框架\n"
     "5. 争议和不同观点（e/acc vs 安全优先）\n\n"
     "对每个路线：理论依据、实际效果、局限。"
     "你的分析对DeepSeek面试中可能的AGI治理讨论有参考价值。",
     ["研究机构和路线覆盖全面","技术路线的分析准确","治理框架信息最新","争议呈现平衡（不站队）","面试可用见解独立标注"],
     {"correctness":0.3,"completeness":0.3,"insight_quality":0.4}, 7, ["STRAT"], "research_report"),

    ("RESEARCH-015", 2, "GitHub Actions vs GitLab CI vs Jenkins",
     "为一个小型团队（10人，GitHub托管代码）做CI/CD工具选型调研。\n"
     "比较：GitHub Actions、GitLab CI、Jenkins、CircleCI、Buildkite。\n"
     "评估：价格、配置复杂度、生态、与代码托管集成度、自托管能力。",
     ["对比表格完整","价格对比基于具体计划","考虑了团队规模","推荐与团队约束一致"],
     {"correctness":0.4,"completeness":0.3,"practicality":0.3}, 3, ["CODE"], "comparison_report"),

    ("RESEARCH-016", 4, "多模态AI应用场景调研",
     "调研多模态AI（文本+图像+音频+视频）在商业场景中的应用现状。\n"
     "覆盖：\n1. 医疗影像分析\n2. 电商视觉搜索\n3. 文档智能处理（OCR+理解）\n"
     "4. 视频内容审核\n5. 多模态客服（图文混合理解）\n\n"
     "每个场景：当前技术成熟度、主要挑战、代表性产品/公司、市场前景。",
     ["5个场景都有深度分析","技术成熟度评估不夸大","挑战分析具体","市场数据有来源"],
     {"correctness":0.4,"completeness":0.3,"source_quality":0.3}, 5, ["STRAT"], "market_landscape"),

    ("RESEARCH-017", 3, "PostgreSQL扩展生态调研",
     "调研PostgreSQL的扩展生态，特别是让PostgreSQL胜任非传统关系型数据库场景的扩展。\n"
     "重点：\n1. PostGIS（地理空间）\n2. pgvector（向量搜索）\n"
     "3. TimescaleDB（时序数据）\n4. Citus（分布式）\n"
     "5. pg_cron + pg_partman（运维自动化）\n"
     "6. AGE（图数据库）\n\n"
     "每个扩展：功能、成熟度、性能、适用场景。回答：PostgreSQL能替代多少专用数据库？",
     ["6个扩展都有深度分析","性能数据有来源","对替代性问题的回答客观平衡","指出了PostgreSQL的局限场景"],
     {"correctness":0.4,"completeness":0.3,"source_quality":0.3}, 4, ["CODE"], "ecosystem_survey"),
]

# ===== MATH DOMAIN (17 tasks) =====
math_tasks = [
    ("MATH-001", 2, "概率问题：骰子游戏期望值",
     "你玩一个游戏：掷三个公平六面骰子。如果三个骰子点数相同（三条），获得100元。"
     "如果恰好两个骰子点数相同（一对），获得10元。如果三个都不同，输掉5元。\n"
     "1. 计算这个游戏的期望值\n2. 确定是否值得玩\n"
     "3. 如果要使游戏公平（期望值为0），三条的奖金应该是多少（保持其他不变）？\n"
     "给出完整计算过程。",
     ["期望值计算过程和结果正确","判断是否值得玩的逻辑正确","奖金调整计算正确","步骤清晰可验证"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 3, [], "math_solution"),

    ("MATH-002", 3, "算法复杂度分析",
     "分析以下算法的精确时间复杂度（非渐进），并给出简化的大O表示：\n\n"
     "算法A：对于输入数组A[0..n-1]，嵌套循环 i从0到n-1，j从i到n-1（步长×2），内部执行O(1)操作。\n"
     "算法B：递归T(n)=3T(n/2)+n^2，T(1)=1。\n"
     "算法C：对于每个长度为k的子串（共n-k+1个），调用O(k)的验证函数。\n\n"
     "对每个算法：1. 建立递推式或求和式 2. 求解 3. 给出大O 4. 举例说明n=100时的操作次数估算。",
     ["三个算法的时间复杂度推导正确","递推式/求和式建模正确","n=100时的数值估算合理","大O表示正确"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 5, ["CODE"], "math_solution"),

    ("MATH-003", 4, "图论：网络可靠性分析",
     "一个通信网络由9个节点和14条边构成。网络的结构如下（邻接表）：\n"
     "0:[1,2,3], 1:[0,4,5], 2:[0,5,6], 3:[0,6], 4:[1,7], 5:[1,2,7,8], 6:[2,3,8], 7:[4,5], 8:[5,6]\n\n"
     "1. 该网络的最小割大小是多少？（即最少切断几条边可以使网络分成两个不连通部分）\n"
     "2. 如果每条边独立以概率p=0.01失效，网络保持连通的概率是多少？（给出计算方法和近似值）\n"
     "3. 如果要使网络连通概率>0.999，每条边的可靠性需要达到多少？（假设独立同分布）\n\n"
     "可以使用算法描述，不需要精确数值计算。",
     ["最小割计算正确","连通概率的计算方法正确","可靠性要求推导合理","步骤可复现"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 5, ["CODE"], "math_solution"),

    ("MATH-004", 1, "贝叶斯推理应用",
     "一个医疗检测的准确率为99%（有病测出阳性的概率）和99%的特异性（没病测出阴性的概率）。\n"
     "该病在人群中的发病率为0.1%。\n"
     "1. 如果一个随机的人检测出阳性，他实际患病的概率是多少？\n"
     "2. 如果检测两次（独立），两次都是阳性，患病的概率是多少？\n"
     "3. 如果发病率变为1%，第一次阳性后的患病概率变化到多少？\n\n"
     "给出完整贝叶斯公式推导。",
     ["贝叶斯公式应用正确","三次计算全部正确","推导步骤完整","用自然语言解释结果（非仅数学符号）"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 3, [], "math_solution"),

    ("MATH-005", 4, "设计一致性哈希算法",
     "设计一致性哈希（Consistent Hashing）的完整数学方案。\n"
     "要求：\n"
     "1. 定义哈希环的数学结构\n"
     "2. 给出节点添加/删除时的数据迁移量公式\n"
     "3. 引入虚拟节点后的负载均衡分析：N个节点各K个虚拟节点，负载方差是多少？\n"
     "4. 证明当K趋于无穷时负载趋于均匀\n\n"
     "给出严谨的数学分析和算法伪代码。",
     ["哈希环的数学定义正确","数据迁移量公式推导正确","负载方差分析有数学支撑","均匀性证明逻辑正确"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 5, ["CODE"], "math_with_pseudocode"),

    ("MATH-006", 3, "排序算法的信息论下界",
     "从信息论角度分析基于比较的排序算法的下界。\n"
     "1. 证明：任何基于比较的排序算法在最坏情况下至少需要ceil(log2(n!))次比较\n"
     "2. 用Stirling近似计算n=100时的下界\n"
     "3. 解释堆排序的O(n log n)为什么接近这个下界\n"
     "4. 讨论：非比较排序（如基数排序）为什么能突破这个下界？",
     ["下界证明完整","Stirling近似应用正确","堆排序分析准确","非比较排序的解释正确"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 4, ["CODE"], "math_proof"),

    ("MATH-007", 5, "设计分布式共识算法的正确性证明",
     "为简化的Paxos协议（单实例，多proposer）给出正确性证明。\n"
     "证明必须包括：\n"
     "1. 安全性（Safety）证明：最多一个值被chosen\n"
     "2. 活性（Liveness）条件分析：在什么条件下协议保证终止？\n"
     "3. 反例构造：如果去掉某个约束，安全性能被违反的构造例子\n\n"
     "使用归纳法和反证法，给出形式化证明。",
     ["安全性证明形式正确","活性分析准确","反例构造有效","归纳法/反证法使用恰当"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 6, ["CODE"], "formal_proof"),

    ("MATH-008", 2, "线性规划建模",
     "一个云服务提供商需要在3个地区部署服务器来满足5个客户区域的需求。\n"
     "已知：\n- 服务器成本（每台/月）：地区1=100, 地区2=120, 地区3=90\n"
     "- 每台服务器可处理1000请求/秒\n"
     "- 5个客户区域的需求（请求/秒）：3000, 2500, 4000, 1500, 2000\n"
     "- 延迟约束：某些服务器-客户配对不可用（延迟过高）\n\n"
     "1. 建立整数线性规划模型（ILP）\n"
     "2. 如果延迟约束为：服务器1不能服务客户3和5，服务器2不能服务客户1，服务器3无限制，求最优解\n"
     "3. 分析LP松弛的解与整数解的差距",
     ["ILP建模正确","最优解计算正确","松弛分析合理","步骤清晰"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 4, ["CODE"], "optimization_solution"),

    ("MATH-009", 3, "马尔可夫链稳态分布",
     "一个Agent在3个状态(s1,s2,s3)间转移，转移矩阵为：\n"
     "P = [[0.3,0.5,0.2],[0.4,0.3,0.3],[0.1,0.6,0.3]]\n\n"
     "1. 计算稳态分布π（即πP=π）\n2. 从初始状态s1出发，经过3步后在s2的概率\n"
     "3. 该马尔可夫链的混合时间（mixing time）大约是多少？（可近似估算）\n"
     "4. 如果Agent在s2时获得reward=10，s1 reward=0，s3 reward=-5，长期平均reward是多少？",
     ["稳态分布计算正确","3步转移概率正确","混合时间估算合理","长期平均reward正确"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 5, ["CODE"], "math_solution"),

    ("MATH-010", 4, "布隆过滤器的假阳性率分析",
     "设计一个布隆过滤器用于10亿个URL的去重。\n"
     "1. 给定假阳性率目标p=0.001，计算最优位数组大小m和哈希函数数量k\n"
     "2. 推导假阳性率公式：p ≈ (1 - e^(-kn/m))^k 的数学过程\n"
     "3. 分析：当插入元素数量从10亿增长到15亿（超出设计容量），假阳性率如何变化？\n"
     "4. 如果内存限制为1GB，能达到的最低假阳性率是多少？",
     ["公式推导完整","参数计算正确","超容量分析正确","内存限制下的计算正确"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 5, ["CODE"], "math_analysis"),

    ("MATH-011", 1, "递归关系求解",
     "求解以下递推式：\n"
     "1. T(n) = T(n-1) + n, T(1)=1\n"
     "2. T(n) = 2T(n/2) + n log n, T(1)=1（主定理）\n"
     "3. T(n) = T(n/3) + T(2n/3) + n, T(1)=1\n\n"
     "对每个递推式：给出求解过程和最终闭合形式，并证明你的答案是正确性（代入验证或归纳法）。",
     ["三个递推式求解正确","证明方法合理","最终闭合形式正确","步骤清晰"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 3, ["CODE"], "math_solution"),

    ("MATH-012", 5, "零知识证明的数学基础",
     "用数学语言描述零知识证明（ZKP）的三个性质：完备性、可靠性、零知识性。\n"
     "给出一个具体的ZKP协议示例（如Schnorr协议或图同构问题的ZKP）：\n"
     "1. 协议步骤描述\n2. 完备性证明\n3. 可靠性证明（soundness error计算）\n"
     "4. 零知识性（构造模拟器）\n\n"
     "使用形式化数学语言，但添加直观解释。",
     ["三个性质的定义准确","协议描述清晰","三个证明逻辑正确","直观解释帮助理解"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 6, ["CODE"], "formal_proof"),

    ("MATH-013", 2, "随机采样算法设计",
     "设计一个从数据流中均匀随机采样k个元素的算法（Reservoir Sampling）。\n"
     "1. 给出算法伪代码\n"
     "2. 用归纳法证明每个元素被选中的概率为k/n\n"
     "3. 扩展到加权采样（每个元素有weight w_i）\n"
     "4. 如果数据流无限（n未知），算法还能保证均匀性吗？",
     ["算法伪代码正确","归纳证明完整","加权扩展正确","无限流分析正确"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 4, ["CODE"], "math_with_pseudocode"),

    ("MATH-014", 3, "信息熵与决策树",
     "给定以下二分类数据集，手动构建最优决策树（使用信息增益作为分裂标准）：\n\n"
     "10个样本，2个特征(X1,X2)，1个标签(Y)。数据如下：\n"
     "(0,0)→0, (0,1)→0, (0,0)→1, (1,0)→0, (1,1)→1,\n"
     "(1,0)→1, (0,1)→1, (1,1)→0, (0,0)→0, (1,1)→1\n\n"
     "1. 计算根节点的熵\n2. 计算每个特征的信息增益\n3. 选择根节点分裂特征\n"
     "4. 递归完成整个树\n5. 计算决策树对训练数据的准确率",
     ["熵计算正确","信息增益计算正确","树结构正确","准确率计算正确"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 4, ["CODE"], "math_solution"),

    ("MATH-015", 4, "排队论在系统设计中的应用",
     "一个API服务处理请求，请求到达服从泊松过程（λ=100 req/s），"
     "每个请求处理时间服从指数分布（均值=8ms）。\n\n"
     "1. 建模为M/M/1队列，计算：平均队列长度、平均等待时间、利用率\n"
     "2. 如果延迟SLA要求P99<50ms，当前配置满足吗？\n"
     "3. 需要增加多少服务器（M/M/c）才能确保P99<30ms？\n"
     "4. 讨论M/M/c模型在实际系统中的局限性（至少3点）",
     ["M/M/1计算正确","SLA分析正确","M/M/c服务器数量估算合理","局限性分析有深度"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 5, ["CODE"], "math_analysis"),

    ("MATH-016", 5, "计算复杂性理论证明",
     "证明以下命题：\n"
     "1. 3-SAT属于NP完全（概述Cook-Levin定理的证明思路）\n"
     "2. 证明子图同构（Subgraph Isomorphism）是NP完全的（归约自Clique）\n"
     "3. 证明2-SAT属于P（给出多项式时间算法）\n\n"
     "对每个命题：给出严谨的证明结构，层次清晰。使用标准复杂度理论符号。",
     ["3-SAT的NP完全证明思路正确","归约正确（Clique≤子图同构）","2-SAT的多项式算法正确","证明结构严谨"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 6, ["CODE"], "formal_proof"),

    ("MATH-017", 3, "HyperLogLog基数估计",
     "分析HyperLogLog算法的数学原理：\n"
     "1. 推导：为什么最长连续零位的期望可以估计基数？\n"
     "2. 证明：n个随机二进制字符串中，最长连续零位≥k的概率\n"
     "3. HLL中使用的调和平均与几何平均/算术平均的区别和优劣\n"
     "4. 计算：标准HLL（2^14=16384个桶）在计数到1亿时的相对误差",
     ["推导过程正确","概率计算正确","调和平均的分析正确","误差计算在合理范围"],
     {"correctness":0.7,"completeness":0.2,"clarity":0.1}, 5, ["CODE"], "math_analysis"),
]

# ===== DESIGN DOMAIN (16 tasks) =====
design = [
    ("DESIGN-001", 3, "设计URL短链系统",
     "设计一个URL短链服务（类似bit.ly），要求：\n"
     "1. 支持每天生成1000万个短链\n2. 短链长度不超过7个字符\n"
     "3. 支持自定义短链（用户指定别名）\n4. 支持访问统计（点击次数、地域、设备）\n"
     "5. 高可用（3个9）\n\n"
     "给出：架构图（文字描述）、数据库设计、API设计、关键算法。讨论短链生成算法（哈希vs自增ID编码）的权衡。",
     ["架构覆盖所有功能需求","数据库设计合理且有扩展性分析","API设计RESTful","短链生成算法的权衡分析完整","容量估算合理"],
     {"correctness":0.3,"design_quality":0.4,"completeness":0.3}, 5, ["CODE","MATH"], "system_design_doc"),

    ("DESIGN-002", 4, "设计实时聊天系统",
     "设计一个支持百万级同时在线的实时聊天系统（类似Discord/微信）。\n"
     "核心功能：\n1. 一对一聊天\n2. 群聊（最多500人）\n"
     "3. 消息已读/未读状态\n4. 离线消息推送\n5. 消息历史搜索\n\n"
     "给出：整体架构、消息流转过程、存储方案（消息/图片/文件不同的存储策略）、扩展方案。"
     "特别讨论：如何保证消息有序性和幂等性。",
     ["架构选型合理","消息有序性和幂等性方案正确","存储分层策略合理","扩展方案可操作","考虑了网络不稳定场景"],
     {"correctness":0.3,"design_quality":0.4,"completeness":0.3}, 6, ["CODE","STRAT"], "system_design_doc"),

    ("DESIGN-003", 2, "设计RESTful API",
     "为一个任务管理系统设计完整的RESTful API。资源包括：\n"
     "项目（Project）、任务（Task）、评论（Comment）、标签（Label）、用户（User）。\n\n"
     "要求：\n1. 完整的CRUD端点\n2. 嵌套资源的设计（如项目下的任务）\n"
     "3. 分页、排序、过滤\n4. 错误码和错误响应格式\n5. API版本管理\n\n"
     "输出完整API文档（OpenAPI/Swagger格式或等价的详细文档）。",
     ["API端点完整且RESTful","嵌套资源设计合理","分页排序过滤设计标准","错误处理完整","文档格式规范"],
     {"correctness":0.3,"design_quality":0.4,"completeness":0.3}, 4, ["CODE"], "api_design_doc"),

    ("DESIGN-004", 5, "设计可扩展的AGI Agent架构",
     "设计一个通用的AGI Agent运行时架构。核心约束：\n"
     "1. 支持多级Agent层次（执行Agent→专家Agent→总管→元认知）\n"
     "2. 可插拔的记忆系统（工作记忆、情景记忆、语义记忆）\n"
     "3. Agent间通信协议（同步RPC+异步消息）\n"
     "4. 工具/能力注册和发现\n"
     "5. 安全沙箱和资源限制\n"
     "6. 可观测性（Agent行为日志、决策追溯、性能监控）\n\n"
     "给出完整架构设计。这不是玩具设计——要考虑Multi-agent协调中的死锁、活锁、资源竞争、"
     "部分失败等分布式系统挑战。这是一个对DeepSeek面试有参考价值的系统设计题。",
     ["架构层次定义清晰且合理","记忆系统插拔设计可行","通信协议考虑了容错","安全沙箱方案可行","可观测性设计完整","讨论了分布式挑战的解决方案"],
     {"correctness":0.3,"design_quality":0.4,"innovation":0.3}, 7, ["STRAT","CODE","RESEARCH"], "architecture_design_doc"),

    ("DESIGN-005", 3, "设计通知系统",
     "为一个SaaS平台设计统一通知系统。需求：\n"
     "1. 多渠道：站内通知、邮件、短信、Webhook、Slack/钉钉/飞书集成\n"
     "2. 用户可配置通知偏好（哪个渠道、什么事件类型）\n"
     "3. 通知聚合（如5分钟内的同类通知合并为一条摘要）\n"
     "4. 高可靠（通知不能丢失）\n5. 模板管理（支持多语言）\n\n"
     "给出：架构设计、数据库模型、通知生命周期、重试和削峰策略。",
     ["多渠道抽象设计合理","用户偏好模型灵活","聚合策略可行","可靠性保证措施明确","模板管理系统设计合理"],
     {"correctness":0.3,"design_quality":0.4,"completeness":0.3}, 5, ["CODE","STRAT"], "system_design_doc"),

    ("DESIGN-006", 4, "设计Feature Flag系统",
     "设计一个企业级Feature Flag（功能开关）系统。要求：\n"
     "1. 多种flag类型：Boolean、Percentage Rollout、Targeted（按用户ID/属性）\n"
     "2. 实时生效（延迟<1秒）\n3. 支持依赖关系（Flag A打开时Flag B才有意义）\n"
     "4. 审计日志（谁在什么时间改了什么flag）\n5. SDK支持多语言\n"
     "6. Kill Switch（紧急关闭某功能的一键开关）\n\n"
     "讨论：如何避免技术债务堆积（flag清理机制）。",
     ["Flag类型覆盖完整","实时生效的方案可行","依赖关系处理正确","审计日志设计完整","Flag清理机制具体可行"],
     {"correctness":0.3,"design_quality":0.4,"completeness":0.3}, 5, ["CODE","STRAT"], "system_design_doc"),

    ("DESIGN-007", 2, "设计数据库Schema：内容平台",
     "为一个内容发布平台设计数据库Schema。业务：\n"
     "用户可以发布文章，文章可以有标签、属于一个分类、可以被其他用户点赞和收藏。"
     "文章可以有多个版本（修订历史）。用户可以关注其他用户。\n\n"
     "给出：ER图（文字描述）、所有表的DDL、关键索引、查询计划分析（对于以下查询："
     "获取一个用户关注的所有作者的最新10篇文章，按发布时间倒序）。\n"
     "使用PostgreSQL。",
     ["ER设计覆盖所有实体和关系","DDL语法正确","索引设计合理","关键查询的查询计划分析正确","考虑了扩展性（分页、大V场景）"],
     {"correctness":0.4,"design_quality":0.4,"completeness":0.2}, 4, ["CODE"], "database_design"),

    ("DESIGN-008", 5, "设计跨数据中心的分布式存储",
     "设计一个支持跨3个数据中心（美东、美西、欧洲）的分布式键值存储系统。\n"
     "要求：\n1. 强一致性（线性一致性）或最终一致性可选\n"
     "2. 容忍单个数据中心故障\n3. 读/写延迟：本地<5ms，跨DC<100ms\n"
     "4. 支持事务（至少单行ACID）\n5. 自动故障转移\n\n"
     "给出：一致性协议选择（Paxos/Raft vs 其他）、数据分片策略、冲突解决、"
     "CAP权衡分析。这不是云服务选型，而是自己设计存储系统。",
     ["一致性协议选择有充分理由","分片策略合理","冲突解决方案具体","CAP分析准确","容灾方案完整"],
     {"correctness":0.3,"design_quality":0.5,"completeness":0.2}, 7, ["CODE","MATH","STRAT"], "architecture_design_doc"),

    ("DESIGN-009", 3, "设计Webhook系统",
     "设计一个可靠的Webhook投递系统。要求：\n"
     "1. 支持注册多个webhook URL（按事件类型订阅）\n"
     "2. 保证投递（at-least-once），接收方幂等去重\n"
     "3. 重试策略（指数退避，最大重试次数可配置）\n"
     "4. 投递状态追踪（成功/失败/待重试）和Dashboard\n"
     "5. 签名验证（HMAC）保证安全\n\n"
     "给出：架构设计、数据库模型、重试状态机、安全设计。",
     ["架构覆盖所有需求","重试状态机设计完整","安全设计（签名验证）正确","状态追踪和Dashboard设计合理"],
     {"correctness":0.3,"design_quality":0.4,"completeness":0.3}, 5, ["CODE"], "system_design_doc"),

    ("DESIGN-010", 4, "设计搜索引擎（文本检索）",
     "为一个文档管理系统设计全文搜索引擎。数据规模：\n"
     "- 1亿个文档，每个平均10KB\n- 每天新增10万文档\n"
     "- 搜索延迟要求P99<200ms\n\n"
     "设计内容：\n1. 倒排索引结构设计\n2. 分词和归一化策略（支持中英文）\n"
     "3. 相关性排序（TF-IDF或BM25）\n4. 索引构建和增量更新\n"
     "5. 分片和副本策略\n\n"
     "可以从零设计，也可以基于Elasticsearch/Meilisearch做架构设计。",
     ["倒排索引设计合理","中英文分词策略正确","排序算法正确","增量更新方案可行","分片策略有计算支撑"],
     {"correctness":0.3,"design_quality":0.5,"completeness":0.2}, 6, ["CODE","MATH"], "system_design_doc"),

    ("DESIGN-011", 2, "设计用户认证系统",
     "为一个微服务架构设计统一的用户认证和授权系统。\n"
     "要求：\n1. OAuth2.0 + OIDC支持\n2. JWT token管理（颁发、刷新、撤销）\n"
     "3. 多租户支持\n4. SSO单点登录\n5. 审计日志\n\n"
     "给出：架构图、token流转过程、安全考量（CSRF、XSS、token存储）、数据库设计。",
     ["OAuth2.0/OIDC流程正确","JWT生命周期管理完整","多租户隔离方案清晰","安全考量覆盖主要攻击向量","架构与微服务兼容"],
     {"correctness":0.3,"design_quality":0.4,"completeness":0.3}, 4, ["CODE","RESEARCH"], "auth_system_design"),

    ("DESIGN-012", 5, "设计代码评审自动化系统",
     "设计一个基于AI的自动化代码评审系统。要求：\n"
     "1. 与GitHub/GitLab PR集成\n2. 多维度评审：正确性、安全性、性能、可维护性、风格\n"
     "3. 评审结果分级：阻塞性(必须修改)、建议性(推荐修改)、参考性(可选)\n"
     "4. 人类评审者的反馈可用来改进AI评审质量\n"
     "5. 支持自定义规则（每个团队可以有自己的编码规范）\n\n"
     "给出：系统架构、AI pipeline设计、反馈回路设计、假阳性/假阴性率控制策略。"
     "特别讨论：如何避免AI评审变成噪声（过多的误报会让人忽略真正重要的发现）。",
     ["系统架构完整且可行","AI pipeline设计合理","反馈回路形成闭环","误报控制策略具体","与现有PR工作流的集成方式自然"],
     {"correctness":0.3,"design_quality":0.4,"innovation":0.3}, 7, ["CODE","STRAT","RESEARCH"], "system_design_doc"),

    ("DESIGN-013", 3, "设计工作流引擎",
     "设计一个通用的工作流（Workflow）引擎。支持：\n"
     "1. DAG定义的工作流（节点=任务，边=依赖）\n2. 条件分支（if/else）\n"
     "3. 并行执行和汇合（fork/join）\n4. 超时和重试\n"
     "5. 人工审批节点\n6. 工作流版本管理\n\n"
     "给出：数据模型、执行引擎设计、状态机、API设计。讨论与Airflow/Temporal的差异化。",
     ["数据模型支持DAG+条件+并行","执行引擎设计合理","状态机覆盖所有状态转换","版本管理方案可行","差异化分析客观"],
     {"correctness":0.3,"design_quality":0.4,"completeness":0.3}, 5, ["CODE"], "system_design_doc"),

    ("DESIGN-014", 1, "设计缓存策略",
     "为一个高并发电商网站设计多级缓存策略。场景：\n"
     "- 商品详情页QPS 10000，商品数据每天更新一次\n"
     "- 用户购物车数据实时性要求高\n"
     "- 商品库存数据需要准实时（延迟<1秒可接受）\n\n"
     "设计：CDN缓存、应用缓存(Redis)、本地缓存的三级策略。"
     "对每种数据类型（商品信息、价格、库存、用户session），给出缓存位置、TTL、失效策略、缓存击穿/雪崩防护。",
     ["三级缓存策略覆盖所有数据类型","TTL设置合理","失效策略正确","缓存击穿和雪崩防护具体","考虑了数据一致性问题"],
     {"correctness":0.3,"design_quality":0.4,"completeness":0.3}, 4, ["CODE","STRAT"], "cache_strategy_doc"),

    ("DESIGN-015", 4, "设计事件驱动架构",
     "为一个电商系统从单体迁移到事件驱动微服务架构做设计。\n"
     "核心事件流：用户下单→扣库存→创建支付→支付确认→通知仓库→发送物流→完成。\n\n"
     "设计：\n1. 事件定义（事件类型、payload schema、版本管理）\n"
     "2. 消息队列/事件总线选型（Kafka vs RabbitMQ vs AWS SQS/SNS vs Redis Streams）\n"
     "3. 事件溯源vs事件通知的选择\n"
     "4. 处理乱序事件、重复事件、丢失事件的策略\n"
     "5. 监控和死信队列\n\n"
     "给出完整的架构设计和关键代码骨架。",
     ["事件定义完整且可扩展","消息队列选型有充分理由","乱序/重复/丢失事件的处理策略正确","监控和死信队列设计合理","架构图清晰"],
     {"correctness":0.3,"design_quality":0.5,"completeness":0.2}, 6, ["CODE","STRAT"], "architecture_design_doc"),

    ("DESIGN-016", 5, "设计AGI系统的自我改进机制",
     "设计一个AI Agent系统能够自我改进的机制。这不是科幻——基于当前LLM的能力边界：\n"
     "1. 错误检测：Agent如何检测自己的输出有误（不需要外部监督）\n"
     "2. 错误分类：将错误按来源分类（知识缺失、推理错误、工具使用错误、通信误解）\n"
     "3. 改进策略：针对每种错误类型，Agent如何自动调整（修改prompt、补充知识、改变策略）\n"
     "4. 改进的持久化：如何让改进跨会话保持（记忆系统存储什么）\n"
     "5. 防止退化：如何确保改进不会导致能力退化（regression test、A/B验证）\n\n"
     "给出完整设计，包括数据结构、算法流程、反馈回路。"
     "对DeepSeek面试有参考价值——展示对AGI系统工程的深度思考。",
     ["错误检测机制可行","错误分类体系合理","改进策略有针对性","持久化方案具体","防止退化机制设计合理","整体设计有理论依据"],
     {"correctness":0.3,"design_quality":0.4,"innovation":0.3}, 7, ["STRAT","CODE","RESEARCH"], "system_design_doc"),
]

# ===== CONTENT DOMAIN (16 tasks) =====
content = [
    ("CONTENT-001", 2, "撰写技术博客：为什么我们选择Go",
     "以技术负责人视角，撰写一篇技术博客，解释团队为什么选择Go语言作为后端主要开发语言。\n"
     "包括：\n1. 决策背景（团队以前用什么，遇到了什么问题）\n"
     "2. 候选语言比较（Go vs Java vs Python vs Rust vs Node.js）\n"
     "3. 关键考量：性能需求、团队学习曲线、生态成熟度、部署运维\n"
     "4. 决策后的效果（1年后的数据对比）\n"
     "5. 如果现在重新选择会不会有不同的决定\n\n"
     "字数：1500-2000字。目标读者：CTO和技术Leader。语气：客观务实，不鼓吹。",
     ["文章结构合理","比较客观（不偏袒）","有数据支撑","语气务实","字数在范围内"],
     {"content_quality":0.4,"structure":0.2,"accuracy":0.2,"readability":0.2}, 4, ["STRAT","CODE"], "blog_post"),

    ("CONTENT-002", 3, "撰写事件复盘报告",
     "基于以下场景撰写一份生产事件复盘（Postmortem）报告：\n\n"
     "事件：某周五晚10点，数据库主库因磁盘满而不可写入，影响50%用户约3小时。"
     "根因：日志表缺少自动清理机制，监控告警阈值设置不合理（95%才告警），"
     "值班人员响应延迟（告警发到了已离职员工的邮箱）。\n\n"
     "报告结构：时间线、影响范围、根因分析（5 Whys）、"
     "改进措施（分技术改进、流程改进、监控改进）、时间线和负责人。"
     "不追究个人责任，聚焦系统改进。语气：专业、客观、无blame。",
     ["时间线清晰准确","5 Whys分析深入（不浅尝辄止）","改进措施具体可执行","有责任人和时间线","语气无blame"],
     {"content_quality":0.4,"structure":0.2,"accuracy":0.2,"readability":0.2}, 4, ["STRAT"], "postmortem_report"),

    ("CONTENT-003", 4, "撰写技术方案文档",
     "为以下需求撰写一份技术方案文档（Technical Design Doc）：\n"
     "需求：将现有的单体后端（Python/Django, 50万行代码）逐步迁移到微服务。\n"
     "第一阶段：将用户认证和订单模块拆分为独立服务。\n\n"
     "文档结构：1. 背景和目标 2. 现状分析 3. 方案设计（含架构图描述） 4. 迁移步骤 5. 风险与缓解 "
     "6. 数据一致性保证 7. 测试策略 8. 回滚方案 9. 时间估算\n\n"
     "目标读者：技术委员会（做架构决策），需要足够的细节来评估可行性。",
     ["文档结构完整","方案具体可操作","风险分析深入","数据一致性方案正确","回滚方案可执行","时间估算合理"],
     {"content_quality":0.4,"structure":0.2,"accuracy":0.2,"readability":0.2}, 5, ["CODE","DESIGN","STRAT"], "technical_design_doc"),

    ("CONTENT-004", 1, "编写README文档",
     "为一个开源项目（AI Agent运行时平台）编写README.md。\n"
     "项目特点：\n- 支持多种LLM后端（OpenAI、Anthropic、本地模型）\n"
     "- Agent可以通过YAML或Python DSL定义\n- 内置记忆系统和工具库\n"
     "- MIT协议开源\n\n"
     "README需要包含：项目简介、快速开始（5分钟上手）、核心特性、安装方法、"
     "简单示例、文档链接、贡献指南概要、License。\n"
     "加分：加入一个badge区域和一张架构示意图（用ASCII art或mermaid）。",
     ["README结构完整","快速开始部分真正可在5分钟内完成","示例代码正确","语气吸引人但不夸大","mermaid图或ASCII art清晰"],
     {"content_quality":0.3,"structure":0.3,"accuracy":0.2,"readability":0.2}, 3, ["CODE"], "readme_doc"),

    ("CONTENT-005", 3, "撰写技术演讲大纲",
     "为一场30分钟的meetup技术分享准备演讲大纲和幻灯片内容结构。\n"
     "主题：从单体到微服务——我们学到了什么\n"
     "听众：中级以上开发者，50-80人。\n\n"
     "大纲应包含：\n1. 每张幻灯片的标题和要点\n"
     "2. 故事线（开头hook、中间展开、结尾takeaway）\n"
     "3. 至少3个真实场景的代码/架构示例\n"
     "4. 互动环节设计（在哪提问、怎么引发讨论）\n"
     "5. 时间分配（30分钟精确到每部分）",
     ["幻灯片结构合理（约20-30张）","故事线有hook和takeaway","示例具体有用","时间分配合理","互动设计自然"],
     {"content_quality":0.4,"structure":0.3,"accuracy":0.1,"readability":0.2}, 4, ["CODE","STRAT"], "presentation_outline"),

    ("CONTENT-006", 4, "撰写技术白皮书摘要",
     "为一篇技术白皮书撰写执行摘要（Executive Summary）。白皮书主题：\n"
     "AI Agent系统在生产环境中的可靠性研究。\n\n"
     "研究内容（你需要合理想象）：\n"
     "- 对50个使用AI Agent的生产系统进行了6个月的可靠性分析\n"
     "- 测量了Agent执行任务的错误率、自恢复率、幻觉率\n"
     "- 提出了一套提升Agent可靠性的架构模式\n\n"
     "执行摘要要求：500-800字，包含：问题背景、研究方法、关键发现（3-5个）、结论和建议。"
     "读者是忙的CTO/VP Engineering，他们可能只读这一页。",
     ["执行摘要字数在范围内","关键发现有数据支撑","结论有可操作建议","读起来像真正的白皮书摘要","适合CTO/VP阅读"],
     {"content_quality":0.5,"structure":0.2,"accuracy":0.15,"readability":0.15}, 4, ["RESEARCH","STRAT"], "executive_summary"),

    ("CONTENT-007", 2, "撰写API变更通知",
     "你的平台需要通知所有API用户：v1版本的3个端点将在3个月后废弃。\n"
     "撰写：\n1. 邮件通知（发给注册开发者）\n2. 变更日志（放在开发者门户）\n"
     "3. v1→v2迁移指南（只涉及3个端点：用户信息、订单列表、支付回调）\n\n"
     "关键：语气要体现出对开发者时间的尊重，不能只是冷冰冰的公告。"
     "迁移指南必须包含清晰的代码对比（v1调用方式 vs v2调用方式）。",
     ["邮件通知语气恰当","变更日志完整","迁移指南代码对比清晰","三个端点都有覆盖","时间线和联系方式明确"],
     {"content_quality":0.3,"structure":0.25,"accuracy":0.25,"readability":0.2}, 3, ["CODE"], "api_migration_guide"),

    ("CONTENT-008", 5, "撰写AGI发展年度综述",
     "撰写一篇关于2024-2025年AGI进展的年度综述文章。\n"
     "覆盖：\n1. 模型能力进展（GPT、Claude、Gemini、DeepSeek、Qwen等）\n"
     "2. Agent系统发展（从ReAct到多Agent协作）\n"
     "3. 基础设施和工具链（LangChain、AutoGen、CrewAI、Multica等）\n"
     "4. 安全和治理进展\n"
     "5. 开源vs闭源的态势变化\n"
     "6. 对2026年的5个预测\n\n"
     "字数：3000-4000字。目标读者：AI从业者。要求：每个论述有具体案例和数据支撑，不空泛。"
     "对DeepSeek面试有参考价值——展示对AGI全局生态的理解。",
     ["覆盖6个方面","每个论述有案例/数据支撑","预测有理有据","不空泛不喊口号","面试可用见解可单独标注"],
     {"content_quality":0.5,"structure":0.2,"accuracy":0.2,"readability":0.1}, 5, ["RESEARCH","STRAT"], "annual_review"),

    ("CONTENT-009", 2, "撰写面试题设计文档",
     "为招聘高级后端工程师设计一套技术面试题目。包括：\n"
     "1. 算法题1道（中等难度，含示例输入输出和评分标准）\n"
     "2. 系统设计题1道（含评判维度和好的答案应该覆盖的要点）\n"
     "3. 代码审查题1道（给一段有问题的代码，列出期望候选人发现的问题）\n"
     "4. 行为面试题3道（含评估维度）\n\n"
     "每道题需要：题目描述、时间分配、评分rubric、好的回答示例要点。",
     ["算法题难度适中且有评分标准","系统设计题评判维度合理","代码审查题的问题点设计合理","行为题评估维度明确","整体难度匹配高级工程师级别"],
     {"content_quality":0.3,"structure":0.3,"accuracy":0.2,"readability":0.2}, 4, ["CODE","DESIGN"], "interview_design_doc"),

    ("CONTENT-010", 3, "撰写架构决策记录(ADR)",
     "为以下3个架构决策撰写ADR（Architecture Decision Record）：\n"
     "1. 选择PostgreSQL而非MongoDB作为主数据库\n"
     "2. 采用事件驱动架构进行服务间通信\n"
     "3. 使用Kubernetes进行容器编排\n\n"
     "每个ADR遵循标准格式：Title、Status、Context、Decision、Consequences。\n"
     "后果部分要特别详细——正面的和负面的都要诚实记录。",
     ["三个ADR格式标准","Context描述充分（为什么需要做这个决策）","Decision部分明确（不模糊）","Consequences诚实（正面+负面）","整体读起来像真实项目文档"],
     {"content_quality":0.3,"structure":0.3,"accuracy":0.2,"readability":0.2}, 3, ["CODE","STRAT"], "adr_documents"),

    ("CONTENT-011", 4, "撰写产品需求文档(PRD)",
     "为以下功能撰写PRD：\n"
     "功能：AI Agent平台的\"Agent市场\"——用户可以发布、发现、复用其他用户创建的Agent模板。\n\n"
     "PRD结构：\n1. 背景和问题\n2. 目标用户和使用场景\n3. 功能需求（P0/P1/P2优先级）\n"
     "4. 非功能需求（安全、性能、扩展性）\n5. 用户故事\n"
     "6. 成功的衡量指标\n7. 与竞品的差异化\n8. 风险和假设\n\n"
     "这不是随便写写——要像真正的PM一样思考用户痛点和商业价值。",
     ["PRD结构完整","用户场景具体非虚构","优先级划分有理有据","成功指标可量化","竞品差异化真实"],
     {"content_quality":0.4,"structure":0.3,"accuracy":0.1,"readability":0.2}, 5, ["STRAT","DESIGN"], "prd_document"),

    ("CONTENT-012", 1, "翻译并本地化技术文档章节",
     "将以下英文技术文档的摘要翻译为中文，并进行本地化（不仅仅是翻译，要适应中文技术阅读习惯）：\n\n"
     "\"This library provides a unified interface for interacting with various LLM providers, "
     "supporting streaming responses, tool calling, and multi-modal inputs. "
     "It handles authentication, rate limiting, and automatic retries transparently. "
     "The architecture is plugin-based, allowing community contributions for new providers. "
     "Performance benchmarks show <50ms overhead compared to direct API calls in most scenarios.\"\n\n"
     "要求：\n1. 自然流畅的中文\n2. 技术术语处理得当（哪些保留英文、哪些翻译）\n"
     "3. 保持原文的技术精确性\n4. 添加中文读者可能需要额外了解的信息（如对国内开发者的相关替代品）",
     ["翻译准确不失真","中文流畅自然","术语处理得当","本地化信息有价值（非堆砌）"],
     {"content_quality":0.4,"accuracy":0.3,"readability":0.2,"localization":0.1}, 2, ["CODE"], "localized_content"),

    ("CONTENT-013", 5, "撰写AGI伦理与治理白皮书大纲",
     "撰写一份关于AGI伦理与治理的白皮书详细大纲。覆盖：\n"
     "1. AGI的定义和当前能力边界\n"
     "2. 核心伦理问题：偏见与公平、透明度与可解释性、隐私、责任归属\n"
     "3. 安全风险：对齐失败、能力涌现、误用与武器化\n"
     "4. 治理框架：各国监管路径对比（EU/US/CN）、行业自律、技术保障\n"
     "5. 推荐行动框架：对不同角色（开发者、部署者、监管者、公众）的建议\n\n"
     "大纲要详细到3级标题，每节标注预计字数和核心论点。对DeepSeek面试中的AGI治理讨论有参考价值。",
     ["大纲结构完整（到3级标题）","覆盖全部5个方面","每节核心论点有实质内容","不同国家监管路径对比客观","对DeepSeek面试的参考价值明确"],
     {"content_quality":0.5,"structure":0.3,"accuracy":0.1,"readability":0.1}, 5, ["STRAT","RESEARCH"], "whitepaper_outline"),

    ("CONTENT-014", 3, "撰写Onboarding文档",
     "为由5名新入职的软件工程师撰写一份2周的Onboarding计划文档。\n"
     "背景：公司是一个50人的SaaS公司，技术栈是Go+React+PostgreSQL+Kubernetes。\n\n"
     "文档包含：\n1. 第1周：环境搭建、代码库概览、架构介绍、第一个小任务\n"
     "2. 第2周：深入模块、参与Code Review、独立完成一个小feature\n"
     "3. 每个阶段的学习资源和mentor指导要点\n"
     "4. 检查点：第1周末和第2周末分别需要达到什么水平\n"
     "5. 新人常见问题和答案（FAQ）",
     ["计划可执行（每天有具体安排）","学习资源具体（链接/文档名）","检查点标准明确","FAQ覆盖真正常见的问题","文档语气友好包容"],
     {"content_quality":0.3,"structure":0.3,"accuracy":0.2,"readability":0.2}, 4, ["CODE"], "onboarding_guide"),

    ("CONTENT-015", 2, "撰写Release Notes",
     "为以下虚构的大版本更新撰写Release Notes：\n"
     "产品：AI Agent开发平台 v2.0\n"
     "主要更新：\n1. 全新的Agent可视化编辑器（拖拽式DAG设计）\n"
     "2. 内置Agent模板市场（50+预置模板）\n"
     "3. Agent协作模式（多Agent可组成squad）\n"
     "4. 性能优化（Agent启动速度提升3倍）\n"
     "5. 破坏性变更：Agent定义格式从YAML迁移到JSON Schema\n"
     "6. 修复：修复了15个已知bug\n"
     "7. 安全：新增Agent执行沙箱\n\n"
     "Release Notes需要包含：亮点、详细变更、升级指南（针对破坏性变更）、已知问题。"
     "语气：专业但有人情味（不是干巴巴的commit log）。",
     ["亮点突出但不浮夸","详细变更覆盖所有更新","升级指南清晰可操作","已知问题诚实列出","语气专业有人情味"],
     {"content_quality":0.3,"structure":0.25,"accuracy":0.25,"readability":0.2}, 3, ["CODE"], "release_notes"),

    ("CONTENT-016", 4, "撰写投资意向书执行摘要",
     "你是一家AI Agent基础设施公司的CTO，需要为A轮融资撰写一份技术部分的执行摘要。\n"
     "公司背景（可以是虚构的，但要合理）：\n"
     "- 产品：面向开发者的AI Agent运行和管理平台\n"
     "- 成立18个月，目前有500+注册开发者，20个付费企业客户\n"
     "- 核心差异化：支持Multi-Agent协作、内置记忆系统、开源核心+云服务\n"
     "- A轮融资目标：500万美元\n\n"
     "技术部分需要包含：\n1. 技术架构概述\n2. 核心技术壁垒（为什么别人不能轻易复制）\n"
     "3. 技术路线图（融到钱后12个月的开发计划）\n4. 技术团队简介\n"
     "5. 关键技术指标（性能、可靠性等）\n\n"
     "语气：自信但不夸大。投资人能看懂的技术语言。",
     ["技术架构描述清晰","技术壁垒分析言之有物（不是笼统的AI）","路线图具体可执行","技术指标有数字","投资人不需AI背景也能理解价值主张"],
     {"content_quality":0.5,"structure":0.2,"accuracy":0.15,"readability":0.15}, 5, ["STRAT","DESIGN"], "pitch_deck_technical"),
]

def make_task(t):
    return {
        "id": t[0],
        "domain": t[0].split("-")[0],
        "difficulty": t[1],
        "title": t[2],
        "description": t[3],
        "acceptance_criteria": t[4],
        "scoring_rubric": t[5],
        "min_steps": t[6],
        "cross_domain": t[7],
        "expected_output_type": t[8]
    }

all_tasks = []
for t in code + strat + research + math_tasks + design + content:
    all_tasks.append(make_task(t))

output = {
    "meta": {
        "version": "1.0.0",
        "created_at": "2026-06-29T04:30:00Z",
        "total_tasks": len(all_tasks),
        "domains": {
            "CODE": {"count": 17, "description": "代码工程：架构设计、bug修复、重构、新功能实现"},
            "STRAT": {"count": 17, "description": "策略决策：技术选型、架构评估、权衡分析"},
            "RESEARCH": {"count": 17, "description": "调研分析：技术调研、竞品分析、可行性评估"},
            "MATH": {"count": 17, "description": "数学逻辑：算法设计、复杂度分析、形式化推理"},
            "DESIGN": {"count": 16, "description": "设计创意：系统设计、API设计、UX方案"},
            "CONTENT": {"count": 16, "description": "内容产出：技术文档、报告撰写、知识整理"}
        },
        "complexity_standards": {
            "min_steps": 3,
            "cross_domain_required": True,
            "ambiguity_present": True,
            "output_verifiable": True,
            "target_baseline_completion": "30-70%"
        }
    },
    "tasks": all_tasks
}

with open('/home/multica/multica_workspaces/1821ae26-62e1-4744-9e0d-3d506c93cc9f/ccfc924b/workdir/experiment/benchmark_tasks.json', 'w') as f:
    json.dump(output, f, ensure_ascii=False, indent=2)

# Print summary
from collections import Counter
domains = Counter(t['domain'] for t in all_tasks)
for d, c in sorted(domains.items()):
    difficulties = Counter(t['difficulty'] for t in all_tasks if t['domain'] == d)
    print(f"{d}: {c} tasks, difficulty distribution: {dict(sorted(difficulties.items()))}")
print(f"\nTotal: {len(all_tasks)} tasks")
print("Written to experiment/benchmark_tasks.json")
