package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"slices"
	"strings"
	_client "will-moss/isaiah/server/_internal/client"
	_io "will-moss/isaiah/server/_internal/io"
	_os "will-moss/isaiah/server/_internal/os"
	_session "will-moss/isaiah/server/_internal/session"
	"will-moss/isaiah/server/_internal/tty"
	"will-moss/isaiah/server/resources"
	"will-moss/isaiah/server/ui"

	"github.com/docker/docker/client"
	"github.com/olahol/melody"
)

// Represent the current server
type Server struct {
	Melody          *melody.Melody
	Docker          *client.Client
	Agents          AgentsArray
	Hosts           HostsArray
	CurrentHostName string
}

// Represent a command handler, used only _internally
// to organize functions in files on a per-resource-type basis
type handler interface {
	RunCommand(*Server, _session.GenericSession, ui.Command)
}

// Primary method for sending messages via websocket
func (server *Server) send(session _session.GenericSession, message []byte) {
	session.Write(message)
}

// Send a notification
func (server *Server) SendNotification(session _session.GenericSession, notification ui.Notification) {
	// If configured, don't show confirmations
	if slices.Contains([]string{ui.TypeInfo, ui.TypeSuccess}, notification.Type) {
		notification.Display = _os.GetEnv("DISPLAY_CONFIRMATIONS") == "TRUE"
	}

	// By default, show errors and warnings
	if slices.Contains([]string{ui.TypeError, ui.TypeWarning}, notification.Type) {
		notification.Display = true
	}

	// When current node is an agent, wrap the notification in a "agent.reply" command
	// and send that to the master node
	if _os.GetEnv("SERVER_ROLE") == "Agent" {
		initiator, _ := session.Get("initiator")

		command := ui.Command{
			Action: "agent.reply",
			Args: ui.JSON{
				"To":           initiator.(string),
				"Notification": notification,
			},
		}

		server.send(session, command.ToBytes())
	} else {
		// Default, when current node is master, simply send the notification
		server.send(session, notification.ToBytes())
	}

}

