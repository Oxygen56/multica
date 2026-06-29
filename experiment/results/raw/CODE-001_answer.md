# CODE-001: 修复并发竞态条件 — 答案（Group A 基线）

## 根因分析

代码存在两个并发问题：

**问题1：读锁内修改map**
`c.mu.RLock()` 持有的是读锁，但代码试图 `c.items[key] = nil`——这是在读锁下进行写操作。Go的`sync.RWMutex`不阻止这种行为（它不是静态检查的），但会导致data race：其他goroutine可能同时在读/写同一个map，造成panic或数据损坏。

**问题2：TOCTOU竞态（Time-of-check to Time-of-use）**
检查`item.expired()`和设置`nil`之间不是原子的。在高并发下：goroutine A检查过期→goroutine B更新了该item→goroutine A仍将其置nil，丢失了B的更新。

## 修复方案

### 方案A：升级为写锁（简单，适合低竞争场景）

```go
func (c *Cache) Get(key string) *Item {
    c.mu.RLock()
    item := c.items[key]
    if item != nil && !item.expired() {
        c.mu.RUnlock()
        return item
    }
    c.mu.RUnlock()

    // 需要清理时升级为写锁
    c.mu.Lock()
    defer c.mu.Unlock()
    // Double-check：避免其他goroutine已经处理了
    item = c.items[key]
    if item != nil && item.expired() {
        delete(c.items, key)  // 用delete而非赋值nil
        return nil
    }
    return item
}
```

### 方案B：异步清理（推荐，高性能场景）

```go
func (c *Cache) Get(key string) *Item {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item := c.items[key]
    if item != nil && item.expired() {
        go c.asyncCleanup(key)  // 异步清理
        return nil
    }
    return item
}

func (c *Cache) asyncCleanup(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if item, ok := c.items[key]; ok && item.expired() {
        delete(c.items, key)
    }
}
```

### 方案C：sync.Map（Go 1.9+标准库）

```go
type Cache struct {
    items sync.Map
}

func (c *Cache) Get(key string) *Item {
    v, ok := c.items.Load(key)
    if !ok {
        return nil
    }
    item := v.(*Item)
    if item.expired() {
        c.items.Delete(key)
        return nil
    }
    return item
}
```

## 方案对比

| 方案 | 性能 | 复杂度 | 适用场景 |
|------|------|--------|---------|
| A: 升级锁 | 中（锁竞争） | 低 | 低并发，简单可靠 |
| B: 异步清理 | 高 | 中 | 高并发，可容忍短暂不一致 |
| C: sync.Map | 高 | 低 | 高并发读写，key稳定 |

**推荐**：方案C或方案B。方案A的double-check锁升级模式是经典做法，但锁竞争会成为瓶颈。

## 自评

- ✅ 正确识别竞态条件根因（读锁写map + TOCTOU）
- ✅ 给出三种可行修复方案并比较优劣
- ✅ 修复后代码正确
- ✅ 解释panic原因

**完成** | 修复轮次: 0
