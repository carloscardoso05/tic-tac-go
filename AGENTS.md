# AGENTS.md

## Project overview

Tic-Tac-Toe multiplayer API. Monorepo with a single Go module at `api/`.

## Module & commands

All Go commands run from `api/`, which is the module root (module name: `api`).

```
# Run all tests
cd api && go test ./...

# Run a single test
cd api && go test ./domain/room -run TestP1WinsHorizontal

# Start dev server with hot reload
cd api && air

# Run server directly
cd api && go run ./cmd/server
```

## Architecture

```
api/
  cmd/server/main.go   --- entrypoint (Gin + websocket)
  domain/room/          --- pure game logic (board, tiles, victory, status)
  event/                --- websocket event parse/send layer
  tmp/                  --- air build output (gitignored)
```

`domain/room` has no framework dependencies and is the only tested package.

## Code generation

`Tile` and `Status` enums in `domain/room/room.go` use `go:generate stringer`. After changing enum values, regenerate:

```
cd api && go generate ./domain/room
```

This updates `tile_string.go` and `status_string.go` (committed, DO NOT EDIT by hand).

## Tests

- Test framework: `github.com/stretchr/testify` (assert style)
- No mocks or external dependencies needed
- Test file: `api/domain/room/room_test.go`
- Helper `newDefaultRoom()` creates a room with two players "Player 1" / "Player 2"

## WebSocket events

Event types defined in `api/event/event.go`. The server entrypoint (`cmd/server/main.go`) currently only partially handles `GameEnded` events — the other event types (`TileMarked`, `UserJoined`, `UserLeft`) have cases but no implementation yet. This is early-stage (WIP).

## Style notes

- Data is passed by value on the stack (`Room`, `Player` structs — no heap pointers). Copying is intentional.
- `Marker` is a type alias `*Player`, not an interface.
- Board is `[9]Marker` (array, not slice). Tile enum values (A=0 through I=8) double as board indices.
