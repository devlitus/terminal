---
id: T-20
status: backlog
agent: Code Expert
depends-on: [T-18]
branch: T-20/perf-tests
---

## T-20 · Performance and Regression Tests

**Context**
Validates the non-functional requirements: 500-block ring buffer cap, no unbounded memory growth on long sessions, render performance at 16ms budget, and layout stability at extreme widths.

**Acceptance criteria**
- [ ] `TestRingBufferCap`: adds 600 blocks, asserts `Len() == 500` and the oldest 100 are gone
- [ ] `BenchmarkBlockViewportRender`: benchmarks `View()` on a 500-block viewport at 120-column width; P99 must be < 16ms (NFR-02)
- [ ] `TestConcurrentRingBufferAdd`: 1000 goroutines each call `Add` once; `go test -race` passes without panic or data race
- [ ] `TestViewportWidth80`: renders viewport at 80-column width; no panic, no layout overflow
- [ ] `TestViewportWidth220`: renders viewport at 220-column width; no panic, no layout overflow
- [ ] `TestMemoryGrowth`: adds and removes 500 blocks in a loop 10 times; calls `runtime.GC()` after each iteration; `HeapAlloc` must not grow more than 10MB above baseline
- [ ] `go test -race -count=1 ./...` passes

**Tech notes**
- NFR-02 (< 16ms keystroke-to-render), NFR-03 (no memory leaks), NFR-04 (80–220 columns)
- Use stdlib `testing` only — no external benchmark frameworks
- Benchmark pattern: `b.ResetTimer()` after setup; call `b.ReportAllocs()`
- Memory growth test uses `runtime.ReadMemStats` — not a formal leak detector, just a sanity check
- Width tests: construct a `Model` and call `Update(tea.WindowSizeMsg{Width: 80, Height: 40})` before calling `View()`

**Completion notes**
*(empty)*
