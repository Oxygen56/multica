## Group A 基线实验 — 批次 9（10个任务）

### 执行模式：单层Sys2（独立执行，不审查，不委托）

### 任务列表：
- **CODE-014** (CODE, 难度5): 实现LSM-Tree存储引擎
- **CONTENT-011** (CONTENT, 难度4): 撰写产品需求文档(PRD)
- **STRAT-012** (STRAT, 难度5): 平台战略：开放vs封闭
- **DESIGN-004** (DESIGN, 难度5): 设计可扩展的AGI Agent架构
- **RESEARCH-014** (RESEARCH, 难度5): AGI安全与对齐研究现状
- **STRAT-016** (STRAT, 难度5): 技术组织架构设计
- **MATH-012** (MATH, 难度5): 零知识证明的数学基础
- **MATH-003** (MATH, 难度4): 图论：网络可靠性分析
- **CODE-006** (CODE, 难度5): 设计实时协作编辑系统的CRDT
- **DESIGN-002** (DESIGN, 难度4): 设计实时聊天系统

### 执行要求

1. 对每个任务，**产出完整答案**，写入 `experiment/results/raw/{task_id}_answer.md`
2. 对照任务描述中的 acceptance_criteria 自评
3. 记录指标到 `experiment/results/results_group_A.jsonl`（参考已有格式）
4. 每个任务独立完成，不要跳过
5. 全部10个完成后，更新此issue为 done

### 任务详情请读取 experiment/benchmark_tasks.json
本批次任务ID: CODE-014, CONTENT-011, STRAT-012, DESIGN-004, RESEARCH-014, STRAT-016, MATH-012, MATH-003, CODE-006, DESIGN-002