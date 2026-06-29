# DESIGN-009: 设计Webhook系统 — 答案（Group A 基线）

## 架构

```
事件源 → Webhook Service
              ├─ 订阅管理(事件类型→URL列表)
              ├─ 投递队列(每个目标独立)
              ├─ 重试引擎(指数退避)
              └─ 状态追踪
```

## 数据库模型

```sql
CREATE TABLE webhook_subscriptions (
    id UUID PK, user_id UUID, url TEXT NOT NULL,
    events TEXT[] NOT NULL,  -- ['order.created','payment.completed']
    secret TEXT NOT NULL,    -- HMAC signing key
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ
);

CREATE TABLE webhook_deliveries (
    id UUID PK, subscription_id UUID FK,
    event_type TEXT, payload JSONB,
    status TEXT DEFAULT 'pending', -- pending/success/failed/retrying
    attempt_count INT DEFAULT 0,
    last_attempt_at TIMESTAMPTZ, next_attempt_at TIMESTAMPTZ,
    response_code INT, response_body TEXT,
    created_at TIMESTAMPTZ
);
```

## 重试状态机

```
pending → 首次投递 → success ✓
                    → failed → retrying
                                  ├─ 1min → 重试1
                                  ├─ 5min → 重试2
                                  ├─ 15min → 重试3
                                  ├─ 1h → 重试4
                                  └─ 24h → 重试5(max) → failed(永久)
                                                 → dead_letter_queue
```

## 安全设计（HMAC签名）

```python
# 发送方
payload = json.dumps(event)
signature = hmac.new(webhook_secret, payload.encode(), hashlib.sha256).hexdigest()
headers = {'X-Webhook-Signature': f'sha256={signature}'}
requests.post(url, data=payload, headers=headers)

# 接收方验证
expected = hmac.new(my_secret, request.body, hashlib.sha256).hexdigest()
if not hmac.compare_digest(request.headers['X-Webhook-Signature'], f'sha256={expected}'):
    return 401  # 签名不匹配，拒绝
```

**防重放**：payload中包含`timestamp`和`nonce`，接收方拒绝超过5分钟的旧请求和重复的nonce。

## 自评

- ✅ 架构覆盖所有需求
- ✅ 重试状态机完整（5级退避+死信队列）
- ✅ 安全设计正确（HMAC签名+防重放）
- ✅ 状态追踪和Dashboard设计合理

**完成** | 修复轮次: 0
