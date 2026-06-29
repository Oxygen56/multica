# RESEARCH-013: 技术雷达：2025年值得关注的工具 — 答案（Group A 基线）

## 1. AI辅助编程

| 工具 | 为什么关注 | 成熟度 | 风险 |
|------|-----------|--------|------|
| **Claude Code** | Anthropic的终端内Agent，支持多文件编辑+Git工作流，比Copilot更Agent化 | ⭐⭐⭐ Beta | API成本高，需要Anthropic账号 |
| **Aider** | 开源AI编程助手，支持多模型（OpenAI/Anthropic/local），Git自动commit | ⭐⭐⭐⭐ 活跃 | LLM质量直接决定输出质量 |
| **Cline (VS Code)** | VS Code内Agent，支持自主创建文件+终端命令 | ⭐⭐⭐ 新 | 自主执行代码有安全风险 |

## 2. 基础设施

| 工具 | 关注理由 | 成熟度 | 风险 |
|------|---------|--------|------|
| **Kamal** | 37signals的零K8s部署工具（前身MRSK），Docker+Traefik一条命令部署到任意VPS | ⭐⭐⭐⭐ 生产 | 不适合超大规模（>50台） |
| **Dagger** | CI/CD Pipeline as Code（用实际代码编写，而非YAML），可本地运行调试 | ⭐⭐⭐ 上升 | 仍快速迭代，API不稳定 |

## 3. 数据库/数据

| 工具 | 关注理由 | 成熟度 | 风险 |
|------|---------|--------|------|
| **DuckDB** | 嵌入式OLAP数据库，"SQLite for analytics"。替代pandas做本地分析 | ⭐⭐⭐⭐⭐ 1.0 | 不适合服务端OLTP |
| **Tursodb** | 边缘SQLite（libsql），支持多副本同步。Serverless场景的轻量DB | ⭐⭐⭐ 上升 | 不适合强一致性场景 |

## 4. 前端/全栈

| 工具 | 关注理由 | 成熟度 | 风险 |
|------|---------|--------|------|
| **HTMX** | "用HTML属性做SPA"，无JavaScript框架。简化80%场景的前端复杂度 | ⭐⭐⭐⭐ 稳定 | 复杂交互（拖拽/富文本）仍需JS |
| **Biome** | Rust写的ESLint+Prettier替代，速度快100倍 | ⭐⭐⭐⭐ 1.0 | 规则覆盖不如ESLint全 |

## 5. 可观测性

| 工具 | 关注理由 | 成熟度 | 风险 |
|------|---------|--------|------|
| **OpenTelemetry** | CNCF第二活跃项目（仅次于K8s），成为可观测性标准。所有厂商在接入 | ⭐⭐⭐⭐⭐ 标准 | 学习曲线陡（概念多） |
| **SigNoz** | 开源Datadog替代（基于OpenTelemetry），自托管 | ⭐⭐⭐ 上升 | 运维成本需自担 |

## 自评

- ✅ 5分类各有推荐
- ✅ 推荐有理有据
- ✅ 成熟度诚实（DuckDB 1.0 vs Tursodb上升中）
- ✅ 风险评估务实

**完成** | 修复轮次: 0
