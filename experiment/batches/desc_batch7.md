## Group A 基线实验 — 批次 7（10个任务）

### 执行模式：单层Sys2（独立执行，不审查，不委托）

### 任务列表：
- **MATH-010** (MATH, 难度4): 布隆过滤器的假阳性率分析
- **CODE-007** (CODE, 难度4): 实现无锁并发队列
- **MATH-016** (MATH, 难度5): 计算复杂性理论证明
- **DESIGN-010** (DESIGN, 难度4): 设计搜索引擎（文本检索）
- **DESIGN-012** (DESIGN, 难度5): 设计代码评审自动化系统
- **MATH-007** (MATH, 难度5): 设计分布式共识算法的正确性证明
- **RESEARCH-016** (RESEARCH, 难度4): 多模态AI应用场景调研
- **DESIGN-006** (DESIGN, 难度4): 设计Feature Flag系统
- **RESEARCH-012** (RESEARCH, 难度4): Agent记忆系统技术调研
- **CODE-010** (CODE, 难度5): 设计分布式事务方案

### 执行要求

1. 对每个任务，**产出完整答案**，写入 `experiment/results/raw/{task_id}_answer.md`
2. 对照任务描述中的 acceptance_criteria 自评
3. 记录指标到 `experiment/results/results_group_A.jsonl`（参考已有格式）
4. 每个任务独立完成，不要跳过
5. 全部10个完成后，更新此issue为 done

### 任务详情请读取 experiment/benchmark_tasks.json
本批次任务ID: MATH-010, CODE-007, MATH-016, DESIGN-010, DESIGN-012, MATH-007, RESEARCH-016, DESIGN-006, RESEARCH-012, CODE-010