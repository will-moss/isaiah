package server

import (
	"context"
	"fmt"
	"io"
	"strings"
	_io "will-moss/isaiah/server/_internal/io"
	_os "will-moss/isaiah/server/_internal/os"
	"will-moss/isaiah/server/_internal/process"
	_session "will-moss/isaiah/server/_internal/session"
	_slices "will-moss/isaiah/server/_internal/slices"
	_strconv "will-moss/isaiah/server/_internal/strconv"
	"will-moss/isaiah/server/_internal/tty"
	"will-moss/isaiah/server/resources"
	"will-moss/isaiah/server/ui"

	"github.com/docker/docker/api/types/filters"
	"github.com/mitchellh/mapstructure"
)

// Placeholder used for internal organization
type Containers struct{}

func (Containers) RunCommand(server *Server, session _session.GenericSession, command ui.Command) {
	switch command.Action {

	// Single - Default menu
	case "container.menu":
		actions := resources.ContainerSingleActions()
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Single - Remove menu
	case "container.menu.remove":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		actions := resources.ContainerRemoveActions(container)
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - Bulk menu
	case "containers.bulk":
		actions := resources.ContainersBulkActions()
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - List
	case "containers.list":
		columns := strings.Split(_os.GetEnv("COLUMNS_CONTAINERS"), ",")
		containers := resources.ContainersList(server.Docker, filters.Args{})

		rows := containers.ToRows(columns)

		// Default communication method - Send all at once
		if _os.GetEnv("SERVER_CHUNKED_COMMUNICATION_ENABLED") != "TRUE" {
			server.SendNotification(
				session,
				ui.NotificationData(ui.NP{
					Content: ui.JSON{"Tab": ui.Tab{Key: "containers", Title: "Containers", Rows: rows, SortBy: _os.GetEnv("SORTBY_CONTAINERS")}}}),
			)
		} else {
			// Chunked communication method, send resources chunk by chunk
			chunkSize := int(_strconv.ParseInt(_os.GetEnv("SERVER_CHUNKED_COMMUNICATION_SIZE"), 10, 64))
			chunkIndex := 1
			chunks := _slices.Chunk(rows, chunkSize)
			for _, c := range chunks {
				server.SendNotification(
					session,
					ui.NotificationDataChunk(ui.NP{
						Content: ui.JSON{
							"Tab":        ui.Tab{Key: "containers", Title: "Containers", Rows: c, SortBy: _os.GetEnv("SORTBY_CONTAINERS")},
							"ChunkIndex": chunkIndex,
						},
					}),
				)
				chunkIndex += 1
			}
		}

	// Bulk - Prune
	case "containers.prune":
		err := resources.ContainersPrune(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "All the unused containers were pruned"}, Follow: "containers.list",
			}),
		)

	// Bulk - Stop
	case "containers.stop":
		task := process.LongTask{
			Function: resources.ContainersStop,
			OnStep: func(id string) {
				server.SendNotification(
					session,
					ui.NotificationInfo(ui.NP{Content: ui.JSON{"Message": fmt.Sprintf("Container %s was stopped", id)}}),
				)
				server.SendNotification(
					session,
					ui.NotificationLoading(),
				)
			},
			OnError: func(err error) {
				server.SendNotification(
					session,
					ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}),
				)
			},
			OnDone: func() {
				server.SendNotification(
					session,
					ui.NotificationSuccess(ui.NP{
						Content: ui.JSON{"Message": "All the containers were stopped"}, Follow: "containers.list",
					}),
				)
			},
		}
		task.RunSync(server.Docker)

	// Bulk - Update
	case "containers.update":
		task := process.LongTask{
			Function: resources.ContainersUpdate,
			OnStep: func(id string) {
				server.SendNotification(
					session,
					ui.NotificationInfo(ui.NP{Content: ui.JSON{"Message": fmt.Sprintf("Container %s was updated", id)}}),
				)
				server.SendNotification(
					session,
					ui.NotificationLoading(),
				)
			},
			OnError: func(err error) {
				server.SendNotification(
					session,
					ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}),
				)
			},
			OnDone: func() {
				server.SendNotification(
					session,
					ui.NotificationSuccess(ui.NP{
						Content: ui.JSON{"Message": "All the containers were updated"}, Follow: "containers.list",
					}),
				)
			},
		}
		task.RunSync(server.Docker)

	// Bulk - Restart
	case "containers.restart":
		task := process.LongTask{
			Function: resources.ContainersRestart,
			OnStep: func(id string) {
				server.SendNotification(
					session,
					ui.NotificationInfo(ui.NP{Content: ui.JSON{"Message": fmt.Sprintf("Container %s was restarted", id)}}),
				)
				server.SendNotification(
					session,
					ui.NotificationLoading(),
				)
			},
			OnError: func(err error) {
				server.SendNotification(
					session,
					ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}),
				)
			},
			OnDone: func() {
				server.SendNotification(
					session,
					ui.NotificationSuccess(ui.NP{
						Content: ui.JSON{"Message": "All the containers were restarted"}, Follow: "containers.list",
					}),
				)
			},
		}
		task.RunSync(server.Docker)

	// Bulk - Remove
	case "containers.remove":
		err := resources.ContainersRemove(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "All the containers were removed"}, Follow: "containers.list",
			}),
		)

	// Single - Pause/Unpause
	case "container.pause":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		information, err := container.Inspect(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		var newState string
		if information.State.Paused {
			err = container.Unpause(server.Docker)
			newState = "unpaused"
		} else {
			err = container.Pause(server.Docker)
			newState = "paused"
		}

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": fmt.Sprintf("The container was succesfully %s", newState)}, Follow: "containers.list",
			}),
		)

	// Single - Stop
	case "container.stop":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		err := container.Stop(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully stopped"}, Follow: "containers.list",
			}),
		)

	// Single - Restart
	case "container.restart":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		err := container.Restart(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully restarted"}, Follow: "containers.list",
			}),
		)

	// Single - Default remove
	case "container.remove.default":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		information, err := container.Inspect(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		if information.State.Running {
			server.SendNotification(
				session,
				ui.NotificationPrompt(ui.NP{
					Content: ui.JSON{
						"Message": "You cannot remove a container unless you force it. Do you want to force it?",
						"Command": "container.remove.force",
					},
				}),
			)
			break
		}

		err = container.Remove(server.Docker, false, false)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully removed"}, Follow: "containers.list",
			}),
		)

	// Single - Force remove
	case "container.remove.force":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		err := container.Remove(server.Docker, true, false)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully removed"}, Follow: "containers.list",
			}),
		)

	// Single - Default remove with volumes
	case "container.remove.default.volumes":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		information, err := container.Inspect(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		if information.State.Running {
			server.SendNotification(
				session,
				ui.NotificationPrompt(ui.NP{
					Content: ui.JSON{
						"Message": "You cannot remove a container unless you force it. Do you want to force it?",
						"Command": "container.remove.force.volumes",
					},
				}),
			)
			break
		}

		err = container.Remove(server.Docker, false, true)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully removed"}, Follow: "containers.list",
			}),
		)

	// Single - Force remove with volumes
	case "container.remove.force.volumes":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		err := container.Remove(server.Docker, true, true)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully removed"}, Follow: "containers.list",
			}),
		)

	// Single - Open shell inside container
	case "container.shell":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		terminal := tty.New(&_io.CustomWriter{WriteFunction: func(p []byte) {
			server.SendNotification(
				session,
				ui.NotificationTty(ui.NP{Content: ui.JSON{"Output": string(p)}}),
			)
		}})
		session.Set("tty", &terminal)

		go func() {
			errs, updates, finished := make(chan error), make(chan string), false
			go container.Shell(server.Docker, &terminal, errs, updates)

			for {
				if finished {
					break
				}

				select {
				case e := <-errs:
					server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": e.Error()}}))
				case u := <-updates:
					server.SendNotification(session, ui.NotificationTty(ui.NP{Content: ui.JSON{"Status": u, "Type": "container"}}))
					finished = u == "exited"
				}

			}
		}()

	// Single - Open in browser
	case "container.browser":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		address, err := container.GetBrowserUrl(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Address": address}}))

	// Single - Rename
	case "container.rename":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		err := container.Rename(server.Docker, command.Args["Name"].(string))

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "Your container was succesfully renamed"},
				Follow:  "containers.list",
			}))

	// Single - Update
	case "container.update":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		if container.IsIsaiah() {
			server.SendNotification(
				session,
				ui.NotificationError(ui.NP{
					Content: ui.JSON{
						"Message": "It seems that you're attempting to update Isaiah from Isaiah itself." +
							" For now, this is not supported. You may rather try to pull the latest image" +
							" and restart the container, or use a 3rd party tool such as WatchTower.",
					},
				}),
			)
			break
		}

		err := container.Update(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "Your container was succesfully updated"},
				Follow:  "containers.list",
			}))

	// Single - Retrieve run command to edit client-side
	case "container.edit.prepare":
		if _os.GetEnv("DOCKER_RUNNING") == "TRUE" {
			server.SendNotification(
				session,
				ui.NotificationError(ui.NP{
					Content: ui.JSON{
						"Message": "It seems that you're running Isaiah inside a Docker container." +
							" In this case, editing containers is unavailable because" +
							" Isaiah is bound to its container and it can't run commands on your hosting system.",
					},
				}),
			)
			break
		}

		if _os.GetEnv("MULTI_HOST_ENABLED") == "TRUE" && !strings.HasPrefix(server.Docker.DaemonHost(), "unix://") {
			server.SendNotification(
				session,
				ui.NotificationError(ui.NP{
					Content: ui.JSON{
						"Message": "It seems that you're running Isaiah inside a multi-host deployment." +
							" In this case, editing a running container is unavailable because" +
							" it requires accessing files on the remote host, which isn't feasible over the raw Docker socket." +
							" You may want to deploy a multi-node setup for that purpose.",
					},
				}),
			)
			return
		}

		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		_command, err := container.GetRunCommand(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationPrompt(ui.NP{
				Content: ui.JSON{
					"RunLocalCommand": true,
					"Input": ui.JSON{
						"Name":         "Edit container",
						"DefaultValue": _command,
						"Type":         "textarea",
						"Placeholder":  "Please fill in the content of your updated docker run command",
					},
					"Command": "_editContainer",
				},
			}),
		)

	// Single - Edit a container (down, and new run command)
	case "container.edit":
		if _os.GetEnv("DOCKER_RUNNING") == "TRUE" {
			server.SendNotification(
				session,
				ui.NotificationError(ui.NP{
					Content: ui.JSON{
						"Message": "It seems that you're running Isaiah inside a Docker container." +
							" In this case, editing containers is unavailable because" +
							" Isaiah is bound to its container and it can't run commands on your hosting system.",
					},
				}),
			)
			break
		}

		if _os.GetEnv("MULTI_HOST_ENABLED") == "TRUE" && !strings.HasPrefix(server.Docker.DaemonHost(), "unix://") {
			server.SendNotification(
				session,
				ui.NotificationError(ui.NP{
					Content: ui.JSON{
						"Message": "It seems that you're running Isaiah inside a multi-host deployment." +
							" In this case, editing a running container is unavailable because" +
							" it requires accessing files on the remote host, which isn't feasible over the raw Docker socket." +
							" You may want to deploy a multi-node setup for that purpose.",
					},
				}),
			)
			return
		}

		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		newCommand := command.Args["Content"].(string)
		if !strings.HasPrefix(newCommand, "docker run") {
			server.SendNotification(
				session,
				ui.NotificationError(ui.NP{
					Content: ui.JSON{
						"Message": "For your own security, you can only run a \"docker run\" command." +
							" Please make sure that your command starts, indeed, with \"docker run\"",
					},
				}),
			)
			break
		}

		task := process.LongTask{
			Function: container.Edit,
			Args:     command.Args, // Expects : { "Content": <string> }
			OnStep: func(update string) {
				server.SendNotification(
					session,
					ui.NotificationInfo(ui.NP{
						Content: ui.JSON{
							"Message": update,
						},
					}),
				)
			},
			OnError: func(err error) {
				server.SendNotification(
					session,
					ui.NotificationSuccess(ui.NP{
						Content: ui.JSON{"Message": fmt.Sprintf("Error: %s", err.Error())}, Follow: "containers.list",
					}),
				)
			},
			OnDone: func() {
				server.SendNotification(
					session,
					ui.NotificationSuccess(ui.NP{
						Content: ui.JSON{"Message": "Your container was succesfully edited (down, up with new command)"}, Follow: "containers.list",
					}),
				)
			},
		}
		task.RunSync(server.Docker)

	// Single - Get inspector tabs
	case "container.inspect.tabs":
		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{"Inspector": ui.JSON{"Tabs": resources.ContainersInspectorTabs()}},
			}),
		)

	// Single - Inspect logs
	case "container.inspect.logs":
		var showTimestamps = command.Args["showTimestamps"].(bool)
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		stream, err := container.GetLogs(
			server.Docker,
			_io.CustomWriter{WriteFunction: func(p []byte) {
				server.SendNotification(
					session,
					ui.NotificationData(ui.NP{
						Content: ui.JSON{
							"Inspector": ui.JSON{
								"Content": ui.InspectorContent{
									ui.InspectorContentPart{Type: "lines", Content: []string{string(p)}},
								},
							},
						},
					}),
				)
			}},
			showTimestamps,
		)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			if _stream, exists := session.Get("stream"); exists {
				(*_stream.(*io.ReadCloser)).Close()
				session.UnSet("stream")
			}
			break
		}

		session.Set("stream", stream)

		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{Content: ui.JSON{}}),
		)

	// Single - Inspect full configuration
	case "container.inspect.config":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		config, err := container.GetConfig(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{
					"Inspector": ui.JSON{
						"Content": config,
					},
				},
			}),
		)

	// Single - Inspect top (running processes)
	case "container.inspect.top":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		processes, err := container.GetTop(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{
					"Inspector": ui.JSON{
						"Content": ui.InspectorContent{
							ui.InspectorContentPart{Type: "table", Content: processes},
						},
					},
				},
			}),
		)

	// Single - Inspect environment variables
	case "container.inspect.env":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		env, err := container.GetEnv(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{
					"Inspector": ui.JSON{
						"Content": ui.InspectorContent{
							ui.InspectorContentPart{Type: "rows", Content: env},
						},
					},
				},
			}),
		)

	// Single - Inspect stats
	case "container.inspect.stats":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		stats, err := container.GetStats(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{
					"Inspector": ui.JSON{
						"Content": stats,
					},
				},
			}),
		)

    case "container.metrics":
		var container resources.Container
        err := mapstructure.Decode(command.Args["Resource"], &container)

        if err != nil{
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
            break
        }

        from, ok := command.Args["From"]

        if !ok{
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": "missing container.metrics mandatory argument \"from\""}}))
            break
        }
        
        idx, ok := from.(float64)
        from_idx := int(idx)


        if !ok{
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": "container.metrics \"from\" argument must be integer "}}))
            break
        }

        // Spawn a new metrics poller, if we don't have one yet
        if !container.IsMetricsPolling(){
            errChan := make(chan  error, 1)
            // Link all metrics poller in one session with one context
            ctxVal, exists := session.Get("metrics-context")
            var ctx context.Context

            if exists{
                ctx = ctxVal.(context.Context)
            } else{
                var cancel context.CancelFunc
                ctx, cancel = context.WithCancel(context.Background())
                session.Set("metrics-context", ctx)
                session.Set("metrics-context-cancel", cancel)
            }

            go container.PollMetrics(server.Docker, ctx, errChan)

            go func() {
                select{
                case e := <- errChan:
                    // based on type of error send correct notification
                    server.SendNotification(
                        session, 
                        ui.NotificationInfo(ui.NP{
                            Content: ui.JSON{
                                "Message": e.Error(),
                            },
                        },
                        ),
                    )
                }
            }()
        }

        metrics := container.GetMetricsFrom(from_idx)
        server.SendNotification(
            session,
            ui.NotificationData(ui.NP{
                Content: ui.JSON{"container.metrics": metrics}}),
        )

	// Command not found
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
