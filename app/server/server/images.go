package server

import (
	"encoding/json"
	"fmt"
	"strings"
	_os "will-moss/isaiah/server/_internal/os"
	"will-moss/isaiah/server/_internal/process"
	_session "will-moss/isaiah/server/_internal/session"
	"will-moss/isaiah/server/resources"
	"will-moss/isaiah/server/ui"

	"github.com/mitchellh/mapstructure"
)

// Placeholder used for internal organization
type Images struct{}

func (Images) RunCommand(server *Server, session _session.GenericSession, command ui.Command) {
	switch command.Action {

	// Single - Default menu
	case "image.menu":
		actions := resources.ImageSingleActions()
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Single - Remove menu
	case "image.menu.remove":
		var volume resources.Volume
		mapstructure.Decode(command.Args["Resource"], &volume)

		actions := resources.ImageRemoveActions(volume)
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - Bulk menu
	case "images.bulk":
		actions := resources.ImagesBulkActions()
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - List
	case "images.list":
		columns := strings.Split(_os.GetEnv("COLUMNS_IMAGES"), ",")
		images := resources.ImagesList(server.Docker)

		rows := images.ToRows(columns)
		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{"Tab": ui.Tab{Key: "images", Title: "Images", Rows: rows}},
			}),
		)

	// Bulk - Prune
	case "images.prune":
		err := resources.ImagesPrune(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "All the unused images were pruned"}, Follow: "images.list",
			}),
		)

	// Bulk - Pull
	case "images.pull":
		images := resources.ImagesList(server.Docker)

		for _, image := range images {
			if image.Version != "latest" {
				continue
			}

			task := process.LongTask{
				Function: resources.ImagePull,
				Args:     map[string]interface{}{"Image": image.Name},
				OnStep: func(update string) {
					metadata := make(map[string]string)
					json.Unmarshal([]byte(update), &metadata)

					message := fmt.Sprintf("Pulling : %s", image.Name)
					message += fmt.Sprintf("<br />Status : %s", metadata["status"])
					if _, ok := metadata["progress"]; ok {
						message += fmt.Sprintf("<br />Progress : %s", metadata["progress"])
					}

					server.SendNotification(
						session,
						ui.NotificationInfo(ui.NP{
							Content: ui.JSON{
								"Message": message,
							},
						}),
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
							Content: ui.JSON{"Message": fmt.Sprintf("The image %s was succesfully pulled", image.Name)}, Follow: "images.list",
						}),
					)
				},
			}
			task.RunSync(server.Docker)
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "All your latest image were succesfully pulled"}, Follow: "images.list",
			}),
		)

	// Single - Default remove
	case "image.remove.default":
		var image resources.Image
		mapstructure.Decode(command.Args["Resource"], &image)

		err := image.Remove(server.Docker, false, true)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The image was succesfully removed"}, Follow: "images.list",
			}),
		)

	// Single - Default remove without deleting untagged parents
	case "image.remove.default.unprune":
		var image resources.Image
		mapstructure.Decode(command.Args["Resource"], &image)

		err := image.Remove(server.Docker, false, false)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The image was succesfully removed"}, Follow: "images.list",
			}),
		)

	// Single - Force remove
	case "image.remove.force":
		var image resources.Image
		mapstructure.Decode(command.Args["Resource"], &image)

		err := image.Remove(server.Docker, true, true)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The image was succesfully removed"}, Follow: "images.list",
			}),
		)

	// Single - Force remove without deleting untagged parents
	case "image.remove.force.unprune":
		var image resources.Image
		mapstructure.Decode(command.Args["Resource"], &image)

		err := image.Remove(server.Docker, true, false)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The image was succesfully removed"}, Follow: "images.list",
			}),
		)

	// Single - Pull
	case "image.pull":
		task := process.LongTask{
			Function: resources.ImagePull,
			Args:     command.Args, // Expects : { "Image": <string> }
			OnStep: func(update string) {
				metadata := make(map[string]string)
				json.Unmarshal([]byte(update), &metadata)

				message := fmt.Sprintf("Pulling : %s", command.Args["Image"])
				message += fmt.Sprintf("<br />Status : %s", metadata["status"])
				if _, ok := metadata["progress"]; ok {
					message += fmt.Sprintf("<br />Progress : %s", metadata["progress"])
				}

				server.SendNotification(
					session,
					ui.NotificationInfo(ui.NP{
						Content: ui.JSON{
							"Message": message,
						},
					}),
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
						Content: ui.JSON{"Message": "The image was succesfully pulled"}, Follow: "images.list",
					}),
				)
			},
		}
		task.RunSync(server.Docker)

	// Single - Get inspector tabs
	case "image.inspect.tabs":
		tabs := resources.ImagesInspectorTabs()
		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{"Inspector": ui.JSON{"Tabs": tabs}},
			}),
		)

	// Single - Inspect full configuration
	case "image.inspect.config":
		var image resources.Image
		mapstructure.Decode(command.Args["Resource"], &image)
		config, err := image.GetConfig(server.Docker)

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

	// Single - Run
	case "image.run":
		var image resources.Image
		mapstructure.Decode(command.Args["Resource"], &image)

		var name string
		name = command.Args["Name"].(string)

		err := image.Run(server.Docker, name)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The image was succesfully used to run a new container"}, Follow: "containers.list",
			}),
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
