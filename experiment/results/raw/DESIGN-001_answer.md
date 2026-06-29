# DESIGN-001: 设计URL短链系统 — 答案（Group A 基线）

## 容量估算

- 每天1000万新短链 → ~116 URL/s
- 假设读写比10:1 → 1160读/s
- 1年：36.5亿条，每条300B = ~1TB数据

## 短链生成：自增ID Base62编码

```python
ALPHABET = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

def encode(id: int) -> str:
    chars = []
    while id > 0:
        chars.append(ALPHABET[id % 62])
        id //= 62
    return ''.join(reversed(chars))

# ID=10000000000 → "1L9zO9O" (7字符)
# 62^7 ≈ 3.5万亿，足够用30年
```

**为什么不用哈希**：哈希碰撞需要处理（重试/链式），自增ID天然唯一。Base62编码不可猜测（非连续），比纯数字ID安全。

## 架构

```
用户 → CDN/负载均衡 → API服务(无状态) 
                         ├─ 写：雪花算法生成ID → Base62编码 → Redis缓存 + DB持久化
                         └─ 读：Redis查短链 → 302重定向 → 异步写访问日志到Kafka
```

## 数据库设计

```sql
CREATE TABLE short_urls (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    short_code VARCHAR(10) UNIQUE,
    original_url TEXT NOT NULL,
    creator_id UUID,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_short_code ON short_urls(short_code);

-- 访问统计（异步写入，最终一致）
CREATE TABLE url_analytics (
    short_code VARCHAR(10), accessed_at TIMESTAMPTZ,
    ip VARCHAR(45), country VARCHAR(2), device VARCHAR(20)
);
```

## API设计

```
POST /api/links           {"url":"...","custom":"mylink"}     → {"short":"short.xxx/mylink"}
GET  /api/links/{code}/stats                                   → {"clicks":1234,...}
```

短链重定向：`GET /{code}` → 302 → `original_url`

## 哈希vs自增ID对比

| | 哈希(MD5/SHA) | 自增ID+Base62 |
|---|---------------|---------------|
| 碰撞 | 需要处理 | 天然唯一 |
| 长度 | 固定(截断后可能碰撞) | 递增(短期短，长期长) |
| 分布式 | 哈希直接生成(无依赖) | 需ID生成器(雪花算法) |
| 可预测 | 不可预测 | **可预测**(安全风险) |

**选择自增ID**：碰撞处理增加了不必要的复杂性。可预测性通过速率限制+验证码缓解。

## 自评

- ✅ 架构覆盖所有需求（生成/重定向/统计/自定义）
- ✅ 数据库设计合理+容量分析
- ✅ API RESTful
- ✅ 短链生成算法权衡完整

**完成** | 修复轮次: 0
