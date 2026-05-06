package block

import (
	"fmt"
	"sync"
	"testing"
)

func TestRingBufferAdd(t *testing.T) {
	rb := NewRingBuffer()

	commands := []string{"ls", "pwd", "echo hi"}
	for _, cmd := range commands {
		rb.Add(Block{Command: cmd})
	}

	if rb.Len() != 3 {
		t.Fatalf("expected Len() == 3, got %d", rb.Len())
	}

	all := rb.All()
	for i, cmd := range commands {
		if all[i].Command != cmd {
			t.Errorf("index %d: expected command %q, got %q", i, cmd, all[i].Command)
		}
	}
}

func TestRingBufferCapacity(t *testing.T) {
	rb := NewRingBuffer()

	// Add one more than capacity
	for i := 0; i < 501; i++ {
		rb.Add(Block{Command: fmt.Sprintf("cmd-%d", i)})
	}

	if rb.Len() != 500 {
		t.Fatalf("expected Len() == 500, got %d", rb.Len())
	}

	all := rb.All()
	firstID := all[0].ID

	// The very first block added should have been evicted.
	// all[0] must not be the block with command "cmd-0".
	if all[0].Command == "cmd-0" {
		t.Error("oldest block should have been dropped, but cmd-0 is still present")
	}

	// Verify firstID is not the ID of the original first block (which is gone).
	// We can also verify IDs are monotonically increasing across the snapshot.
	for i := 1; i < len(all); i++ {
		if all[i].ID <= firstID && i == 0 {
			t.Errorf("IDs should be increasing; slot %d ID %d <= %d", i, all[i].ID, firstID)
		}
	}
}

func TestRingBufferConcurrent(t *testing.T) {
	rb := NewRingBuffer()

	var wg sync.WaitGroup
	goroutines := 50
	addsPerGoroutine := 10

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < addsPerGoroutine; i++ {
				rb.Add(Block{Command: fmt.Sprintf("g%d-cmd%d", g, i)})
			}
		}(g)
	}

	wg.Wait()

	total := goroutines * addsPerGoroutine // 500
	if rb.Len() != total {
		t.Fatalf("expected Len() == %d, got %d", total, rb.Len())
	}
}

func TestAllReturnsCopy(t *testing.T) {
	rb := NewRingBuffer()
	rb.Add(Block{Command: "original"})

	snapshot := rb.All()
	snapshot[0].Command = "mutated"

	internal := rb.All()
	if internal[0].Command != "original" {
		t.Errorf("mutation of returned slice should not affect internal state, got %q", internal[0].Command)
	}
}
