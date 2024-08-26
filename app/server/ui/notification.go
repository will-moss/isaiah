package ui

import (
	_json "will-moss/isaiah/server/_internal/json"
)

type JSON map[string]interface{}

// Represent a notification sent to the web browser
type Notification struct {
	Category string                 // The top-most category of notification
	Type     string                 // The type of notification (among success, error, warning, and info)
	Title    string                 // The title of the notification (as displayed to the end user)
	Content  map[string]interface{} // The content of the notification (JSON string)
	Follow   string                 // The command the client should run when they receive the notification
	Display  bool                   // Whether or not the notification should be shown to the end user
}

type NotificationParams struct {
	Content map[string]interface{}
	Follow  string
	Type    string
}

type NP = NotificationParams

const (
	TypeSuccess = "success"
	TypeError   = "error"
	TypeWarning = "warning"
	TypeInfo    = "info"
)

const (
	CategoryInit         = "init"          // Notification sent at first connection established
	CategoryInitChunk    = "init-chunk"    // Notification sent at first connection established, in chunked communication
	CategoryRefresh      = "refresh"       // Notification sent when requesting new data for Docker / UI resources
	CategoryRefreshChunk = "refresh-chunk" // Notification sent when requesting new data for Docker / UI resources, in chunked communication
	CategoryLoading      = "loading"       // Notification sent to let the user know that the server is loading
	CategoryReport       = "report"        // Notification sent to let the user know something (message, error)
	CategoryPrompt       = "prompt"        // Notification sent to ask confirmation from the user
	CategoryTty          = "tty"           // Notification sent to instruct about TTY status / output
	CategoryAuth         = "auth"          // Notification sent to instruct about authentication
)

func NotificationInit(p NotificationParams) Notification {
	return Notification{Category: CategoryInit, Type: TypeSuccess, Content: p.Content, Follow: p.Follow}
}

func NotificationInitChunk(p NotificationParams) Notification {
	return Notification{Category: CategoryInitChunk, Type: TypeSuccess, Content: p.Content, Follow: p.Follow}
}

func NotificationError(p NotificationParams) Notification {
	return Notification{Category: CategoryReport, Type: TypeError, Title: "Error", Content: p.Content, Follow: p.Follow}
}

func NotificationData(p NotificationParams) Notification {
	return Notification{Category: CategoryRefresh, Type: TypeInfo, Content: p.Content, Follow: p.Follow}
}
func NotificationDataChunk(p NotificationParams) Notification {
	return Notification{Category: CategoryRefreshChunk, Type: TypeInfo, Content: p.Content, Follow: p.Follow}
}

func NotificationInfo(p NotificationParams) Notification {
	return Notification{Category: CategoryReport, Type: TypeInfo, Title: "Information", Content: p.Content, Follow: p.Follow}
}

func NotificationSuccess(p NotificationParams) Notification {
	return Notification{Category: CategoryReport, Type: TypeSuccess, Title: "Success", Content: p.Content, Follow: p.Follow}
}

func NotificationPrompt(p NotificationParams) Notification {
	return Notification{Category: CategoryPrompt, Type: TypeInfo, Title: "Confirm", Content: p.Content}
}

func NotificationAuth(p NotificationParams) Notification {
	return Notification{Category: CategoryAuth, Type: p.Type, Title: "Authentication", Content: p.Content}
}

func NotificationTty(p NotificationParams) Notification {
	return Notification{Category: CategoryTty, Type: TypeInfo, Content: p.Content}
}

func NotificationLoading() Notification {
	return Notification{Category: CategoryLoading, Type: TypeInfo}
}

func (n Notification) ToBytes() []byte {
	v := _json.Marshal(n)
	return []byte(v)
}
