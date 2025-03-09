package advent

import (
	"math/rand"

	"github.com/andrewsjg/goAdventure/dungeon"
)

//TODO: Implement settings

func NewGame() Game {

	newGame := Game{}
	seedval := rand.Int()

	newGame.LcgX = int32(seedval) % LCG_M

	if newGame.LcgX < 0 {
		newGame.LcgX += LCG_M + newGame.LcgX
	}

	// Generate the Z'ZZZ word
	for i := 0; i < 5; i++ {
		newGame.LcgX = getNextLCGValue(newGame.LcgX)

		if newGame.LcgX < 0 {
			newGame.LcgX = -newGame.LcgX
		}

		newGame.Zzword[i] = byte('A' + (26 * newGame.LcgX / LCG_M))
	}

	// Make the second character an apostrophe
	newGame.Zzword[1] = '\''
	newGame.Zzword[5] = '\x00'

	for i := 1; i < dungeon.NDWARVES; i++ {
		newGame.Dwarves[i].Loc = dungeon.DwarfLocs[i-1]

	}

	return newGame
}

func getNextLCGValue(lcgX int32) int32 {
	return (LCG_A*lcgX + LCG_C) % LCG_M
}
