package ringbuf

import (
	"testing"
)


func TestRingBuffer_Table(t *testing.T) {
	tests := []struct {
		name         string
		size         int
		addItems     []int
		getCount     uint64
		wantValues   []int
		wantCount    uint64
	}{
		{
			name:       "on initial request",
			size:       5,
			addItems:   []int{10, 20, 30},
			getCount:   0,
			wantValues: []int{10, 20, 30},
			wantCount:  3,
		},
		{
			name:       "Add exactly buffer size",
			size:       4,
			addItems:   []int{1, 2, 3, 4},
			getCount:   0,
			wantValues: []int{1, 2, 3, 4},
			wantCount:  4,
		},
		{
			name:       "Add more than buffer size, wrap around",
			size:       3,
			addItems:   []int{5, 6, 7, 8},
			getCount:   0,
			wantValues: []int{6, 7, 8},
			wantCount:  4,
		},
		{
			name:       "Get from non-zero count, partial read",
			size:       5,
			addItems:   []int{1, 2, 3, 4, 5, 6},
			getCount:   3,
			wantValues: []int{4, 5, 6},
			wantCount:  6,
		},
		{
			name:       "Get after full wrap, count skips all",
			size:       3,
			addItems:   []int{1, 2, 3, 4, 5, 6},
			getCount:   6,
			wantValues: []int{},
			wantCount:  6,
		},
		{
			name:       "Get after wrap, count in middle",
			size:       4,
			addItems:   []int{1, 2, 3, 4, 5, 6},
			getCount:   2,
			wantValues: []int{3, 4, 5, 6},
			wantCount:  6,
		},
		{
			name:       "Get one before overwrite",
			size:       5,
			addItems:   []int{1, 2, 3, 4, 5, 6},
			getCount:   2,
			wantValues: []int{3, 4, 5, 6},
			wantCount:  6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRingBuffer(tt.size)
			for _, v := range tt.addItems {
				rb.Add(v)
			}
			gotValues, gotCount := rb.GetFromCount(tt.getCount)
			if !equalSlices(gotValues, tt.wantValues) {
				t.Errorf("GetFromCount(%d) = %v, want %v", tt.getCount, gotValues, tt.wantValues)
			}
			if gotCount != tt.wantCount {
				t.Errorf("GetFromCount(%d) count = %d, want %d", tt.getCount, gotCount, tt.wantCount)
			}
		})
	}
}

func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
