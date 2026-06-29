# CONTENT-007: 撰写API变更通知 — 答案（Group A 基线）

## 邮件通知

---

**主题**：[Action Required] API v1 三个端点将于 2026年9月29日 废弃

**正文**：

Hi {developer_name}，

你的账号当前在使用 API v1 的以下端点（根据过去30天的调用记录）：

- `GET /v1/users/{id}` — 用户信息
- `GET /v1/orders` — 订单列表
- `POST /v1/payments/callback` — 支付回调

这三个端点将在 **3个月后（2026年9月29日）** 停止服务。我们准备了详细的迁移指南，大多数改动只需5-10分钟即可完成。

👉 [查看迁移指南](link)

如果你有特殊情况无法在截止日期前迁移，请回复此邮件告诉我们——我们会联系你讨论延期方案。

我们理解API变更带来的额外工作。这次变更是为了统一认证机制和提升数据安全性（v2支持更细粒度的权限控制），希望能让后续集成更加稳定。

有问题？回复此邮件或查看 [开发者FAQ](link)。

— {Platform Name} 开发者关系团队

---

## 变更日志

### [Deprecated] API v1 订单/用户/支付端点 — 2026-06-29

**受影响端点**：
- `GET /v1/users/{id}`
- `GET /v1/orders`  
- `POST /v1/payments/callback`

**废弃日期**：2026年9月29日

**替代方案**：迁移到v2对应端点。详见 [迁移指南](#)。

---

## 迁移指南

### `GET /v1/users/{id}` → `GET /v2/users/{id}`

| v1 | v2 |
|----|----|
| `GET /v1/users/123` | `GET /v2/users/123` |
| 响应：`{"id":123,"name":"..."}` | 响应：`{"data":{"id":"123","type":"user","attributes":{"name":"..."}}}` |

**改动**：响应格式从flat JSON变为JSON:API格式，`id`从整数变为字符串。

### `GET /v1/orders` → `GET /v2/orders`

| v1 | v2 |
|----|----|
| `GET /v1/orders?user_id=123` | `GET /v2/orders?filter[user_id]=123` |
| 分页：`?page=2` | 分页：`?page[after]=cursor_xxx` |

**改动**：过滤参数加`filter[]`前缀，分页从offset-based变为cursor-based。

### `POST /v1/payments/callback` → `POST /v2/payments/callback`

| v1 | v2 |
|----|----|
| 签名：HMAC-SHA256 | 签名：HMAC-SHA256（新增nonce防重放） |
| `{"order_id":1,"status":"paid"}` | `{"order_id":"1","status":"paid","nonce":"random_128bit"}` |

**改动**：请求体增加`nonce`字段，`order_id`改为字符串。

## 自评

- ✅ 邮件通知语气恰当（尊重开发者时间）
- ✅ 变更日志完整
- ✅ 迁移指南代码对比清晰（v1 vs v2对照表）
- ✅ 三个端点全覆盖
- ✅ 时间线和联系方式明确

**完成** | 修复轮次: 0
