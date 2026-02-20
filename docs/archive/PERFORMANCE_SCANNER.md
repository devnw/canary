# Scanner Performance Notes

Target: <10s wall clock on ~50k text files; memory RSS <= 512 MiB.

## Current Benchmark Snapshot

`BenchmarkScannerLarge` (5k mixed small files) baseline shows ~2.7s/op on high core CPU (Threadripper). Scaling roughly linearly, 50k similar files would project near ~27s unless mitigated. Real-world variance (I/O caching, SSD) may reduce effective time. Optimization required.

## Hotspots (Likely)

1. `filepath.WalkDir` sequential traversal
2. Per-file allocation of 1MB scanner buffer
3. Regex match (`scannerCanaryRe`) on every line
4. Map growth for aggregated keys (many small maps)

## Immediate Low-Risk Improvements

1. Reuse a global `[]byte` buffer pool for scanner allocations to cut GC churn.
2. Short-circuit file scan if first 16KB contains no `CANARY:` substring (pre-filter with `Index` before full line tokenization).
3. Replace heavy regex with a simpler prefix scan followed by manual parsing (state machine) for speed.
4. Use `WalkDir` concurrency: fan out file paths to worker goroutines (bounded, e.g. GOMAXPROCS\*2) reading & parsing; aggregate via channels.

## Medium-Risk Improvements

1. Memory-map large files (`os.ReadFile` vs line scan) only when size < 256KB to reduce syscall overhead.
2. Precompile a smaller set of aspect/status validations; avoid repeated `strings.ToUpper` by normalizing once.

## Measurement Strategy

1. Add micro-benchmarks: parsing single line with regex vs manual parser.
2. Add benchmark for scanning directory with varying CANARY density (sparse vs dense) to evaluate pre-filter benefits.
3. Use `go test -bench . -benchmem` to capture alloc counts before/after changes.
4. Optionally integrate `pprof` (`go test -bench BenchmarkScannerLarge -cpuprofile cpu.out -memprofile mem.out`).

## Acceptance Gate After Optimization

Add CI guard that fails if benchmark exceeds threshold (configurable) on standard runner. (Note: results can flake; consider percentile or allow slack.)

## Next Steps

1. Implement buffer pool + substring pre-filter.
2. Introduce concurrency with controlled worker pool.
3. Replace regex with manual parser; retain regex behind feature flag for fallback.
4. Benchmark and document deltas here.

---

Maintainer: canary
Updated: 2025-11-02
