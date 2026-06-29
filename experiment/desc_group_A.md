## 任务：Group A 基线实验 — 100任务执行（单层Sys2）

### 实验配置

- **组别**：A组（基线）
- **架构**：Sys2 only — 仅总管，无监察官，无agent池，无记忆系统
- **模式**：独立执行，不审查，不委托

### 执行说明

1. 加载实验任务文件：`experiment/benchmark_tasks.json`（100个任务）
2. 使用实验框架：`python3 experiment/runner.py --group A --prepare` 初始化批次
3. 按随机化顺序逐任务执行：
   - 读取任务描述
   - 独立产出完整答案/交付物
   - 对照 acceptance_criteria 自评
   - 记录指标到 `experiment/results/results_group_A.jsonl`
4. 每完成一个任务，更新结果文件
5. 全部100个任务完成后，运行 `python3 experiment/runner.py --group A --summary` 产出汇总

### 关键约束

- **不委托**：所有任务由总管直接执行，不使用Sys1 agent池
- **不审查**：产出不经过监察官审查，直接记录（模拟单层架构）
- **记录所有指标**：completion_rate, first_pass_rate, avg_fix_rounds, error_density, time_seconds, token_consumption
- **保留原始产出**：每个任务的回答内容保存到 `experiment/results/raw/` 目录

### 验收标准

- 100个任务全部执行完毕
- 每个任务有完整的指标记录
- 原始产出可追溯
- 汇总统计产出
