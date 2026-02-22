package model

import (
	"errors"
)

const (
	BoardSize = 15
	EmptyCell = 0
)

type GameState int

const (
	GameStateWaiting GameState = iota
	GameStatePlaying
	GameStateFinished
)

var (
	ErrNotYourTurn     = errors.New("not your turn")
	ErrInvalidMove     = errors.New("invalid move")
	ErrCellOccupied    = errors.New("cell already occupied")
	ErrGameNotStarted  = errors.New("game not started")
	ErrGameAlreadyOver = errors.New("game already over")
	ErrInvalidPosition = errors.New("invalid position")
)

type Game struct {
	RoomID    int64
	Board     [][]int
	Players   []int64
	Current   int
	State     GameState
	Winner    int64
	WinLine   []int
	MoveCount int
}

func NewGame(roomID int64, players []int64) *Game {
	board := make([][]int, BoardSize)
	for i := range board {
		board[i] = make([]int, BoardSize)
	}

	return &Game{
		RoomID:  roomID,
		Board:   board,
		Players: players,
		Current: 0,
		State:   GameStatePlaying,
		Winner:  0,
		WinLine: nil,
	}
}

func (g *Game) CurrentPlayer() int64 {
	if len(g.Players) == 0 {
		return 0
	}
	return g.Players[g.Current]
}

func (g *Game) NextTurn() {
	g.Current = (g.Current + 1) % len(g.Players)
}

func (g *Game) IsValidPosition(x, y int) bool {
	return x >= 0 && x < BoardSize && y >= 0 && y < BoardSize
}

func (g *Game) IsEmpty(x, y int) bool {
	if !g.IsValidPosition(x, y) {
		return false
	}
	return g.Board[x][y] == EmptyCell
}

func (g *Game) MakeMove(playerID int64, x, y int) error {
	if g.State != GameStatePlaying {
		return ErrGameNotStarted
	}

	if g.Winner != 0 {
		return ErrGameAlreadyOver
	}

	if g.CurrentPlayer() != playerID {
		return ErrNotYourTurn
	}

	if !g.IsValidPosition(x, y) {
		return ErrInvalidPosition
	}

	if !g.IsEmpty(x, y) {
		return ErrCellOccupied
	}

	playerIndex := g.Current + 1
	g.Board[x][y] = playerIndex
	g.MoveCount++

	if g.CheckWin(x, y, playerIndex) {
		g.Winner = playerID
		g.State = GameStateFinished
	} else if g.MoveCount >= BoardSize*BoardSize {
		g.State = GameStateFinished
	} else {
		g.NextTurn()
	}

	return nil
}

func (g *Game) CheckWin(x, y, player int) bool {
	directions := [][2]int{
		{1, 0},
		{0, 1},
		{1, 1},
		{1, -1},
	}

	for _, dir := range directions {
		if g.checkDirection(x, y, dir[0], dir[1], player) {
			return true
		}
	}

	return false
}

func (g *Game) checkDirection(x, y, dx, dy, player int) bool {
	count := 1
	line := []int{x*BoardSize + y}

	for i := 1; i < 5; i++ {
		nx, ny := x+dx*i, y+dy*i
		if !g.IsValidPosition(nx, ny) || g.Board[nx][ny] != player {
			break
		}
		count++
		line = append(line, nx*BoardSize+ny)
	}

	for i := 1; i < 5; i++ {
		nx, ny := x-dx*i, y-dy*i
		if !g.IsValidPosition(nx, ny) || g.Board[nx][ny] != player {
			break
		}
		count++
		line = append(line, nx*BoardSize+ny)
	}

	if count >= 5 {
		g.WinLine = line
		return true
	}

	return false
}

func (g *Game) IsFinished() bool {
	return g.State == GameStateFinished
}

func (g *Game) IsDraw() bool {
	return g.State == GameStateFinished && g.Winner == 0
}

func (g *Game) GetBoardCopy() [][]int {
	board := make([][]int, BoardSize)
	for i := range board {
		board[i] = make([]int, BoardSize)
		copy(board[i], g.Board[i])
	}
	return board
}

func (g *Game) Forfeit(playerID int64) int64 {
	if g.State != GameStatePlaying {
		return 0
	}

	for i, p := range g.Players {
		if p != playerID {
			g.Winner = p
			g.State = GameStateFinished
			return g.Winner
		}
		if i == len(g.Players)-1 {
			break
		}
	}

	return 0
}
