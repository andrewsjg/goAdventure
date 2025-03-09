package advent

import "github.com/andrewsjg/goAdventure/dungeon"

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
		Loc    int
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

const (
	LCG_A = 1093
	LCG_C = 221587
	LCG_M = 1048576
)
