# POC - Impact of `defer` on Mutex Performance in Go

## 🎯 Objective

This POC (Proof of Concept) demonstrates why using `defer` with mutexes is a **bad practice** that can drastically impact the performance of a Go application under concurrent load.

## 🚨 The Problem

Many Go developers systematically use `defer` to unlock mutexes:

```go
mu.Lock()
defer mu.Unlock()  // ❌ BAD PRACTICE
// ... long processing ...
```

This approach keeps the mutex locked for the **entire duration of the function**, creating a critical bottleneck.

## ✅ The Solution

Unlock the mutex immediately after the critical section:

```go
mu.Lock()
// ... fast critical operation ...
mu.Unlock()  // ✅ GOOD PRACTICE

// ... long processing WITHOUT the mutex ...
```

## 📊 Benchmark Results

The tests compare two identical HTTP servers, with the only difference being how mutexes are handled:

### Average Latency per Request

| Concurrency | Bad Server (defer) | Good Server | **Improvement** |
|-------------|--------------------|-------------|------------------|
| 1 goroutine | 11.12 ms           | 11.20 ms    | -0.7%            |
| 10 goroutines | **106.97 ms**   | 11.89 ms    | **88.9%**        |
| 50 goroutines | **419.25 ms**   | 13.84 ms    | **96.7%**        |
| 100 goroutines | **558.35 ms**  | 16.17 ms    | **97.1%**        |

### Throughput (requests/second)

- **Bad Server**: ~87–90 req/s (constant, regardless of concurrency)
- **Good Server**:
  - 1 goroutine: 88 req/s
  - 10 goroutines: **687 req/s**
  - 50 goroutines: 367 req/s
  - 100 goroutines: 316 req/s

## 🔍 Analysis

1. **No concurrency** (1 goroutine): Same performance, no contention
2. **With concurrency**: The "bad" server becomes a **bottleneck**, as only one goroutine can proceed at a time
3. **Exponential impact**: The higher the concurrency, the worse the degradation (up to **97% slower**)

## 🏗️ Project Structure

- `bad_server.go`: HTTP server using mutex + defer (port 8081)
- `good_server.go`: HTTP server with optimized mutex usage (port 8082)
- `benchmark_test.go`: Comparative load tests
- `run_benchmark.sh`: Benchmark automation script

## 🚀 Installation and Execution

### Prerequisites

- Go 1.21 or higher  
- Git (to clone the project)

### Installation

1. Clone the repository:
```bash
git clone <repo-url>
cd poc
```

2. Install dependencies:
```bash
go mod download
go mod tidy
```

### Running the Benchmarks

#### Method 1: Automatic Script (Recommended)

The `run_benchmark.sh` script automatically handles server startup and benchmarking:

```bash
chmod +x run_benchmark.sh
./run_benchmark.sh
```

The script will:
1. Start both servers in the background
2. Check that they respond properly
3. Run each benchmark for 10 seconds
4. Display comparative latency results
5. Gracefully stop the servers

#### Method 2: Manual Execution

If you prefer to control each step:

1. **Terminal 1** – Start the "bad" server:
```bash
go run bad_server.go
# Server listens on http://localhost:8081
```

2. **Terminal 2** – Start the "good" server:
```bash
go run good_server.go
# Server listens on http://localhost:8082
```

3. **Terminal 3** – Verify that servers are running:
```bash
# Test the "bad" server
curl http://localhost:8081/stats

# Test the "good" server
curl http://localhost:8082/stats
```

4. **Terminal 3** – Run the benchmarks:
```bash
# Full benchmarks (10 seconds per test)
go test -bench=. -benchtime=10s benchmark_test.go

# Quick version (1 second per test)
go test -bench=. -benchtime=1s benchmark_test.go

# Latency-only test
go test -run TestLatencyComparison -v benchmark_test.go
```

### Understanding the Results

The benchmarks output:
- **ns/op**: Nanoseconds per operation
- **ms/req**: Milliseconds per request (easier to read)
- **req/s**: Requests per second (throughput)

As concurrency increases, the difference between the two approaches becomes more apparent.

## 💡 Key Takeaways

1. **Only use `defer` with mutexes for very short operations**
2. **Release mutexes as soon as possible** to enable parallelism
3. **Copy required data**, then unlock the mutex before processing
4. **Performance impact can be catastrophic** (up to 97% degradation)

## 📝 Conclusion

This POC proves that systematically using `defer` with mutexes is an anti-pattern that can effectively turn your application into a single-threaded system, negating all the benefits of Go's concurrency model.