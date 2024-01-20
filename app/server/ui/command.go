package ui

import _json "will-moss/isaiah/server/_internal/json"

// Represent a command sent by the web browser
type Command struct {
	Action    string
	Args      map[string]interface{}
	Agent     string
	Initiator string
	Sequence  int32
}

func (c Command) ToBytes() []byte {
	v := _json.Marshal(c)
	return []byte(v)
}
