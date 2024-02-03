package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/olahol/melody"

	_client "will-moss/isaiah/server/_internal/client"
	_fs "will-moss/isaiah/server/_internal/fs"
	_json "will-moss/isaiah/server/_internal/json"
	_os "will-moss/isaiah/server/_internal/os"
	_session "will-moss/isaiah/server/_internal/session"
	_strconv "will-moss/isaiah/server/_internal/strconv"
	"will-moss/isaiah/server/_internal/tty"
	"will-moss/isaiah/server/server"
	"will-moss/isaiah/server/ui"
)

//go:embed client/*
var clientAssets embed.FS

//go:embed default.env
var defaultEnv string

// Perform checks to ensure the server is ready to start
// Returns an error if any condition isn't met
func performVerifications() error {

	// 1. Ensure Docker CLI is available
	if _os.GetEnv("DOCKER_RUNNING") != "TRUE" {
		cmd := exec.Command("docker", "version")
		_, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("Failed Verification : Access to Docker CLI -> %s", err)
		}
	}

	// 2. Ensure Docker socket is reachable
	if _os.GetEnv("MULTI_HOST_ENABLED") != "TRUE" {
		c, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			return fmt.Errorf("Failed Verification : Access to Docker socket -> %s", err)
		}
		defer c.Close()
	}

	// 3. Ensure server port is available
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", _os.GetEnv("SERVER_PORT")))
	if err != nil {
		return fmt.Errorf("Failed Verification : Port binding -> %s", err)
	}
	defer l.Close()

	// 4. Ensure certificate and private key are provided
	if _os.GetEnv("SSL_ENABLED") == "TRUE" {
		if _, err := os.Stat("./certificate.pem"); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("Failed Verification : Certificate file missing -> Please put your certificate.pem file next to the executable")
		}
		if _, err := os.Stat("./key.pem"); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("Failed Verification : Private key file missing -> Please put your key.pem file next to the executable")
		}
	}

	// 5. Ensure master node is available if current node is an agent
	if _os.GetEnv("SERVER_ROLE") == "Agent" {
		h, err := net.DialTimeout("tcp", _os.GetEnv("MASTER_HOST"), 5*time.Second)
		if err != nil {
			return fmt.Errorf("Failed Verification : Master node is unreachable -> %s", err)
		}
		defer h.Close()
	}

	// 6. Ensure an agent name is provided if current node is an agent
	if _os.GetEnv("SERVER_ROLE") == "Agent" {
		if _os.GetEnv("AGENT_NAME") == "" {
			return fmt.Errorf("Failed Verification : You must provide a name for your Agent node")
		}
	}

	// 7. Ensure docker_hosts file is available when multi-host is enabled
	if _os.GetEnv("MULTI_HOST_ENABLED") == "TRUE" {
		if _, err := os.Stat("docker_hosts"); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("Failed Verification : docker_hosts file is missing. Please put it next to the executable")
		}
	}

	// 8. Ensure every host is reachable if multi-host is enabled, and docker_hosts is well-formatted
	if _os.GetEnv("MULTI_HOST_ENABLED") == "TRUE" {
		raw, err := os.ReadFile("docker_hosts")
		if err != nil {
			return fmt.Errorf("Failed Verification : docker_hosts file can't be read -> %s", err)
		}
		if len(raw) == 0 {
			return fmt.Errorf("Failed Verification : docker_hosts file is empty.")
		}

		lines := strings.Split(string(raw), "\n")
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}

			parts := strings.Split(line, " ")
			if len(parts) != 2 {
				return fmt.Errorf("Failed Verification : docker_hosts file isn't properly formatted. Line : -> %s", line)
			}

			c, err := client.NewClientWithOpts(client.WithHost(parts[1]))
			if err != nil {
				return fmt.Errorf("Failed Verification : Access to Docker host -> %s", err)
			}

			_, err = c.Ping(context.Background())
			if err != nil {
				return fmt.Errorf("Failed Verification : Access to Docker host -> %s", err)
			}

			c.Close()
		}
	}

	return nil
}

