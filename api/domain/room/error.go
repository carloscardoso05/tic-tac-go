package room

import "fmt"

type ErrTileAlreadyMarked struct {
	tile   Tile
	marker Marker
}

func (e ErrTileAlreadyMarked) Error() string {
	return fmt.Sprintf("The tile %v is already marked by player %v", e.tile, e.marker.Name)
}

type ErrNotPlayerTurn struct {
	player string
}

func (e ErrNotPlayerTurn) Error() string {
	return fmt.Sprintf("It's not the player %s turn", e.player)
}

type ErrMarkedWithNil struct {
	tile Tile
}

func (e ErrMarkedWithNil) Error() string {
	return fmt.Sprintf("Can't mark the tile %v with nil", e.tile)
}
