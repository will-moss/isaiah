package main

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"github.com/olahol/melody"

	_client "will-moss/isaiah/server/_internal/client"
	_fs "will-moss/isaiah/server/_internal/fs"
	_os "will-moss/isaiah/server/_internal/os"
	_strconv "will-moss/isaiah/server/_internal/strconv"
	"will-moss/isaiah/server/_internal/tty"
	"will-moss/isaiah/server/server"
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
	c, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("Failed Verification : Access to Docker socket -> %s", err)
	}
	defer c.Close()

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

	return nil
}

// Entrypoint
func main() {
	// Automatically discover the Docker host on the machine
	discoveredHost, err := _client.DiscoverDockerHost()
	if err != nil {
		log.Print(err.Error())
		return
	}
	os.Setenv("DOCKER_HOST", discoveredHost)

	// Load default settings via default.env file (workaround since the file is embed)
	defaultSettings, _ := godotenv.Unmarshal(defaultEnv)
	for k, v := range defaultSettings {
		if _os.GetEnv(k) == "" {
			os.Setenv(k, v)
		}
	}

	// Load custom settings via .env file
	err = godotenv.Overload(".env")
	if err != nil {
		log.Print("No .env file provided, will continue with system env")
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
	server := server.Server{
		Melody: melody.New(),
		Docker: _client.NewClientWithOpts(client.FromEnv),
	}
	server.Melody.Config.MaxMessageSize = _strconv.ParseInt(_os.GetEnv("SERVER_MAX_READ_SIZE"), 10, 64)

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

	if _os.GetEnv("DEV_ENABLED") != "TRUE" {
		// Set up static file serving for all the front-end files
		http.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(serverRoot))))
	} else {
		http.Handle("/", http.FileServer(http.Dir("./client")))
	}

	// Set up an endpoint to handle Websocket connections with Melody
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.Melody.HandleRequest(w, r)
	})

	// WS - Handle first user connecion
	server.Melody.HandleConnect(func(session *melody.Session) {
		server.Handle(session)
	})

	// WS - Handle user commands
	server.Melody.HandleMessage(func(session *melody.Session, message []byte) {
		go server.Handle(session, message)
	})

	// WS - Handle user disconnection
	server.Melody.HandleDisconnect(func(s *melody.Session) {
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
	})

	log.Printf("Server starting on port %s", _os.GetEnv("SERVER_PORT"))

	// Start the server
	if _os.GetEnv("SSL_ENABLED") == "TRUE" {
		http.ListenAndServeTLS(fmt.Sprintf(":%s", _os.GetEnv("SERVER_PORT")), "certificate.pem", "key.pem", nil)
	} else {
		http.ListenAndServe(fmt.Sprintf(":%s", _os.GetEnv("SERVER_PORT")), nil)
	}
}
