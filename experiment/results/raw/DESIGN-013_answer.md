# DESIGN-013: 设计工作流引擎 — 答案（Group A 基线）

## 数据模型

```sql
CREATE TABLE workflows (id UUID PK, name TEXT, version INT, definition JSONB NOT NULL);
-- definition: {"nodes":[...], "edges":[...], "conditions":{...}}

CREATE TABLE workflow_runs (id UUID PK, workflow_id FK, status TEXT, input JSONB, started_at, completed_at);
CREATE TABLE task_instances (id UUID PK, run_id FK, node_id TEXT, type TEXT, -- task/condition/fork/join/approval
                             status TEXT, -- pending/running/completed/failed/skipped
                             input JSONB, output JSONB, retry_count INT DEFAULT 0,
                             timeout_seconds INT, started_at, completed_at);
```

## DAG执行引擎

```python
class WorkflowEngine:
    def execute(self, workflow_id: str, input: dict) -> str:
        wf = self.load_workflow(workflow_id)
        run = self.create_run(workflow_id, input)
        
        ready = self._get_ready_nodes(wf, run_id=run.id)  # 入度为0的节点
        while ready:
            for node in ready:
                if node.type == 'fork':
                    for branch in node.branches:
                        self._enqueue_task(run.id, branch)
                elif node.type == 'join':
                    if self._all_branches_done(run.id, node):
                        self._enqueue_task(run.id, node)
                elif node.type == 'condition':
                    result = self._evaluate_condition(node.expression, run.context)
                    next_node = node.true_branch if result else node.false_branch
                    self._enqueue_task(run.id, next_node)
                elif node.type == 'approval':
                    self._wait_for_approval(run.id, node)  # 暂停，等待人工
                else:  # task
                    self._enqueue_task(run.id, node)
            ready = self._get_ready_nodes(wf, run.id)
        return run.id
```

## 状态机

```
pending → running → completed
                  → failed → retrying (retry_count < max) → running
                           → failed (永久) → 触发补偿/告警
                  → timed_out → retrying/failed
                  → waiting_approval (人工节点)
```

## 差异化：vs Airflow / Temporal

| | 本引擎 | Airflow | Temporal |
|---|--------|---------|----------|
| 定位 | 应用内嵌工作流 | 数据Pipeline调度 | 微服务编排 |
| 触发 | API调用 | 定时cron | SDK调用 |
| 人工审批 | ✅ 一等公民 | ⚠️ Sensor | ✅ Signal |
| 部署 | 嵌入应用 | 独立集群 | 独立集群 |
| 复杂度 | 低 | 中 | 高 |

## 自评

- ✅ 数据模型支持DAG+条件+并行
- ✅ 执行引擎设计合理（fork/join/condition/approval）
- ✅ 状态机完整
- ✅ 版本管理可行
- ✅ 差异化分析客观（内嵌 vs 独立集群）

**完成** | 修复轮次: 0
