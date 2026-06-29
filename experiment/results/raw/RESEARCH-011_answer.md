# RESEARCH-011: Python异步编程最佳实践 — 答案（Group A 基线）

## 1. async/await 使用模式

```python
# ✅ 正确：异步调用链完整
async def fetch_data(url: str) -> dict:
    async with aiohttp.ClientSession() as session:
        async with session.get(url) as resp:
            return await resp.json()

# ❌ 错误：在async函数中调同步阻塞代码
async def bad_fetch(url: str) -> dict:
    return requests.get(url).json()  # 阻塞事件循环！

# ✅ 正确：用线程池执行阻塞代码
async def good_fetch(url: str) -> dict:
    return await asyncio.to_thread(requests.get, url).json()
```

**关键模式**：`asyncio.gather()` 并发执行多个协程；`asyncio.create_task()` 创建后台任务。

## 2. 事件循环管理

```python
# ✅ 每个线程一个事件循环
async def main():
    loop = asyncio.get_running_loop()
    # 不要手动创建/销毁loop

# ✅ 嵌套async不需要新loop
# asyncio.run() 自动管理

# ⚠️ 避免在async函数中调用 asyncio.run()
# 会导致 "cannot be called from a running event loop" 错误
```

## 3. 并发 vs 并行

| 场景 | 工具 | 原因 |
|------|------|------|
| I/O密集型（HTTP请求） | `asyncio` | 单线程协程，高并发低开销 |
| CPU密集型（计算） | `concurrent.futures.ProcessPoolExecutor` | 绕过GIL |
| 混合场景 | `asyncio.to_thread()` | 阻塞I/O放线程池 |
| 大量文件操作 | `aiofiles` | 异步文件I/O |

## 4. 第三方库兼容性

| 同步库 | 异步替代 |
|--------|---------|
| `requests` | `httpx` (async模式) 或 `aiohttp` |
| `psycopg2` | `asyncpg` 或 `databases` |
| `redis-py` | `redis.asyncio` (Redis 4.2+) |
| `boto3` | `aioboto3` |
| `sqlalchemy` | `sqlalchemy[asyncio]` (1.4+) |
| `pymongo` | `motor` |

## 5. 调试和性能分析

```bash
# 检测未awaited的协程
python -W default -m asyncio

# 检测慢回调（>100ms）
PYTHONASYNCIODEBUG=1 python app.py

# 性能分析
import asyncio
loop = asyncio.get_event_loop()
loop.set_debug(True)  # 检测慢回调、未关闭资源
loop.slow_callback_duration = 0.1  # 100ms阈值
```

## 反模式清单

| 反模式 | 问题 | 修复 |
|--------|------|------|
| `time.sleep()` 在async中 | 阻塞事件循环 | `await asyncio.sleep()` |
| 用`open()`读大文件 | 阻塞 | `aiofiles.open()` |
| 忘记`await`协程 | 协程不执行 | 启用`-W default`警告 |
| 在循环中`await`逐个请求 | 串行化 | `asyncio.gather()` |
| 捕获`Exception`吞掉`CancelledError` | 任务无法取消 | 重新raise `CancelledError` |
| 共享可变状态不加锁 | 竞态（虽单线程但有await点） | `asyncio.Lock()` |

## 自评

- ✅ 覆盖5个方面
- ✅ 最佳实践有代码示例
- ✅ 反模式有具体例子说明危害
- ✅ 工具推荐有版本依据

**完成** | 修复轮次: 0
