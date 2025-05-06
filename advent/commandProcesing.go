package advent

import (
	"strings"
	"unicode"

	"github.com/andrewsjg/goAdventure/dungeon"
)

// This take the command and tokenises it into a command structure.
func (g *Game) tokeniseCommand(command string) Command {
	cmd := Command{}

	if countWords(command) > 2 {

		g.OutputType = 1
		g.rspeak(int32(dungeon.TWO_WORDS))
		return cmd
	}

	/* (ESR) In oldstyle mode, simulate the uppercasing and truncating
	 * effect on raw tokens of packing them into sixbit characters, 5
	 * to a 32-bit word.  This is something the FORTRAN version did
	 * because archaic FORTRAN had no string types.  Don Wood's
	 * mechanical translation of 2.5 to C retained the packing and
	 * thus this misfeature.
	 *
	 * It's philosophically questionable whether this is the right
	 * thing to do even in oldstyle mode.  On one hand, the text
	 * mangling was not authorial intent, but a result of limitations
	 * in their tools. On the other, not simulating this misbehavior
	 * goes against the goal of making oldstyle as accurate as
	 * possible an emulation of the original UI.
	 */

	// Leaving this in here in case I want to implement 'oldstyle' mode later.
	/*
			 if (settings.oldstyle) {
				cmd->word[0].raw[TOKLEN + TOKLEN] =
				    cmd->word[1].raw[TOKLEN + TOKLEN] = '\0';
				for (size_t i = 0; i < strlen(cmd->word[0].raw); i++) {
					cmd->word[0].raw[i] = toupper(cmd->word[0].raw[i]);
				}
				for (size_t i = 0; i < strlen(cmd->word[1].raw); i++) {
					cmd->word[1].raw[i] = toupper(cmd->word[1].raw[i]);
				}
		}
	*/
	words := []Command_Word{}

	for _, word := range SplitWords(command) {
		tmpWord := Command_Word{}
		tmpWord.Raw = strings.ToUpper(word)

		words = append(words, tmpWord)
	}

	cmd.Word = words

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

		cmd := "LOOK STREAM MOUNTAIN"
		g.tokeniseCommand(cmd)

	} else {

		// We just got some input from the user

		// Put the input into a command structure
		cmd := g.tokeniseCommand(command)

		if g.Settings.EnableDebug {
			if len(cmd.Word) == 1 {

				g.Output = cmd.Word[0].Raw
			} else if len(cmd.Word) == 2 {
				g.Output = cmd.Word[0].Raw + " " + cmd.Word[1].Raw
			} else {
				//g.DescribeLocation()
			}
		}

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

func SplitWords(input string) []string {
	// Use FieldsFunc to split the string based on word boundaries.
	words := strings.FieldsFunc(input, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	return words
}
