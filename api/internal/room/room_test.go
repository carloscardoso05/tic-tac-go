package room

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	P1 Player = Player{"Player 1"}
	P2 Player = Player{"Player 2"}
)

func newDefaultRoom() *Room {
	return NewRoom(P1, P2)
}

func TestNewRoomContainsOnlyNilMarkers(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()
	for _, marker := range room.board {
		assert.Nil(marker)
	}
}

func TestNewRoomCurrentPlayerIsP1(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()
	assert.Equal(P1, *room.CurrentPlayer())
}

func TestNewRoomStatusIsNotFinished(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()
	assert.Equal(NOT_FINISHED, room.Status())
}

func TestMarkRotatesPlayer(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()
	room.Mark(A)
	assert.Equal(P2, *room.CurrentPlayer())
}

func TestMarkDoesNotMarksAlreadyMarkedTile(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()

	assert.NoError(room.Mark(A))
	assert.ErrorAs(room.Mark(A), &ErrTileAlreadyMarked{})
}

func TestMarkWorks(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()

	assert.NoError(room.Mark(A))
	assert.NoError(room.Mark(B))

	assert.Equal(P1, *room.Board()[A])
	assert.Equal(P2, *room.Board()[B])
}

func TestP1WinsHorizontal(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()

	assert.NoError(room.Mark(A)) // P1
	assert.NoError(room.Mark(D)) // P2
	assert.NoError(room.Mark(B)) // P1
	assert.NoError(room.Mark(E)) // P2
	assert.NoError(room.Mark(C)) // P1

	assert.Equal(P1_WON, room.Status())
}

func TestP2WinsHorizontal(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()

	assert.NoError(room.Mark(D)) // P1
	assert.NoError(room.Mark(A)) // P2
	assert.NoError(room.Mark(E)) // P1
	assert.NoError(room.Mark(B)) // P2
	assert.NoError(room.Mark(G)) // P1
	assert.NoError(room.Mark(C)) // P2

	assert.Equal(P2_WON, room.Status())
}

func TestP1WinsVertical(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()

	assert.NoError(room.Mark(A)) // P1
	assert.NoError(room.Mark(B)) // P2
	assert.NoError(room.Mark(D)) // P1
	assert.NoError(room.Mark(C)) // P2
	assert.NoError(room.Mark(G)) // P1

	assert.Equal(P1_WON, room.Status())
}

func TestP1WinsDiagonal(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()

	assert.NoError(room.Mark(A)) // P1
	assert.NoError(room.Mark(B)) // P2
	assert.NoError(room.Mark(E)) // P1
	assert.NoError(room.Mark(C)) // P2
	assert.NoError(room.Mark(I)) // P1

	assert.Equal(P1_WON, room.Status())
}

func TestP1WinsDiagonal2(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()

	assert.NoError(room.Mark(C)) // P1
	assert.NoError(room.Mark(A)) // P2
	assert.NoError(room.Mark(E)) // P1
	assert.NoError(room.Mark(B)) // P2
	assert.NoError(room.Mark(G)) // P1

	assert.Equal(P1_WON, room.Status())
}

func TestDraw(t *testing.T) {
	assert := assert.New(t)
	room := newDefaultRoom()

	assert.NoError(room.Mark(A)) // P1
	assert.NoError(room.Mark(C)) // P2
	assert.NoError(room.Mark(B)) // P1
	assert.NoError(room.Mark(D)) // P2
	assert.NoError(room.Mark(F)) // P1
	assert.NoError(room.Mark(E)) // P2
	assert.NoError(room.Mark(G)) // P1
	assert.NoError(room.Mark(I)) // P2
	assert.NoError(room.Mark(H)) // P1

	assert.Equal(DRAW, room.Status())
}
