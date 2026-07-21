# AGENTS.md

## Project overview

Tic-Tac-Toe multiplayer API. Monorepo with a single Go module at `api/`.

## Module & commands

All Go commands run from `api/`, which is the module root (module name: `api`).

```
# Run all tests
cd api && go test ./...

# Run a single test
cd api && go test ./internal/room -run TestP1WinsHorizontal

# Start dev server with hot reload
cd api && air

# Run server directly
cd api && go run ./cmd/server
```

## Architecture

```
api/
  cmd/server/main.go        --- entrypoint (Gin bootstrap + connection loop)
  event/                     --- message types, parse/send, dispatch router
  internal/
    room/                     --- pure game logic (board, tiles, victory, status)
    lobby/                    --- room hub: create, join, mark, leave, broadcast
    transport/               --- websocket accept/read/write wrappers
    client/                  --- connected player state (ID, name, conn)
  tmp/                       --- air build output (gitignored)
```

`internal/room` has no framework dependencies and is the only tested package.

### Flow

Client connects via `GET /ws` → WebSocket upgrade → connection loop reads `event.Message` → `event.Router` dispatches by `msg.Type` string to a lobby handler → lobby calls domain logic and broadcasts to room members.

## Code generation

`Tile` and `Status` enums in `internal/room/room.go` use `go:generate stringer`. After changing enum values, regenerate:

```
cd api && go generate ./internal/room
```

This updates `tile_string.go` and `status_string.go` (committed, DO NOT EDIT by hand).

## Event protocol

- Messages are flat JSON: `{"type":"mark","data":{"room":"uuid","tile":0}}`
- `EventType` is a `string` (not `uint`). Defined in `api/event/event.go`.
- `event.Message.Data` is `json.RawMessage` — human-readable on the wire, no base64 encoding.
- Dispatch table in `event/router.go`: register handlers with `router.Handle("create", handler)`.
- Handler signature: `func(ctx context.Context, cl *client.Client, data json.RawMessage) error`
- `ctx` is **not** stored in the `Client` struct — it's passed per operation (Go convention).

## Tests

- Test framework: `github.com/stretchr/testify` (assert style)
- No mocks or external dependencies needed
- Test file: `api/internal/room/room_test.go`
- Helper `newDefaultRoom()` creates a room with two players "Player 1" / "Player 2"

## Style notes

- Data is passed by value on the stack (`Room`, `Player` structs — no heap pointers). Copying is intentional.
- `Marker` is a type alias `*Player`, not an interface.
- Board is `[9]Marker` (array, not slice). Tile enum values (A=0 through I=8) double as board indices.
