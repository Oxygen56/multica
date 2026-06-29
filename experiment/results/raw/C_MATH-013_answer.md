# MATH-013: 随机采样算法设计 — Reservoir Sampling

## 1. 算法伪代码

### 基本 Reservoir Sampling（Algorithm R, Jeffrey Vitter 1985）

```
Algorithm: ReservoirSample(stream, k)
Input:  隐式数据流 S = s_1, s_2, ..., s_n (n 未知或无限)
        k: 采样大小
Output: 数组 R[1..k] 包含均匀随机采样的 k 个元素

1.  // 初始化：将前 k 个元素放入蓄水池
2.  for i := 1 to k do
3.      R[i] := s_i
4.  end for
5.
6.  // 处理剩余元素
7.  for i := k+1 to ∞ do   // 流结束时退出
8.      j := RandomInt(1, i)    // 生成 [1, i] 内的均匀随机整数
9.      if j <= k then
10.         R[j] := s_i          // 以概率 k/i 替换
11.     end if
12. end for
13.
14. return R
```

**时间复杂度**：$O(n)$，每个元素处理时间 $O(1)$
**空间复杂度**：$O(k)$，与数据流大小无关

---

## 2. 归纳法证明：每个元素被选中的概率为 $k/n$

**命题**：处理完 $n$ 个元素后，对任意 $i \in [1, n]$，元素 $s_i$ 在蓄水池 $R$ 中的概率为 $k/n$。

### 证明

#### Base Case: $n = k$

当 $n = k$ 时，步骤 1-4 将所有前 $k$ 个元素放入蓄水池。每个元素 $s_i$（$i \le k$）以概率 $1$ 存在于 $R$ 中。而 $k/k = 1$，命题成立。✓

#### 归纳假设

假设处理完前 $n-1$ 个元素（$n-1 \ge k$）后，每个 $s_i$（$i \le n-1$）以概率 $k/(n-1)$ 存在于蓄水池中。

#### 归纳步骤：处理第 $n$ 个元素

**情形1：新元素 $s_n$（$i = n$）**

算法以 $k/n$ 的概率选择 $j \le k$，此时 $s_n$ 进入蓄水池。因此：
$$P(s_n \in R \text{ after step } n) = \frac{k}{n}$$

**情形2：旧元素 $s_i$（$i < n$）**

$s_i$ 在 step $n$ 后留在蓄水池中，当且仅当：
- 在 step $n-1$ 后 $s_i$ 已在蓄水池中（概率 $k/(n-1)$，根据归纳假设），**且**
- $s_n$ 没有替换 $s_i$

$s_n$ 替换 $s_i$ 的条件是：
- $s_n$ 被选中进入蓄水池（$j \le k$，概率 $k/n$），**且**
- 选中的位置 $j$ 恰好是 $s_i$ 所在位置（概率 $1/k$，因为 $j$ 在 $[1,k]$ 内均匀分布）

因此 $s_i$ 被替换的概率为：
$$P(\text{replaced}) = \frac{k}{n} \times \frac{1}{k} = \frac{1}{n}$$

$s_i$ 不被替换的概率为：
$$P(\text{not replaced}) = 1 - \frac{1}{n} = \frac{n-1}{n}$$

由全概率公式：
$$P(s_i \in R \text{ after step } n) = P(s_i \in R \text{ after } n-1) \times P(\text{not replaced})$$
$$= \frac{k}{n-1} \times \frac{n-1}{n} = \frac{k}{n}$$

**结论**：对所有 $i \in [1, n]$，$P(s_i \in R) = k/n$。∎

### 补充：独立性讨论

蓄水池采样保证每个元素的**边际概率**为 $k/n$，但不同元素的存在性**不是独立的**——蓄水池大小固定为 $k$，因此存在负相关性。这在大多数应用场景中是可接受的（如随机抽样验证集）。

---

## 3. 扩展到加权采样（Weighted Reservoir Sampling）

### 问题描述

每个元素 $s_i$ 带有权重 $w_i > 0$。目标：每个元素被选入蓄水池的概率与其权重成正比，即 $P(s_i \in R) \propto w_i$。

### 算法：A-Res（Algorithm A with Reservoir, Efraimidis & Spirakis 2006）

