package advent

import (
	"math/rand"

	"github.com/andrewsjg/goAdventure/dungeon"
)

type Game struct {
	LcgX         int32
	Abbnum       int32
	Bonus        int32
	Chloc        int32
	Chloc2       int32
	Clock1       int32
	Clock2       int32
	Clshnt       bool
	Closed       bool
	Closng       bool
	Lmwarn       bool
	Novice       bool
	Panic        bool
	Wzdark       bool
	Blooded      bool
	Conds        int32
	Detail       int32
	Dflag        int32
	Dkill        int32
	Dtotal       int32
	Foobar       int32
	Holdng       int32
	Igo          int32
	Iwest        int32
	Knfloc       int32
	Limit        int32
	Loc          int32
	Newloc       int32
	Numdie       int32
	Oldloc       int32
	Oldlc2       int32
	Oldobj       int32
	Saved        int32
	Tally        int32
	Thresh       int32
	Seenbigwords bool
	Trnluz       int32
	Turns        int32
	Zzword       [5 + 1]byte
	Locs         [dungeon.NLOCATIONS + 1]struct {
		Abbrev int32
		Atloc  int32
	}
	Dwarves [dungeon.NDWARVES + 1]struct {
		Seen   int32
		Loc    int32
		Oldloc int32
	}
	Objects [dungeon.NOBJECTS + 1]struct {
		Found bool
		Fixed int32
		Prop  int32
		Place int32
	}
	Hints [dungeon.NHINTS]struct {
		Used bool
		Lc   int32
	}
	Link [dungeon.NOBJECTS*2 + 1]int32
}

//TODO: Implement settings

func NewGame() Game {

	newGame := Game{}
	seedval := rand.Int()

	newGame.LcgX = int32(seedval) % LCG_M

	if newGame.LcgX < 0 {
		newGame.LcgX += LCG_M + newGame.LcgX
	}

	for i := 0; i < 5; i++ {
		newGame.Zzword[i] = byte('A' + (26 + getNextLCGValue(newGame.LcgX)%LCG_M))
	}

	newGame.Zzword[1] = '\''
	newGame.Zzword[5] = '\x00'

	return newGame
}

func getNextLCGValue(lcgX int32) int32 {
	return (LCG_A*lcgX + LCG_C) % LCG_M
}
