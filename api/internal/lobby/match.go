package lobby

import (
	"api/internal/event"

	"github.com/google/uuid"
)

type Tile int

/*
A B C
D E F
G H I
*/
//go:generate stringer -type=Tile
const (
	A Tile = iota
	B
	C
	D
	E
	F
	G
	H
	I
)

var victoryPatterns [8][3]Tile = [8][3]Tile{
	{A, B, C},
	{D, E, F},
	{G, H, I},
	{A, D, G},
	{B, E, H},
	{C, F, I},
	{A, E, I},
	{C, E, G},
}

//go:generate stringer -type=Status
type Status uint

const (
	NOT_FINISHED Status = iota
	DRAW
	P1_WON
	P2_WON
)

type Match struct {
	ID      uuid.UUID
	board   [9]*event.Client
	current uint
	P1      *event.Client
	P2      *event.Client
}

func NewMatch(p1, p2 *event.Client) *Match {
	return &Match{
		ID:      uuid.New(),
		P1:      p1,
		P2:      p2,
		board:   [9]*event.Client{},
		current: 0,
	}
}

func (m *Match) CurrentClient() *event.Client {
	if m.current == 0 {
		return m.P1
	}
	return m.P2
}

func (m *Match) Board() [9]*event.Client {
	return m.board
}

func (m *Match) Status() Status {
	for _, line := range victoryPatterns {
		if m.board[line[0]] != nil &&
			m.board[line[0]] == m.board[line[1]] &&
			m.board[line[1]] == m.board[line[2]] {
			if m.board[line[0]] == m.P1 {
				return P1_WON
			}
			return P2_WON
		}
	}
	for _, c := range m.board {
		if c == nil {
			return NOT_FINISHED
		}
	}
	return DRAW
}

func (m *Match) Mark(tile Tile) error {
	current := m.CurrentClient()
	if m.board[tile] != nil {
		return ErrTileAlreadyMarked{tile, m.board[tile].Name}
	}
	m.board[tile] = current
	if m.Status() == NOT_FINISHED {
		m.current = (m.current + 1) % 2
	}
	return nil
}
