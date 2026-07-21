package lobby

import "fmt"

type ErrTileAlreadyMarked struct {
	tile Tile
	name string
}

func (e ErrTileAlreadyMarked) Error() string {
	return fmt.Sprintf("The tile %v is already marked by %s", e.tile, e.name)
}
