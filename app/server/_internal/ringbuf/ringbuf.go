package ringbuf

import "sync"

// RingBuffer provides a thread-safe circular buffer for Metric points.
// The 'head' field indicates the position in the buffer where the next element will be written.
// The 'count' field tracks the total number of elements ever added to the buffer, which helps
// distinguish between buffer states (e.g., empty, full, or partially filled) and supports
// operations that depend on the history of insertions, such as determining the oldest or newest
// elements.
type RingBuffer[T any] struct {
	data    []T
	head    int
	count   uint64
	rbMutex sync.RWMutex
}

func NewRingBuffer[T any](size int) *RingBuffer[T] {
	return &RingBuffer[T]{
		data:  make([]T, size),
		head:  0,
		count: 0,
	}
}

func (r *RingBuffer[T]) Add(item T) {
	r.rbMutex.Lock()
	r.data[r.head] = item
	r.head = (r.head + 1) % len(r.data)
	r.count++
	r.rbMutex.Unlock()
}

func (r *RingBuffer[T]) GetFromCount(count uint64) ([]T, uint64) {
	r.rbMutex.RLock()
	defer r.rbMutex.RUnlock()
	if count > r.count {
		return []T{}, r.count
	} else if count < 0 {
		return []T{}, r.count
	}

	itemsToRead := r.count - count
	// if from last time to now, we added more than whole array, head is made full circle
	if itemsToRead >= uint64(len(r.data)) {
		// To be safe from modifications of this slice
		out := append([]T{}, r.data[r.head:]...)
		out = append(out, r.data[:r.head]...)
		return out, r.count
	} else {
		// retrieve index of position when we had this count
		idx := int(count % uint64(len(r.data)))

		if idx <= r.head {
			out := append([]T{}, r.data[idx:r.head]...)
			return out, r.count
		} else {
			out := append([]T{}, r.data[idx:]...)
			out = append(out, r.data[:r.head]...)
			return out, r.count
		}
	}
}
