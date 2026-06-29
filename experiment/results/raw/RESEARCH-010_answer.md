# RESEARCH-010: WebAssembly应用场景调研 — 答案（Group A 基线）

## 1. Serverless/FaaS中的Wasm

**现状**：Cloudflare Workers、Fastly Compute@Edge、Fermyon Spin都使用Wasm作为执行运行时。Wasm冷启动<1ms（vs 容器~100ms），适合边缘计算和事件驱动函数。

**代表产品**：Wasmedge（CNCF沙箱）、Wasmtime（Bytecode Alliance）。  
**成熟度**：⭐⭐⭐ 生产可用（Cloudflare Workers日处理万亿请求），但生态仍在早期。

## 2. 插件系统

**Envoy Proxy**：使用Wasm扩展过滤器（取代Lua）。优势：沙箱隔离（插件崩溃不影响Envoy主进程）、多语言支持（Rust/Go/C++编译到Wasm）。

**Zed编辑器**：使用Wasm插件系统。插件在沙箱中运行，不能访问文件系统除非显式授权。

**成熟度**：⭐⭐⭐⭐ Envoy的Wasm扩展已在生产中使用。

## 3. 边缘计算

Fastly、Cloudflare、Netlify都支持Wasm边缘函数。相比JavaScript（V8 isolates），Wasm提供：更好的性能（接近原生）、多语言支持、更小的运行时。

**成熟度**：⭐⭐⭐⭐ Cloudflare Workers支撑大量生产流量。

## 4. 区块链智能合约

Polkadot、NEAR、Solana使用Wasm作为智能合约运行时。替代EVM（以太坊虚拟机），提供更好的性能和语言灵活性。

**成熟度**：⭐⭐⭐ Polkadot/NEAR已主网上线。但Wasm合约总数远少于EVM合约。

## 成熟度总结

| 场景 | 成熟度 | 趋势 |
|------|--------|------|
| Serverless/FaaS | ⭐⭐⭐ | 上升（2024-2025快速增长） |
| 插件系统 | ⭐⭐⭐⭐ | 稳定（Envoy生产验证） |
| 边缘计算 | ⭐⭐⭐⭐ | 稳定（CDN厂商广泛采用） |
| 区块链 | ⭐⭐⭐ | 上升（非EVM链首选） |

**2年趋势**：Wasm正在成为"轻量级容器替代品"。不是替代Docker，而是替代"需要更快冷启动+更强隔离"的场景。WASI（WebAssembly System Interface）Preview 2的发布使Wasm能标准化地访问文件系统、网络等系统资源。

## 自评

- ✅ 4场景有具体案例
- ✅ 成熟度评估有依据
- ✅ 趋势预测有理有据
- ✅ 讨论了Wasm vs Docker差异化（非替代，互补）

**完成** | 修复轮次: 0
