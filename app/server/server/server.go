package server

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	_io "will-moss/isaiah/server/_internal/io"
	_os "will-moss/isaiah/server/_internal/os"
	"will-moss/isaiah/server/_internal/tty"
	"will-moss/isaiah/server/resources"
	"will-moss/isaiah/server/ui"

	"github.com/docker/docker/client"
	"github.com/olahol/melody"
)

// Represent the current server
type Server struct {
	Melody *melody.Melody
	Docker *client.Client
}

// Represent a command handler, used only _internally
// to organize functions in files on a per-resource-type basis
type handler interface {
	RunCommand(*Server, *melody.Session, ui.Command)
}

// Primary method for sending messages via websocket
func (server *Server) send(session *melody.Session, message []byte) {
	session.Write(message)
}

// Send a notification
func (server *Server) SendNotification(session *melody.Session, notification ui.Notification) {
	// If configured, don't show confirmations
	if slices.Contains([]string{ui.TypeInfo, ui.TypeSuccess}, notification.Type) {
		notification.Display = _os.GetEnv("DISPLAY_CONFIRMATIONS") == "TRUE"
	}

	// By default, show errors and warnings
	if slices.Contains([]string{ui.TypeError, ui.TypeWarning}, notification.Type) {
		notification.Display = true
	}

	server.send(session, notification.ToBytes())
}

// Same as handler.RunCommand
func (server *Server) runCommand(session *melody.Session, command ui.Command) {
	switch command.Action {
	// Command : Initialization
	case "init":
		var tabs []ui.Tab

		containers := resources.ContainersList(server.Docker)
		images := resources.ImagesList(server.Docker)
		volumes := resources.VolumesList(server.Docker)
		networks := resources.NetworksList(server.Docker)

		if len(containers) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_CONTAINERS"), ",")
			rows := containers.ToRows(columns)
			tabs = append(tabs, ui.Tab{Key: "containers", Title: "Containers", Rows: rows})
		}

		if len(images) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_IMAGES"), ",")
			rows := images.ToRows(columns)
			tabs = append(tabs, ui.Tab{Key: "images", Title: "Images", Rows: rows})
		}

		if len(volumes) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_VOLUMES"), ",")
			rows := volumes.ToRows(columns)
			tabs = append(tabs, ui.Tab{Key: "volumes", Title: "Volumes", Rows: rows})
		}

		if len(networks) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_NETWORKS"), ",")
			rows := networks.ToRows(columns)
			tabs = append(tabs, ui.Tab{Key: "networks", Title: "Networks", Rows: rows})
		}

		server.SendNotification(session, ui.NotificationInit(ui.NotificationParams{Content: ui.JSON{"Tabs": tabs}}))

	// Command : Open shell on the server
	case "shell":
		terminal := tty.New(_io.CustomWriter{WriteFunction: func(p []byte) {
			server.SendNotification(
				session,
				ui.NotificationTty(ui.NotificationParams{Content: ui.JSON{"Output": string(p)}}),
			)
		}})
		session.Set("tty", &terminal)

		errs, updates, finished := make(chan error), make(chan string), false
		go _os.OpenShell(&terminal, errs, updates)

		for {
			if finished {
				break
			}

			select {
			case e := <-errs:
				server.SendNotification(session, ui.NotificationError(ui.NotificationParams{Content: ui.JSON{"Message": e.Error()}}))
			case u := <-updates:
				server.SendNotification(session, ui.NotificationTty(ui.NotificationParams{Content: ui.JSON{"Status": u, "Type": "system"}}))
				finished = u == "exited"
			}
		}

	// Command : Run a command inside the currently-opened shell (can be a container shell, or a system shell)
	case "shell.command":
		command := command.Args["Command"].(string)
		shouldQuit := command == "exit"
		terminal, exists := session.Get("tty")

		if exists != true {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": "No tty opened"}}))
			break
		}

		var err error
		if shouldQuit {
			(terminal.(*tty.TTY)).ClearAndQuit()
			session.UnSet("tty")
		} else {
			err = (terminal.(*tty.TTY)).RunCommand(command)
		}

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

	// Command : Not found
	default:
		server.SendNotification(
			session,
			ui.NotificationError(ui.NP{
				Content: ui.JSON{
					"Message": fmt.Sprintf("This command is unknown, unsupported, or not implemented yet : %s", command.Action),
				},
			}),
		)
	}
}

// Main function (dispatch a message to the appropriate handler, and run it)
func (server *Server) Handle(session *melody.Session, message ...[]byte) {
	// On first connection
	if len(message) == 0 {
		// Dev-only : If authentication is disabled
		//            - set client's session authenticated by default
		//            - send confirmation to the client
		if _os.GetEnv("AUTHENTICATION_ENABLED") != "TRUE" {
			session.Set("authenticated", true)
			server.SendNotification(session, ui.NotificationAuth(ui.NP{
				Type: ui.TypeSuccess,
				Content: ui.JSON{
					"Authentication": ui.JSON{
						"Spontaneous": true,
						"Message":     "Your are now authenticated",
					},
				},
			}),
			)
		}

		// Normal case : Do nothing
		return
	}

	// Decode the received command
	var command ui.Command
	err := json.Unmarshal(message[0], &command)

	if err != nil {
		server.SendNotification(session, ui.NotificationError(ui.NotificationParams{Content: ui.JSON{"Message": err.Error()}}))
		return
	}

	// By default, prior to running the command, close the current stream if any's still open
	if stream, exists := session.Get("stream"); exists {
		(*stream.(*io.ReadCloser)).Close()
		session.UnSet("stream")
	}

	// Dispatch the command to the appropriate handler
	var h handler
	if authenticated, _ := session.Get("authenticated"); authenticated != true ||
		strings.HasPrefix(command.Action, "auth") {
		h = Authentication{}
	} else {
		// Let the client know the server is processing their input
		server.SendNotification(session, ui.NotificationLoading())

		switch true {
		case strings.HasPrefix(command.Action, "image"):
			h = Images{}
		case strings.HasPrefix(command.Action, "container"):
			h = Containers{}
		case strings.HasPrefix(command.Action, "volume"):
			h = Volumes{}
		case strings.HasPrefix(command.Action, "network"):
			h = Networks{}
		default:
			h = nil
		}
	}

	if h != nil {
		h.RunCommand(server, session, command)
	} else {
		server.runCommand(session, command)
	}
}
