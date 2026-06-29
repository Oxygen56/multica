## 任务：Group C 完整架构实验 — 100任务执行（五层+D）

### 实验配置

- **组别**：C组（完整五层+D）
- **架构**：Sys0-4全层 + 双层记忆系统
- **组件**：Sys0-执行Agent × 2, Sys1-战术Agent池 × 5, Sys2-总管, Sys3-监察官, Sys4-记录Agent
- **模式**：Sys0分解 → Sys1池并行执行 → Sys2协调汇总 → Sys3审查 → Sys4记录

### 执行说明

1. 加载实验任务文件：`experiment/benchmark_tasks.json`（100个任务）
2. 使用实验框架：`python3 experiment/runner.py --group C --prepare` 初始化批次
3. 对每个任务：
   - Sys0（执行Agent）将任务分解为子任务
   - Sys1池并行执行子任务（利用5个战术Agent）
   - Sys2（总管）汇总、协调、整合产出
   - Sys3（监察官）审查最终产出
   - Sys4（记录Agent）记录所有中间步骤和决策
   - 利用工作记忆（跨会话上下文）和情景记忆（关键决策检索）
4. 指标记录到 `experiment/results/results_group_C.jsonl`
5. 全部完成后运行 `python3 experiment/runner.py --group C --summary`

### 关键约束

- **全层参与**：五层架构全部参与，不得跳过任何层
- **记忆系统启用**：每个任务的上下文和关键决策写入记忆系统
- **并行执行**：Sys1池利用4-5个并发agent并行处理子任务
- **审查和记录**：Sys3审查 + Sys4完整记录
- **记录所有指标**：包括token消耗（会显著高于A/B组）

### 验收标准

- 100个任务全部执行完毕，五层架构全链路参与
- 每个任务有完整的指标记录和中间步骤记录
- 记忆系统中有完整的决策追溯
- Sys4记录完整可复现
