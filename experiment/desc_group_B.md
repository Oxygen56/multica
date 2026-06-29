## 任务：Group B 两层实验 — 100任务执行（Sys2+Sys3）

### 实验配置

- **组别**：B组（两层）
- **架构**：Sys2 + Sys3 — 总管执行 + 监察官审查
- **模式**：总管产出 → 监察官审查 → 修改 → 再审查 → 交付

### 执行说明

1. 加载实验任务文件：`experiment/benchmark_tasks.json`（100个任务）
2. 使用实验框架：`python3 experiment/runner.py --group B --prepare` 初始化批次
3. 对每个任务：
   - 总管独立产出答案
   - 提交给监察官审查
   - 根据审查意见修改（如有）
   - 监察官通过后记录指标
   - 记录审查轮次、修改内容
4. 指标记录到 `experiment/results/results_group_B.jsonl`
5. 全部完成后运行 `python3 experiment/runner.py --group B --summary`

### 关键约束

- **监察官审查**：每个产出必须经监察官审查。监察官检查：事实准确性、逻辑完整性、是否满足acceptance_criteria
- **记录审查数据**：每次审查的意见和修改轮次
- **不委托执行**：总管自己执行任务，不使用Sys1 agent池
- **记录所有指标**：包括审查轮次（avg_fix_rounds）

### 验收标准

- 100个任务全部执行完毕，全部经过监察官审查
- 每个任务有完整的指标记录（含审查轮次）
- 与Group A使用相同的100个任务（便于对照）
