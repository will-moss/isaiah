package ui

// Represent a menu action (row) in the web browser
type MenuAction struct {
	Label            string
	Command          string
	Prompt           string
	Key              string
	RequiresResource bool
	RunLocally       bool
}
