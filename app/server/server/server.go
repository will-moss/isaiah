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
	_slices "will-moss/isaiah/server/_internal/slices"
	_strconv "will-moss/isaiah/server/_internal/strconv"
	"will-moss/isaiah/server/_internal/tty"
	"will-moss/isaiah/server/resources"
	"will-moss/isaiah/server/ui"

	"github.com/docker/docker/api/types/filters"
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
	StatsManager    *resources.ContainerStatsManager
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

		tabs_enabled := strings.Split(strings.ToLower(_os.GetEnv("TABS_ENABLED")), ",")

		containers := resources.ContainersList(server.Docker, filters.Args{})
		images := resources.ImagesList(server.Docker)
		volumes := resources.VolumesList(server.Docker)
		networks := resources.NetworksList(server.Docker)
		stacks := resources.StacksList(server.Docker)
		agents := server.Agents.ToStrings()
		hosts := server.Hosts.ToStrings()

		if len(stacks) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_STACKS"), ",")
			rows := stacks.ToRows(columns)

			if slices.Contains(tabs_enabled, "stacks") {
				tabs = append(tabs, ui.Tab{Key: "stacks", Title: "Stacks", Rows: rows, SortBy: _os.GetEnv("SORTBY_STACKS")})
			}
		}

		if len(containers) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_CONTAINERS"), ",")
			rows := containers.ToRows(columns)

			if slices.Contains(tabs_enabled, "containers") {
				tabs = append(tabs, ui.Tab{Key: "containers", Title: "Containers", Rows: rows, SortBy: _os.GetEnv("SORTBY_CONTAINERS")})
			}
		}

		if len(images) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_IMAGES"), ",")
			rows := images.ToRows(columns)

			if slices.Contains(tabs_enabled, "images") {
				tabs = append(tabs, ui.Tab{Key: "images", Title: "Images", Rows: rows, SortBy: _os.GetEnv("SORTBY_IMAGES")})
			}
		}

		if len(volumes) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_VOLUMES"), ",")
			rows := volumes.ToRows(columns)

			if slices.Contains(tabs_enabled, "volumes") {
				tabs = append(tabs, ui.Tab{Key: "volumes", Title: "Volumes", Rows: rows, SortBy: _os.GetEnv("SORTBY_VOLUMES")})
			}
		}

		if len(networks) > 0 {
			columns := strings.Split(_os.GetEnv("COLUMNS_NETWORKS"), ",")
			rows := networks.ToRows(columns)

			if slices.Contains(tabs_enabled, "networks") {
				tabs = append(tabs, ui.Tab{Key: "networks", Title: "Networks", Rows: rows, SortBy: _os.GetEnv("SORTBY_NETWORKS")})
			}
		}

		// Default communication method - Send all at once
		if _os.GetEnv("SERVER_CHUNKED_COMMUNICATION_ENABLED") != "TRUE" {
			if command.Action == "init" {
				server.SendNotification(
					session,
					ui.NotificationInit(ui.NotificationParams{
						Content: ui.JSON{
							"Tabs":   tabs,
							"Agents": agents,
							"Hosts":  hosts,
						},
					}))
			} else if command.Action == "enumerate" {
				// `enumerate` is used only in the context of the `Jump` command
				server.SendNotification(
					session,
					ui.NotificationData(ui.NotificationParams{
						Content: ui.JSON{"Enumeration": tabs, "Host": command.Host},
					}))
			}
		} else {
			// Chunked communication method, send resources chunk by chunk
			chunkSize := int(_strconv.ParseInt(_os.GetEnv("SERVER_CHUNKED_COMMUNICATION_SIZE"), 10, 64))
			chunkIndex := 1
			if command.Action == "init" {
				// First, send the Agents and Hosts
				server.SendNotification(
					session,
					ui.NotificationInit(ui.NotificationParams{
						Content: ui.JSON{
							"Agents":     agents,
							"Hosts":      hosts,
							"ChunkIndex": -1,
						},
					}))

				// Then, send the resources by chunks
				for _, t := range tabs {
					chunks := _slices.Chunk(t.Rows, chunkSize)
					for _, c := range chunks {
						server.SendNotification(
							session,
							ui.NotificationInitChunk(ui.NotificationParams{
								Content: ui.JSON{
									"Tab":        ui.Tab{Key: t.Key, Title: t.Title, Rows: c, SortBy: t.SortBy},
									"ChunkIndex": chunkIndex,
								},
							}),
						)
						chunkIndex += 1
					}
				}
			} else if command.Action == "enumerate" {
				for _, t := range tabs {
					chunks := _slices.Chunk(t.Rows, chunkSize)
					for _, c := range chunks {
						server.SendNotification(
							session,
							ui.NotificationDataChunk(ui.NotificationParams{
								Content: ui.JSON{
									"Host":        command.Host,
									"Enumeration": ui.Tab{Key: t.Key, Title: t.Title, Rows: c, SortBy: t.SortBy},
									"ChunkIndex":  chunkIndex,
								},
							}),
						)
						chunkIndex += 1
					}
				}
			}
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
		if _os.GetEnv("DOCKER_RUNNING") == "TRUE" {
			server.SendNotification(
				session,
				ui.NotificationError(ui.NP{
					Content: ui.JSON{
						"Message": "It seems that you're running Isaiah inside a Docker container." +
							" In this case, opening a system shell isn't available because" +
							" Isaiah is bound to its container and it can't access the shell on your hosting system.",
					},
				}),
			)
			break
		}

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
					"Preferences": server.GetPreferences(),
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
		case strings.HasPrefix(command.Action, "stack"):
			h = Stacks{}
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

func (s *Server) GetPreferences() ui.Preferences {
	var preferences = make(ui.Preferences, 0)
	for k, v := range _os.GetFullEnv() {
		if strings.HasPrefix(k, "CLIENT_PREFERENCE_") {
			preferences[k] = v
		}
	}

	return preferences
}
