<p align="center">
    <h1 align="center">Isaiah</h1>
    <p align="center">
      Self-hostable clone of lazydocker for the web.<br />Manage your Docker fleet with ease
    </p>
    <p align="center">
      <a href="#table-of-contents">Table of Contents</a> -
      <a href="#deployment-and-examples">Install</a> -
      <a href="#configuration">Configure</a>
    </p>
</p>

| | | |
|:-------------------------:|:-------------------------:|:-------------------------:|
|<img width="1604" src="/assets/CAPTURE-1.png"/> |  <img width="1604" src="/assets/CAPTURE-2.png"/> | <img width="1604" src="/assets/CAPTURE-3.png" /> |
|<img width="1604" src="/assets/CAPTURE-4.png"/> |  <img width="1604" src="/assets/CAPTURE-5.png"/> | <img width="1604" src="/assets/CAPTURE-6.png" /> |
|<img width="1604" src="/assets/CAPTURE-7.png"/> |  <img width="1604" src="/assets/CAPTURE-8.png"/> | <img width="1604" src="/assets/CAPTURE-9.png" /> |
|<img width="1604" src="/assets/CAPTURE-10.png"/> |  <img width="1604" src="/assets/CAPTURE-11.png"/> | <img width="1604" src="/assets/CAPTURE-12.png"/> |
|<img width="1604" src="/assets/CAPTURE-13.png"/> |  <img width="1604" src="/assets/CAPTURE-14.png"/> | <img width="1604" src="/assets/CAPTURE-15.png"/> |

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Deployment and Examples](#deployment-and-examples)
  * [Deploy with Docker](#deploy-with-docker)
  * [Deploy with Docker Compose](#deploy-with-docker-compose)
  * [Deploy as a standalone application](#deploy-as-a-standalone-application)
    + [Using an existing binary](#using-an-existing-binary)
    + [Building the binary manually](#building-the-binary-manually)
- [Multi-node deployment](#multi-node-deployment)
  * [General information](#general-information)
  * [Setup](#setup)
- [Multi-host deployment](#multi-host-deployment)
  * [General information](#general-information-1)
  * [Setup](#setup-1)
- [Forward Proxy Authentication / Trusted SSO](#forward-proxy-authentication--trusted-sso)
- [Configuration](#configuration)
- [Theming](#theming)
- [Troubleshoot](#troubleshoot)
- [Security](#security)
- [Disclaimer](#disclaimer)
- [Contribute](#contribute)
- [Credits](#credits)

## Introduction

Isaiah is a self-hostable service that enables you to manage all your Docker resources on a remote server. It is an attempt at recreating the [lazydocker](https://github.com/jesseduffield/lazydocker) command-line application from scratch, while making it available as a web application without compromising on the features.


## Features

Isaiah has all these features implemented :
- For stacks :
    - Bulk update
    - Up, Down, Pause, Unpause, Restart, Update
    - Create and Edit stacks using `docker-compose.yml` files in your browser
    - Inspect (live logs, `docker-compose.yml`, services)
- For containers :
    - Bulk stop, Bulk remove, Prune
    - Remove, Pause, Unpause, Restart, Rename, Update, Edit, Open in browser
    - Open a shell inside the container (from your browser)
    - Inspect (live logs, stats, env, full configuration, top)
- For images :
    - Prune
    - Remove
    - Run (create and start a container using the image)
    - Open on Docker Hub
    - Pull a new image (from Docker Hub)
    - Bulk pull all latest images (from Docker Hub)
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
- Built-in authentication by master password (supplied raw or sha256-hashed)
- Built-in authentication by forward proxy authentication headers (e.g. Authelia / Trusted SSO)
- Built-in terminal emulator (with support for opening a shell on the server)
- Responsive for Desktop, Tablet, and Mobile
- Support for multiple layouts
- Support for custom CSS theming (with variables for colors already defined)
- Support for keyboard navigation
- Support for mouse navigation
- Support for search through Docker resources and container logs
- Support for ascending and descending sort by any supported field
- Support for customizable user settings (line-wrap, timestamps, prompt, etc.)
- Support for custom Docker Host / Context.
- Support for extensive configuration with `.env`
- Support for HTTP and HTTPS
- Support for standalone / proxy / multi-node / multi-host deployment

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

# Option 3 : Quick run with default settings
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

- `docker-compose.agent.yml`: A sample setup with Isaiah operating as an Agent in a multi-node deployment.

- `docker-compose.host.yml`: A sample setup with Isaiah expecting to communicate with other hosts in a multi-host deployment.

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
in your `/usr/[local/]bin/` directory to ease running it. Later on, you can run :

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

The local install script will try to perform a production build on your machine, and move `isaiah` to your `/usr/[local/]bin/` directory
to ease running it. In more details, the following actions are performed :
- Local install of Babel, LightningCSS, Less, and Terser
- Prefixing, Transpilation, and Minification of CSS and JS assets
- Building of the Go source code into a single-file executable (with CSS and JS embed)
- Cleaning of the artifacts generated during the previous steps
- Removal of the previous `isaiah` executable, if any in `/usr/[local/]bin/`
- Moving the new `isaiah` executable in `/usr/[local/]bin` with `755` permissions.

If you encounter any issue during this process, please feel free to tweak the install script or reach out by opening an issue.

## Multi-node deployment

Using Isaiah, you can manage multiple nodes with their own distinct Docker resources from a single dashboard.

Before delving into that part, please get familiar with the general information below.

### General information

You may find these information useful during your setup and reading :
- Isaiah distinguishes two types of nodes : `Master` and `Agent`.
- The word `node` refers to any machine (virtual or not) holding Docker resources.
- The `Master` node has three responsabilities :
  - Serving the web interface.
  - Managing the Docker resources inside the environment on which it is already installed.
  - Acting as a central proxy between the client (you) and the remote Agent nodes.
- The `Master` node has the following characteristics :
  - There should be only one Master node in a multi-node deployment.
  - The Master node should be the only part of your deployment that is publicly exposed on the network.
- The `Agent` nodes have the following characteristics :
  - They are headless instances of Isaiah, and they can't exist without a Master node.
  - As with the Master node, they have their own authentication if you don't disable it explicitly.
  - On startup, they perform registration with their Master node using as a Websocket client
  - For as long as the Master node is alive, a Websocket connection remains established between them.
  - The Agent node should never be publicly exposed on the network.
  - The Agent node never communicates with the client (you). Everything remains between the nodes.
  - There is no limit to how many Agent nodes can connect to a Master node.

In other words, one `Master` acts as a `Proxy` between the `Client` and the `Agents`.<br />
For example, when a `Client` wants to stop a Docker container inside an `Agent`, the `Client` first requests it from `Master`.
Then, `Master` forwards it to the designated `Agent`.
When the `Agent` has finished, they reply to `Master`, and `Master` forwards that response to the initial `Client`.

Schematically, it looks like this :
- Client ------------> Master : Stop container C-123 on Agent AG-777
- Master ------------> Agent  : Stop container C-123
- Agent  ------------> Master : Container C-123 was stopped
- Master ------------> Client : Container C-123 was stopped on Agent AG-777

Now that we understand how everything works, let's see how to set up a multi-node deployment.

### Setup

First, please ensure the following :
- Your `Master` node is running, exposed on the network, and available in your web browser
- Your `Agent` node has Isaiah installed and configured with the following settings :
  - `SERVER_ROLE` equal to `Agent`
  - `MASTER_HOST` configured to reach the `Master` node
  - `MASTER_SECRET` equal to the `AUTHENTICATION_SECRET` setting on the `Master` node, or empty when authentication is disabled
  - `AGENT_NAME` equal to a unique string of your choice

Then, launch Isaiah on each `Agent` node, and you should see logs indicating whether connection with `Master` was established. Eventually, you will see `Master` or `The name of your agent` in the lower right corner of your screen as agents register.

If encounter any issue, please read the [Troubleshoot](#troubleshoot) section.

> You may want to note that you don't need to expose ports on the machine / Docker container running Isaiah when it is configured as an Agent.

## Multi-host deployment

Using Isaiah, you can manage multiple hosts with their own distinct Docker resources from a single dashboard.

Before delving into that part, please get familiar with the general information below.

### General information

The big difference between multi-node and multi-host deployments is that you won't need to install Isaiah on every single node
if you are using multi-host. In this setup, Isaiah is installed only on one server, and communicates with other Docker hosts
directly over TCP / Unix sockets. It makes it easier to manage multiple remote Docker environments without having to setup Isaiah
on all of them.

Please note that, in a multi-host setup, there must be a direct access between the main host (where Isaiah is running)
and the other ones. Usually, they should be on the same network, or visible through a secured gateway / VPN / filesystem mount.

Let's see how to set up a multi-host deployment.

### Setup

In order to help you get started, a [sample file](/app/sample.docker_hosts) was created.

First, please ensure the following :
- Your `Master` host is running, exposed on the network, and available in your web browser
- Your `Master` host has the setting `MULTI_HOST_ENABLED` set to `true`.
- Your `Master` host has access to the other Docker hosts over TCP / Unix socket.

Second, please create a `docker_hosts` file next to Isaiah's executable, using the sample file cited above:
- Every line should contain two strings separated by a single space.
- The first string is the name of your host, and the second string is the path to reach it.
- The path to your host should look like this : [PROTOCOL]://[URI]
- Example 1 : Local unix:///var/run/docker.sock
- Example 2 : Remote tcp://my-domain.tld:4382

> If you're using Docker, you can mount the file at the root of the filesystem, as in :<br />
`docker ... -v my_docker_hosts:/docker_hosts ...`

Finally, launch Isaiah on the Master host, and you should see logs indicating whether connection with remote hosts was established.
Eventually, you will see `Master` with `The name of your host` in the lower right corner of your screen.

## Forward Proxy Authentication / Trusted SSO

If you wish to deploy Isaiah behind a secure proxy or authentication portal, you must configure Forward Proxy Authentication.

This will enable you to :
- Log in, once and for all, into your authentication portal.
- Connect to Isaiah without having to type your `AUTHENTICATION_SECRET` every time.
- Protect Isaiah using your authentication proxy rather than the current mechanism (cleartext / hashed password).
- Manage the access to Isaiah from your authentication portal rather than through your `.env` configuration.

Before proceeding, please ensure the following :
- Your proxy supports HTTP/2 and Websockets.
- Your proxy can communicate with Isaiah on the network.
- Your proxy forwards authentication headers to Isaiah (but not to the browser).

<blockquote>
  <br />
  For example, if you're using Nginx Proxy Manager (NPM), you should do the following :
  <br /><br />
  <ul>
    <li>In the tab "Details"</li>
    <ul>
      <li>Tick the box "Websockets support"</li>
      <li>Tick the box "HTTP/2 support"</li>
      <li>Tick the box "Block common exploits"</li>
      <li>Tick the box "Force SSL"</li>
   </ul>
   <br />
   <li>In the tab "Advanced"</li>
   <ul>
      <li>In your custom location block, add the lines :</li>
      <ul>
        <li>proxy_set_header Upgrade $http_upgrade;</li>
        <li>proxy_set_header Connection "upgrade";</li>
      </ul>
    </ul>
  </ul>
  <br />
</blockquote>

Then, configure Isaiah using the following variables :
- Set `FORWARD_PROXY_AUTHENTICATION_ENABLED` to `true`.
- Set `FORWARD_PROXY_AUTHENTICATION_HEADER_KEY` to the name of the forwarded authentication header your proxy sends to Isaiah.
- Set `FORWARD_PROXY_AUTHENTICATION_HEADER_VALUE` to the value of the header that Isaiah should expect (or use `*` if all values are accepted).

> By default, Isaiah is configured to work with Authelia out of the box. Hence, you can just set `FORWARD_PROXY_AUTHENTICATION_ENABLED` to `true` and be done with it.

If everything was properly set up, you will encounter the following flow :
- Navigate to `isaiah.your-domain.tld`.
- Get redirected to `authentication-portal.your-domain.tld`.
- Fill in your credentials.
- Authentication was successful.
- Get redirected to `isaiah.your-domain.tld`.
- Isaiah **does not** prompt you for the password, you're automatically logged in.


## Configuration

To run Isaiah, you will need to set the following environment variables in a `.env` file located next to your executable :

> **Note :** Regular environment variables provided on the commandline work too

| Parameter               | Type      | Description                | Default |
| :---------------------- | :-------- | :------------------------- | ------- |
| `SSL_ENABLED`           | `boolean` | Whether HTTPS should be used in place of HTTP. When configured, Isaiah will look for `certificate.pem` and `key.pem` next to the executable for configuring SSL. Note that if Isaiah is behind a proxy that already handles SSL, this should be set to `false`. | False        |
| `SERVER_PORT`           | `integer` | The port Isaiah listens on. | 3000        |
| `SERVER_MAX_READ_SIZE`  | `integer` | The maximum size (in bytes) per message that Isaiah will accept over Websocket. Note that, in a multi-node deployment, you may need to incrase the value of that setting. (Shouldn't be modified, unless your server randomly restarts the Websocket session for no obvious reason) | 100000        |
| `AUTHENTICATION_ENABLED`| `boolean` | Whether a password is required to access Isaiah. (Recommended) | True |
| `AUTHENTICATION_SECRET` | `string`  | The master password used to secure your Isaiah instance against malicious actors. | one-very-long-and-mysterious-secret        |
| `AUTHENTICATION_HASH`   | `string`  | The master password's hash (sha256 format) used to secure your Isaiah instance against malicious actors. Use this setting instead of `AUTHENTICATION_SECRET` if you feel uncomfortable providing a cleartext password. | Empty    |
| `DISPLAY_CONFIRMATIONS` | `boolean` | Whether the web interface should display a confirmation message after every succesful operation. | True |
| `TABS_ENABLED`          | `string`  | Comma-separated list of tabs to display in the interface. (Case-insensitive) (Available: Stacks, Containers, Images, Volumes, Networks) | stacks,containers,images,volumes,networks |
| `COLUMNS_CONTAINERS`    | `string`  | Comma-separated list of fields to display in the `Containers` panel. (Case-sensitive) (Available: ID, State, ExitCode, Name, Image, Created) | State,ExitCode,Name,Image |
| `COLUMNS_IMAGES`        | `string`  | Comma-separated list of fields to display in the `Images` panel. (Case-sensitive) (Available: ID, Name, Version, Size) | Name,Version,Size |
| `COLUMNS_VOLUMES`       | `string`  | Comma-separated list of fields to display in the `Volumes` panel. (Case-sensitive) (Available: Name, Driver, MountPoint) | Driver,Name |
| `COLUMNS_NETWORKS`      | `string`  | Comma-separated list of fields to display in the `Networks` panel. (Case-sensitive) (Available: ID, Name, Driver) | Driver,Name |
| `SORTBY_CONTAINERS`     | `string`  | Field used to sort the rows in the `Containers` panel. (Case-sensitive) (Available: ID, State, ExitCode, Name, Image, Created) | Empty |
| `SORTBY_IMAGES`         | `string`  | Field used to sort the rows in the `Images` panel. (Case-sensitive) (Available: ID, Name, Version, Size) | Empty |
| `SORTBY_VOLUMES`        | `string`  | Field used to sort the rows in the `Volumes` panel. (Case-sensitive) (Available: Name, Driver, MountPoint) | Empty |
| `SORTBY_NETWORKS`       | `string`  | Field used to sort the rows in the `Networks` panel. (Case-sensitive) (Available: Id, Name, Driver) | Empty |
| `CONTAINER_HEALTH_STYLE`| `string`  | Style used to display the containers' health state. (Available: long, short, icon)| long |
| `CONTAINER_LOGS_TAIL`   | `integer` | Number of lines to retrieve when requesting the last container logs | 50 |
| `CONTAINER_LOGS_SINCE`  | `string`  | The amount of time from now to use for retrieving the last container logs | 60m |
| `STACKS_DIRECTORY`      | `string`  | The path to the directory that will be used to store the `docker-compose.yml` files generated while creating and editing stacks. It must be a valid path to an existing and writable directory. | `.` (current directory) |
| `TTY_SERVER_COMMAND`    | `string`  | The command used to spawn a new shell inside the server where Isaiah is running | `/bin/sh -i` |
| `TTY_CONTAINER_COMMAND` | `string`  | The command used to spawn a new shell inside the containers that Isaiah manages | `/bin/sh -c eval $(grep ^$(id -un): /etc/passwd \| cut -d : -f 7-) -i` |
| `CUSTOM_DOCKER_HOST`    | `string`  | The host to use in place of the one defined by the DOCKER_HOST default variable | Empty |
| `CUSTOM_DOCKER_CONTEXT` | `string`  | The Docker context to use in place of the current Docker context set on the system | Empty |
| `SKIP_VERIFICATIONS`    | `boolean` | Whether Isaiah should skip startup verification checks before running the HTTP(S) server. (Not recommended) | False        |
| `SERVER_ROLE`           | `string`  | For multi-node deployments only. The role of the current instance of Isaiah. Can be either `Master` or `Agent` and is case-sensitive. | Master        |
| `MASTER_HOST`           | `string`  | For multi-node deployments only. The host used to reach the Master node, specifying the IP address or the hostname, and the port if applicable (e.g. my-server.tld:3000). | Empty        |
| `MASTER_SECRET`         | `string`  | For multi-node deployments only. The secret password used to authenticate on the Master node. Note that it should equal the `AUTHENTICATION_SECRET` setting on the Master node. | Empty        |
| `AGENT_NAME`            | `string`  | For multi-node deployments only. The name associated with the Agent node as it is displayed on the web interface. It should be unique for each Agent. | Empty        |
| `MULTI_HOST_ENABLED`    | `boolean` | Whether Isaiah should be run in multi-host mode. When enabled, make sure to have your `docker_hosts` file next to the executable. | False        |
| `FORWARD_PROXY_AUTHENTICATION_ENABLED`    | `boolean` | Whether Isaiah should accept authentication headers from a forward proxy. | False        |
| `FORWARD_PROXY_AUTHENTICATION_HEADER_KEY` | `string` | The name of the authentication header sent by the forward proxy after a succesful authentication. | Remote-User        |
| `FORWARD_PROXY_AUTHENTICATION_HEADER_VALUE` | `string` | The value accepted by Isaiah for the authentication header. Using `*` means that all values are accepted (except emptiness). This parameter can be used to enforce that only a specific user or group can access Isaiah (e.g. `admins` or `john`). | * |
| `CLIENT_PREFERENCE_XXX` | `string` | Please read [this troubleshooting paragraph](#the-web-interface-does-not-save-my-preferences). These settings enable you to define your client preferences on the server, for when your browser can't use the `localStorage` due to limitations, or private browsing. | Empty |

> **Note :** Boolean values are case-insensitive, and can be represented via "ON" / "OFF" / "TRUE" / "FALSE" / 0 / 1.

> **Note :** To sort rows in reverse using the `SORTBY_` parameters, prepend your field with the minus symbol, as in `-Name`

> **Note :** Use either `AUTHENTICATION_SECRET` or `AUTHENTICATION_HASH` but not both at the same time.

> **Note** : You can generate a sha256 hash using an online tool, or using the following commands :
**On OSX** : `echo -n your-secret | shasum -a 256`
**On Linux** : `echo -n your-secret | sha256sum`

Additionally, once Isaiah is fully set up and running, you can open the Parameters Manager by pressing the `X` key.
Using this interface, you can toggle the following options based on your preferences :

| Parameter               | Description                |
| :---------------------- | :------------------------- |
| `enableMenuPrompt`      | Whether an extra prompt should warn you before trying to stop / pause / restart a Docker container. |
| `enableLogLinesWrap`    | Whether log lines streamed from Docker containers should be wrapped (as opposed to extend beyond your screen). |
| `enableTimestampDisplay`| Whether log lines' timestamps coming from Docker containers should be displayed. |
| `enableOverviewOnLaunch`| Whether an overview panel should show first before anything when launching Isaiah in your browser. |
| `enableLogLinesStrippedBackground`| Whether alternated log lines should have a brighter background to enhance readability. |
| `enableJumpFuzzySearch` | Whether, in Jump mode, fuzzy search should be used, as opposed to default substring search. |
| `enableSyntaxHightlight`| Whether syntax highlighting should be enabled (when viewing docker-compose.yml files). |

> Note : You must have Isaiah open in your browser and be authenticated to access these options. Once set up, these options will be saved to your localStorage.


## Theming

You can customize Isaiah's web interface using your own custom CSS. At runtime, Isaiah will look for a file named `custom.css` right next to the executable.
If this file exists, it will be loaded in your browser and it will override any existing CSS rule.

In order to help you get started, a [sample file](/app/sample.custom.css) was created.
It shows how to modify the CSS variables responsible for the colors of the interface. (All the values are the ones used by default)
You can copy that file, update it, and rename it to `custom.css`.

If you're using Docker, you should mount a `custom.css` file at the root of your container's filesystem.
Example : `docker ... -v my-custom.css:/custom.css ...`

Finally, you will find below a table that describes what each CSS color variable means :

| Variable                | Description                |
| :---------------------- | :--------                  |
| `color-terminal-background` | Background of the interface |
| `color-terminal-base` | Texts of the interface |
| `color-terminal-accent` | Elements that are interactive or must catch the attention |
| `color-terminal-accent-selected` | Panel's title when the panel is in focus |
| `color-terminal-hover` | Panel's rows that are in focus / hover |
| `color-terminal-border` | Panels' borders color |
| `color-terminal-danger` | The color used to convey danger / failure |
| `color-terminal-warning` | Connection indicator when connection is lost |
| `color-terminal-accent-alternative` | Connection indicator when connection is established |
| `color-terminal-log-row-alternative` | The color used as background for each odd row in the logs tab |
| `color-terminal-json-key` | The color used to distinguish keys from values in the inspector when displaying a long configuration |
| `color-terminal-json-value` | The color used to distinguish values from keys in the inspector when displaying a long configuration |
| `color-terminal-cell-failure` | Container health state when exited |
| `color-terminal-cell-success` | Container health state when running |
| `color-terminal-cell-paused` | Container health state when paused |

On a side note, creating custom layouts using only CSS isn't implemented yet as it requires interaction with Javascript.
That said, implementing this feature should be quick and simple since the way layouts are managed currently is already modular.

Ultimately, please note that Isaiah already comes with three themes : dawn, moon, and the default one.
The first two themes are based on Rosé Pine, and new themes may be implemented later.

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
- Handle up and down arrow keys to cycle through commands history locally

Therefore it appears that, unless we use a VNC-like solution, the emulation can neither be enhanced nor use keyboard-based features (such as tab completion).

Unless a contributor points the project in the right direction, and as far as my skills go, I personally believe that the current implementation has reached its maximum potential.

I leave here a few ideas that I believe could be implemented, but may require more knowledge, time, testing :
- Convert escape sequences to CSS colors
- Wrap every command in a "block" (begin - command - end) to easily distinguish user-sent commands from output
- Sending to the real terminal the key presses captured from the web (a.k.a sending key presses to a running process)

Ultimately, please also note that in a multi-node / multi-host setup, the extra network latency and unexpected buffering from remote terminals may cause additional display artifacts.

#### An error happens when spawning a new shell on the server / inside a Docker container

The default commands used to spawn a shell, although being more or less standard, may not fit your environment.
In this case, please edit the `TTY_SERVER_COMMAND` and `TTY_CONTAINER_COMMAND` settings to define a command that works better in your setup.

Also, please note that if you have deployed Isaiah using Docker, trying to open a system shell (`S` key) will not work.
Isaiah being confined to its Docker container, it won't be able to open a shell out of it (on your hosting system).

#### The connection with the remote server randomly stops or restarts

This is a known incident that happens when the Websocket server receives a data message that exceeds its maximum read size.
You should be able to fix that by updating the `SERVER_MAX_READ_SIZE` setting to a higher value (default is 100,000 bytes).
This operation shouldn't cause any problem or impact performances.

#### I can neither click nor use the keyboard, nothing happens

In such a case, please check the icon in the lower right corner.
If you see an orange warning symbol, it means that the connection with the server was lost.
When the connection is lost, all inputs are disabled, until the connection is reestablished (a new attempt is performed every second).

#### The interface is stuck loading indefinitely

This incident arises when a crash occurs while inside a shell or performing a Docker command.
The quickest "fix" for that is to refresh your browser tab (Ctrl+R/Cmd+R).

The real "fix" (if any) could be to implement a "timeout" (client-side or server-side) after which, the "loading" state is automatically discarded

If you encounter this incident consistently, please reach out by opening an issue so we look deeper into that part

#### The web interface seems to randomly crash and restart

If you haven't already, please read about the `SERVER_MAX_READ_SIZE` setting in the [Configuration](#configuration) section.

That incident occurs when the Websocket messages sent from the client to the server are too big.
The server's reaction to overly large messages sent over Websocket is to close the connection with the client.
When that happens, Isaiah (as a client in your browser) automatically reopens a connection with the server, hence explaining the "crash-restart" cycle.

#### The web interface does not save my preferences

First, please ensure that your browser supports the `localStorage` API.

Second, please ensure that you're not using the `private browsing` or `incognito` or `anonymous browsing` mode. This mode will
turn off the `localStorage`, hence disabling the user preferences saved by Isaiah in your browser.

If you wish to use Isaiah inside a private browser window while still having your preferences stored somewhere, use the
`CLIENT_PREFERENCE_XXX` settings in your deployment. These settings will be stored server-side, and understood by your browser
without ever using `localStorage`, hence circumventing the limitation of the private browsing mode.

All the preferences described in the second table of [Configuration](#configuration) are available server-side, using their uppercased-underscore counterpart.
See below :
- `theme` becomes `CLIENT_PREFERENCE_THEME`
- `enableOverviewOnLaunch` becomes `CLIENT_PREFERENCE_ENABLE_OVERVIEW_ON_LAUNCH`
- `enableMenuPrompt` becomes `CLIENT_PREFERENCE_ENABLE_MENU_PROMPT`
- `enableLogLinesWrap` becomes `CLIENT_PREFERENCE_ENABLE_LOG_LINES_WRAP`
- `enableJumpFuzzySearch` becomes `CLIENT_PREFERENCE_ENABLE_JUMP_FUZZY_SEARCH`
- `enableTimestampDisplay` becomes `CLIENT_PREFERENCE_ENABLE_TIMESTAMP_DISPLAY`
- `enableLogLinesStrippedBackground` becomes `CLIENT_PREFERENCE_ENABLE_LOG_LINES_STRIPPED_BACKGROUND`

#### A feature that works on desktop is missing from the mobile user interface

Please note that you can horizontally scroll the mobile controls located in the bottom part of your screen to reveal all of them.
If, for any reason, you still encounter a case when a feature is missing on your mobile device, please open an issue
indicating the browser you're using, your screen's viewport size, and the model of your phone.

#### In a multi-node deployment, the agent's registration with master is stuck loading indefinitely

This issue arises when the authentication settings between Master and Agent nodes are incompatible.<br />
To fix it, please make sure that :
- When authentication is enabled on Master, the Agent has a `MASTER_SECRET` setting defined.
- When authentication is disabled on Master, the Agent has no `MASTER_SECRET` setting defined.

Also don't forget to restart your nodes when changing settings.

#### Something else

Please feel free to open an issue, explaining what happens, and describing your environment.

## Security

Due to the very nature of Isaiah, I can't emphasize enough how important it is to harden your server :
- Always enable the authentication (with `AUTHENTICATION_ENABLED` and `AUTHENTICATION_SECRET` settings) unless you have your own authentication mechanism built into a proxy.
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
- [Fuse](https://github.com/krisk/fuse) : For the amazing fuzzy-search library.
- The countless others!

And don't forget to mention Isaiah if it makes your life easier!


