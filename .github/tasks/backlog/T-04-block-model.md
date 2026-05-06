---
id: T-04
status: backlog
agent: Code Expert
depends-on: [T-01, T-02]
branch: T-04/block-model
---

## T-04 · Block Data Model

**Context**
Defines the core data structures for command blocks and the ring buffer. This is the shared model that UI components render, the executor writes into, and the ACP layer annotates with AI cards.

**Acceptance criteria**
- [ ] `internal/block/block.go` defines `Block` struct with fields: `ID uint64`, `Command string`, `Dir string`, `StartedAt time.Time`, `Duration time.Duration`, `ExitCode int`, `State BlockState`, `Output []string`, `AICard *AICard`
- [ ] `BlockState` is a typed string constant set: `StateRunning`, `StateSuccess`, `StateFailed`
- [ ] `AICard` struct has fields: `Prompt string`, `Tokens []string`, `Streaming bool`, `Accepted bool`, `Dismissed bool`
- [ ] `internal/block/ringbuffer.go` defines `RingBuffer` with capacity 500, methods `Add(Block)`, `All() []Block`, `Len() int`, protected by `sync.RWMutex`
- [ ] `RingBuffer.Add` drops the oldest block when capacity is exceeded (circular overwrite)
- [ ] `Block.ID` is assigned by an atomic counter inside `RingBuffer.Add` — callers do not set it
- [ ] `go test -race ./internal/block/...` passes including: cap-at-500 test, insertion-order test, concurrent-Add race test

**Tech notes**
- PRD OQ-2 (§14): ring buffer cap N=500
- `Output []string` stores one element per output line; rendering joins with `\n`
- This package must NOT import any UI, Bubble Tea, or Lip Gloss packages — it is pure data
- Use `sync/atomic` for the ID counter: `var nextID atomic.Uint64`; call `nextID.Add(1)` in `Add`
- `All()` must return a copy of the slice to prevent callers from mutating internal state

**Completion notes**
*(empty)*
