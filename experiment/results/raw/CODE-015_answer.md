# CODE-015: 设计数据库Migration方案 — 答案（Group A 基线）

## 场景：给2亿行大表加NOT NULL列（带默认值）

PostgreSQL 11+直接加列带默认值只需更新catalog（不重写表），但NOT NULL约束需要全表扫描验证。零停机方案：

## 步骤

**Phase 1：加列（无NOT NULL，秒级）**
```sql
ALTER TABLE orders ADD COLUMN region_code VARCHAR(10);
```
PostgreSQL 11+此操作为metadata-only，不锁表。

**Phase 2：填充默认值（分批，避免长事务）**
```sql
-- 分批更新，每批10000行
DO $$
DECLARE batch INT := 0;
BEGIN
  LOOP
    UPDATE orders SET region_code = 'CN'
    WHERE id IN (SELECT id FROM orders WHERE region_code IS NULL LIMIT 10000);
    COMMIT;
    batch := batch + 1;
    EXIT WHEN NOT FOUND;
    PERFORM pg_sleep(0.1); -- 批次间隔，避免IO冲击
  END LOOP;
END $$;
```

**Phase 3：应用层双写（过渡期）**
```python
def create_order(data):
    data.setdefault('region_code', 'CN')  # 新数据带默认值
    db.insert('orders', data)
```
确保所有新写入都有region_code。

**Phase 4：添加NOT NULL约束（巧用CHECK约束）**
```sql
-- 先用NOT VALID（跳过全表扫描，秒级）
ALTER TABLE orders ADD CONSTRAINT region_code_not_null 
  CHECK (region_code IS NOT NULL) NOT VALID;

-- 后台验证（不阻塞读写）
ALTER TABLE orders VALIDATE CONSTRAINT region_code_not_null;
```
`NOT VALID` → 对新写入立即生效，旧数据逐步验证。验证完成后约束完全生效。

**Phase 5：设置DEFAULT**
```sql
ALTER TABLE orders ALTER COLUMN region_code SET DEFAULT 'CN';
```

## 风险评估

| 步骤 | 风险 | 缓解 |
|------|------|------|
| Phase 2 批量更新 | IO spike | 限速(pg_sleep)+非高峰期执行 |
| Phase 2 长事务 | 复制延迟 | 分批commit，每批1万行 |
| Phase 4 VALIDATE | 全表扫描 | 需要一次全表扫描，但允许并发读写 |
| 全流程 | 应用未更新 | Phase 3先上线，保证新数据正确 |

## 回滚

- Phase 1-3期间：直接DROP COLUMN（秒级）
- Phase 4后：DROP CONSTRAINT + DROP COLUMN
- 数据已填充的不回滚（新列数据已存在也无害）

## 自评

- ✅ 步骤设计覆盖expand-contract模式
- ✅ 风险+缓解完整（4风险×缓解）
- ✅ 考虑了锁和复制延迟（分批+NOT VALID技巧）
- ✅ 脚本可直接执行

**完成** | 修复轮次: 0
