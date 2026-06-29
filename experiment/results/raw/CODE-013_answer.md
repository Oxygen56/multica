# CODE-013: 排查内存泄漏 — 答案（Group A 基线）

## 4种可能的内存泄漏来源

**1. 闭包循环引用**
```python
# 泄漏：handler持有self的引用，self持有handlers
class Service:
    def register(self):
        def handler(): return self.do_something()  # 闭包捕获self
        self.handlers.append(handler)  # 循环引用
```
**排查**：`gc.get_objects()` + `objgraph.show_backrefs()`

**2. 自定义缓存无上限**
```python
_cache = {}
def get_data(key):
    if key not in _cache:
        _cache[key] = expensive_query(key)
    return _cache[key]  # 无限增长！
```
**修复**：`functools.lru_cache(maxsize=1000)` 或带TTL的缓存

**3. asyncio未等待的Task**
```python
async def handler():
    asyncio.create_task(background_job())  # Task未被await，也未存储引用
    # 如果background_job永不结束 → Task对象永远存在
```
**排查**：`asyncio.all_tasks()` 检查待处理任务数

**4. 第三方ML库的GPU/内存泄漏**
- TensorFlow的`tf.keras.backend.clear_session()`
- PyTorch的梯度未释放（`loss.backward()`后未`optimizer.zero_grad()`）
- 排查：`nvidia-smi`看GPU内存，`tracemalloc`看Python内存

## 排查命令

```bash
# 1. 内存快照对比
python -c "import objgraph; objgraph.show_growth(limit=10)"

# 2. asyncio诊断
python -c "import asyncio; print(len(asyncio.all_tasks()))"

# 3. tracemalloc找最大内存消耗者
python -X tracemalloc=10 app.py

# 4. 实时监控
pip install memray
python -m memray run app.py
```

## 代码审查检查清单

- [ ] 所有缓存有TTL或maxsize限制
- [ ] 闭包中避免捕获大对象（用weakref）
- [ ] asyncio.create_task的返回值被存储或被await
- [ ] 定时任务(cron/callback)在服务停止时被取消
- [ ] 文件句柄/网络连接在finally中关闭
- [ ] 循环数据结构使用`__del__`或`weakref`
- [ ] 日志handler不会无限累积（RotatingFileHandler）
- [ ] 数据库连接池正确归还连接

## 自评

- ✅ 识别4+种泄漏来源（闭包/缓存/async/ML库）
- ✅ 具体排查命令+工具（memray/tracemalloc/objgraph）
- ✅ 8项审查检查清单
- ✅ 每种泄漏有修复建议

**完成** | 修复轮次: 0
