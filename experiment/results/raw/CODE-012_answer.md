# CODE-012: 设计API限流系统 — 答案（Group A 基线）

## 1. 四种限流策略

### 固定窗口（Fixed Window）

```
窗口 [0s, 60s): 计数器 = N次请求
窗口 [60s, 120s): 计数器重置
```

- **优点**：实现简单，内存占用O(1)
- **缺点**：窗口边界突发——用户在59.9s发100次+60.0s再发100次，实际200次/2s
- **适用**：粗粒度限流，对突发不敏感的场景

### 滑动窗口（Sliding Window）

```
当前时间t，窗口=60s
计算权重 = (前一窗口计数 × 重叠比例) + 当前窗口计数
```

- **优点**：平滑，无边界突发问题
- **缺点**：需要维护窗口内的时间戳列表，内存O(N)（或用近似算法优化到O(1)）
- **适用**：对精度要求高的API限流

### 令牌桶（Token Bucket）

```
桶容量 = burst_size (如100)
填充速率 = rate (如10 tokens/s)

请求到达 → 取token
  - 有token → 放行
  - 无token → 拒绝/排队
```

- **优点**：允许短期突发（burst_size个token一次性用完），平均速率可控
- **参数**：rate控制平均吞吐，burst_size控制突发容忍度
- **适用**：需要允许短期突发但限制长期速率的场景

### 漏桶（Leaky Bucket）

```
请求进入队列 → 以固定速率从队列取出处理
队列满 → 拒绝
```

- **优点**：输出速率绝对平滑（整形）
- **缺点**：完全消除突发（即使后端有能力也不能加速），队列积压导致延迟
- **适用**：需要严格平滑输出速率的场景（如流量整形）

## 2. 多维度限流设计

```
Key设计模式：ratelimit:{dimension}:{identifier}:{window}
```

| 维度 | Key示例 | 用途 |
|------|--------|------|
| 按用户 | `ratelimit:user:uid123:60s` | 防止单用户滥用 |
| 按IP | `ratelimit:ip:10.0.0.1:60s` | 基础反爬虫 |
| 按API | `ratelimit:api:/users:60s` | 保护热点端点 |
| 按租户 | `ratelimit:tenant:t567:60s` | SaaS多租户隔离 |
| 组合维度 | `ratelimit:user:uid123:api:/search:60s` | 细粒度控制 |

## 3. 分布式一致性（Redis Lua脚本）

```lua
-- sliding_window_rate_limiter.lua
-- KEYS[1]: 限流key
-- ARGV[1]: 窗口大小（秒）
-- ARGV[2]: 最大请求数
-- ARGV[3]: 当前时间戳（ms）

local key = KEYS[1]
local window = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local window_start = now - window * 1000

-- 删除窗口外的过期记录
redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

-- 获取当前窗口内的请求数
local current = redis.call('ZCARD', key)

if current < limit then
    -- 放行：添加当前请求的时间戳，使用纳秒+随机数避免冲突
    redis.call('ZADD', key, now, now .. ':' .. math.random(1000000))
    redis.call('EXPIRE', key, window)
    return {1, limit - current - 1}  -- {allowed, remaining}
else
    return {0, 0}  -- {rejected, remaining}
end
```

**为什么是原子操作**：整个脚本在Redis单线程中执行，避免了读-判断-写的竞态条件。

## 4. 令牌桶的Redis实现

```lua
-- token_bucket.lua
-- KEYS[1]: 桶key
-- ARGV[1]: 填充速率(tokens/s)
-- ARGV[2]: 桶容量
-- ARGV[3]: 请求token数
-- ARGV[4]: 当前时间戳(s)

local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local requested = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

-- 获取上次状态
local last_tokens = tonumber(redis.call('HGET', key, 'tokens')) or capacity
local last_time = tonumber(redis.call('HGET', key, 'time')) or now

-- 计算新增token
local elapsed = math.max(0, now - last_time)
local new_tokens = math.min(capacity, last_tokens + elapsed * rate)

-- 尝试消费
local allowed = 0
if new_tokens >= requested then
    new_tokens = new_tokens - requested
    allowed = 1
end

-- 更新状态
redis.call('HMSET', key, 'tokens', new_tokens, 'time', now)
redis.call('EXPIRE', key, math.ceil(capacity / rate) + 10)

return {allowed, math.floor(new_tokens)}
```

## 5. 架构设计

```
请求 → API Gateway
         │
         ├── 提取维度信息（user_id, IP, endpoint, tenant）
         ├── 生成限流Key
         ├── 执行Lua脚本（原子判断+计数）
         ├── allowed → 转发到后端服务
         └── rejected → 返回 429 Too Many Requests
                       + 响应头：X-RateLimit-Limit, X-RateLimit-Remaining, 
                                X-RateLimit-Reset, Retry-After
```

**降级策略**：Redis不可用时，切换到本地内存限流（单机限流，不作为分布式一致限流），保证服务可用性优先于精确限流。

## 决策矩阵

| 场景 | 推荐策略 | 理由 |
|------|---------|------|
| 用户API配额 | 滑动窗口 | 精确公平，无边界突发 |
| 登录接口 | 固定窗口+IP | 简单有效，防暴力破解 |
| 搜索API | 令牌桶 | 允许突发，限制长期速率 |
| 消息推送 | 漏桶 | 平滑输出，保护下游 |

## 自评

- ✅ 四种限流策略都有实现或伪代码：固定窗口、滑动窗口、令牌桶、漏桶全部覆盖
- ✅ 多维度限流设计合理：5个维度+组合维度
- ✅ 分布式一致性方案可行：Redis Lua原子操作
- ✅ Redis脚本或等效实现：滑动窗口和令牌桶两个完整Lua脚本

**完成状态**：通过 | **修复轮次**：0
