<p align="center">
<h1 align="center">Isaiah</h1>
<p align="center">Self-hostable clone of lazydocker for the web</p>
</p>

| | | |
|:-------------------------:|:-------------------------:|:-------------------------:|
|<img width="1604" src="/assets/CAPTURE-1.png"/> |  <img width="1604" src="/assets/CAPTURE-2.png"/> | <img width="1604" src="/assets/CAPTURE-3.png" /> |
|<img width="1604" src="/assets/CAPTURE-4.png"/> |  <img width="1604" src="/assets/CAPTURE-5.png"/> | <img width="1604" src="/assets/CAPTURE-6.png" /> |
|<img width="1604" src="/assets/CAPTURE-7.png"/> |  <img width="1604" src="/assets/CAPTURE-8.png"/> | <img width="1604" src="/assets/CAPTURE-9.png" /> |
|<img width="1604" src="/assets/CAPTURE-10.png"/> |  <img width="1604" src="/assets/CAPTURE-11.png"/> | <img width="1604" src="/assets/CAPTURE-12.png"/> |


## Introduction

Isaiah is a self-hostable service that enables you to manage all your Docker resources on a remote server. It is an attempt at recreating the [lazydocker](https://github.com/jesseduffield/lazydocker) command-line application from scratch, while making it available as a web application without compromising on the features.


## Features

Isaiah has all these features implemented :
- For containers :
    - Bulk stop, Bulk remove, Prune
    - Remove, Pause, Unpause, Restart, Open in browser
    - Open a shell inside the container (from your browser)
    - Inspect (live logs, stats, env, full configuration, top)
- For images :
    - Prune
    - Remove
    - Run (create and start a container using the image)
    - Open on Docker Hub
    - Pull a new image (from Docker Hub)
    - Inspect (full configuration, layers)
- For volumes :
    - Prune
    - Remove
    - Browse volume files (from your browser, via shell)
    - Inspect (full configuration)
- For networks :
    - Prune
    - Remove
    - Inspect (full configuration)
- Built-in automatic Docker host discovery
- Built-in authentication by master password
- Built-in terminal emulator (with support for opening a shell on the server)
- Support for multiple layouts
- Support for color-theming
- Support for keyboard navigation
- Support for mouse navigation
- Support for custom Docker Host / Context.
- Support for extensive configuration with `.env`
- Support for HTTP and HTTPS
- Support for standalone / proxy deployment

On top of these, one may appreciate the following characteristics :
- Written in Go (for the server) and Vanilla JS (for the client)
- Holds in a ~4 MB single file executable
- Holds in a ~4 MB Docker image
- Works exclusively over Websocket, with very little bandwidth usage
- Uses the official Docker SDK for 100% of the Docker features

For more information, read about [Configuration](#configuration) and [Deployment](#deployment-and-examples).


## Deployment and Examples


### Deploy with Docker

You can run Isaiah with Docker on the command line very quickly.

You can use the following commands :

```sh
# Create a .env file
touch .env

# Edit .env file ...

# Option 1 : Run Isaiah attached to the terminal (useful for debugging)
docker run \
  --env-file .env \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -p <YOUR-PORT-MAPPING> \
  mosswill/isaiah

# Option 2 : Run Isaiah as a daemon
docker run \
  -d \
  --env-file .env \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -p <YOUR-PORT-MAPPING> \
  mosswill/isaiah

# Option 3 : Quick run with default variables
docker run -v /var/run/docker.sock:/var/run/docker.sock:ro -p 3000:3000 mosswill/isaiah
```

### Deploy with Docker Compose

To help you get started quickly, multiple example `docker-compose` files are located in the ["examples/"](examples) directory.

Here's a description of every example :

- `docker-compose.simple.yml`: Run Isaiah as a front-facing service on port 80., with environment variables supplied in the `docker-compose` file directly.

- `docker-compose.volume.yml`: Run Isaiah as a front-facing service on port 80, with environment variables supplied as a `.env` file mounted as a volume.

- `docker-compose.ssl.yml`:  Run Isaiah as a front-facing service on port 443, listening for HTTPS requests, with certificate and private key provided as mounted volumes.

- `docker-compose.proxy.yml`: A full setup with Isaiah running on port 80, behind a proxy listening on port 443.

- `docker-compose.traefik.yml`: A sample setup with Isaiah running on port 80, behind a Traefik proxy listening on port 443.

When your `docker-compose` file is on point, you can use the following commands :
```sh
# Option 1 : Run Isaiah in the current terminal (useful for debugging)
docker-compose up

# Option 2 : Run Isaiah in a detached terminal (most common)
docker-compose up -d

# Show the logs written by Isaiah (useful for debugging)
docker logs <NAME-OF-YOUR-CONTAINER>
```

> Warning : Always make sure that your Docker Unix socket is mounted, else Isaiah won't be able to communicate with the Docker API.

### Deploy as a standalone application

You can deploy Isaiah as a standalone application, either by downloading an existing binary that fits your architecture,
or by building the binary yourself on your machine.

#### Using an existing binary

An install script was created to help you install Isaiah in one line, from your terminal :

> As always, check the content of every file you pipe in bash

```sh
curl https://raw.githubusercontent.com/will-moss/isaiah/master/scripts/remote-install.sh | bash
```

This script will try to automatically download a binary that matches your operating system and architecture, and put it
in your `/usr/bin/` directory to ease running it. Later on, you can run :

```sh
# Create a new .env file
touch .env

# Edit .env file ...

# Run Isaiah
isaiah
```

In case you feel uncomfortable running the install script, you can head to the `Releases`, find the binary that meets your system, and install it yourself.

#### Building the binary manually

In this case, make sure that your system meets the following requirements :
- You have Go 1.21 installed
- You have Node 20+ installed along with npm and npx

When all the prerequisites are met, you can run the following commands in your terminal :

> As always, check the content of everything you run inside your terminal

```sh
# Retrieve the code
git clone https://github.com/will-moss/isaiah
cd isaiah

# Run the local install script
./scripts/local-install.sh

# Move anywhere else, and create a dedicated directory
cd ~
mkdir isaiah-config
cd isaiah-config

# Create a new .env file
touch .env

# Edit .env file ...

# Option 1 : Run Isaiah in the current terminal
isaiah

# Option 2 : Run Isaiah as a background process
isaiah &

# Option 3 : Run Isaiah using screen
screen -S isaiah
isaiah
<CTRL+A> <D>

# Optional : Remove the cloned repository
# cd <back to the cloned repository>
# rm -rf ./isaiah
```

The local install script will try to perform a production build on your machine, and move `isaiah` to your `/usr/bin/` directory
to ease running it. In more details, the following actions are performed :
- Local install of Babel, LightningCSS, Less, and Terser
- Prefixing, Transpilation, and Minification of CSS and JS assets
- Building of the Go source code into a single-file executable (with CSS and JS embed)
- Cleaning of the artifacts generated during the previous steps
- Removal of the previous `isaiah` executable, if any in `/usr/bin/`
- Moving the new `isaiah` executable in `/usr/bin` with `755` permissions.

If you encounter any issue during this process, please feel free to tweak the install script or reach out by opening an issue.

## Configuration

To run Isaiah, you will need to set the following environment variables in a `.env` file located next to your executable :

> **Note :** Regular environment variables provided on the commandline work too

| Parameter               | Type      | Description                | Default |
| :---------------------- | :-------- | :------------------------- | ------- |
| `SSL_ENABLED`           | `boolean` | Whether HTTPS should be used in place of HTTP. When configured, Isaiah will look for `certificate.pem` and `key.pem` next to the executable for configuring SSL. Note that if Isaiah is behind a proxy that already handles SSL, this should be set to `false`. | False        |
| `SERVER_PORT`           | `integer` | The port Isaiah listens on. | 3000        |
| `SERVER_MAX_READ_SIZE`  | `integer` | The maximum size (in bytes) per message that Isaiah will accept over Websocket. (Shouldn't be modified, unless your server randomly restarts the Websocket session for no obvious reason) | 1024        |
| `AUTHENTICATION_ENABLED`| `boolean` | Whether a master password is required to access Isaiah. (Recommended) | True |
| `AUTHENTICATION_SECRET` | `string`  | The master password used to secure your Isaiah instance against malicious actors. | one-very-long-and-mysterious-secret        |
| `DISPLAY_CONFIRMATIONS` | `boolean` | Whether the web interface should display a confirmation message after every succesful operation. | True |
| `COLUMNS_CONTAINERS`    | `string`  | Comma-separated list of fields to display in the `Containers` panel. (Case-sensitive) (Available: ID, State, ExitCode, Name, Image) | State,ExitCode,Name,Image |
| `COLUMNS_IMAGES`        | `string`  | Comma-separated list of fields to display in the `Images` panel. (Case-sensitive) (Available: ID, Name, Version, Size) | Name,Version,Size |
| `COLUMNS_VOLUMES`       | `string`  | Comma-separated list of fields to display in the `Volumes` panel. (Case-sensitive) (Available: Name, Driver, MountPoint) | Driver,Name |
| `COLUMNS_NETWORKS`      | `string`  | Comma-separated list of fields to display in the `Networks` panel. (Case-sensitive) (Available: ID, Name, Driver) | Driver,Name |
| `CONTAINER_HEALTH_STYLE`| `string`  | Style used to display the containers' health state. (Available: long, short, icon)| long |
| `CONTAINER_LOGS_TAIL`   | `integer` | Number of lines to retrieve when requesting the last container logs | 50 |
| `CONTAINER_LOGS_SINCE`  | `string`  | The amount of time from now to use for retrieving the last container logs | 60m |
| `TTY_SERVER_COMMAND`    | `string`  | The command used to spawn a new shell inside the server where Isaiah is running | `/bin/sh -i` |
| `TTY_CONTAINER_COMMAND` | `string`  | The command used to spawn a new shell inside the containers that Isaiah manages | `/bin/sh -c eval $(grep ^$(id -un): /etc/passwd \| cut -d : -f 7-) -i` |
| `CUSTOM_DOCKER_HOST`    | `string`  | The host to use in place of the one defined by the DOCKER_HOST default variable | Empty |
| `CUSTOM_DOCKER_CONTEXT` | `string`  | The Docker context to use in place of the current Docker context set on the system | Empty |
| `SKIP_VERIFICATIONS`    | `boolean` | Whether Isaiah should skip startup verification checks before running the HTTP(S) server. (Not recommended) | False        |

> **Note :** Boolean values are case-insensitive, and can be represented via "ON" / "OFF" / "TRUE" / "FALSE" / 0 / 1.

## Troubleshoot

Should you encounter any issue running Isaiah, please refer to the following common problems with their solutions.

#### Isaiah is unreachable over HTTP / HTTPS

Please make sure that the following requirements are met :

- If Isaiah runs as a standalone application without proxy :
    - Make sure your server / firewall accepts incoming connections on Isaiah's port.
    - Make sure your DNS configuration is correct. (Usually, such record should suffice : `A isaiah XXX.XXX.XXX.XXX` for `https://isaiah.your-server-tld`)
    - Make sure your `.env` file is well configured according to the [Configuration](#configuration) section.

- If Isaiah runs on Docker :
    - Perform the previous (standalone) verifications first.
    - Make sure you mounted your server's Docker Unix socket onto the container that runs Isaiah (/var/run/docker.sock)
    - Make sure your Docker container is accessible remotely

- If Isaiah runs behind a proxy :
    - Perform the previous (standalone) verifications first.
    - Make sure that `SERVER_PORT` (Isaiah's port) are well set in `.env`.
    - Check your proxy forwarding rules.

In any case, the crucial part is [Configuration](#configuration) and making sure your Docker / Proxy setup is correct as well.

#### The emulated shell behaves unconsistently or displays unexpected characters

Please note that the emulated shell works by performing the following steps :
- Open a headless terminal on the remote server / inside the remote Docker container.
- Capture standard output, standard error, and bind standard input to the web interface.
- Display standard output and standard error on the web interface as they are streamed over Websocket from the terminal.

According to this implementation, the remote terminal never receives key presses. It only receives commands.

Also, the following techniques are used to try to enhance the user experience on the web interface :
- Enable clearing the shell (HTML) screen via "Ctrl+L" (while the real terminal remains untouched)
- Enable quitting the (HTML) shell via "Ctrl+D" (by sending an "exit" command to the real terminal)
- Handle "command mirror" by appending "# ISAIAH" to every command sent by the user (to distinguish it from command output)
- Handle both "\r" and "\n" newline characters
- Use a time-based approach to detect when a command is finished if it doesn't output anything that shows clear ending
- Remove all escape sequences meant for coloring the terminal output

Therefore it appears that, unless we use a VNC-like solution, the emulation can neither be enhanced nor use keyboard-based features (such as tab completion).

Unless a contributor points the project in the right direction, and as far as my skills go, I personally believe that the current implementation has reached its maximum potential.

I leave here a few ideas that I believe could be implemented, but may require more knowledge, time, testing :
- Convert escape sequences to CSS colors
- Wrap every command in a "block" (begin - command - end) to easily distinguish user-sent commands from output
- Sending to the real terminal the key presses captured from the web (a.k.a sending key presses to a running process)

#### An error happens when spawning a new shell on the server / inside a Docker container

The default commands used to spawn a shell, although being more or less standard, may not fit your environment.
In this case, please edit the `TTY_SERVER_COMMAND` and `TTY_CONTAINER_COMMAND` variables to define a command that works better in your setup.

#### Isaiah doesn't work on mobile

As of now, the web interface isn't responsive. For this reason, a piece of code was intentionally added to prevent any user
from using Isaiah on mobile, so as to spare the unpleasant usage. If this was a mistake on our end, please feel free to open an issue.
Making it an option, or disabling it altogether shouldn't take too long. Also note that the piece of code that is responsible for
disabling Isaiah on mobile devices is easy to bypass as it is 100% CSS-based (removing/editing a CSS class is all it takes, so a browser extension could do it)

#### The connection with the remote server randomly stops or restarts

This is a known incident that happens when the Websocket server receives a data message that exceeds its maximum read size.
You should be able to fix that by setting the `SERVER_MAX_READ_SIZE` variable to a higher value (default is 1024 bytes).
This operation shouldn't cause any problem or impact performances.

#### I can neither click nor use the keyboard, nothing happens

In such a case, please check the icon in the lower right corner.
If you see an orange warning symbol, it means that the connection with the server was lost.
When the connection is lost, all inputs are disabled, until the connection is reestablished (a new attempt is performed every second).

#### Something else

Please feel free to open an issue, explaining what happens, and describing your environment.

## Security

Due to the very nature of Isaiah, I can't emphasize enough how important it is to harden your server :
- Always enable the authentication (with `AUTHENTICATION_ENABLED` and `AUTHENTICATION_SECRET` variables) unless you have your own authentication mechanism built into a proxy.
- Always use a long and secure password to prevent any malicious actor from taking over your Isaiah instance.
- You may also consider putting Isaiah on a private network accessible only through a VPN.

Keep in mind that any breach or misconfiguration on your end could allow a malicious actor to fully take over your server.

## Disclaimer

I believe that, although we're both in the open-source sphere and have all the best intentions, it is important to state the following :

- Isaiah isn't a competitor or any attempt at replacing the lazydocker project. Funnily enough, I'm myself more comfortable running lazydocker through SSH rather than in a browser.
- I've browsed almost all the open issues on lazydocker, and tried to implement and improve what I could (hence the `TTY_CONTAINER_COMMAND` variable, as an example, or even the Image pulling feature).
- Isaiah was built from absolute zero (for both the server and the client), and was ultimately completed using knowledge from lazydocker that I'm personally missing (e.g. the container states and icons).
- Before creating Isaiah, I tried to "serve lazydocker over websocket" (trying to send keypresses to the lazydocker process, and retrieving the output via Websocket), but didn't succeed, hence the full rewrite.
- I also tried to start Isaiah from the lazydocker codebase and implement a web interface on top of it, but it seemed impractical or simply beyond my skills, hence the full rewrite.

Ultimately, thanks to the people behind lazydocker both for the amazing project (that I'm using daily) and for paving the way for Isaiah.

PS : Please also note that Isaiah isn't exactly 100% feature-equivalent with lazydocker (e.g. charts are missing)
PS2 : What spurred me to build Isaiah in the first place is a bunch of comments on the Reddit self-hosted community, stating that Portainer and other available solutions were too heavy or hard to use. A Redditor said that having lazydocker over the web would be amazing, so I thought I'd do just that.

## Contribute

This is one of my first ever open-source projects, and I'm not a Docker / Github / Docker Hub / Git guru yet.

If you can help in any way, please do! I'm looking forward to learning from you.

From the top of my head, I'm sure there's already improvement to be made on :
- Terminology (using the proper words to describe technical stuff)
- Coding practices (e.g. writing better comments, avoiding monkey patches)
- Shell emulation (e.g. improving on what's done already)
- Release process (e.g. making explicit commits, pushing Docker images properly to Docker Hub)
- Github settings (e.g. using discussions, wiki, etc.)
- And more!

## Credits

Hey hey ! It's always a good idea to say thank you and mention the people and projects that help us move forward.

Big thanks to the individuals / teams behind these projects :
- [laydocker](https://github.com/jesseduffield/lazydocker) : Isaiah wouldn't exist if Lazydocker hadn't been created prior, and to say that it is an absolutely incredible and very advanced project is an understatement.
- [Heroicons](https://github.com/tailwindlabs/heroicons) : For the great icons.
- [Melody](https://github.com/olahol/melody) : For the awesome Websocket implementation in Go.
- [GoReleaser](https://github.com/goreleaser/goreleaser) : For the amazing release tool.
- The countless others!

And don't forget to mention Isaiah if it makes your life easier!

