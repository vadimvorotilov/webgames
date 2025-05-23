package mnkgame

import (
	"errors"
	"fmt"
)

type Game struct {
	Board       [][]Cell
	Status      Status
	InARowToWin int
	History     []Position
}

func (g *Game) Width() int {
	return len(g.Board[0])
}

func (g *Game) Height() int {
	return len(g.Board)
}

type Board [][]Cell

type Player int

const (
	PlayerX Player = 1
	PlayerO Player = 2
)

type Cell int

const (
	CellEmpty Cell = iota
	CellX
	CellO
)

type Status int

const (
	StatusTurnX Status = iota
	StatusTurnO
	StatusWinX
	StatusWinO
	StatusDraw
)

var statusName = map[Status]string{
	StatusTurnX: "Turn X",
	StatusTurnO: "Turn O",
	StatusWinX:  "Win X",
	StatusWinO:  "Win O",
	StatusDraw:  "Draw",
}

func (s Status) String() string {
	return statusName[s]
}

type Position struct {
	X int
	Y int
}

func NewGame(width int, height int, InARowToWin int) *Game {
	board := make(Board, height)
	for i := range board {
		board[i] = make([]Cell, width)
	}

	return &Game{
		Board:       board,
		InARowToWin: InARowToWin,
	}
}

func MakeTurn(g *Game, pos Position) error {
	if pos.X < 0 || pos.X >= g.Width() ||
		pos.Y < 0 || pos.Y >= g.Height() ||
		g.Board[pos.Y][pos.X] != CellEmpty {
		return errors.New("Position is out of bounds or cell is already occupied")
	}

	// set cell or return error if game has already ended
	var cell Cell
	switch g.Status {
	case StatusTurnX:
		cell = CellX
	case StatusTurnO:
		cell = CellO
	default:
		return fmt.Errorf("Game has already ended with status: %s", g.Status)
	}

	g.Board[pos.Y][pos.X] = cell
	g.History = append(g.History, pos)

	win := checkWin(g, pos, cell)

	if win {
		if cell == CellX {
			g.Status = StatusWinX
		} else {
			g.Status = StatusWinO
		}

		return nil
	}

	// check draw condition
	if len(g.History) == g.Width()*g.Height() {
		g.Status = StatusDraw
		return nil
	}

	// switch the player for the next move
	if cell == CellX {
		g.Status = StatusTurnO
	} else {
		g.Status = StatusTurnX
	}

	return nil
}

func checkWin(g *Game, pos Position, cell Cell) bool {
	window := 1
	delta := 1

	// check vertical
	for window < g.InARowToWin {
		if pos.Y+delta >= 0 && pos.Y+delta < g.Height() && // out of bounds
			cell == g.Board[pos.Y+delta][pos.X] { // equality
			window++

			if delta > 0 {
				delta++
			} else {
				delta--
			}

		} else if delta > 0 { // change direction
			delta = -1
		} else { // no win
			delta = 1
			window = 1
			break
		}
	}
	if window == g.InARowToWin {
		return true
	}

	// // check horizontal
	for window < g.InARowToWin {
		if pos.X+delta >= 0 && pos.X+delta < g.Width() && // out of bounds
			cell == g.Board[pos.Y][pos.X+delta] { // equality
			window++
			if delta > 0 {
				delta++
			} else {
				delta--
			}
		} else if delta > 0 { // change direction
			delta = -1
		} else { // no win
			delta = 1
			window = 1
			break
		}
	}
	if window == g.InARowToWin {
		return true
	}

	// // check diagonal-top-left-bottom-right
	// for window < g.InARowToWin {
	// 	if pos.x+delta >= 0 && pos.x+delta < len(g.board[0]) &&
	// 		pos.y+delta >= 0 && pos.y+delta < len(g.board[0]) &&
	// 		cell == g.board[pos.y+delta][pos.x+delta] { // check out of bounds and equality
	// 		window++
	// 		delta++
	// 	} else if delta > 0 { // change direction
	// 		delta = -1
	// 	} else { // no win
	// 		delta = 1
	// 		window = 1
	// 		break
	// 	}
	// }
	// if window == g.InARowToWin {
	// 	return true
	// }

	// // check diagonal-top-right-bottom-left
	// for window < g.InARowToWin {
	// 	if pos.x+delta >= 0 && pos.x+delta < len(g.board[0]) &&
	// 		pos.y-delta >= 0 && pos.y-delta < len(g.board[0]) &&
	// 		cell == g.board[pos.y-delta][pos.x+delta] { // check out of bounds and equality
	// 		window++
	// 		delta++
	// 	} else if delta > 0 { // change direction
	// 		delta = -1
	// 	} else { // no win
	// 		delta = 1
	// 		window = 1
	// 		break
	// 	}
	// }

	return window == g.InARowToWin
}
