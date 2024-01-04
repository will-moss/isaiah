package json

import (
	"encoding/json"
)

// Alias for json.Marshal, without returning any error
func Marshal(v any) []byte {
	r, _ := json.Marshal(v)
	return r
}
