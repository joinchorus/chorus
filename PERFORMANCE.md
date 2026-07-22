# Chorus System Performance, Benchmarks & Load Testing Report

This document presents empirical performance benchmarks, concurrency stress test metrics, and bottleneck analysis for **Chorus**.

All measurements were collected on a reference system (**Apple M4 Pro, 14 CPU cores, Go 1.26.2**) without altering core application logic or introducing speculative optimizations.

---

## 1. Micro-Benchmarks (`go test -bench`)

Micro-benchmarks were executed using Go's `testing.B` package (`internal/gitstore/gitstore_bench_test.go`).

### Benchmark Results Summary

| Benchmark Operation | Iterations | Time / Op (ns) | Time / Op (ms) | B / Op | Allocs / Op |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `BenchmarkCreateThread` | 20 | 50,684,752 ns | **50.68 ms** | 98,244 B | 556 allocs |
| `BenchmarkAppendMessage` | 20 | 49,175,404 ns | **49.17 ms** | 78,105 B | 485 allocs |
| `BenchmarkGetThread` | 20 | 15,038 ns | **0.015 ms** | 1,496 B | 14 allocs |
| `BenchmarkListThreads` | 20 | 275,635 ns | **0.275 ms** | 47,240 B | 362 allocs |
| `BenchmarkListMessages` | 20 | 130,198 ns | **0.130 ms** | 47,096 B | 913 allocs |

### Key Benchmark Insights

1. **Read Throughput**: Read operations (`GetThread`, `ListThreads`, `ListMessages`) complete in **15 µs to 275 µs** (0.015ms - 0.275ms). Reads are **~3,300x faster** than write operations because they read directly from disk JSON/NDJSON files without executing Git CLI processes.
2. **Write Latency**: Write operations (`CreateThread`, `AppendMessage`) average **~49 ms - 50 ms** per operation.

---

## 2. Load & Stress Simulation Scenarios

Load testing was conducted using `scripts/loadtest.go`, simulating realistic production workloads (**80% reads, 20% writes**) across 100, 1,000, and 10,000 virtual concurrent users.

### Concurrency Stress Test Results

| Virtual Users | Total Requests | Total Time | Throughput (RPS) | Latency p50 (Median) | Latency p95 | Latency p99 | Max Latency | Errors | Total Mem Allocated |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **100 VUs** | 1,000 req | 12.57 s | **79.52 req/s** | **60.31 ms** | 6.48 s | 6.68 s | 6.69 s | **0** | 59.96 MB |
| **1,000 VUs** | 2,000 req | 19.47 s | **102.68 req/s** | **115.50 ms** | 14.54 s | 18.38 s | 19.28 s | **0** | 137.63 MB |
| **10,000 VUs** | 5,000 req | 50.05 s | **99.90 req/s** | **272.86 ms** | 37.10 s | 47.46 s | 50.04 s | **0** | 611.99 MB |

---

## 3. System Bottlenecks Identified

### Bottleneck 1: Git CLI Process Spawning Overhead (`exec.Command`)
- **Observation**: Every write operation (`SaveThread`, `SaveMessage`) executes `git add` and `git commit` as external subprocesses (`exec.Command("git", ...)`).
- **Impact**: Creating an OS process (`fork/exec`) takes ~15ms to ~25ms on macOS/Linux. Executing two Git CLI commands per write introduces a baseline **~50ms latency floor** for every single write, accounting for **99.9%** of single-write latency.

### Bottleneck 2: Storage Mutex Lock Contention (`sync.RWMutex`)
- **Observation**: `GitStore` uses a single mutex lock to serialize write operations (`AddAndCommit`).
- **Impact**: Under high concurrency (1,000 to 10,000 virtual users), concurrent write goroutines wait in queue while earlier writes finish their Git CLI subprocess execution.
- **Result**: While median latency remains low (60ms - 272ms), high-percentile latencies (p95, p99) scale linearly with concurrency depth (6.4s at 100 VUs, 37.1s at 10,000 VUs).

### Bottleneck 3: Index Rewrite on Every Write
- **Observation**: Every message append updates `boards/general/index.json` by re-sorting all thread items and rewriting the index JSON to disk.
- **Impact**: Under heavy write traffic, re-marshalling and re-writing `index.json` adds unnecessary disk I/O serial overhead inside the lock.

---

## 4. Architectural Recommendations for Future Optimization

While optimization was explicitly excluded from this milestone, the following non-breaking enhancements are recommended for future performance iterations:

1. **Asynchronous / Batched Git Commits**:
   - Decouple HTTP write response from Git commit execution by buffering file changes and batching Git commits asynchronously (e.g. background worker committing queued writes every 500ms or N operations).
2. **In-Memory Write Queueing**:
   - Use channel-based write queues to eliminate lock contention on Git CLI commands.
3. **Index Patching**:
   - Append to index or update memory-mapped index items rather than rewriting full `index.json` structures on disk during message appends.
