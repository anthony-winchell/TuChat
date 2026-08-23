# AGENTS.md

Go chat app: TCP server + bubbletea TUI client sharing a JSON protocol. Module is named **`tuchat`** (not go-chat) — imports are `tuchat/client`, `tuchat/client/tui`, `tuchat/protocol`, `tuchat/server`.

## Commands

```sh
go build ./...
go vet ./...
go test ./server              # only package with tests (~4s)
go test ./server -run TestServerAddClientDuplicate   # single test
```

Run the two binaries (both are `package main`):

```sh
go run ./client
go run ./server               # MUST be run from server/ dir (see below)
```

## Gotchas

- **Server state paths are CWD-relative**: it loads/saves `config.json` and writes chat logs to `logs/<room>.log`. Running from the repo root scatters state files at the root. Run it from `server/`.
- **State persists only on graceful shutdown**: SIGINT/SIGTERM both trigger `SaveConfig` + clean shutdown (`server/main.go`). A hard kill (`kill -9`, crash) loses config changes since the last successful save. `config.json` and `/logs` are gitignored.
- **Ports are hardcoded**: server listens on `:8080` (`server/main.go`), client dials `localhost:8080` (`client/tui/connection.go:10` const `serverAddr`). Change both together.
- `AGENTS.md` itself is gitignored — treat it as local notes.

## Architecture

- `protocol/message.go`: ALL client↔server traffic is one flat `Message` struct discriminated by `Type` string, with `omitempty` fields. A new feature usually starts here, then touches both `server/messaging.go` and `client/tui/update_server.go`.
- Client TUI follows bubbletea's Elm architecture, split by responsibility: `update_*.go` (message handling per area: auth/chat/sidebar/server/connection), `view_*.go` (rendering per screen), `model.go` (single `Model` holding sub-state structs), `messages.go` (custom tea.Msg types), `connection.go` (network I/O wrapped as `tea.Cmd`).
- Server concurrency convention: shared maps guarded by `sync.RWMutex`; never expose fields directly — use accessor getters (`s.Name()`) and copy-out snapshot methods (`usersSnapshot()`, `clientsSnapshot()` in `snapshots.go`).
- Sentinel errors live in `server/main.go` (`ErrUsernameTaken`, etc.); match them with `errors.Is`.
- Server slash commands are registered in a registry (`InitializeCommands` in `commands.go`, implementations in `commands_user.go` / `commands_room.go` / `commands_server.go`).
- **Wire format**: newline-delimited JSON — `json.Encoder.Encode` appends `\n`; server reads via `bufio.Scanner` with `maxMessageSize` cap (1 MiB, `server/server.go`). Oversized message = disconnect (stream position is lost after `ErrTooLong`; never continue the read loop). The TUI client still reads with `json.Decoder`, which handles the `\n` framing fine.
- **Auth hardening** (`server/server.go` consts): 2-min pre-auth read deadline (`authTimeout`) covering both server-password and login/register phases; combined attempt cap (`maxAuthAttempts` = 3) then disconnect.
- **Auth errors are uniform by design**: every auth failure returns exactly `protocol.AuthInvalidCredentials` (shared const used by server, client matcher in `update_auth.go`, and tests). Don't reintroduce user-existence leaks or string-match old error text; see `TestServerAuthenticateUserErrorsAreIndistinguishable`.
- **Usernames are canonically lowercase**: normalized at every entry point (`NewUser`/`RestoreUser`/`FindUser`/`findClient`, `Server.SetOwner`, `Room.SetOwner`, `Room.RestoreAdmin`). Nicknames keep their case. New username touchpoints must normalize too — mixed-case entries from old configs are handled on load.
- **`SaveConfig` is atomic + self-locked**: writes `config.json.tmp` then `os.Rename`s over the target, guarded by its own `saveMu` (separate from `s.mu` because RWMutex isn't reentrant). Don't call it while holding `s.mu`.
- Connection goroutines get a `recover()` guard in `Start()` (`server/server.go`) — a panic in one client handler logs a stack trace instead of killing the server.

## Testing

Stdlib `testing` only — no testify/gomock. Shared helpers: `newTestServer()` (`server_users_test.go`), `newTestClient(t, username)` (`room_test.go`). Tests construct servers directly; they don't bind real ports.
