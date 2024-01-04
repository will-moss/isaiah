package ui

// Represent a command sent by the web browser
type Command struct {
	Action string
	Args   map[string]interface{}
}
