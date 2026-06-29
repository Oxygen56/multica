## Group A 基线实验 — 批次 8（10个任务）

### 执行模式：单层Sys2（独立执行，不审查，不委托）

### 任务列表：
- **STRAT-014** (STRAT, 难度4): 多产品线资源分配
- **MATH-015** (MATH, 难度4): 排队论在系统设计中的应用
- **CONTENT-016** (CONTENT, 难度4): 撰写投资意向书执行摘要
- **CONTENT-013** (CONTENT, 难度5): 撰写AGI伦理与治理白皮书大纲
- **STRAT-002** (STRAT, 难度4): 微服务拆分边界决策
- **MATH-005** (MATH, 难度4): 设计一致性哈希算法
- **CONTENT-006** (CONTENT, 难度4): 撰写技术白皮书摘要
- **STRAT-008** (STRAT, 难度5): 构建vs购买vs开源决策框架
- **RESEARCH-002** (RESEARCH, 难度4): LLM推理优化技术综述
- **DESIGN-016** (DESIGN, 难度5): 设计AGI系统的自我改进机制

### 执行要求

1. 对每个任务，**产出完整答案**，写入 `experiment/results/raw/{task_id}_answer.md`
2. 对照任务描述中的 acceptance_criteria 自评
3. 记录指标到 `experiment/results/results_group_A.jsonl`（参考已有格式）
4. 每个任务独立完成，不要跳过
5. 全部10个完成后，更新此issue为 done

### 任务详情请读取 experiment/benchmark_tasks.json
本批次任务ID: STRAT-014, MATH-015, CONTENT-016, CONTENT-013, STRAT-002, MATH-005, CONTENT-006, STRAT-008, RESEARCH-002, DESIGN-016