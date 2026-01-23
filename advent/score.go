package advent

import (
	"fmt"
	"os"

	"github.com/andrewsjg/goAdventure/dungeon"
)

// I hate this. Make it part of the game struct?
var mxscr = 0

const (
	EXIT_SUCCESS = 0
)

func (g *Game) terminate(mode Termination) {
	points := g.score(mode)

	// Autosave if enabled
	if err := g.AutoSave(); err != nil {
		if g.Settings.EnableDebug {
			fmt.Printf("DEBUG: Autosave failed: %s\n", err.Error())
		}
	}

	if points+int(g.Trnluz)+1 >= mxscr && g.Trnluz != 0 {
		g.rspeak((int32(dungeon.TOOK_LONG)))
	}

	if points+int(g.Saved)+1 >= mxscr && g.Saved != 0 {
		g.rspeak((int32(dungeon.WITHOUT_SUSPENDS)))
	}

	g.rspeak(int32(dungeon.TOTAL_SCORE), points, mxscr, g.Turns, g.Turns)

	for i := 1; i < dungeon.NCLASSES; i++ {
		if dungeon.Classes[i].Threshold >= points {
			g.speak(dungeon.Classes[i].Message)

			if i < dungeon.NCLASSES {
				nxt := dungeon.Classes[i].Threshold + 1 - points
				g.rspeak(int32(dungeon.NEXT_HIGHER), nxt, nxt)
			} else {
				g.rspeak(int32(dungeon.NO_HIGHER))
			}

			os.Exit(EXIT_SUCCESS)
		}
	}

	g.rspeak(int32(dungeon.OFF_SCALE))
	os.Exit(EXIT_SUCCESS)
}

// GetScore returns the current score without side effects (for UI display)
func (g *Game) GetScore() int {
	return g.calculateScore(false)
}

func (g *Game) score(mode Termination) int {
	score := g.calculateScore(mode == EndGame)

	/* Return to score command if that's where we came from. */
	if mode == ScoreGame {
		g.rspeak(int32(dungeon.GARNERED_POINTS), score, mxscr, g.Turns, g.Turns)
	}

	return score
}

func (g *Game) calculateScore(endGame bool) int {

	score := 0

	/*  The present scoring algorithm is as follows:
	 *     Objective:          Points:        Present total possible:
	 *  Getting well into cave   25                    25
	 *  Each treasure < chest    12                    60
	 *  Treasure chest itself    14                    14
	 *  Each treasure > chest    16                   224
	 *  Surviving             (MAX-NUM)*10             30
	 *  Not quitting              4                     4
	 *  Reaching "game.closng"   25                    25
	 *  "Closed": Quit/Killed    10
	 *            Klutzed        25
	 *            Wrong way      30
	 *            Success        45                    45
	 *  Came to Witt's End        1                     1
	 *  Round out the total       2                     2
	 *                                       TOTAL:   430
	 *  Points can also be deducted for using hints or too many turns, or
	 * for saving intermediate positions. */

	/*  First tally up the treasures.  Must be in building and not broken.
	 *  Give the poor guy 2 points just for finding each treasure. */

	for i := 1; i <= dungeon.NOBJECTS; i++ {
		if !dungeon.Objects[i].Is_Treasure {
			continue
		}

		if dungeon.Objects[i].Inventory != "" {
			k := 12

			if i == dungeon.CHEST {
				k = 14
			}

			if i > dungeon.CHEST {
				k = 16
			}

			if !g.objectIsStashed(i) && !g.objectIsNotFound(i) {
				score += 2
			}

			if g.Objects[i].Place == int32(dungeon.LOC_BUILDING) &&
				g.objectIsFound(i) {
				score += k - 2
			}

			mxscr += k
		}
	}

	/*  Now look at how the player finished and how far they got.  NDEATHS and
	 *  game.numdie tell us how well they survived.  game.dflag will tell us
	 *  if they ever got suitably deep into the cave.  game.closng still
	 *  indicates whether they reached the endgame.  And if they got as far as
	 *  "cave closed" (indicated by "game.closed"), then bonus is zero for
	 *  mundane exits or 133, 134, 135 if he blew it (so to speak). */

	score += (dungeon.NDEATHS - int(g.Numdie)) * 10
	mxscr += dungeon.NDEATHS * 10

	if endGame {
		score += 4
	}
	mxscr += 4

	if g.Dflag != 0 {
		score += 25
	}
	mxscr += 25

	if g.Closing {
		score += 25
	}
	mxscr += 25

	if g.Closed {
		if g.Bonus == None {
			score += 10
		}

		if g.Bonus == Splatter {
			score += 25
		}

		if g.Bonus == Defeat {
			score += 30
		}

		if g.Bonus == Victory {
			score += 45
		}
	}
	mxscr += 45

	/* Did the player come to Witt's End as they should? */

	if g.Objects[dungeon.MAGAZINE].Place == int32(dungeon.LOC_WITTSEND) {
		score += 1
	}
	mxscr += 1

	// Round it off
	score += 2
	mxscr += 2

	/* Deduct for hints/turns/saves. Hints < 4 are special; see database
	 * desc. */

	for i := 0; i < dungeon.NHINTS; i++ {
		if g.Hints[i].Used {
			score = score - dungeon.Hints[i].Penalty
		}
	}

	if g.Novice {
		score -= 5
	}

	if g.Clshnt {
		score -= 10
	}

	score = score - int(g.Trnluz) - int(g.Saved)

	return score
}