// Same as handler.RunCommand
func (server *Server) runCommand(session _session.GenericSession, command ui.Command) {
	switch command.Action {
	case "init", "enumerate":
		var tabs []ui.Tab

		containers := resources.ContainersList(server.Docker)
		images := resources.ImagesList(server.Docker)
		volumes := resources.VolumesList(server.Docker)
		networks := resources.NetworksList(server.Docker)
		agents := server.Agents.ToStrings()
		hosts := server.Hosts.ToStrings()

		if len(containers) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_CONTAINERS"), ",")
			rows := containers.ToRows(columns)
			tabs = append(tabs, ui.Tab{Key: "containers", Title: "Containers", Rows: rows, SortBy: _os.GetEnv("SORTBY_CONTAINERS")})
		}

		if len(images) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_IMAGES"), ",")
			rows := images.ToRows(columns)
			tabs = append(tabs, ui.Tab{Key: "images", Title: "Images", Rows: rows, SortBy: _os.GetEnv("SORTBY_IMAGES")})
		}

		if len(volumes) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_VOLUMES"), ",")
			rows := volumes.ToRows(columns)
			tabs = append(tabs, ui.Tab{Key: "volumes", Title: "Volumes", Rows: rows, SortBy: _os.GetEnv("SORTBY_VOLUMES")})
		}

		if len(networks) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_NETWORKS"), ",")
			rows := networks.ToRows(columns)
			tabs = append(tabs, ui.Tab{Key: "networks", Title: "Networks", Rows: rows, SortBy: _os.GetEnv("SORTBY_NETWORKS")})
		}

		if command.Action == "init" {
			server.SendNotification(
				session,
				ui.NotificationInit(ui.NotificationParams{
					Content: ui.JSON{"Tabs": tabs, "Agents": agents, "Hosts": hosts},
				}))
		} else if command.Action == "enumerate" {
			// `enumerate` is used only in the context of the `Jump` command
			server.SendNotification(
				session,
				ui.NotificationData(ui.NotificationParams{
					Content: ui.JSON{"Enumeration": tabs, "Host": command.Host},
				}))
		}

	// Command : Agent-only - Clear TTY / Stream
	case "clear":
		if _os.GetEnv("SERVER_ROLE") != "Agent" {
			break
		}

		// Clear user tty if there's any open
		if terminal, exists := session.Get("tty"); exists {
			(terminal.(*tty.TTY)).ClearAndQuit()
			session.UnSet("tty")
		}

		// Clear user read stream if there's any open
		if stream, exists := session.Get("stream"); exists {
			(*stream.(*io.ReadCloser)).Close()
			session.UnSet("stream")
		}

	// Command : Open shell on the server
	case "shell":
		terminal := tty.New(_io.CustomWriter{WriteFunction: func(p []byte) {
			server.SendNotification(
				session,
				ui.NotificationTty(ui.NotificationParams{Content: ui.JSON{"Output": string(p)}}),
			)

		}})
		session.Set("tty", &terminal)

		go func() {
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
		}()

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

	// Command : Get a global overview of the server and all other hosts / nodes
	case "overview":
		overview := ui.Overview{Instances: make(ui.OverviewInstanceArray, 0)}

		serverName := "Master"
		if _os.GetEnv("SERVER_ROLE") == "Agent" {
			serverName = _os.GetEnv("AGENT_NAME")
		}

		// Case when : Standalone
		if _os.GetEnv("MULTI_HOST_ENABLED") != "TRUE" && len(server.Agents) == 0 {
			dockerVersion, _ := server.Docker.ServerVersion(context.Background())
			instance := ui.OverviewInstance{
				Server: ui.OverviewServer{
					CountCPU:  runtime.NumCPU(),
					AmountRAM: _os.VirtualMemory().Total,
					Name:      serverName,
					Role:      _os.GetEnv("SERVER_ROLE"),
				},
				Docker: ui.OverviewDocker{
					Version: dockerVersion.Version,
					Host:    server.Docker.DaemonHost(),
				},
				Resources: ui.OverviewResources{
					Containers: ui.JSON{"Count": resources.ContainersCount(server.Docker)},
					Images:     ui.JSON{"Count": resources.ImagesCount(server.Docker)},
					Volumes:    ui.JSON{"Count": resources.VolumesCount(server.Docker)},
					Networks:   ui.JSON{"Count": resources.NetworksCount(server.Docker)},
				},
			}
			overview.Instances = append(overview.Instances, instance)
		} else if _os.GetEnv("MULTI_HOST_ENABLED") != "TRUE" && len(server.Agents) > 0 {
			// Case when : Multi-agent

			// First : Append current server
			dockerVersion, _ := server.Docker.ServerVersion(context.Background())
			instance := ui.OverviewInstance{
				Server: ui.OverviewServer{
					CountCPU:  runtime.NumCPU(),
					AmountRAM: _os.VirtualMemory().Total,
					Name:      serverName,
					Role:      _os.GetEnv("SERVER_ROLE"),
					Agents:    server.Agents.ToStrings(),
				},
				Docker: ui.OverviewDocker{
					Version: dockerVersion.Version,
					Host:    server.Docker.DaemonHost(),
				},
				Resources: ui.OverviewResources{
					Containers: ui.JSON{"Count": resources.ContainersCount(server.Docker)},
					Images:     ui.JSON{"Count": resources.ImagesCount(server.Docker)},
					Volumes:    ui.JSON{"Count": resources.VolumesCount(server.Docker)},
					Networks:   ui.JSON{"Count": resources.NetworksCount(server.Docker)},
				},
			}
			overview.Instances = append(overview.Instances, instance)

			// After : Do nothing more, the client will request an overview from each agent

		} else if _os.GetEnv("MULTI_HOST_ENABLED") == "TRUE" {
			// Case when : Multi-host
			originalHost := server.CurrentHostName
			for _, h := range server.Hosts {
				server.SetHost(h[0])

				dockerVersion, _ := server.Docker.ServerVersion(context.Background())
				instance := ui.OverviewInstance{
					Server: ui.OverviewServer{
						Name: h[0],
						Host: h[1],
						Role: "Master",
					},
					Docker: ui.OverviewDocker{
						Version: dockerVersion.Version,
						Host:    server.Docker.DaemonHost(),
					},
					Resources: ui.OverviewResources{
						Containers: ui.JSON{"Count": resources.ContainersCount(server.Docker)},
						Images:     ui.JSON{"Count": resources.ImagesCount(server.Docker)},
						Volumes:    ui.JSON{"Count": resources.VolumesCount(server.Docker)},
						Networks:   ui.JSON{"Count": resources.NetworksCount(server.Docker)},
					},
				}

				if strings.HasPrefix(h[1], "unix://") {
					instance.Server.CountCPU = runtime.NumCPU()
					instance.Server.AmountRAM = _os.VirtualMemory().Total
				}

				overview.Instances = append(overview.Instances, instance)
			}
			server.SetHost(originalHost)
		}

		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Overview": overview}}))

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
func (server *Server) Handle(session _session.GenericSession, message ...[]byte) {
	// Dev-only : Set authenticated by default if authentication is disabled
	if _os.GetEnv("AUTHENTICATION_ENABLED") != "TRUE" {
		session.Set("authenticated", true)
	}

	// On first connection
	if len(message) == 0 {
		// Dev-only : If authentication is disabled
		//            - send spontaneous auth confirmation to the client
		if _os.GetEnv("AUTHENTICATION_ENABLED") != "TRUE" {
			server.SendNotification(session, ui.NotificationAuth(ui.NP{
				Type: ui.TypeSuccess,
				Content: ui.JSON{
					"Authentication": ui.JSON{
						"Spontaneous": true,
						"Message":     "Your are now authenticated",
					},
				},
			}))
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

	if command.Action == "" {
		return
	}

	// If the command is meant to be forwarded to the final client, locally store the "initiator" field
	if _os.GetEnv("SERVER_ROLE") == "Agent" && command.Initiator != "" {
		session.Set("initiator", command.Initiator)

		// Set "authenticated" to true when authentication is disabled
		// Why once again? Because now, we have an "initiator" field, so "authenticated" is per-client
		if _os.GetEnv("AUTHENTICATION_ENABLED") != "TRUE" {
			session.Set("authenticated", true)
		}
	}

	// By default, prior to running any command, close the current stream if any's still open
	if stream, exists := session.Get("stream"); exists {
		(*stream.(*io.ReadCloser)).Close()
		session.UnSet("stream")
	}

	// If the command is meant to be run by an agent, forward it, no further action
	if _os.GetEnv("SERVER_ROLE") == "Master" && command.Agent != "" {
		allSessions, _ := server.Melody.Sessions()
		for index := range allSessions {
			s := allSessions[index]

			agent, ok := s.Get("agent")

			if !ok {
				continue
			}

			if agent.(Agent).Name != command.Agent {
				continue
			}

			clientId, _ := session.Get("id")

			// Remove Agent from the Command to prevent infinite forwarding
			command.Agent = ""

			// Append initial client's id to enable reverse response routing (from agent to initial client)
			command.Initiator = clientId.(string)

			// Send the command to the agent
			s.Write(command.ToBytes())

			break
		}

		// Let the client know the agent is processing their input
		if !strings.HasPrefix(command.Action, "auth") {
			server.SendNotification(session, ui.NotificationLoading())
		}
		return
	}

	// When multi-host is enabled, set the appropriate host before interacting with Docker
	if _os.GetEnv("MULTI_HOST_ENABLED") == "TRUE" {
		if command.Host != "" {
			server.SetHost(command.Host)
		}
	}

	// # - Dispatch the command to the appropriate handler
	var h handler

	if authenticated, _ := session.Get("authenticated"); authenticated != true ||
		strings.HasPrefix(command.Action, "auth") {
		h = Authentication{}
	} else {
		// Let the client know the server is processing their input
		// + Disable sending "loading" notifications for agent nodes, as Master does it already
		if _os.GetEnv("SERVER_ROLE") == "Master" {
			server.SendNotification(session, ui.NotificationLoading())
		}

		switch true {
		case strings.HasPrefix(command.Action, "image"):
			h = Images{}
		case strings.HasPrefix(command.Action, "container"):
			h = Containers{}
		case strings.HasPrefix(command.Action, "volume"):
			h = Volumes{}
		case strings.HasPrefix(command.Action, "network"):
			h = Networks{}
		case strings.HasPrefix(command.Action, "agent"):
			h = Agents{}
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

func (s *Server) SetHost(name string) {
	var correspondingHost []string
	for _, v := range s.Hosts {
		if v[0] == name {
			correspondingHost = v
			break
		}
	}

	s.Docker = _client.NewClientWithOpts(client.WithHost(correspondingHost[1]))
	s.CurrentHostName = name
}
