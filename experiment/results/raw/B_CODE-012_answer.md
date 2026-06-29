# Group B: CODE-012 — 设计API限流系统

## 审查记录

**审查轮次**：1
**发现的问题**：
1. 令牌桶Lua脚本中`elapsed * rate`可能产生浮点数——Redis Lua中整数运算更安全
2. 缺少"限流降级时如何通知客户端"的设计——429响应之外，用户需要知道当前limit和reset时间

**修改**：
- 令牌桶计算改用整数：`new_tokens = math.min(capacity, last_tokens + math.floor(elapsed * rate))`
- 补充限流响应标准：`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After`
- 降级策略新增：Redis不可用时本地令牌桶兜底

**审查结论**：通过（1轮修改）。修改后的方案更加健壮。
