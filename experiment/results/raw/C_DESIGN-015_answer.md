# 电商系统事件驱动微服务架构设计

> **文档版本**: v1.0 | **日期**: 2026-06-29 | **任务编号**: DESIGN-015
>
> **设计目标**: 将一个电子商务系统从单体架构迁移到事件驱动微服务架构，覆盖完整订单生命周期事件流：用户下单 -> 扣库存 -> 创建支付 -> 支付确认 -> 通知仓库 -> 发送物流 -> 完成。

---

## 目录

1. [架构概览与核心原则](#1-架构概览与核心原则)
2. [事件定义体系](#2-事件定义体系)
3. [消息队列/事件总线选型](#3-消息队列事件总线选型)
4. [事件溯源 vs 事件通知的选择](#4-事件溯源-vs-事件通知的选择)
5. [乱序/重复/丢失事件处理策略](#5-乱序重复丢失事件处理策略)
6. [监控与死信队列设计](#6-监控与死信队列设计)
7. [部署拓扑与扩展性](#7-部署拓扑与扩展性)

---

## 1. 架构概览与核心原则

### 1.1 整体架构图

```
                              ┌──────────────────────────────────────────┐
                              │            API Gateway (Kong)             │
                              │        Rate Limit / Auth / Route          │
                              └──────┬──────┬──────┬──────┬──────────────┘
                                     │      │      │      │
                              ┌──────▼──────▼──────▼──────▼──────────────┐
                              │          Backend-for-Frontend             │
                              │     (GraphQL Federation / REST Agg.)      │
                              └──────────────────┬───────────────────────┘
                                                 │
         ┌───────────────────────────────────────┼───────────────────────────────────────┐
         │                          EVENT BUS (Kafka Cluster)                            │
         │                                                                               │
         │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
         │  │  orders  │  │inventory │  │ payments │  │notifica- │  │logistics │       │
         │  │   .evt   │  │  .evt    │  │  .evt    │  │tions.evt │  │  .evt    │       │
         │  │ (p=12,r=3)│  │(p=6,r=3) │  │(p=6,r=3) │  │(p=3,r=2) │  │(p=6,r=3) │       │
         │  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘       │
         └────────┼─────────────┼─────────────┼─────────────┼─────────────┼─────────────┘
                  │             │             │             │             │
    ┌─────────────▼──┐ ┌────────▼───┐ ┌────────▼───┐ ┌────────▼───┐ ┌────────▼───┐
    │  Order Service │ │ Inventory  │ │  Payment   │ │Notification│ │ Logistics  │
    │                │ │  Service   │ │  Service   │ │  Service   │ │  Service   │
    │ - Create Order │ │- Reserve   │ │- Create Pay│ │- Email     │ │- Allocate  │
    │ - Order State  │ │- Deduct    │ │- Confirm   │ │- SMS       │ │- Track     │
    │ - Order Query  │ │- Restore   │ │- Refund    │ │- Push      │ │- Deliver   │
    └───────┬────────┘ └─────┬──────┘ └─────┬──────┘ └─────┬──────┘ └─────┬──────┘
            │                │              │              │              │
            ▼                ▼              ▼              ▼              ▼
    ┌───────────┐   ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐
    │ PostgreSQL│   │ PostgreSQL│  │ PostgreSQL│  │  MongoDB  │  │ PostgreSQL│
    │ (Orders)  │   │(Inventory)│  │(Payments) │  │(Templates)│  │(Logistics)│
    └───────────┘   └───────────┘  └───────────┘  └───────────┘  └───────────┘
            │                │              │              │              │
            └────────────────┼──────────────┼──────────────┼──────────────┘
                             │              │              │
              ┌──────────────▼──────────────▼──────────────▼──────────────┐
              │              CDC Connector (Debezium)                     │
              │        Outbox Table -> Kafka (exactly-once)               │
              └──────────────────────────────────────────────────────────┘
                                         │
                          ┌──────────────▼───────────────────────────────┐
                          │              DEAD LETTER QUEUE                │
                          │  ┌──────────────────────────────────────┐    │
                          │  │ dlq.orders  dlq.inventory dlq.pay... │    │
                          │  │ (retention=30d, p=1, r=1)            │    │
                          │  └──────────────────────────────────────┘    │
                          │              │                                │
                          │  ┌───────────▼───────────────────────┐       │
                          │  │      DLQ Consumer / Dashboard      │       │
                          │  │   - Manual replay                  │       │
                          │  │   - Auto-retry with backoff        │       │
                          │  │   - Alert -> PagerDuty             │       │
                          │  └───────────────────────────────────┘       │
                          └──────────────────────────────────────────────┘
                                         │
                          ┌──────────────▼───────────────────────────────┐
                          │          OBSERVABILITY STACK                  │
                          │  ┌─────────┐ ┌──────────┐ ┌──────────────┐  │
                          │  │Prometheus│ │Grafana   │ │Elasticsearch │  │
                          │  │+ AlertM. │ │Dashboards│ │+ Kibana      │  │
                          │  └─────────┘ └──────────┘ └──────────────┘  │
                          │  ┌──────────┐ ┌───────────────────────────┐  │
                          │  │ Jaeger   │ │  Schema Registry (Apicurio)│  │
                          │  │ (Tracing)│ │  Event Version Management  │  │
                          │  └──────────┘ └───────────────────────────┘  │
                          └──────────────────────────────────────────────┘
```

### 1.2 核心设计原则

| 原则 | 说明 | 实现手段 |
|------|------|----------|
| **异步优先** | 服务间通过事件通信，不进行同步RPC调用 | Kafka作为事件总线，每个服务独立消费 |
| **最终一致性** | 跨服务数据对齐不要求实时强一致 | Outbox模式 + 幂等消费 + 补偿事务 |
| **事件不可变** | 事件一旦发布，不可修改 | Append-only topic + Schema Registry版本管理 |
| **独立部署** | 每个服务独立代码库、独立数据库、独立扩缩容 | Kubernetes Deployment per service |
| **故障隔离** | 一个服务故障不影响其他服务继续运行 | 断路器 + 死信队列 + 缓冲区 |

---

## 2. 事件定义体系

### 2.1 事件命名规范

采用 `{domain}.{entity}.{action}[.v{version}]` 的命名结构，确保事件名称自描述、可路由：

```
{domain}.{entity}.{action}
```

- **domain**: 业务域 (order, payment, inventory, logistics, notification)
- **entity**: 实体名词 (order, payment, stock, shipment)
- **action**: 过去式动词 (created, confirmed, deducted, delivered)

### 2.2 完整事件目录

| # | 事件类型 | 生产者 | 消费者 | 关键度 | 场景 |
|---|----------|--------|--------|--------|------|
| 1 | `order.order.created.v1` | Order Service | Inventory, Payment, Notification | P0 | 用户下单成功 |
| 2 | `inventory.stock.reserved.v1` | Inventory Service | Order, Notification | P1 | 库存预留成功 |
| 3 | `inventory.stock.deducted.v1` | Inventory Service | Order, Payment | P0 | 库存已扣减 |
| 4 | `inventory.stock.reservation_failed.v1` | Inventory Service | Order | P0 | 库存不足，订单需取消 |
| 5 | `payment.payment.created.v1` | Payment Service | Order, Notification | P0 | 支付单已创建 |
| 6 | `payment.payment.confirmed.v1` | Payment Service | Order, Inventory, Logistics, Notification | P0 | 支付已到账 |
| 7 | `payment.payment.failed.v1` | Payment Service | Order, Inventory | P0 | 支付失败，需恢复库存 |
| 8 | `payment.payment.refunded.v1` | Payment Service | Order, Inventory | P1 | 退款完成 |
| 9 | `logistics.shipment.created.v1` | Logistics Service | Notification, Order | P1 | 物流单已创建 |
| 10 | `logistics.shipment.dispatched.v1` | Logistics Service | Notification, Order | P1 | 已发货 |
| 11 | `logistics.shipment.delivered.v1` | Logistics Service | Order | P1 | 已签收 |
| 12 | `order.order.completed.v1` | Order Service | Notification | P0 | 订单生命周期结束 |
| 13 | `order.order.cancelled.v1` | Order Service | Inventory, Payment | P0 | 订单取消（含补偿） |
| 14 | `notification.notification.sent.v1` | Notification Service | (审计消费者) | P2 | 通知已发送 |

### 2.3 事件信封(Envelope)设计

所有事件遵循统一的外层信封结构，携带元数据便于路由、追踪和版本管理：

```json
{
  "$schema": "https://api.example.com/schemas/event-envelope.v1.json",
  "title": "Event Envelope v1",
  "type": "object",
  "required": ["event_id", "event_type", "timestamp", "payload", "metadata"],
  "properties": {
    "event_id": {
      "type": "string",
      "format": "uuid",
      "description": "全局唯一事件ID，用于去重和幂等"
    },
    "correlation_id": {
      "type": "string",
      "format": "uuid",
      "description": "关联ID，串联同一订单生命周期内所有事件"
    },
    "event_type": {
      "type": "string",
      "pattern": "^[a-z]+\\.[a-z]+\\.[a-z_]+\\.[a-z]+\\.[a-z]+\\.[a-z_]+\\.[a-z]+\\.[a-z]+\\.v[0-9]+$",
      "description": "事件类型全限定名，如 order.order.created.v1"
    },
    "event_version": {
      "type": "string",
      "pattern": "^v[0-9]+$",
      "description": "事件schema版本号"
    },
    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "事件产生时间 (RFC3339, UTC)"
    },
    "source": {
      "type": "object",
      "properties": {
        "service_name": { "type": "string" },
        "service_version": { "type": "string" },
        "instance_id": { "type": "string" }
      },
      "description": "事件来源服务标识"
    },
    "payload": {
      "type": "object",
      "description": "业务载荷，schema由event_type决定"
    },
    "metadata": {
      "type": "object",
      "properties": {
        "trace_id": {
          "type": "string",
          "description": "分布式追踪ID (W3C Trace Context)"
        },
        "span_id": {
          "type": "string",
          "description": "当前Span ID"
        },
        "idempotency_key": {
          "type": "string",
          "description": "幂等键，消费端用它去重"
        },
        "sequence_number": {
          "type": "integer",
          "description": "同一实体的事件序列号，用于检测乱序"
        },
        "tenant_id": {
          "type": "string",
          "description": "多租户标识"
        }
      }
    }
  }
}
```

### 2.4 核心事件Payload Schema

#### 2.4.1 order.order.created.v1

```json
{
  "$schema": "https://api.example.com/schemas/order.created.v1.json",
  "title": "Order Created Event",
  "type": "object",
  "required": ["order_id", "user_id", "items", "total_amount", "currency"],
  "properties": {
    "order_id": {
      "type": "string",
      "format": "uuid"
    },
    "user_id": {
      "type": "string",
      "format": "uuid"
    },
    "items": {
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "object",
        "required": ["sku", "quantity", "unit_price"],
        "properties": {
          "sku": { "type": "string" },
          "product_name": { "type": "string" },
          "quantity": { "type": "integer", "minimum": 1 },
          "unit_price": {
            "type": "object",
            "properties": {
              "amount": { "type": "string", "pattern": "^[0-9]+\\.[0-9]{2}$" },
              "currency": { "type": "string", "enum": ["CNY", "USD", "EUR"] }
            }
          }
        }
      }
    },
    "total_amount": {
      "type": "object",
      "properties": {
        "amount": { "type": "string", "pattern": "^[0-9]+\\.[0-9]{2}$" },
        "currency": { "type": "string", "enum": ["CNY", "USD", "EUR"] }
      }
    },
    "shipping_address": {
      "type": "object",
      "required": ["country", "city", "line1", "postal_code"],
      "properties": {
        "country": { "type": "string", "minLength": 2, "maxLength": 2 },
        "province": { "type": "string" },
        "city": { "type": "string" },
        "district": { "type": "string" },
        "line1": { "type": "string" },
        "line2": { "type": "string" },
        "postal_code": { "type": "string" },
        "recipient_name": { "type": "string" },
        "phone": { "type": "string" }
      }
    },
    "coupon_code": { "type": "string" },
    "discount_amount": {
      "type": "object",
      "properties": {
        "amount": { "type": "string" },
        "currency": { "type": "string" }
      }
    },
    "channel": {
      "type": "string",
      "enum": ["app", "web", "miniprogram", "h5"]
    },
    "created_at": { "type": "string", "format": "date-time" }
  }
}
```

#### 2.4.2 inventory.stock.reserved.v1

```json
{
  "$schema": "https://api.example.com/schemas/stock.reserved.v1.json",
  "title": "Stock Reserved Event",
  "type": "object",
  "required": ["order_id", "reservation_id", "reserved_items", "reserved_at"],
  "properties": {
    "order_id": { "type": "string", "format": "uuid" },
    "reservation_id": { "type": "string", "format": "uuid" },
    "reserved_items": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["sku", "quantity_reserved", "warehouse_id"],
        "properties": {
          "sku": { "type": "string" },
          "quantity_reserved": { "type": "integer", "minimum": 1 },
          "quantity_remaining": { "type": "integer", "minimum": 0 },
          "warehouse_id": {
            "type": "string",
            "description": "分配的仓库ID"
          },
          "reservation_expires_at": {
            "type": "string",
            "format": "date-time",
            "description": "预留过期时间，超时自动释放"
          }
        }
      }
    },
    "reserved_at": { "type": "string", "format": "date-time" }
  }
}
```

#### 2.4.3 payment.payment.confirmed.v1

```json
{
  "$schema": "https://api.example.com/schemas/payment.confirmed.v1.json",
  "title": "Payment Confirmed Event",
  "type": "object",
  "required": ["payment_id", "order_id", "amount_paid", "payment_method", "confirmed_at"],
  "properties": {
    "payment_id": { "type": "string", "format": "uuid" },
    "order_id": { "type": "string", "format": "uuid" },
    "amount_paid": {
      "type": "object",
      "properties": {
        "amount": { "type": "string", "pattern": "^[0-9]+\\.[0-9]{2}$" },
        "currency": { "type": "string" }
      }
    },
    "payment_method": {
      "type": "string",
      "enum": ["alipay", "wechat_pay", "union_pay", "credit_card", "debit_card"]
    },
    "transaction_id": {
      "type": "string",
      "description": "第三方支付网关交易流水号"
    },
    "gateway_response": {
      "type": "object",
      "description": "支付网关原始响应（脱敏后）",
      "properties": {
        "gateway": { "type": "string" },
        "status_code": { "type": "string" },
        "message": { "type": "string" }
      }
    },
    "confirmed_at": { "type": "string", "format": "date-time" }
  }
}
```

#### 2.4.4 logistics.shipment.dispatched.v1

```json
{
  "$schema": "https://api.example.com/schemas/shipment.dispatched.v1.json",
  "title": "Shipment Dispatched Event",
  "type": "object",
  "required": ["shipment_id", "order_id", "carrier", "tracking_number", "dispatched_at"],
  "properties": {
    "shipment_id": { "type": "string", "format": "uuid" },
    "order_id": { "type": "string", "format": "uuid" },
    "warehouse_id": { "type": "string" },
    "carrier": {
      "type": "object",
      "properties": {
        "name": { "type": "string", "enum": ["SF", "YTO", "ZTO", "STO", "EMS", "DHL", "FedEx"] },
        "carrier_code": { "type": "string" }
      }
    },
    "tracking_number": { "type": "string" },
    "packages": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "package_id": { "type": "string" },
          "weight_kg": { "type": "number" },
          "dimensions_cm": {
            "type": "object",
            "properties": {
              "length": { "type": "number" },
              "width": { "type": "number" },
              "height": { "type": "number" }
            }
          }
        }
      }
    },
    "estimated_delivery_at": { "type": "string", "format": "date-time" },
    "dispatched_at": { "type": "string", "format": "date-time" }
  }
}
```

#### 2.4.5 order.order.cancelled.v1

```json
{
  "$schema": "https://api.example.com/schemas/order.cancelled.v1.json",
  "title": "Order Cancelled Event",
  "type": "object",
  "required": ["order_id", "cancel_reason", "cancelled_at"],
  "properties": {
    "order_id": { "type": "string", "format": "uuid" },
    "cancel_reason": {
      "type": "string",
      "enum": [
        "user_requested",
        "payment_timeout",
        "stock_insufficient",
        "fraud_detected",
        "system_error"
      ]
    },
    "cancel_initiator": {
      "type": "string",
      "enum": ["user", "system", "admin"]
    },
    "compensation_required": {
      "type": "boolean",
      "description": "是否需要触发补偿事务（释放库存、退款等）"
    },
    "compensation_status": {
      "type": "string",
      "enum": ["pending", "in_progress", "completed", "failed"]
    },
    "cancelled_at": { "type": "string", "format": "date-time" },
    "remark": { "type": "string" }
  }
}
```

### 2.5 事件版本管理策略

采用 **Schema Registry (Apicurio Registry)** 集中管理所有事件的 Avro/JSON Schema，版本策略如下：

| 版本变更类型 | 兼容性级别 | 示例 | 处理规则 |
|-------------|-----------|------|----------|
| 新增可选字段 | FORWARD | payload新增 `gift_wrap: boolean?` | 旧消费者忽略新字段，无影响 |
| 删除可选字段 | BACKWARD | 移除某个不再使用的可选字段 | 旧生产者仍可发送，新消费者忽略 |
| 新增必填字段 | FULL | 新增 `tax_id`(required) | 需要新topic(v2)或协调升级所有消费者 |
| 修改字段类型 | NONE(不兼容) | `quantity`: int->string | 必须创建新版本topic，双写过渡 |

**版本升级流程:**
1. 在Schema Registry注册新版本schema
2. 生产者升级，双写到新旧topic（过渡期）
3. 消费者逐批升级，切换到新topic
4. 确认所有消费者升级完成后，下线旧topic

---

## 3. 消息队列/事件总线选型

### 3.1 候选方案对比

| 维度 | Apache Kafka | RabbitMQ | AWS SQS + SNS | Redis Streams |
|------|-------------|----------|---------------|---------------|
| **消息模型** | 分布式日志 (partitioned log) | 智能代理 (exchange/queue) | 托管队列 + 发布订阅 | 追加日志流 |
| **吞吐量** | 百万条/秒 (单broker 100k+) | 万条/秒 (单节点 ~20k) | 弹性 (受限于AWS配额) | 十万条/秒 (内存限制) |
| **延迟** | 毫秒级 (端到端 ~5-15ms) | 微秒级 (~1ms) | 毫秒-秒级 (~10-100ms) | 亚毫秒级 |
| **消息持久化** | 磁盘持久化, 可配置保留期 (默认7d) | 内存+磁盘, 消费后删除 | 自动持久化, max 14天保留 | 内存+定期RDB/AOF, 消费后删除 |
| **消息回溯** | **支持** (基于offset, 可任意回退重放) | **不支持** (消费即删除) | **不支持** | **支持** (基于ID范围) |
| **消费模式** | 拉模式 (pull, 消费者控制速率) | 推模式为主 (push) + 拉 | 拉模式 (long polling) | 拉模式为主 |
| **消息顺序** | 分区内有序 | 单队列内有序 | FIFO队列有序 (限300 TPS) | 流内有序 |
| **投递语义** | at-least-once / exactly-once (事务) | at-least-once (ack) | at-least-once (FIFO支持exactly-once) | at-least-once |
| **运维复杂度** | 高 (ZooKeeper/KRaft, partition管理) | 中 | 低 (全托管) | 低 (Redis运维) |
| **生态系统** | Kafka Connect, KSQL, Schema Registry, 大量Connector | 插件丰富, 管理UI内置 | SQS + Lambda天然集成 | Redis生态, 需额外组件 |
| **事件溯源支持** | **原生支持** (不可变日志, CQRS天然适配) | 不适合 (消息短暂) | 不适合 (超时删除) | 部分适合 (AOF持久化) |
| **死信队列** | 需自行实现 (独立topic + consumer) | **内置DLX** (死信交换机) | **内置DLQ** (redrive policy) | 需自行实现 |
| **多消费者组** | **原生支持** (不同group独立消费) | 通过fanout exchange | SNS fanout -> multiple SQS | 通过consumer group (Redis 5.0+) |
| **部署方式** | 自托管/K8s/Confluent Cloud | 自托管/K8s/CloudAMQP | AWS原生托管 | 自托管/K8s/Redis Cloud |
| **成本** | 中等 (自托管需3+ broker) | 低 (资源消耗小) | 按量付费 (大规模成本高) | 低 (内存成本) |

### 3.2 选型结论：Apache Kafka

**选择 Kafka**，理由如下：

#### 决定性因素

1. **事件回溯能力** (权重最高)。电商场景中，支付回调丢失、库存对账异常时，需要回溯历史事件重放修复数据。RabbitMQ和SQS消息消费后即删除，无法回溯。Kafka的不可变持久化日志天然支持任意时间点的事件重放。

2. **事件溯源架构兼容性**。本设计采用了事件溯源+事件通知的混合模式。Kafka的partitioned log是实现事件溯源的基础设施——每个聚合根的事件流天然对应一个partition，保证顺序和完整性。

3. **多消费者组独立消费**。同一个"订单创建"事件，Inventory Service、Payment Service、Notification Service需要独立消费、独立管理offset、独立处理速率。Kafka的consumer group模型完美匹配这一需求，RabbitMQ需要额外配置fanout exchange来模拟。

4. **生态集成**。Kafka Connect提供开箱即用的CDC Connector (Debezium)，直接从数据库Outbox表捕获事件写入Kafka，避免在应用代码中做"写数据库+发消息"的双重操作，是实现Transactional Outbox模式的最优解。

5. **水平扩展性**。订单量增长时，按order_id hash分区即可线性扩展。每个partition独立处理，增加partition = 增加并行度。

#### 不选的方案及排除理由

- **RabbitMQ**：不支持消息回溯，事件无法重放；消息消费后删除导致审计追踪困难。推送模式下消费者过载会丢失消息。
- **AWS SQS+SNS**：FIFO队列限制300 TPS，大促期间无法满足需求；消息最长保留14天，历史事件丢失；云供应商锁定。
- **Redis Streams**：数据持久化依赖内存+AOF，大流量下内存成本高昂；消费者组功能相对不成熟；缺乏Schema Registry等配套工具。

### 3.3 Kafka集群拓扑

```
                     ┌─────────────────────────────────────────┐
                     │         Kafka Cluster (3 Brokers)        │
                     │         replication-factor=3             │
                     │         min.insync.replicas=2            │
                     │                                          │
                     │  Broker-1           Broker-2    Broker-3 │
                     │  ┌─────────┐  ┌─────────┐  ┌─────────┐  │
                     │  │orders-0 │  │orders-1 │  │orders-2 │  │
                     │  │paymnt-0 │  │paymnt-1 │  │paymnt-2 │  │
                     │  │invtry-0 │  │invtry-1 │  │invtry-2 │  │
                     │  │logis-0  │  │logis-1  │  │logis-2  │  │
                     │  │dlq-0    │  │dlq-1    │  │dlq-2    │  │
                     │  └─────────┘  └─────────┘  └─────────┘  │
                     └─────────────────────────────────────────┘
                                       │
                    ┌──────────────────┼──────────────────┐
                    │                  │                  │
              ┌─────▼─────┐    ┌──────▼──────┐   ┌──────▼──────┐
              │Consumer   │    │Consumer     │   │Consumer     │
              │Group:     │    │Group:       │   │Group:       │
              │inventory  │    │payment      │   │logistics    │
              │(3 members)│    │(2 members)  │   │(2 members)  │
              └───────────┘    └─────────────┘   └─────────────┘
```

**关键配置参数：**

| 参数 | 值 | 理由 |
|------|-----|------|
| `replication.factor` | 3 | 容忍2个broker故障 |
| `min.insync.replicas` | 2 | 至少2个副本确认写入 |
| `acks` | all | 生产者等待所有ISR确认 |
| `enable.idempotence` | true | 生产者幂等，防止重复写入 |
| `retention.ms` | 604800000 (7d) | 事件保留7天，DLQ保留30天 |
| `segment.bytes` | 1073741824 (1GB) | 日志分段大小 |
| `compression.type` | snappy | 平衡压缩比与CPU开销 |

---

## 4. 事件溯源 vs 事件通知的选择

### 4.1 决策：混合模式

本架构采用**事件通知为主，事件溯源为辅**的混合模式，而非纯事件溯源。

#### 选择理由

| 对比维度 | 纯事件溯源 (Event Sourcing) | 事件通知 (Event Notification) | 本架构选择 |
|----------|--------------------------|-------------------------------|-----------|
| **状态重建** | 完全从事件流重建 | 数据库存储当前状态 | **通知**：数据库存储当前状态（查询效率高） |
| **查询复杂度** | 需要CQRS + 读模型投影 | 直接查询数据库，简单高效 | **通知**：业务查询路径短 |
| **开发复杂度** | 高（事件建模、投影、快照） | 中等 | **通知**：团队学习成本可控 |
| **调试难度** | 高（事件链长，重建耗时） | 中等 | **通知**：直接查数据库状态 |
| **审计追踪** | 原生支持 | 需额外事件日志 | **溯源**：关键流程记录完整事件链 |
| **补偿/回滚** | 逆向事件追加 | Saga编排器或手动补偿 | **溯源**：事件链支持精确回滚定位 |

#### 具体应用

- **事件通知 (主力)**：Order Service、Payment Service等核心业务服务通过事件通知其他服务"发生了什么"，各自维护自己的数据副本。查询操作直接走数据库，不做事件重建。
- **事件溯源 (辅助)**：以下场景启用事件溯源记录：
  - 订单状态变更的完整审计日志（order_id粒度的事件流）
  - 支付对账（可回溯的支付事件链）
  - 库存变更历史（用于库存差异排查）
- **Outbox模式**：所有服务使用Transactional Outbox模式保证"业务操作+事件发布"的原子性，避免双写不一致。

### 4.2 Outbox模式实现

```
   ┌─────────────────────────────────┐
   │       Service Transaction        │
   │                                  │
   │  BEGIN;                          │
   │    INSERT INTO orders (...);     │
   │    INSERT INTO outbox (          │
   │      event_id,                   │
   │      event_type,                 │
   │      payload_json,               │
   │      created_at                  │
   │    );                            │
   │  COMMIT;                         │
   │                                  │
   │  ┌───────────────────────────┐   │
   │  │  Debezium CDC Connector   │   │
   │  │  1. 监测outbox表binlog     │   │
   │  │  2. 解析变更记录            │   │
   │  │  3. 写入Kafka topic        │   │
   │  │  4. 标记outbox记录为已发送   │   │
   │  └───────────┬───────────────┘   │
   └──────────────┼───────────────────┘
                  │
          ┌───────▼───────┐
          │  Kafka Topic  │
          └───────────────┘
```

**Outbox表结构** (每个服务一个):

```sql
CREATE TABLE outbox (
    id            BIGSERIAL PRIMARY KEY,
    event_id      UUID NOT NULL UNIQUE,
    event_type    VARCHAR(255) NOT NULL,
    aggregate_id  UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    payload       JSONB NOT NULL,
    metadata      JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at  TIMESTAMPTZ,
    retry_count   INT DEFAULT 0,
    status        VARCHAR(20) DEFAULT 'PENDING'
);

CREATE INDEX idx_outbox_status_created ON outbox(status, created_at);
```

选择Debezium CDC而非应用层轮询outbox表的原因：Debezium基于MySQL binlog / PostgreSQL WAL的实时捕获，延迟低（毫秒级），且不会给数据库增加轮询负载。

---

## 5. 乱序/重复/丢失事件处理策略

### 5.1 三种异常的统一处理框架

```
                              事件进入消费者
                                    │
                          ┌─────────▼─────────┐
                          │  1. 重复检测        │
                          │  idempotency_key   │
                          │  查Redis去重表      │
                          └──────┬────────────┘
                                 │
                    ┌────────────┼────────────┐
                    │ 已处理      │ 未处理      │
                    ▼            ▼             │
              直接ACK      ┌─────────▼─────────┐
                           │  2. 乱序检测        │
                           │  sequence_number   │
                           │  与预期序列号比对    │
                           └──────┬────────────┘
                                  │
                 ┌────────────────┼────────────────┐
                 │ 正常顺序        │ 跳号(丢失)       │ 超前(乱序)
                 ▼                ▼                 ▼
           ┌──────────┐   ┌──────────────┐   ┌──────────────┐
           │ 3. 处理   │   │ 等待窗口      │   │ 暂存缓冲区    │
           │ 业务逻辑  │   │ + 缺失事件    │   │ 等待前序事件  │
           │          │   │ 查询/拉取    │   │              │
           └──────────┘   └──────┬───────┘   └──────┬───────┘
                                 │                  │
                          ┌──────▼───────┐   ┌──────▼───────┐
                          │ 超时未到达    │   │ 前序事件到达  │
                          │ 告警 + DLQ   │   │ 按序处理      │
                          └──────────────┘   └──────────────┘
```

### 5.2 重复事件处理

**产生原因**: 生产者重试、Kafka at-least-once投递、消费者rebalance导致重复消费。

**处理策略**:

1. **幂等消费（核心机制）**：每个消费者在处理事件前，以 `event_id` 为key查询去重表。

```python
# 消费端幂等处理伪代码
async def handle_event(event: EventEnvelope):
    event_id = event["event_id"]
    
    # 1. 检查去重缓存
    if await redis.exists(f"dedup:{event_id}"):
        logger.info(f"Duplicate event {event_id}, skipping")
        return  # 直接ACK，不重复处理
    
    # 2. 处理业务逻辑（数据库事务）
    async with db.transaction():
        # 2a. 插入去重记录（利用唯一约束防并发重复）
        try:
            await db.execute(
                "INSERT INTO idempotency_keys (event_id, processed_at) VALUES ($1, NOW())",
                event_id
            )
        except UniqueViolation:
            return  # 并发重复，另一个实例已处理
        
        # 2b. 执行业务逻辑
        await process_order_created(event["payload"])
    
    # 3. 标记已处理（本地缓存+分布式缓存）
    await redis.setex(f"dedup:{event_id}", 86400 * 7, "1")  # 7天TTL
```

2. **数据库唯一约束兜底**：`idempotency_keys`表的`event_id`唯一索引作为最后防线，防止Redis故障导致重复处理。

3. **Kafka生产者幂等**：`enable.idempotence=true`，Kafka broker端去重（基于生产者ID+序列号），减少进入topic的重复消息。

### 5.3 乱序事件处理

**产生原因**: 网络抖动、分区再均衡、同一聚合根事件被路由到不同partition。

**处理策略**:

1. **聚合根路由**：确保同一order_id的事件始终路由到同一partition。选择order_id作为Kafka消息key：

```
Producer: key=order_id -> hash(order_id) % partition_count = 固定partition
```

这样就可以保证单个订单的所有事件在partition内有序。

2. **序列号检测**：消费者维护 `{aggregate_id: expected_sequence_number}` 映射。

```python
class SequenceTracker:
    def __init__(self, redis_client):
        self.redis = redis_client
    
    async def check_sequence(self, aggregate_id: str, seq_num: int, event: dict) -> SequenceResult:
        expected = await self.redis.get(f"seq:{aggregate_id}")
        
        if expected is None:
            # 首次处理，记录当前序列号
            await self.redis.set(f"seq:{aggregate_id}", seq_num)
            return SequenceResult.IN_ORDER
        
        expected = int(expected)
        
        if seq_num == expected + 1:
            # 正常顺序
            await self.redis.incr(f"seq:{aggregate_id}")
            return SequenceResult.IN_ORDER
        elif seq_num <= expected:
            # 乱序(滞后事件)
            return SequenceResult.OUT_OF_ORDER_LATE
        else:
            # 跳号(seq_num > expected + 1)，可能丢失中间事件
            return SequenceResult.GAP_DETECTED
```

3. **缓冲区重排**：对于提前到达的事件（seq_num > expected + 1），暂存到Redis Sorted Set（score为序列号），等待前序事件：

```python
# 暂存超前事件
await redis.zadd(f"buffer:{aggregate_id}", {event_id: seq_num})

# 设置等待超时(如30秒)
await redis.expire(f"buffer:{aggregate_id}", 30)

# 当前序事件到达后，从缓冲区拉取连续事件按序处理
pending = await redis.zrangebyscore(f"buffer:{aggregate_id}", 
                                     min=expected+1, max='+inf')
```

4. **超时兜底**：如果等待窗口（默认30秒）内缺失事件仍未到达：
   - 记录WARN级别日志，标注缺失的sequence_number
   - 发送到DLQ，人工判断是否补偿
   - 对于可容忍的业务场景（如通知），跳过继续处理；对于关键场景（如支付），阻塞等待。

### 5.4 丢失事件处理

**产生原因**: 生产者崩溃、网络分区、partition leader切换、消费者处理失败未commit offset。

**处理策略**:

1. **Outbox模式防生产丢失**：事件写入数据库outbox表与业务操作在同一事务内，避免"业务成功但事件未发布"。Debezium CDC持续监听outbox表，发现PENDING记录即发布。如果CDC故障恢复后会自动追赶未处理的记录。

2. **端到端事件追踪**：在事件链起始点注入correlation_id和预期事件数量。独立的审计消费者按correlation_id聚合，检测事件链完整性：

```python
# 审计消费者：检测订单完整事件链
expected_chain = [
    "order.order.created",
    "inventory.stock.reserved",
    "inventory.stock.deducted", 
    "payment.payment.created",
    "payment.payment.confirmed",
    "logistics.shipment.created",
    "logistics.shipment.dispatched",
    "logistics.shipment.delivered",
    "order.order.completed"
]

# 定时扫描：超过合理时间窗口（如1小时）仍未完整的事件链 → 告警
```

3. **定期对账任务**：每天凌晨执行全量对账：
   - 扫描订单表状态为"已支付"但超过24小时未出现物流事件的记录
   - 扫描库存扣减但支付事件缺失超过30分钟的记录
   - 对账不一致 → 自动补偿或人工介入

4. **消费者容错**：
   - 消费者处理失败时不提交offset，让Kafka重新投递
   - 可重试错误（临时故障、超时）→ 指数退避重试（最多3次）
   - 不可重试错误（数据校验失败、业务规则违反）→ 直接路由到DLQ
   - offset手动提交（`enable.auto.commit=false`），在处理成功后提交

---

## 6. 监控与死信队列设计

### 6.1 死信队列 (DLQ) 架构

```
   ┌──────────────────────────────────────────────────────────────────┐
   │                     Dead Letter Queue System                      │
   │                                                                   │
   │  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐             │
   │  │dlq.orders   │   │dlq.inventory│   │dlq.payments │   ...       │
   │  │ret=30d      │   │ret=30d      │   │ret=30d      │             │
   │  │p=1 r=2      │   │p=1 r=2      │   │p=1 r=2      │             │
   │  └──────┬──────┘   └──────┬──────┘   └──────┬──────┘             │
   │         │                 │                 │                     │
   │         └─────────────────┼─────────────────┘                     │
   │                           │                                       │
   │              ┌────────────▼────────────┐                          │
   │              │    DLQ Processor        │                          │
   │              │                         │                          │
   │              │  ┌───────────────────┐  │                          │
   │              │  │ 1. 分类 (Reason)  │  │                          │
   │              │  │ 2. 重试策略判断    │  │                          │
   │              │  │ 3. 执行重放/告警   │  │                          │
   │              │  │ 4. 持久化结果      │  │                          │
   │              │  └───────────────────┘  │                          │
   │              └────────────┬────────────┘                          │
   │                           │                                       │
   │              ┌────────────▼────────────┐                          │
   │              │   DLQ Dashboard (UI)    │                          │
   │              │   - 查看未处理事件       │                          │
   │              │   - 手动重放/跳过        │                          │
   │              │   - 原因分析/统计        │                          │
   │              │   - 批量操作             │                          │
   │              └─────────────────────────┘                          │
   └──────────────────────────────────────────────────────────────────┘
```

### 6.2 DLQ事件进入规则

| 失败类型 | 重试次数 | 退避策略 | 进入DLQ? | 附加操作 |
|----------|---------|----------|----------|----------|
| 临时网络超时 | 3 | 指数退避 (1s, 5s, 25s) | 是（3次后） | 监控告警 |
| 数据库连接池耗尽 | 3 | 指数退避 (2s, 10s, 50s) | 是（3次后） | 扩容通知 |
| Schema校验失败 | 0 | 不重试 | 是（立即） | 开发修复schema |
| 业务规则冲突 | 0 | 不重试 | 是（立即） | 人工判断补偿 |
| 下游服务503 | 5 | 指数退避 + 抖动 (max 60s) | 是（5次后） | 断路器打开通知 |
| 数据不存在(如查不到order) | 3 | 固定间隔(10s) | 是（3次后） | 可能为乱序，待前序事件到达 |

### 6.3 DLQ事件结构（增加错误上下文）

```json
{
  "original_event": { /* 完整的原始事件envelope */ },
  "dlq_metadata": {
    "failed_at": "2026-06-29T10:30:00Z",
    "failed_reason": "INVENTORY_SERVICE_UNAVAILABLE",
    "failed_message": "Connection refused: inventory-service:5432",
    "retry_count": 5,
    "max_retries": 5,
    "last_retry_at": "2026-06-29T10:35:00Z",
    "retry_backoff_ms": [1000, 5000, 25000, 50000, 60000],
    "source_consumer_group": "inventory-service",
    "source_topic": "orders.events",
    "source_partition": 3,
    "source_offset": 152843,
    "stack_trace": "java.net.ConnectException: ...",
    "severity": "HIGH"
  },
  "resolution": {
    "status": "PENDING",
    "assigned_to": null,
    "resolved_at": null,
    "resolution_type": null,
    "comment": null
  }
}
```

### 6.4 监控指标体系

#### 6.4.1 四大黄金信号

| 信号类别 | 指标 | 采集方式 | 告警阈值 |
|----------|------|---------|---------|
| **延迟 (Latency)** | 端到端事件处理延迟 | 消费端在event metadata中记录produce_time，处理完成后计算差值 | P99 > 5s 告警 |
| **流量 (Traffic)** | 每个topic的事件速率 (events/sec) | Kafka JMX metrics (`kafka.server:type=BrokerTopicMetrics,name=MessagesInPerSec`) | 偏离基线±50% 告警 |
| **错误 (Errors)** | 消费失败率、DLQ增长率 | 消费者记录处理失败计数，/ DLQ topic的`MessagesInPerSec` | 失败率 > 1% 告警 |
| **饱和度 (Saturation)** | 消费者lag | `kafka.consumer:type=consumer-fetch-manager-metrics,client-id=...` 的 `records-lag-max` | lag > 10000 告警 |

#### 6.4.2 业务级监控

| 指标 | Prometheus Metric名称 | 说明 |
|------|----------------------|------|
| 订单创建到支付确认时间 | `order_to_payment_duration_seconds` | 直方图，按channel分 |
| 订单生命周期完整率 | `order_lifecycle_completion_ratio` | 已完成/总创建，滚动24h |
| 事件链完整性 | `event_chain_completeness_ratio` | 按correlation_id聚合 |
| 库存扣减到发货时间 | `stock_deducted_to_dispatched_duration_seconds` | 直方图 |
| 补偿事务触发次数 | `compensation_transaction_total` | 按reason分 |
| DLQ堆积量 | `dlq_pending_events_total` | 按topic/reason分 |

#### 6.4.3 Grafana仪表板布局

```
┌───────────────────────────────────────────────────────────────┐
│  Row 1: Overview                                               │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐         │
│  │Events/sec│ │Consumer  │ │Error Rate│ │DLQ Pending│         │
│  │ (by top.)│ │Lag (max) │ │ (by svc.)│ │ (by topic)│         │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘         │
├───────────────────────────────────────────────────────────────┤
│  Row 2: Event Flow Timeline                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Sankey Diagram: Orders -> Reserve -> Deduct -> Pay      │  │
│  │  -> Confirm -> Ship -> Deliver -> Complete                │  │
│  │  (各阶段转化率+耗时标注)                                    │  │
│  └──────────────────────────────────────────────────────────┘  │
├───────────────────────────────────────────────────────────────┤
│  Row 3: Per-Service Detail (可切换Tab)                         │
│  ┌──────────────────────┐ ┌──────────────────────┐            │
│  │ Order Service        │ │ Payment Service      │            │
│  │ - 处理延迟 P50/P90/P99│ │ - 支付成功率           │            │
│  │ - 事件发布速率        │ │ - 网关调用延迟         │            │
│  │ - Outbox积压量        │ │ - 重复事件率           │            │
│  └──────────────────────┘ └──────────────────────┘            │
├───────────────────────────────────────────────────────────────┤
│  Row 4: DLQ Operations                                         │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  DLQ队列深度趋势图 + 失败原因分布饼图 + 最近失败事件列表     │  │
│  └──────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
```

### 6.5 告警规则

| 告警名称 | 条件 | 严重度 | 通知渠道 |
|----------|------|--------|----------|
| ConsumerLagHigh | `kafka_consumer_lag > 10000` for 5m | critical | PagerDuty + 飞书 |
| DLQGrowRate | `rate(dlq_events_total[5m]) > 10` | critical | PagerDuty |
| EventChainBroken | `event_chain_completeness < 0.95` for 15m | warning | 飞书 |
| PaymentTimeout | `order_to_payment_duration > 1800s` (30min) | critical | PagerDuty + 飞书 |
| OutboxBacklog | `outbox_pending > 1000` for 5m | warning | 飞书 |
| ServiceDown | `up{job="*service*"} == 0` for 1m | critical | PagerDuty |

---

## 7. 部署拓扑与扩展性

### 7.1 Kubernetes部署拓扑

```
┌────────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                           │
│                                                                 │
│  ┌──────────────────────┐    ┌──────────────────────────────┐  │
│  │   Control Plane      │    │        Worker Nodes           │  │
│  │   - API Server       │    │                               │  │
│  │   - Scheduler        │    │  ┌─────────────────────────┐  │  │
│  │   - Controller Mgr   │    │  │ Node-1                   │  │  │
│  └──────────────────────┘    │  │  order-svc (x2 pods)    │  │  │
│                               │  │ payment-svc (x2 pods)   │  │  │
│  ┌──────────────────────┐    │  │ kafka-broker-1          │  │  │
│  │   Namespace: event   │    │  └─────────────────────────┘  │  │
│  │   - Strimzi Operator │    │                               │  │
│  │   - Kafka CR         │    │  ┌─────────────────────────┐  │  │
│  │   - KafkaConnect     │    │  │ Node-2                   │  │  │
│  │   - Schema Registry  │    │  │  inventory-svc (x2 pods) │  │  │
│  └──────────────────────┘    │  │  logistics-svc (x2 pods) │  │  │
│                               │  │  kafka-broker-2          │  │  │
│  ┌──────────────────────┐    │  └─────────────────────────┘  │  │
│  │   Namespace: monitor │    │                               │  │
│  │   - Prometheus       │    │  ┌─────────────────────────┐  │  │
│  │   - Grafana          │    │  │ Node-3                   │  │  │
│  │   - Jaeger           │    │  │  notification-svc (x2)   │  │  │
│  └──────────────────────┘    │  │  kafka-broker-3          │  │  │
│                               │  │  dlq-processor           │  │  │
│  ┌──────────────────────┐    │  └─────────────────────────┘  │  │
│  │   Namespace: db      │    └──────────────────────────────┘  │
│  │   - PostgreSQL (HA)  │                                       │
│  │   - Redis Cluster    │                                       │
│  └──────────────────────┘                                       │
└────────────────────────────────────────────────────────────────┘
```

### 7.2 扩展性设计要点

| 扩展维度 | 策略 | 实现 |
|----------|------|------|
| **事件量增长** | 水平增加partition | 初期12 partition (orders topic)，按需扩展到48 |
| **服务实例增长** | HPA (Horizontal Pod Autoscaler) | 基于CPU + consumer lag双指标自动扩缩 |
| **新事件类型** | Schema Registry追加，新topic | 不影响已有topic，渐进式添加 |
| **新消费者** | 注册新consumer group | 独立offset，从earliest或latest开始消费 |
| **多数据中⼼** | MirrorMaker 2.0跨集群复制 | 异步复制，offset自动翻译 |
| **多租户** | tenant_id路由到独立partition | Kafka header注入tenant_id |

---

## 总结

本设计为电商系统从单体到事件驱动微服务的完整迁移方案：

1. **事件定义**：14种核心事件类型覆盖完整订单生命周期，统一envelope+独立payload schema，Schema Registry管理版本兼容性。

2. **消息队列**：选择Apache Kafka，核心决策因素为事件回溯能力、事件溯源兼容性、多消费者组独立消费、CDC生态集成、水平扩展性。

3. **混合模式**：事件通知为主（服务间协调）、事件溯源为辅（审计对账），通过Outbox+Debezium CDC保证事件发布的原子性和可靠性。

4. **异常处理**：三层防护——重复事件通过幂等消费+数据库唯一约束消除，乱序事件通过partition路由+序列号追踪+缓冲区重排+超时兜底处理，丢失事件通过Outbox保证+端到端追踪+定期对账发现。

5. **监控运维**：DLQ系统按失败类型分类处理，Grafana仪表板覆盖四大黄金信号+业务指标，PagerDuty+飞书告警联动。

---

**设计自信度**: 0.88 (基本可靠)

不确定部分：
- 具体partition数量需根据实际流量压测确定
- DLQ自动重试的退避参数需在生产环境调优
- MirrorMaker跨集群复制的延迟和一致性需实测验证

<!-- answer_complete -->
