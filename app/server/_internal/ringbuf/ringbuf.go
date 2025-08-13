package ringbuf

import "sync"

// RingBuffer provides a thread-safe circular buffer for storing integers.
// The 'head' field indicates the position in the buffer where the next element will be written.
// The 'count' field tracks the total number of elements ever added to the buffer, which helps
// distinguish between buffer states (e.g., empty, full, or partially filled) and supports
// operations that depend on the history of insertions, such as determining the oldest or newest
// elements.
type RingBuffer struct {
	data    []float64
	head    int
	count   uint64
	rbMutex sync.RWMutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data:  make([]float64, size),
		head:  0,
		count: 0,
	}
}

func (r *RingBuffer) Add(item float64) {
	r.rbMutex.Lock()
	r.data[r.head] = item
	r.head = (r.head + 1) % len(r.data)
	r.count++
	r.rbMutex.Unlock()
}

func (r *RingBuffer) GetFromCount(count uint64) ([]float64, uint64) {
	r.rbMutex.RLock()
	defer r.rbMutex.RUnlock()
	if count > r.count {
		return []float64{}, r.count
	} else if count < 0 {
		return []float64{}, r.count
	}

	itemsToRead := r.count - count
	// if from last time to now, we added more than whole array, head is made full circle
	if itemsToRead >= uint64(len(r.data)) {
		// To be safe from modifications of this slice
		out := append([]float64(nil), r.data[r.head:]...)
		out = append(out, r.data[:r.head]...)
		return out, r.count
	} else {
		// retrieve index of position when we had this count
		idx := int(count % uint64(len(r.data)))

		if idx <= r.head {
			out := append([]float64(nil), r.data[idx:r.head]...)
			return out, r.count
		} else {
			out := append([]float64(nil), r.data[idx:]...)
			out = append(out, r.data[:r.head]...)
			return out, r.count
		}
	}
}
