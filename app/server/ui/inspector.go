package ui

type InspectorContent []InspectorContentPart

type InspectorContentPart struct {
	Type    string // One of "rows", "json", "table", "lines", "code", "plot"
	Content interface{}
}
