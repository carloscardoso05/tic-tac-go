# AGENTS.md

## Project overview

Tic-Tac-Toe multiplayer API. Monorepo with a single Go module at `api/`.

## Module & commands

All Go commands run from `api/`, which is the module root (module name: `api`).

```
# Run all tests
cd api && go test ./...

# Run a single test
cd api && go test ./internal/lobby -run TestP1WinsHorizontal

# Start dev server with hot reload
cd api && air

# Run server directly
cd api && go run ./cmd/server
```

## Architecture

```
api/
  cmd/server/main.go        --- entrypoint (Gin bootstrap + connection loop)
  internal/
    event/
      event.go               --- message types, Client, parse/send, Read/Write
      router.go               --- dispatch table (Handler + Router)
    lobby/
      match.go                --- Tile, Status, Board, Match, Mark, victory
      lobby.go                --- Lobby (hub), handlers, Disconnect, broadcast
      error.go                --- domain errors
      match_test.go           --- tests
      tile_string.go          --- generated (DO NOT EDIT)
      status_string.go         --- generated (DO NOT EDIT)
  tmp/                        --- air build output (gitignored)
```

`internal/lobby` has no framework dependencies and is the only tested package.

### Flow

Client connects via `GET /ws` → WebSocket upgrade → connection loop reads `event.Message` → `event.Router` dispatches by `msg.Type` string to a lobby handler → lobby calls match logic and broadcasts to both players.

## Code generation

`Tile` and `Status` enums in `internal/lobby/match.go` use `go:generate stringer`. After changing enum values, regenerate:

```
cd api && go generate ./internal/lobby
```

This updates `tile_string.go` and `status_string.go` (committed, DO NOT EDIT by hand).

## Event protocol

- Messages are flat JSON: `{"type":"mark","data":{"room":"uuid","tile":0}}`
- `EventType` is a `string`. Defined in `api/internal/event/event.go`.
- `event.Message.Data` is `json.RawMessage` — human-readable on the wire, no base64 encoding.
- Dispatch table in `event/router.go`: register handlers with `router.Handle("create", handler)`.
- Handler signature: `func(ctx context.Context, cl *event.Client, data json.RawMessage) error`
- `ctx` is **not** stored in the `Client` struct — it's passed per operation (Go convention).

## Tests

- Test framework: `github.com/stretchr/testify` (assert style)
- No mocks or external dependencies needed
- Test file: `api/internal/lobby/match_test.go`
- Helper `newDefaultMatch()` creates a match with two clients "Player 1" / "Player 2"

## Style notes

- Data is passed by value on the stack. Copying is intentional.
- Board is `[9]*Client` (array, not slice). Tile enum values (A=0 through I=8) double as board indices.
