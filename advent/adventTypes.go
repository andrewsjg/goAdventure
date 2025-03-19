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
	Seedval      int
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

	CARRIED        = -1
	STATE_NOTFOUND = -1
	STATE_FOUND    = 0
)

func (g *Game) Drop(object, where int32) {

	/*  Place an object at a given loc, prefixing it onto the game atloc
	 * list.  Decr game.holdng if the object was being toted. No state
	 * change on the object. */

	if object > dungeon.NOBJECTS {
		g.Objects[object-dungeon.NOBJECTS].Fixed = where
	} else {
		if g.Objects[object].Place == CARRIED {
			if object != int32(dungeon.BIRD) {
				/* The bird has to be weightless.  This ugly
				 * hack (and the corresponding code in the carry
				 * function) brought to you by the fact that
				 * when the bird is caged, we need to be able to
				 * either 'take bird' or 'take cage' and have
				 * the right thing happen.
				 */
				g.Holdng--
			}
			g.Objects[object].Place = where
		}

		if where == int32(dungeon.LOC_NOWHERE) || where == CARRIED {
			return
		}

		g.Link[object] = g.Locs[where].Atloc
		g.Locs[where].Atloc = object

	}
}
