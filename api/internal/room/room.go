package room

import (
	"sync"
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
	//
	{A, D, G},
	{B, E, H},
	{C, F, I},
	//
	{A, E, I},
	{C, E, G},
}

type Player struct {
	Name string
}

type Marker *Player

type Board [9]Marker

//go:generate stringer -type=Status
type Status uint

const (
	NOT_FINISHED Status = iota
	DRAW
	P1_WON
	P2_WON
)

type Room struct {
	board              Board
	currentPlayerIndex uint
	players            [2]Player
	playerMutex        sync.Mutex
}

func NewRoom(p1, p2 Player) *Room {
	return &Room{players: [2]Player{p1, p2}}
}

func (r *Room) CurrentPlayer() *Player {
	return &r.players[r.currentPlayerIndex]
}

func (r *Room) Board() Board {
	return r.board
}

func (r *Room) rotatePlayer() {
	r.currentPlayerIndex = (r.currentPlayerIndex + 1) % 2
}

func (r *Room) Status() Status {
	if r.checkVictory(&r.players[0]) {
		return P1_WON
	}
	if r.checkVictory(&r.players[1]) {
		return P2_WON
	}
	for _, m := range r.board {
		if m == nil {
			return NOT_FINISHED
		}
	}
	return DRAW
}

func (r *Room) checkPattern(marker Marker, line [3]Tile) bool {
	return marker != nil &&
		r.board[line[0]] == marker &&
		r.board[line[1]] == marker &&
		r.board[line[2]] == marker
}

func (r *Room) checkVictory(marker Marker) bool {
	for _, pattern := range victoryPatterns {
		if r.checkPattern(marker, pattern) {
			return true
		}
	}
	return false
}

func (r *Room) Mark(tile Tile) error {
	marker := r.CurrentPlayer()
	if r.board[tile] != nil {
		return ErrTileAlreadyMarked{tile, marker}
	}
	r.board[tile] = marker
	if !r.checkVictory(marker) {
		r.rotatePlayer()
	}
	return nil
}
