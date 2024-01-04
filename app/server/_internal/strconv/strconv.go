package strconv

import "strconv"

// Alias of strconv.ParseInt, without returning any error
func ParseInt(s string, base int, bitSize int) int64 {
	i, _ := strconv.ParseInt(s, base, bitSize)
	return i
}
