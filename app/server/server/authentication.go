package server

import (
	"crypto/sha256"
	"fmt"
	_os "will-moss/isaiah/server/_internal/os"
	_session "will-moss/isaiah/server/_internal/session"
	"will-moss/isaiah/server/ui"
)

type Authentication struct{}

func (Authentication) RunCommand(server *Server, session _session.GenericSession, command ui.Command) {
	switch command.Action {

	// Command : Authenticate the client by password
	case "auth.login":
		if _os.GetEnv("AUTHENTICATION_ENABLED") != "TRUE" {
			session.Set("authenticated", true)
			server.SendNotification(session, ui.NotificationAuth(ui.NP{
				Type: ui.TypeSuccess,
				Content: ui.JSON{
					"Authentication": ui.JSON{
						"Message": "Your are now authenticated",
					},
					"Preferences": server.GetPreferences(),
				},
			}))
			break
		}

		password := command.Args["Password"]

		// Authentication against raw password
		if _os.GetEnv("AUTHENTICATION_HASH") == "" {
			if password != _os.GetEnv("AUTHENTICATION_SECRET") {
				session.Set("authenticated", false)
				server.SendNotification(
					session,
					ui.NotificationAuth(ui.NP{
						Type: ui.TypeError,
						Content: ui.JSON{
							"Authentication": ui.JSON{
								"Message": "Invalid password",
							},
						},
					}),
				)
				break
			}
		}

		// Authentication against hashed password
		if _os.GetEnv("AUTHENTICATION_HASH") != "" {
			hasher := sha256.New()
			hasher.Write([]byte(password.(string)))
			hashed := fmt.Sprintf("%x", hasher.Sum(nil))

			if hashed != _os.GetEnv("AUTHENTICATION_HASH") {
				session.Set("authenticated", false)
				server.SendNotification(
					session,
					ui.NotificationAuth(ui.NP{
						Type: ui.TypeError,
						Content: ui.JSON{
							"Authentication": ui.JSON{
								"Message": "Invalid password",
							},
						},
					}),
				)
				break
			}
		}

		session.Set("authenticated", true)

		server.SendNotification(
			session,
			ui.NotificationAuth(ui.NP{
				Type: ui.TypeSuccess,
				Content: ui.JSON{
					"Authentication": ui.JSON{
						"Message": "Your are now authenticated",
					},
					"Preferences": server.GetPreferences(),
				},
			}),
		)

	// Command : Log out the client
	case "auth.logout":
		session.Set("authenticated", false)

	// Command not found
	default:
		server.SendNotification(
			session,
			ui.NotificationAuth(ui.NP{
				Type: ui.TypeError,
				Content: ui.JSON{
					"Authentication": ui.JSON{
						"Message": "You are not authenticated yet",
					},
				},
			}),
		)
	}
}
