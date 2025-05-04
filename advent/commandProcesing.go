package advent

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/andrewsjg/goAdventure/dungeon"
)

func (g *Game) getCommand(command string) Command {
	cmd := Command{}

	return cmd
}

/* Pre-processes a command input to see if we need to tease out a few specific
 * cases:
 * - "enter water" or "enter stream":
 *   weird specific case that gets the user wet, and then kicks us back to get
 * another command
 * - <object> <verb>:
 *   Irregular form of input, but should be allowed. We switch back to <verb>
 * <object> form for further processing.
 * - "grate":
 *   If in location with grate, we move to that grate. If we're in a number of
 * other places, we move to the entrance.
 * - "water plant", "oil plant", "water door", "oil door":
 *   Change to "pour water" or "pour oil" based on context
 * - "cage bird":
 *   If bird is present, we change to "carry bird"
 *
 * Returns true if pre-processing is complete, and we're ready to move to the
 * primary command processing, false otherwise. */

func (g *Game) preProcessCommand(command string) string {
	return ""
}

func (g *Game) ProcessCommand(command string) error {

	var err error
	//g.Output = fmt.Sprintf("CMD: %s LOC: %d", command, g.Loc)

	cmd := strings.ToUpper(command)

	// Game start condition
	// If this is the start of a new game and the command is yes
	// then the player has asked for instructions. This is kind of a kludge.
	// Can probably improve by using AskQuestion.

	if g.Settings.NewGame && strings.Contains(cmd, "Y") {
		g.Output = dungeon.Arbitrary_Messages[dungeon.CAVE_NEARBY]
		g.Novice = true
		g.Limit = NOVICELIMIT // Numner of turns allowed for a novice player

		// Reset new game flag since the game has now progressed
		g.Settings.NewGame = false

	} else if g.Settings.NewGame && strings.Contains(cmd, "N") {
		g.Output = dungeon.Arbitrary_Messages[dungeon.NO_MESSAGE]
		// Reset new game flag since the game has now progressed
		g.Settings.NewGame = false
		g.DescribeLocation()

	} else if g.Settings.EnableDebug && cmd == "ZZTEST" {

		wantHint := func(response string, game *Game) string {
			if !strings.Contains(strings.ToUpper(response), "Y") {
				err := game.speak(dungeon.Arbitrary_Messages[dungeon.OK_MAN])

				if err != nil {
					fmt.Println("Error: ", err.Error())
					return fmt.Sprintf("Error: %s", err.Error())
				}

				return ""
			}

			game.speak(dungeon.Hints[1].Hint)

			if game.Hints[1].Used && game.Limit > WARNTIME {
				game.Limit += int32(WARNTIME * dungeon.Hints[1].Penalty)
			}

			return dungeon.Hints[1].Hint
		}

		hintQuestion := func(response string, game *Game) string {

			if !strings.Contains(strings.ToUpper(response), "Y") {

				err := game.speak(dungeon.Arbitrary_Messages[dungeon.OK_MAN])

				if err != nil {
					fmt.Println("Error: ", err.Error())
					return fmt.Sprintf("Error: %s", err.Error())
				}

				return ""
			}

			game.rspeak(int32(dungeon.HINT_COST), dungeon.Hints[1].Penalty)

			game.Output = game.Output + "\n\n" + dungeon.Arbitrary_Messages[dungeon.WANT_HINT]

			game.AskQuestion(game.Output, wantHint)

			return response
		}

		g.AskQuestion(dungeon.Hints[1].Question, hintQuestion)

		/*	} else if g.QueryFlag {
			// Game has asked a question. The command will be the response
			fmt.Println("QueryFlag set. Query Response: ", cmd)
			g.QueryResponse = cmd
			g.QueryFlag = false */

	} else {

		// We just got some input from the user

		// Put the input into a command structure
		//cmd := g.getCommand(command)

		if g.Closed {
			/*  If closing time, check for any stashed
			* objects being toted and unstash them.  This
			* way objects won't be described until they've
			* been picked up and put down separate from
			* their respective piles. */

			if (g.objectIsFound(dungeon.OYSTER) || g.objectIsStashed(dungeon.OYSTER)) && g.toting(dungeon.OYSTER) {

				g.pSpeak(int32(dungeon.OYSTER), Look, true, 1)
			}

			for i := 1; i <= dungeon.NOBJECTS; i++ {
				if g.toting(i) && (g.objectIsNotFound(i) || g.objectIsStashed(i)) {
					g.Objects[i].Prop = g.objectStashed(i)
				}
			}
		}

		/* Check to see if the room is dark. */
		g.Wzdark = g.dark()

		/* If the knife is not here it permanently disappears.
		* Possibly this should fire if the knife is here but
		* the room is dark? */

		if g.Knfloc > int32(dungeon.LOC_NOWHERE) && g.Knfloc != g.Loc {
			g.Knfloc = int32(dungeon.LOC_NOWHERE)
		}

		// Check for hints
		g.CheckHints()

		/* Every input, check "foobar" flag. If zero,
		 * nothing's going on. If pos, make neg. If neg,
		 * the player skipped a word, so make it zero.
		 */
		if g.Foobar > 0 {
			g.Foobar = -g.Foobar
		} else {
			g.Foobar = 0
		}

	}

	return err
}

func (g *Game) AskQuestion(query string, callback func(response string, game *Game) string) {
	g.QueryFlag = true
	g.QueryResponse = ""

	g.Output = query
	g.OnQueryResponse = callback

}

// Utility functions

func countWords(input string) int {
	// Use FieldsFunc to split the string based on word boundaries.
	words := strings.FieldsFunc(input, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	return len(words)
}