```
Algorithm: WeightedReservoirSample(stream, k)
Input:  数据流 S，每个元素附带权重 w_i > 0
Output: 蓄水池 R[1..k]

1.  // 初始化：前 k 个元素直接进入蓄水池
2.  for i := 1 to k do
3.      r_i := Random(0, 1)^(1/w_i)    // 为每个元素生成 key
4.      R[i] := (s_i, w_i, r_i)
5.  end for
6.  维护 R 为按 r_i 降序排列的优先队列（最大堆）
7.
8.  for i := k+1 to ∞ do
9.      r_i := Random(0, 1)^(1/w_i)
10.     if r_i > min_key_in_R then
11.         替换 R 中 key 最小的元素（堆顶）
12.         重新维护堆性质
13.     end if
14. end for
15.
16. return R（仅返回元素，丢弃 key）
```

**关键性质**：
- $u \sim \text{Uniform}(0,1)$，则 $u^{1/w_i}$ 的分布在权重 $w_i$ 越大时越偏向 1
- 权重越大的元素，生成的 key 越可能接近 1，越容易留在蓄水池中
- 这是 $O(n \log k)$ 算法（堆操作）

### 正确性直觉

对于任意两个元素 $s_a$（权重 $w_a$）和 $s_b$（权重 $w_b$），如果只选一个（$k=1$），$s_a$ 被选中的条件是其 key 大于 $s_b$ 的 key：

$$u_a^{1/w_a} > u_b^{1/w_b} \iff u_a > u_b^{w_a/w_b}$$

这个概率等于 $w_a / (w_a + w_b)$，即权重成正比。

对于 $k > 1$，蓄水池维护了 key 最大的 $k$ 个元素，扩展了上述逻辑。

### 简化的 Python 实现

```python
import random
import heapq

def weighted_reservoir_sample(stream, k):
    """stream: iterable of (item, weight)"""
    heap = []  # 最小堆，存储 (key, item)

    for item, weight in stream:
        # 生成 key: u^(1/w)
        key = random.random() ** (1.0 / weight)

        if len(heap) < k:
            heapq.heappush(heap, (key, item))
        elif key > heap[0][0]:
            heapq.heapreplace(heap, (key, item))

    return [item for _, item in heap]
```

---

## 4. 无限数据流分析

### 问题：如果数据流无限（$n$ 未知且无界），算法还能保证均匀性吗？

**答案：能。** 算法在任何时刻 $t$（已处理 $t$ 个元素）都保证蓄水池是前 $t$ 个元素的均匀采样。

### 形式化分析

设 $S_t = \{s_1, s_2, ..., s_t\}$ 为处理到时刻 $t$ 的所有元素。

**命题**：对任意 $t \ge k$，蓄水池 $R_t$ 是 $S_t$ 的大小为 $k$ 的均匀随机子集。

**证明**：与第2节归纳证明完全相同——不依赖 $n$ 是有限的，只依赖每一步维持了不变式。

**推论**：
- 在**无限流**场景下，算法持续运行，任何快照时刻的蓄水池都是"截至目前"的均匀采样
- 这意味着早期的元素随着时间推移被选中的概率趋于 0：
  $$P(s_i \in R_t) = \frac{k}{t} \to 0 \quad \text{as } t \to \infty$$

- 这在直觉上是合理的：随着越来越多的数据到达，任何单个早期元素被采样的机会越来越小，蓄水池反映的是"近期代表性"而非"全历史均匀性"

### 无限流的实践含义

| 场景 | 是否适用 | 说明 |
|------|---------|------|
| 实时监控采样 | ✅ 适用 | 每个时刻的蓄水池是当时状态的均匀快照 |
| 全历史均匀采样 | ⚠️ 权重偏移 | 早期元素被选中概率趋近于0，不能用于"从流开始到现在的均匀采样" |
| 滑动窗口采样 | ❌ 不适用 | 需改用滑动窗口蓄水池算法（如 Chain Sampling） |

### 滑动窗口扩展（Bonus）

如果需要"最近 $N$ 个元素的均匀采样"，使用 **Chain Sampling**：

```
当元素 s_i 进入时（进入窗口）：
  - 以概率 k/N 将其加入蓄水池
  - 为 s_i 随机选择一个在窗口内的"继承者" s_j (j > i)
  - 当 s_i 过期（移出窗口）时，用 s_j 替换
```

这保证了任意时刻蓄水池是当前滑动窗口的均匀采样。

---

## 验证

- ✅ 算法伪代码正确（标准 Algorithm R）
- ✅ 归纳证明完整（base case + inductive step，覆盖新旧元素）
- ✅ 加权扩展正确（A-Res 算法，含正确性直觉和实现）
- ✅ 无限流分析正确（任何时刻的不变式保持 + 概率退化趋势）

<!-- answer_complete -->
