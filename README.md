# st4 — Statusmate CLI

🌐 **English** · [Русский](README.RU.md)

`st4` is a command-line tool for managing [Statusmate](https://statusmate.top) status pages:
incidents, maintenances, components, teams and subscribers.

## Installation

### Homebrew (macOS / Linux)

```bash
brew install statusmate/tap/st4
```

If the tap is not connected yet, you can add it explicitly:

```bash
brew tap statusmate/tap
brew install st4
```

Verify the installation:

```bash
st4 version
```

### Updating via Homebrew

```bash
brew update
brew upgrade st4
```

Upgrade only this formula without touching other packages:

```bash
brew upgrade statusmate/tap/st4
```

### Building from source

```bash
go build -o st4 .
```

## Quick start

```bash
# Log in (the default server is statusmate.top)
st4 login

# Pick a default status page
st4 config use-status-page

# Check who you are logged in as
st4 whoami

# List status pages
st4 ls

# List incidents / components / maintenances
st4 ls i
st4 ls c
st4 ls m

# Create an incident / maintenance
st4 create-incident --help
st4 create-maintenance --help
```

## Configuration

`st4` stores its configuration **separately for each server**. This lets you work with
several Statusmate installations (for example, `statusmate.top` and your own self-hosted
server) without switching profiles by hand.

### Where the config lives

The `authrc` file is stored as JSON at:

```
~/.st4/<server-domain>/authrc
```

For example, for the default server `statusmate.top`:

```
~/.st4/statusmate.top/authrc
```

Print the exact path to the config:

```bash
st4 config path
```

Show the current config contents:

```bash
st4 config show
```

### What the config holds

| Field                  | Description                         |
| ---------------------- | ----------------------------------- |
| `token`                | Auth token (created by `st4 login`) |
| `default_status_page`  | Default status page                 |
| `default_release_page` | Default release page                |
| `default_team`         | Default team                        |
| `api`                  | API server address                  |

Default values are applied to commands automatically, so you don't have to pass
`--status-page` or `--team` every time.

### Setting defaults

```bash
# Pick the default status page (interactive list)
st4 config use-status-page

# Show the current default status page
st4 config current-page

# Pick the default team (interactive list)
st4 config use-team

# Show the current default team
st4 config current-team
```

### Choosing a server

The default server is `statusmate.top`. You can override it with the `--server` flag
or the `ST4_SERVER` environment variable:

```bash
# Via flag
st4 --server=statusmate.top login

# Via environment variable
export ST4_SERVER=statusmate.top
st4 login
```

Each server keeps its own configuration directory under `~/.st4/`, so authentication
and defaults never overlap between servers.

## Main commands

```bash
st4 login                 # log in (2FA supported)
st4 whoami                # current user info
st4 ls                    # list status pages
st4 ls i                  # list incidents
st4 ls c                  # list components
st4 ls m                  # list maintenances
st4 ls t                  # list teams
st4 ls u                  # list team members
st4 create-incident       # create an incident
st4 create-maintenance    # schedule a maintenance
st4 status                # component status tree
st4 tui                   # interactive TUI dashboard
st4 open                  # open a status page in the browser
st4 config ...            # manage configuration (see above)
st4 version               # tool version
```

List all commands and flags:

```bash
st4 --help
st4 <command> --help
```

## License

See the [LICENSE](LICENSE) file.
