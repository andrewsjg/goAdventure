package advent

import (
	"math/rand"

	"github.com/andrewsjg/goAdventure/dungeon"
)

//TODO: Implement settings

func NewGame(seed int) Game {

	newGame := Game{}

	seedval := seed

	if seedval == 0 {
		seedval = rand.Int()
	}

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

	for i := 1; i < dungeon.NOBJECTS; i++ {
		newGame.Objects[i].Place = int32(dungeon.LOC_NOWHERE)
	}

	for i := 1; i < dungeon.NLOCATIONS; i++ {
		if !(dungeon.Locations[i].Description.Big == "" || dungeon.TKey[i] == 0) {
			k := dungeon.TKey[i]

			if dungeon.Travel[k].Motion == dungeon.HERE {
				dungeon.Conditions[i] |= (1 << dungeon.COND_FORCED)
			}
		}
	}

	/*  Set up the game.locs atloc and game.link arrays.
	 *  We'll use the DROP subroutine, which prefaces new objects on the
	 *  lists.  Since we want things in the other order, we'll run the
	 *  loop backwards.  If the object is in two locs, we drop it twice.
	 *  Also, since two-placed objects are typically best described
	 *  last, we'll drop them first. */

	for i := dungeon.NOBJECTS; i >= 1; i-- {
		if dungeon.Objects[i].Fixd > 0 {

			// TODO: Fix these int type casts.
			newGame.Drop(int32(i+dungeon.NOBJECTS), int32(dungeon.Objects[i].Fixd))
			newGame.Drop(int32(i), int32(dungeon.Objects[i].Plac))
		}
	}

	for i := 1; i < dungeon.NOBJECTS; i++ {
		k := dungeon.NOBJECTS + 1 - i
		newGame.Objects[k].Fixed = int32(dungeon.Objects[k].Fixd)
		if dungeon.Objects[k].Plac != 0 && dungeon.Objects[k].Fixd <= 0 {
			newGame.Drop(int32(k), int32(dungeon.Objects[k].Plac))
		}
	}

	/*  Treasure props are initially STATE_NOTFOUND, and are set to
	 *  STATE_FOUND the first time they are described.  game.tally
	 *  keeps track of how many are not yet found, so we know when to
	 *  close the cave.
	 *  (ESR) Non-trreasures are set to STATE_FOUND explicity so we
	 *  don't rely on the value of uninitialized storage. This is to
	 *  make translation to future languages easier. */

	for obj := 1; obj < dungeon.NOBJECTS; obj++ {
		if dungeon.Objects[obj].Is_Treasure {
			newGame.Tally++

			if dungeon.Objects[obj].Inventory != "" {
				newGame.Objects[obj].Prop = STATE_NOTFOUND
			}
		} else {
			newGame.Objects[obj].Prop = STATE_FOUND
		}

		newGame.Conds = setBit(dungeon.COND_HBASE)
	}

	newGame.Seedval = seedval
	return newGame
}

func setBit(bit int32) int32 {
	return 1 << bit
}

func getNextLCGValue(lcgX int32) int32 {
	return (LCG_A*lcgX + LCG_C) % LCG_M
}
