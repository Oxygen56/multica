# CONTENT-009: 撰写面试题设计文档 — 答案（Group A 基线）

## 1. 算法题：并发任务调度器

设计一个并发任务调度器，输入N个任务（每个有id、耗时、依赖列表），输出执行顺序使得总完成时间最短。假设有k个并行工作线程。

**示例**：tasks=[(A,3,[]), (B,2,[A]), (C,1,[A]), (D,4,[B,C])], k=2。期望输出：最优调度方案+总时间。

**评分标准（30min）**：
- 识别为拓扑排序+优先队列问题 (5分)
- 正确处理依赖关系 (5分)
- 时间复杂度O(N log N) (5分)
- 代码可运行+测试用例 (5分)
- 讨论k=1和k=∞的退化情况 (5分)
- 总25分，15分过关

## 2. 系统设计题：设计一个分布式配置中心

支持：配置实时推送（<1s延迟）、灰度发布（10%→50%→100%）、版本回滚、审计日志。1000+服务实例。

**评判维度（45min）**：
- 数据模型（配置/版本/灰度规则）✓
- 推送机制（长轮询 vs WebSocket vs gRPC stream）✓
- 一致性保证（配置变更的原子性）✓
- 高可用（配置中心自身挂了怎么办）✓
- 灰度策略的设计细节 ✓
- 好的答案应覆盖：Etcd/Consul方案对比、客户端缓存兜底、推送vs拉取取舍

## 3. 代码审查题：找问题

```python
def process_orders(orders):
    results = []
    for order in orders:
        user = db.query(f"SELECT * FROM users WHERE id={order.user_id}")
        if user['vip']:
            discount = 0.8
        total = order['amount'] * discount if 'discount' in dir() else order['amount']  # line 7
        db.execute(f"UPDATE orders SET total={total} WHERE id={order['id']}")
        results.append(total)
    return results
```

**期望发现的7个问题**：
1. SQL注入（f-string拼接）
2. N+1查询（循环内DB查询）
3. `dir()`检查变量存在——hack，应用`discount = 1.0`初始化
4. 无事务保护（批量UPDATE中途失败）
5. 无输入验证（orders是否为None？user为None？）
6. 缺少`vip`字段检查（可能KeyError）
7. 性能：应批量UPDATE而非逐条

## 4. 行为面试题

| 题目 | 评估维度 |
|------|---------|
| "描述一次你推动的技术决策遭遇团队强烈反对的经历，你怎么处理的？" | 沟通能力、技术说服力、妥协智慧 |
| "你发现生产环境一个bug，修复只需5分钟但需要绕过正常发布流程，你怎么做？" | 流程意识、风险判断、伦理决策 |
| "你如何保持技术不落伍？举一个最近学的新技术并应用到工作中的例子。" | 学习能力、主动性、知识内化 |

## 自评

- ✅ 算法题中等难度+明确评分rubric
- ✅ 系统设计题评判维度合理
- ✅ 代码审查题7个问题点设计合理
- ✅ 行为题评估维度明确

**完成** | 修复轮次: 0
