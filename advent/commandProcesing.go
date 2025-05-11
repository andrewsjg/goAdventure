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

		g.tspeak(int32(dungeon.TWO_WORDS))
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

		tmpWord = getVocabMetaData(tmpWord.Raw)

		words = append(words, tmpWord)
	}

	cmd.Word = words

	return cmd
}

func (g *Game) getVocabMetaData(rawWord string) Command_Word {

	word := Command_Word{}
	word.Raw = rawWord
	word.ID = WORD_EMPTY
	word.WordType = NO_WORD_TYPE

	if rawWord == "" {
		return word
	}

	refNum := getMotionVocabID(rawWord, g.Settings.OldStyle)

	if refNum != WORD_NOT_FOUND {
		word.ID = refNum
		word.WordType = MOTION
		return word
	}

	refNum = getObjectVocabID(rawWord)

	if refNum != WORD_NOT_FOUND {
		word.ID = refNum
		word.WordType = OBJECT
		return word
	}

	refNum = getActionVocabID(rawWord, g.Settings.OldStyle)
	if refNum != WORD_NOT_FOUND {
		word.ID = refNum
		word.WordType = ACTION
		return word
	}

	// TODO: Test this
	if strings.Compare(strings.ToUpper(rawWord), strings.ToUpper(string(g.Zzword[:]))) == 0 {
		word.ID = dungeon.PART
		word.WordType = NUMERIC
		return word
	}

	return word
}

func getMotionVocabID(rawWord string, oldStyle bool) int {
	for i := 0; i < dungeon.NMOTIONS; i++ {
		for j := 0; j < dungeon.Motions[i].Words.N; j++ {
			motionWord := dungeon.Motions[i].Words.Strs[j]
			// Compare up to TOKLEN characters, case-insensitive
			maxLen := TOKLEN
			if len(rawWord) < TOKLEN {
				maxLen = len(rawWord)
			}
			if len(motionWord) < maxLen {
				maxLen = len(motionWord)
			}
			if strings.EqualFold(rawWord[:maxLen], motionWord[:maxLen]) &&
				(len(rawWord) > 1 ||
					!strings.ContainsRune(dungeon.Ignore, rune(rawWord[0])) ||
					!oldStyle) {
				return i
			}
		}
	}

	return WORD_NOT_FOUND
}

func getObjectVocabID(rawWord string) int {
	for i := 0; i < dungeon.NOBJECTS+1; i++ {
		for j := 0; j < dungeon.Objects[i].Words.N; j++ {
			if strnCaseCmpEqual(rawWord, dungeon.Objects[i].Words.Strs[j], TOKLEN) {
				return i
			}
		}
	}

	return WORD_NOT_FOUND
}

func getActionVocabID(rawWord string, oldStyle bool) int {
	for i := 0; i < dungeon.NACTIONS; i++ {
		for j := 0; j < dungeon.Actions[i].Words.N; j++ {
			if strnCaseCmpEqual(rawWord, dungeon.Actions[i].Words.Strs[j], TOKLEN) &&
				(len(rawWord) > 1 || !strings.ContainsRune(dungeon.Ignore, rune(rawWord[0]))) ||
				!oldStyle {

				return i
			}
		}
	}

	return WORD_NOT_FOUND
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

	} else if cmd == "" {
		g.DescribeLocation()

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

func strnCaseCmpEqual(s1, s2 string, n int) bool {

	s1 = strings.ToUpper(s1)
	s2 = strings.ToUpper(s2)

	// Limit the strings to n characters
	if len(s1) > n {
		s1 = s1[:n]
	}
	if len(s2) > n {
		s2 = s2[:n]
	}

	// Compare case-insensitively
	return strings.EqualFold(s1, s2)
}
