# DESIGN-005: 设计通知系统 — 答案（Group A 基线）

## 架构

```
事件源(下单/评论/@mention) → 通知服务 → [路由引擎]
                                          ├─ 优先级判断
                                          ├─ 聚合窗口
                                          ├─ 用户偏好过滤
                                          └─ 渠道适配器
                                              ├─ 站内通知(WS推送)
                                              ├─ 邮件(SendGrid/SES)
                                              ├─ 短信(Twilio)
                                              ├─ Webhook
                                              └─ 钉钉/飞书/Slack
```

## 数据库模型

```sql
CREATE TABLE notifications (
    id UUID PK, user_id UUID NOT NULL, 
    type VARCHAR(50),  -- comment/mention/assign/status_change
    title TEXT, body TEXT,
    resource_type VARCHAR(50), resource_id UUID,  -- 关联的实体
    is_read BOOLEAN DEFAULT false,
    channel VARCHAR(20),  -- in_app/email/sms/webhook
    status VARCHAR(20) DEFAULT 'pending',  -- pending/sent/failed
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_notif_user_unread ON notifications(user_id, created_at DESC) WHERE NOT is_read;

CREATE TABLE user_preferences (
    user_id UUID, event_type VARCHAR(50), channel VARCHAR(20),
    enabled BOOLEAN DEFAULT true,
    PRIMARY KEY(user_id, event_type, channel)
);

CREATE TABLE notification_templates (
    id UUID PK, event_type VARCHAR(50), channel VARCHAR(20),
    locale VARCHAR(10) DEFAULT 'zh-CN',
    subject_template TEXT, body_template TEXT  -- 使用{{variable}}占位
);
```

## 通知生命周期

```
事件触发 → 查询用户偏好 → 模板渲染 → 聚合判断 → 渠道投递 → 状态更新
                                            ↓
                                urgency=high? → 即时发送
                                urgency=normal → 进入聚合窗口(5min)
                                urgency=low → 进入每日摘要
```

## 重试和削峰

**重试策略**（邮件/SMS/Webhook失败时）：
- 第1次重试: 1min后
- 第2次: 5min后
- 第3次: 15min后
- 第4次: 1h后 → 标记failed，通知管理员

**削峰**：
- Redis队列缓冲突发通知（如促销活动触发百万通知）
- 按优先级分队列：high(即时)/normal(按速率)/low(夜间批处理)
- 渠道限流：邮件<100/s, SMS<10/s

## 聚合逻辑

```python
def should_aggregate(notification, pending_window):
    # 同一用户+同一事件类型+5分钟内 → 聚合
    recent = pending_window.find(
        user_id=notification.user_id,
        event_type=notification.event_type,
        created_after=now() - timedelta(minutes=5)
    )
    if recent and len(recent) >= 3:
        return AggregateStrategy.DIGEST  # 合并为摘要
    if recent:
        return AggregateStrategy.APPEND  # 追加到待发送通知
    return AggregateStrategy.SEND_NOW
```

摘要模板："你在过去5分钟内收到了5条新评论"（而非5条单独通知）

## 自评

- ✅ 多渠道抽象设计合理（5渠道+适配器模式）
- ✅ 用户偏好模型灵活（事件×渠道矩阵）
- ✅ 聚合策略可行（5分钟窗口+计数阈值+摘要模板）
- ✅ 可靠性保证（4级重试+死信+削峰）
- ✅ 模板管理系统设计合理（多语言支持）

**完成** | 修复轮次: 0