// Entrypoint
func main() {
	// Load default settings via default.env file (workaround since the file is embed)
	defaultSettings, _ := godotenv.Unmarshal(defaultEnv)
	for k, v := range defaultSettings {
		if _os.GetEnv(k) == "" {
			os.Setenv(k, v)
		}
	}

	// Load custom settings via .env file
	err := godotenv.Overload(".env")
	if err != nil {
		log.Print("No .env file provided, will continue with system env")
	}

	if _os.GetEnv("MULTI_HOST_ENABLED") != "TRUE" {
		// Automatically discover the Docker host on the machine
		discoveredHost, err := _client.DiscoverDockerHost()
		if err != nil {
			log.Print(err.Error())
			return
		}
		os.Setenv("DOCKER_HOST", discoveredHost)
	}

	// Perform initial verifications
	if _os.GetEnv("SKIP_VERIFICATIONS") != "TRUE" {
		// Ensure everything is ready for our app
		log.Print("Performing verifications before starting")
		err = performVerifications()
		if err != nil {
			log.Print("Error performing initial verifications, abort\n")
			log.Print(err)
			return
		}
	}

	// Set up everything (Melody instance, Docker client, Server settings)
	var _server server.Server
	if _os.GetEnv("MULTI_HOST_ENABLED") != "TRUE" {
		_server = server.Server{
			Melody: melody.New(),
			Docker: _client.NewClientWithOpts(client.FromEnv),
		}
	} else {
		_server = server.Server{
			Melody: melody.New(),
		}

		// Populate server's known hosts when multi-host is enabled
		_server.Hosts = make(server.HostsArray, 0)
		var firstHost string

		raw, _ := os.ReadFile("docker_hosts")
		lines := strings.Split(string(raw), "\n")
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			parts := strings.Split(line, " ")

			_server.Hosts = append(_server.Hosts, []string{parts[0], parts[1]})

			if len(firstHost) == 0 {
				firstHost = parts[0]
			}
		}

		// Set default Docker client on the first known host
		_server.SetHost(firstHost)
	}
	_server.Melody.Config.MaxMessageSize = _strconv.ParseInt(_os.GetEnv("SERVER_MAX_READ_SIZE"), 10, 64)

	// Disable client when current node is an agent
	if _os.GetEnv("SERVER_ROLE") != "Agent" {

		// Load embed assets as a filesystem
		serverRoot := _fs.Sub(clientAssets, "client")

		// Set up static file serving for the CSS theming
		http.HandleFunc("/assets/css/custom.css", func(w http.ResponseWriter, r *http.Request) {
			if _, err := os.Stat("custom.css"); errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(200)
				return
			}

			http.ServeFile(w, r, "custom.css")
		})

		// Use on-disk assets rather than embedded ones when in development
		if _os.GetEnv("DEV_ENABLED") != "TRUE" {
			// Set up static file serving for all the front-end files
			http.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(serverRoot))))
		} else {
			http.Handle("/", http.FileServer(http.Dir("./client")))
		}
	}

	// Set up an endpoint to handle Websocket connections with Melody
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		_server.Melody.HandleRequest(w, r)
	})

	// WS - Handle first user connecion
	_server.Melody.HandleConnect(func(session *melody.Session) {
		session.Set("id", uuid.NewString())
		_server.Handle(session)
	})

	// WS - Handle user commands
	_server.Melody.HandleMessage(func(session *melody.Session, message []byte) {
		// go _server.Handle(session, message)
		_server.Handle(session, message)
	})

	// WS - Handle user disconnection
	_server.Melody.HandleDisconnect(func(s *melody.Session) {
		// When current node is master
		if _os.GetEnv("SERVER_ROLE") == "Master" {
			// Clear user tty if there's any open
			if terminal, exists := s.Get("tty"); exists {
				(terminal.(*tty.TTY)).ClearAndQuit()
				s.UnSet("tty")
			}

			// Clear user read stream if there's any open
			if stream, exists := s.Get("stream"); exists {
				(*stream.(*io.ReadCloser)).Close()
				s.UnSet("stream")
			}

			// Unregister the agent node if applicable
			if agent, exists := s.Get("agent"); exists {
				newAgents := make(server.AgentsArray, 0)
				for _, _agent := range _server.Agents {
					if (agent.(server.Agent)).Name != _agent.Name {
						newAgents = append(newAgents, _agent)
					}
				}
				_server.Agents = newAgents

				s.UnSet("agent")

				// Notify all the clients about the agent's disconnection
				notification := ui.NotificationData(ui.NotificationParams{Content: ui.JSON{"Agents": _server.Agents.ToStrings()}})
				_server.Melody.Broadcast(notification.ToBytes())
			}
		}

	})

	// When current node is an agent, perform agent registration procedure with the master node
	if _os.GetEnv("SERVER_ROLE") == "Agent" {
		log.Print("Initiating registration with master node")

		var response ui.Notification

		// 1. Establish connection with Master node
		masterAddress := url.URL{Scheme: "ws", Host: _os.GetEnv("MASTER_HOST"), Path: "/ws"}
		connection, _, err := websocket.DefaultDialer.Dial(masterAddress.String(), nil)
		if err != nil {
			log.Print("Error establishing connection to the master node")
			log.Print(err)
			return
		}

		if _os.GetEnv("MASTER_SECRET") != "" {
			log.Print("Performing authentication")

			// 2. Send authentication command
			authCommand := ui.Command{Action: "auth.login", Args: ui.JSON{"Password": _os.GetEnv("MASTER_SECRET")}}
			err = connection.WriteMessage(websocket.TextMessage, _json.Marshal(authCommand))
			if err != nil {
				log.Print("Error sending authentication command to the master node")
				log.Print(err)
				return
			}

			// 3. Verify that authentication was succesful
			err = connection.ReadJSON(&response)
			if err != nil {
				log.Print("Error decoding authentication response from the master node")
				log.Print(err)
				return
			}

			if response.Type != ui.TypeSuccess {
				log.Print("Authentication with master node unsuccesful")
				log.Print("Please check your MASTER_SECRET setting and restart")
				return
			}

			// Quirk : When authentication is disabled, the server has already initially sent an auth success
			//         Trying to empty / vaccuum the message queue proves unfeasible with Gorilla Websocket
			//         Hence we must undergo the following code to skip authentication in that case
			spontaneous, ok := response.Content["Authentication"].(map[string]interface{})["Spontaneous"]
			if ok && spontaneous.(bool) {
				connection.ReadMessage()
			}
		} else {
			log.Print("No authentication secret was provided, skipping authentication")
			// Quirk : Same as above
			connection.ReadMessage()
		}

		// 4. Send registration command
		registrationCommand := ui.Command{
			Action: "agent.register",
			Args: ui.JSON{
				"Resource": server.Agent{
					Name: _os.GetEnv("AGENT_NAME"),
				},
			},
		}
		err = connection.WriteMessage(websocket.TextMessage, _json.Marshal(registrationCommand))
		if err != nil {
			log.Print("Error sending registration command to the master node")
			log.Print(err)
			return
		}

		// Quirk : Skip loading indicator
		connection.ReadMessage()

		// 5. Verify that registration was succesful
		response = ui.Notification{}
		err = connection.ReadJSON(&response)
		if err != nil {
			log.Print("Error decoding registration response from the master node")
			log.Print(err)
			return
		}
		if response.Type != ui.TypeSuccess {
			log.Print("Registration with master node unsuccesful")
			log.Print("Please check your settings and connectivity")
			log.Printf("Error : %s", response.Content["Message"])
			return
		}

		log.Print("Connection with master node is established")

		// Workaround : Create a tweaked reimplementation of melody.Session to reuse existing code
		session := _session.Create(connection)

		// 6. Process the commands as they are received
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Print(err)
				break
			}

			_server.Handle(session, message)
		}

		// 7. Clear all opened TTY / Stream instances when applicable
		session.UnSet("initiator")

		// Clear all users' tty if there's any open
		for k := range session.Keys {
			if strings.HasSuffix(k, "tty") {
				(session.Keys[k].(*tty.TTY)).ClearAndQuit()
				session.UnSet(k)
			}
		}

		// Clear all users' read stream if there's any open
		for k := range session.Keys {
			if strings.HasSuffix(k, "stream") {
				(*session.Keys[k].(*io.ReadCloser)).Close()
				session.UnSet(k)
			}
		}

		return
	}

	// When current node is master, start the HTTP server
	if _os.GetEnv("SERVER_ROLE") == "Master" {
		log.Printf("Server starting on port %s", _os.GetEnv("SERVER_PORT"))
		if _os.GetEnv("SSL_ENABLED") == "TRUE" {
			http.ListenAndServeTLS(fmt.Sprintf(":%s", _os.GetEnv("SERVER_PORT")), "certificate.pem", "key.pem", nil)
		} else {
			http.ListenAndServe(fmt.Sprintf(":%s", _os.GetEnv("SERVER_PORT")), nil)
		}
	}
}
