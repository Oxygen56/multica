## Group A 基线实验 — 批次 6（10个任务）

### 执行模式：单层Sys2（独立执行，不审查，不委托）

### 任务列表：
- **CONTENT-008** (CONTENT, 难度5): 撰写AGI发展年度综述
- **DESIGN-008** (DESIGN, 难度5): 设计跨数据中心的分布式存储
- **CODE-017** (CODE, 难度4): 实现共识算法Raft的领导者选举
- **STRAT-004** (STRAT, 难度5): AGI架构设计决策
- **STRAT-006** (STRAT, 难度4): 数据迁移策略
- **RESEARCH-007** (RESEARCH, 难度4): 实时协作底层技术调研
- **DESIGN-015** (DESIGN, 难度4): 设计事件驱动架构
- **CONTENT-014** (CONTENT, 难度3): 撰写Onboarding文档
- **STRAT-010** (STRAT, 难度4): On-Call事件响应流程设计
- **RESEARCH-009** (RESEARCH, 难度5): DeepSeek技术栈深度调研

### 执行要求

1. 对每个任务，**产出完整答案**，写入 `experiment/results/raw/{task_id}_answer.md`
2. 对照任务描述中的 acceptance_criteria 自评
3. 记录指标到 `experiment/results/results_group_A.jsonl`（参考已有格式）
4. 每个任务独立完成，不要跳过
5. 全部10个完成后，更新此issue为 done

### 任务详情请读取 experiment/benchmark_tasks.json
本批次任务ID: CONTENT-008, DESIGN-008, CODE-017, STRAT-004, STRAT-006, RESEARCH-007, DESIGN-015, CONTENT-014, STRAT-010, RESEARCH-009