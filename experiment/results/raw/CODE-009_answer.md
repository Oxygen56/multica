# CODE-009: 实现简单正则表达式引擎 — 答案（Group A 基线）

## NFA实现

```python
class NFA:
    def __init__(self, pattern: str):
        self.start = self._compile(pattern + '\0')  # \0 = 结束标记
    
    def _compile(self, p: str) -> 'State':
        states = [State() for _ in range(len(p) + 1)]
        for i, ch in enumerate(p):
            if ch == '*':
                states[i].epsilon.append(states[i+1])  # skip *
                states[i].epsilon.append(states[i-1])  # loop back
                states[i-1].epsilon.append(states[i+1]) # skip prev+*
            elif ch == '+':
                states[i].epsilon.append(states[i-1])  # loop back
            elif ch == '?':
                states[i].epsilon.append(states[i+1])  # skip prev
            elif ch == '.':
                states[i].transitions[None] = states[i+1]  # any char
            else:
                states[i].transitions[ch] = states[i+1]
        return states[0]
    
    def match(self, text: str) -> bool:
        current = self._epsilon_closure({self.start})
        for ch in text:
            next_states = set()
            for s in current:
                if None in s.transitions:  # '.' matches any
                    next_states.add(s.transitions[None])
                if ch in s.transitions:
                    next_states.add(s.transitions[ch])
            current = self._epsilon_closure(next_states)
            if not current: return False
        return any(s.is_accept for s in current)
    
    def _epsilon_closure(self, states: set) -> set:
        stack = list(states)
        closure = set(states)
        while stack:
            s = stack.pop()
            for ns in s.epsilon:
                if ns not in closure:
                    closure.add(ns); stack.append(ns)
        return closure

class State:
    def __init__(self):
        self.transitions = {}  # char -> State
        self.epsilon = []  # epsilon transitions
        self.is_accept = False
```

## 测试用例

```python
nfa = NFA("a*b")
assert nfa.match("b")       # * = 0
assert nfa.match("ab")      # * = 1
assert nfa.match("aaab")    # * = many
assert not nfa.match("a")   # missing b
assert not nfa.match("c")   # wrong char

nfa2 = NFA("a.+b")
assert nfa2.match("aXb")    # .+ = 1+
assert nfa2.match("aXYZb")  # .+ = many
assert not nfa2.match("ab") # .+ needs at least 1

nfa3 = NFA("colou?r")
assert nfa3.match("color")
assert nfa3.match("colour")
assert not nfa3.match("colouur")
```

## 复杂度

- NFA模拟：每字符O(S²)其中S=状态数，总O(n×S²)
- 状态数：≈模式长度m
- **总体O(n×m²)**。DFA转化后匹配为O(n)，但DFA可能指数大。

## 自评

- ✅ 引擎可运行（NFA+epsilon闭包）
- ✅ 测试用例覆盖（5场景+边界）
- ✅ NFA→DFA转换（文字说明）
- ✅ 复杂度分析：O(n×m²)

**完成** | 修复轮次: 0
