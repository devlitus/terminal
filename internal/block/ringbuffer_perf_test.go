package block

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

// TestRingBufferCap adds 600 blocks (100 more than capacity) and verifies
// that the oldest 100 are evicted: Len()==500 and cmd-0 is gone.
func TestRingBufferCap(t *testing.T) {
	rb := NewRingBuffer()

	for i := 0; i < 600; i++ {
		rb.Add(Block{Command: fmt.Sprintf("cmd-%d", i)})
	}

	if got := rb.Len(); got != 500 {
		t.Fatalf("expected Len() == 500, got %d", got)
	}

	all := rb.All()
	if all[0].Command == "cmd-0" {
		t.Error("cmd-0 should have been evicted but is still present at index 0")
	}
}

// TestConcurrentRingBufferAdd spawns 1000 goroutines each calling Add once.
// The race detector verifies there are no data races.
func TestConcurrentRingBufferAdd(t *testing.T) {
	rb := NewRingBuffer()

	var wg sync.WaitGroup
	const goroutines = 1000
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			rb.Add(Block{Command: "x"})
		}()
	}

	wg.Wait()
}

// TestMemoryGrowth repeatedly fills and resets the ring buffer, checking that
// heap allocation does not grow by more than 10 MB above the baseline.
func TestMemoryGrowth(t *testing.T) {
	var baseline runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&baseline)

	for iter := 0; iter < 10; iter++ {
		rb := NewRingBuffer()
		for i := 0; i < 500; i++ {
			rb.Add(Block{Command: fmt.Sprintf("block-%d", i)})
		}
		rb.Reset()
		runtime.GC()
	}

	var final runtime.MemStats
	runtime.ReadMemStats(&final)

	const maxGrowthBytes = 10 * 1024 * 1024 // 10 MB
	if final.HeapAlloc > baseline.HeapAlloc+maxGrowthBytes {
		t.Fatalf("heap grew too much: baseline=%d final=%d diff=%d (max allowed %d)",
			baseline.HeapAlloc, final.HeapAlloc,
			final.HeapAlloc-baseline.HeapAlloc,
			uint64(maxGrowthBytes))
	}
}
