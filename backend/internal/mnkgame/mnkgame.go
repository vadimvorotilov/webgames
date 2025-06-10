package mnkgame

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type GameID string
type PlayerID string

type Game struct {
	ID        GameID
	PlayerXID PlayerID
	PlayerOID PlayerID
	Board     [][]Cell
	Status    Status
	WinRow    int
	History   []Position
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

func (c Cell) String() string {
	switch c {
	case 0:
		return " "
	case 1:
		return "X"
	case 2:
		return "O"
	}

	return ""
}

const (
	CellEmpty Cell = iota
	CellX
	CellO
)

type Status int

const (
	StatusOpponent Status = iota
	StatusTurnX
	StatusTurnO
	StatusWinX
	StatusWinO
	StatusDraw
)

var statusName = map[Status]string{
	StatusOpponent: "Need opponent",
	StatusTurnX:    "Turn X",
	StatusTurnO:    "Turn O",
	StatusWinX:     "Win X",
	StatusWinO:     "Win O",
	StatusDraw:     "Draw",
}

func (s Status) String() string {
	return statusName[s]
}

type Position struct {
	X int
	Y int
}

type CreateGameParams struct {
	PlayerXID string
	Width     int
	Height    int
	WinRow    int
}

type MakeTurnParams struct {
	X int
	Y int
}

var gamesRepository = map[GameID]*Game{
	"9f4ef5fb-d1ce-4ecd-aa7c-3c5ba02bc0a7": {
		ID:        "9f4ef5fb-d1ce-4ecd-aa7c-3c5ba02bc0a7",
		PlayerXID: "74273137-5d5b-48b1-910c-9718afae8ae6",
		PlayerOID: "",
		WinRow:    3,
		Board: [][]Cell{
			{0, 0, 0},
			{0, 0, 0},
			{0, 0, 0},
		},
	},
}

func CreateGame(params CreateGameParams) *Game {
	board := make(Board, params.Height)
	for i := range board {
		board[i] = make([]Cell, params.Width)
	}

	game := &Game{
		ID:        GameID(uuid.NewString()),
		PlayerXID: PlayerID(params.PlayerXID),
		Board:     board,
		WinRow:    params.WinRow,
	}

	gamesRepository[game.ID] = game
	return game
}

func NewGame(width int, height int, winRow int) *Game {
	board := make(Board, height)
	for i := range board {
		board[i] = make([]Cell, width)
	}

	return &Game{
		ID:     GameID(uuid.NewString()),
		Board:  board,
		WinRow: winRow,
	}
}

func BecomeOpponent(game *Game, playerID PlayerID) error {
	if game.Status != StatusOpponent {
		return fmt.Errorf("invalid game status: %s", game.Status)
	}

	game.PlayerOID = playerID
	game.Status = StatusTurnX

	return nil
}

func FindGame(id GameID) (*Game, bool) {
	game, ok := gamesRepository[id]
	return game, ok
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
	for window < g.WinRow {
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
	if window == g.WinRow {
		return true
	}

	// // check horizontal
	for window < g.WinRow {
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
	if window == g.WinRow {
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

	return window == g.WinRow
}
