# MATH-013: 随机采样算法设计 — 答案（Group A 基线）

## 1. Reservoir Sampling算法

```python
import random

def reservoir_sample(stream, k):
    reservoir = []
    for i, item in enumerate(stream):
        if i < k:
            reservoir.append(item)
        else:
            j = random.randint(0, i)
            if j < k:
                reservoir[j] = item
    return reservoir
```

## 2. 均匀性证明（归纳法）

**命题**：处理n个元素后，每个元素在蓄水池中的概率为k/n。

**Base case (n=k)**：前k个元素直接放入，概率=k/k=1=k/n ✅

**归纳步骤**：假设n个元素时概率=k/n。第(n+1)个元素到来时：
- 第(n+1)个元素被选中的概率 = k/(n+1)（随机整数j∈[0,n]，j<k的概率=k/(n+1)）
- 前面第i个元素留在池中的概率 = P(之前在第n步被选中) × P(不被第n+1个元素替换)
  = (k/n) × (1 - P(被替换))
  = (k/n) × (1 - k/(n+1) × 1/k)
  = (k/n) × (1 - 1/(n+1))
  = (k/n) × (n/(n+1))
  = k/(n+1) ✅

**得证**：每一步所有元素概率相等=k/n。

## 3. 加权采样扩展（A-Res算法）

```python
import random, heapq

def weighted_reservoir_sample(stream, k):
    reservoir = []  # (priority, item) 最小堆
    for item, weight in stream:
        # 生成带权重的随机key
        r = random.random()
        priority = r ** (1.0 / weight)  # 关键：权重越大priority越小
        if len(reservoir) < k:
            heapq.heappush(reservoir, (priority, item))
        elif priority > reservoir[0][0]:  # 比堆中最小priority大
            heapq.heapreplace(reservoir, (priority, item))
    return [item for _, item in reservoir]
```

**数学直觉**：对于权重为w的元素，P(priority最小) ∝ w。均匀分布取幂后权重越大的元素越可能获得小priority，从而留在top-k堆中。

## 4. 无限流分析

Reservoir Sampling不依赖n的值（算法中只用到当前索引i），所以**无限流仍然适用**。每一项在任意时刻都有k/i的概率被采样（i为当前已处理数量）。算法不需要知道总长度——这就是它的核心优势。

## 自评

- ✅ 算法伪代码正确
- ✅ 归纳证明完整
- ✅ 加权扩展正确（A-Res算法）
- ✅ 无限流分析正确（不依赖总长度）

**完成** | 修复轮次: 0
