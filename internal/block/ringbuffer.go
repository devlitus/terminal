package block

import (
	"sync"
	"sync/atomic"
)

const capacity = 500

var nextID atomic.Uint64

// RingBuffer is a fixed-capacity circular buffer of Blocks.
// Oldest entries are overwritten when the buffer is full.
type RingBuffer struct {
	mu   sync.RWMutex
	buf  [capacity]Block
	head int // next write position
	size int
}

// NewRingBuffer returns an empty RingBuffer with capacity 500.
func NewRingBuffer() *RingBuffer {
	return &RingBuffer{}
}

// Add assigns an ID to b, then writes it into the buffer.
// When full, the oldest entry is overwritten.
func (r *RingBuffer) Add(b Block) {
	b.ID = nextID.Add(1)
	r.mu.Lock()
	r.buf[r.head] = b
	r.head = (r.head + 1) % capacity
	if r.size < capacity {
		r.size++
	}
	r.mu.Unlock()
}

// All returns a copy of all blocks in insertion order (oldest first).
func (r *RingBuffer) All() []Block {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Block, r.size)
	if r.size < capacity {
		// buffer not yet full: data starts at index 0
		copy(out, r.buf[:r.size])
	} else {
		// buffer is full: oldest entry is at head
		n := copy(out, r.buf[r.head:])
		copy(out[n:], r.buf[:r.head])
	}
	return out
}

// Len returns the number of blocks currently stored.
func (r *RingBuffer) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.size
}
