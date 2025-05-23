package mnkgame

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGame(t *testing.T) {
	game := NewGame(15, 15, 5)
	assert.Equal(t, StatusTurnX, game.Status)
}

func TestVerticalWin(t *testing.T) {
	g := NewGame(3, 3, 3)

	require.Nil(t, MakeTurn(g, Position{X: 0, Y: 0}))
	require.Nil(t, MakeTurn(g, Position{X: 2, Y: 0}))
	require.Nil(t, MakeTurn(g, Position{X: 0, Y: 1}))
	require.Nil(t, MakeTurn(g, Position{X: 2, Y: 1}))
	require.Nil(t, MakeTurn(g, Position{X: 0, Y: 2}))
	require.Equal(t, StatusWinX.String(), g.Status.String())
}

func TestHorizontalWin(t *testing.T) {
	g := NewGame(3, 3, 3)

	require.Nil(t, MakeTurn(g, Position{X: 0, Y: 0}))
	require.Nil(t, MakeTurn(g, Position{X: 1, Y: 1}))
	require.Nil(t, MakeTurn(g, Position{X: 1, Y: 0}))
	require.Nil(t, MakeTurn(g, Position{X: 0, Y: 1}))
	require.Nil(t, MakeTurn(g, Position{X: 2, Y: 2}))
	require.Nil(t, MakeTurn(g, Position{X: 2, Y: 1}))

	require.Equal(t, StatusWinO.String(), g.Status.String())
}

// func TestDiagonalTopLeftBottomRightWin(t *testing.T) {
// 	game := NewGame(5, 5, 3)
// 	MakeTurn(game, Position{x: 1, y: 1})
// 	MakeTurn(game, Position{x: 2, y: 0})
// 	MakeTurn(game, Position{x: 0, y: 0})
// 	MakeTurn(game, Position{x: 1, y: 0})
// 	MakeTurn(game, Position{x: 2, y: 2})
// 	assert.Equal(t, StatusWinX, game.Status)
// }

// func TestDiagonalTopRightBottomLeftWin(t *testing.T) {
// 	game := NewGame(5, 5, 3)
// 	MakeTurn(game, Position{x: 1, y: 1})
// 	MakeTurn(game, Position{x: 0, y: 0})
// 	MakeTurn(game, Position{x: 2, y: 0})
// 	MakeTurn(game, Position{x: 1, y: 0})
// 	MakeTurn(game, Position{x: 0, y: 2})
// 	assert.Equal(t, StatusWinX, game.Status)
// }
