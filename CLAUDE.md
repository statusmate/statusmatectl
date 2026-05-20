# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o st4 .

# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.txt -covermode=atomic ./...

# Run a single test
go test ./pkg/api/... -run TestFunctionName

# Run the binary directly
go run . <command>
```

## Architecture

This is a CLI tool (`st4`) for managing Statusmate status pages. It is built with [Cobra](https://github.com/spf13/cobra) and communicates with the Statusmate REST API.

### Package layout

- **`cmd/`** — Cobra commands. Each file defines one or more `*cobra.Command` values and registers them in `init()`. The root command is `RootCmd` in `cmd/cmd.go` (binary name `st4`). The `ls` subcommand (`cmd/ls.go`) acts as a namespace for short-alias list commands.
- **`pkg/api/`** — API client and domain models. `client.go` defines `Client` with generic `Get`/`Post`/`Patch`/`Delete` methods. Domain types (`Incident`, `Component`, `StatusPage`, etc.) live in their own files and attach methods to `Client`.
- **`pkg/printer/`** — Output formatting. `PrintTableConfig` controls table vs. JSON output. Each entity type has its own print file. The `detail-incident.go` file handles `key=value` summary output (designed to be parsed with `awk -F=`).
- **`pkg/format/`** — Custom INI-like marshal/unmarshal for interactive editing. Struct fields are tagged with `` `format:"field_name"` `` and map to `[field_name]` sections. Used to serialize structs to a temp file for editor-based editing.
- **`pkg/editor/`** — Opens `$EDITOR` (default: `vim`) with a temp file and returns the edited content.

### Auth flow

Credentials are stored per-server in `~/.st4/<sanitized-server-domain>/authrc` as JSON (`AuthRC` struct). `InitClientCommandContextCobra` (in `cmd/root.go`) loads this file and sets `client.Token`. The `--server` flag (default: `statusmate.top`) controls which profile is used.

### Adding a new command

1. Create `cmd/<name>.go`, define a `*cobra.Command`, and register it with `RootCmd.AddCommand(...)` in `init()`.
2. Use `InitClientCommandContextCobra(command)` to get an authenticated `*api.Client`.
3. Add the corresponding API method to the relevant file in `pkg/api/`.

### Component impact format

When specifying affected components (e.g. `--components`), the format is `"<impact> <component-name>"`. Impact shorthands: `o`/`op` = operational, `u`/`um` = under_maintenance, `d`/`dp` = degraded_performance, `p`/`po` = partial_outage, `m`/`mo` = major_outage.

### Output format

List commands support `--format table|json` (via `PrintTableConfig`). Summary output after create/update uses `key=value` lines for awk-friendliness.

### Pagination

`NewAllPaginatedRequest(filter)` requests up to 1000 results (size=1000, page=0). Use `PaginatedRequestFilter` (a `map[string]any`) to pass server-side filters like `status_page` ID or `status`.

## API Reference

Full API documentation (endpoints, schemas, enums) is in `.claude/skills/statusmate-api.md`. Consult it when implementing or debugging any HTTP calls. The live OpenAPI schema is at `https://devstatusmate.ru/api/schema/`.
