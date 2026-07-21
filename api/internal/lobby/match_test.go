package lobby

import (
	"api/internal/event"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newDefaultMatch() *Match {
	p1 := event.NewClient(nil, "Player 1")
	p2 := event.NewClient(nil, "Player 2")
	return NewMatch(p1, p2)
}

func TestNewMatchContainsOnlyNilMarkers(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()
	for _, c := range m.board {
		assert.Nil(c)
	}
}

func TestNewMatchCurrentClientIsP1(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()
	assert.Equal(m.P1, m.CurrentClient())
}

func TestNewMatchStatusIsNotFinished(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()
	assert.Equal(NOT_FINISHED, m.Status())
}

func TestMarkRotatesClient(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()
	m.Mark(A)
	assert.Equal(m.P2, m.CurrentClient())
}

func TestMarkDoesNotMarkAlreadyMarkedTile(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()

	assert.NoError(m.Mark(A))
	assert.ErrorAs(m.Mark(A), &ErrTileAlreadyMarked{})
}

func TestMarkWorks(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()

	assert.NoError(m.Mark(A))
	assert.NoError(m.Mark(B))

	assert.Equal(m.P1, m.Board()[A])
	assert.Equal(m.P2, m.Board()[B])
}

func TestP1WinsHorizontal(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()

	assert.NoError(m.Mark(A))
	assert.NoError(m.Mark(D))
	assert.NoError(m.Mark(B))
	assert.NoError(m.Mark(E))
	assert.NoError(m.Mark(C))

	assert.Equal(P1_WON, m.Status())
}

func TestP2WinsHorizontal(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()

	assert.NoError(m.Mark(D))
	assert.NoError(m.Mark(A))
	assert.NoError(m.Mark(E))
	assert.NoError(m.Mark(B))
	assert.NoError(m.Mark(G))
	assert.NoError(m.Mark(C))

	assert.Equal(P2_WON, m.Status())
}

func TestP1WinsVertical(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()

	assert.NoError(m.Mark(A))
	assert.NoError(m.Mark(B))
	assert.NoError(m.Mark(D))
	assert.NoError(m.Mark(C))
	assert.NoError(m.Mark(G))

	assert.Equal(P1_WON, m.Status())
}

func TestP1WinsDiagonal(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()

	assert.NoError(m.Mark(A))
	assert.NoError(m.Mark(B))
	assert.NoError(m.Mark(E))
	assert.NoError(m.Mark(C))
	assert.NoError(m.Mark(I))

	assert.Equal(P1_WON, m.Status())
}

func TestP1WinsDiagonal2(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()

	assert.NoError(m.Mark(C))
	assert.NoError(m.Mark(A))
	assert.NoError(m.Mark(E))
	assert.NoError(m.Mark(B))
	assert.NoError(m.Mark(G))

	assert.Equal(P1_WON, m.Status())
}

func TestDraw(t *testing.T) {
	assert := assert.New(t)
	m := newDefaultMatch()

	assert.NoError(m.Mark(A))
	assert.NoError(m.Mark(C))
	assert.NoError(m.Mark(B))
	assert.NoError(m.Mark(D))
	assert.NoError(m.Mark(F))
	assert.NoError(m.Mark(E))
	assert.NoError(m.Mark(G))
	assert.NoError(m.Mark(I))
	assert.NoError(m.Mark(H))

	assert.Equal(DRAW, m.Status())
}
