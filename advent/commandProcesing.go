package advent

import (
	"fmt"
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
		tmpWord = g.getVocabMetaData(tmpWord.Raw)

		words = append(words, tmpWord)
	}

	cmd.Word = words

	// Not sure this is going to matter the way we will
	// process commands in this version.
	cmd.CmdState = TOKENISED
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

			if strnCaseCmpEqual(rawWord, dungeon.Motions[i].Words.Strs[j], TOKLEN) &&
				(len(rawWord) > 1 || !strings.Contains(strings.ToUpper(rawWord), strings.ToUpper(dungeon.Ignore)) ||
					!oldStyle) {
				return i
			}
		}
	}

	return WORD_NOT_FOUND
}

/*
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
*/

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
				(len(rawWord) > 1 || !strings.ContainsRune(dungeon.Ignore, rune(rawWord[0])) ||
					!oldStyle) {

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

func (g *Game) preProcessCommand(command *Command) bool {

	if command.Word[0].WordType == MOTION && command.Word[0].ID == dungeon.ENTER &&
		(command.Word[1].ID == dungeon.STREAM || command.Word[1].ID == dungeon.WATER) {

		if g.LiqLoc() == int32(dungeon.WATER) {
			// TODO: t speak or rspeak?
			g.tspeak(int32(dungeon.FEET_WET))

		} else {

			g.rspeak(int32(dungeon.WHERE_QUERY))

		}
	} else {
		if command.Word[0].WordType == OBJECT {
			if command.Word[1].WordType == ACTION {
				stage := command.Word[0]
				command.Word[0] = command.Word[1]
				command.Word[1] = stage
			}

			if command.Word[0].ID == dungeon.GRATE {
				command.Word[0].WordType = MOTION

				if g.Loc == int32(dungeon.LOC_START) ||
					g.Loc == int32(dungeon.LOC_VALLEY) ||
					g.Loc == int32(dungeon.LOC_SLIT) {
					command.Word[0].ID = dungeon.DEPRESSION
				}

				if g.Loc == int32(dungeon.LOC_COBBLE) ||
					g.Loc == int32(dungeon.LOC_AWKWARD) ||
					g.Loc == int32(dungeon.LOC_BIRDCHAMBER) ||
					g.Loc == int32(dungeon.LOC_PITTOP) {
					command.Word[0].ID = dungeon.ENTRANCE
				}

			}

			if (command.Word[0].ID == dungeon.WATER || command.Word[0].ID == dungeon.OIL) &&
				(command.Word[1].ID == dungeon.PLANT || command.Word[1].ID == dungeon.DOOR) {

				if g.at(int32(command.Word[1].ID)) {
					command.Word[1] = command.Word[0]
					command.Word[0].ID = dungeon.POUR
					command.Word[0].WordType = ACTION
					command.Word[0].Raw = "pour"
				}
			}

			if command.Word[0].ID == dungeon.CAGE && command.Word[1].ID == dungeon.BIRD && g.here(dungeon.CAGE) &&
				g.here(dungeon.BIRD) {
				command.Word[0].ID = dungeon.CARRY
				command.Word[0].WordType = ACTION
			}

		}

		/* If no word type is given for the first word, we assume it's a
		* motion. */

		if command.Word[0].WordType == NO_WORD_TYPE {
			command.Word[0].WordType = MOTION
		}

		command.CmdState = PREPROCESSED
		return true
	}

	return false
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

		// Should write this as a test

		// No Word  == 0
		// Motion   == 1
		// Object   == 2
		// Action   == 3
		// Numeric  == 4

		cmd := "Carry stream" // 3
		// cmd := "Enter stream" // 1

		//cmd := "LAMP STREAM" // 2

		tokCmd := g.tokeniseCommand(cmd)

		g.Output = fmt.Sprintf("ZZTEST: %d, %s\n", tokCmd.Word[0].WordType, tokCmd.Word[1].Raw)

	} else if cmd == "" {
		g.DescribeLocation()

	} else {

		// We just got some input from the user

		/*
			if g.Settings.EnableDebug {
				if len(cmd.Word) == 1 {

					g.Output = cmd.Word[0].Raw
				} else if len(cmd.Word) == 2 {
					g.Output = cmd.Word[0].Raw + " " + cmd.Word[1].Raw
				} else {
					//g.DescribeLocation()
				}
			} */

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

		g.Turns++
		// Put the input into a command structure and pre-process
		cmd := g.tokeniseCommand(command)
		g.preProcessCommand(&cmd)

		/* check if game is closed, and exit if it is */
		if g.closeCheck() {
			// TODO: Handle game exit
			return nil
		}

		// Contine to process the command

		// TODO: Revisit

		for cmd.CmdState == PREPROCESSED {
			cmd.CmdState = PROCESSING

			if cmd.Word[0].ID == WORD_NOT_FOUND {
				g.sspeak(dungeon.DONT_KNOW, cmd.Word[0].Raw)
				cmd.Word = []Command_Word{} // Clear the command
			}

			/* Give user hints of shortcuts */
			if strnCaseCmpEqual(cmd.Word[0].Raw, "WEST", len("WEST")) {

				g.Iwest++
				if g.Iwest == 10 {
					g.rspeak(int32(dungeon.W_IS_WEST))
				}
			}

			if strnCaseCmpEqual(cmd.Word[0].Raw, "GO", len("GO")) && cmd.Word[1].ID != WORD_EMPTY {
				g.Igo++
				if g.Igo == 10 {
					g.rspeak(int32(dungeon.GO_UNNEEDED))
				}
			}

			switch cmd.Word[0].WordType {
			case MOTION:
				g.PlayerMove(cmd.Word[0].ID)
				cmd.CmdState = EXECUTED
				continue

			case OBJECT:
				cmd.Part == 0 // unknown
				cmd.Obj = cmd.Word[0].ID
				break 
			case ACTION:
				

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

func (g *Game) closeCheck() bool {

	/* If a turn threshold has been met, apply penalties and tell
	 * the player about it. */

	for i := 0; i < dungeon.NTHRESHOLDS; i++ {
		if g.Turns == int32(dungeon.Turn_Thresholds[i].Threshold+1) {
			g.Trnluz += int32(dungeon.Turn_Thresholds[i].Point_loss)
			g.speak(dungeon.Turn_Thresholds[i].Message)
		}
	}

	if g.Tally == 0 && indeep(g.Loc) && g.Loc != int32(dungeon.LOC_Y2) {
		g.Clock1--
	}

	/*  When the first warning comes, we lock the grate, destroy
	 *  the bridge, kill all the dwarves (and the pirate), remove
	 *  the troll and bear (unless dead), and set "closng" to
	 *  true.  Leave the dragon; too much trouble to move it.
	 *  from now until clock2 runs out, he cannot unlock the
	 *  grate, move to any location outside the cave, or create
	 *  the bridge.  Nor can he be resurrected if he dies.  Note
	 *  that the snake is already gone, since he got to the
	 *  treasure accessible only via the hall of the mountain
	 *  king. Also, he's been in giant room (to get eggs), so we
	 *  can refer to it.  Also also, he's gotten the pearl, so we
	 *  know the bivalve is an oyster.  *And*, the dwarves must
	 *  have been activated, since we've found chest. */

	if g.Clock1 == 0 {
		g.Objects[dungeon.GRATE].Prop = dungeon.GRATE_CLOSED
		g.Objects[dungeon.FISSURE].Prop = dungeon.UNBRIDGED

		for i := 1; i <= dungeon.NDWARVES; i++ {
			g.Dwarves[i].Seen = false
			g.Dwarves[i].Loc = int32(dungeon.LOC_NOWHERE)
		}

		g.destroy(int32(dungeon.TROLL))

		g.move(int32(dungeon.TROLL), IS_FREE)
		g.move(int32(dungeon.TROLL2), int32(dungeon.Objects[dungeon.TROLL].Fixd))
		g.juggle(int32(dungeon.CHAIN))

		if g.Objects[dungeon.BEAR].Prop != dungeon.BEAR_DEAD {
			g.destroy(int32(dungeon.BEAR))
		}

		g.Objects[dungeon.CHAIN].Prop = dungeon.CHAIN_HEAP
		g.Objects[dungeon.CHAIN].Fixed = IS_FREE
		g.Objects[dungeon.AXE].Prop = dungeon.AXE_HERE
		g.Objects[dungeon.AXE].Fixed = IS_FREE
		g.rspeak(int32(dungeon.CAVE_CLOSING))
		g.Clock1 = -1
		g.Closing = true

		return g.Closed
	} else if g.Clock1 < 0 {
		g.Clock2--
	}

	if g.Clock2 == 0 {
		/*  Once the player is panicked, and clock2 has run out, we come here
		 *  to set up the storage room.  The room has two locs,
		 *  hardwired as LOC_NE and LOC_SW.  At the ne end, we
		 *  place empty bottles, a nursery of plants, a bed of
		 *  oysters, a pile of lamps, rods with stars, sleeping
		 *  dwarves, and the player.  At the sw end we place grate over
		 *  treasures, snake pit, covey of caged birds, more rods, and
		 *  pillows.  A mirror stretches across one wall.  Many of the
		 *  objects come from known locations and/or states (e.g. the
		 *  snake is known to have been destroyed and needn't be
		 *  carried away from its old "place"), making the various
		 *  objects be handled differently.  We also drop all other
		 *  objects the player might be carrying (lest he has some which
		 *  could cause trouble, such as the keys).  We describe the
		 *  flash of light and trundle back. */

		g.put(int32(dungeon.BOTTLE), int32(dungeon.LOC_NE), int32(dungeon.EMPTY_BOTTLE))
		g.put(int32(dungeon.PLANT), int32(dungeon.LOC_NE), int32(dungeon.PLANT_THIRSTY))
		g.put(int32(dungeon.OYSTER), int32(dungeon.LOC_NE), int32(STATE_FOUND))
		g.put(int32(dungeon.LAMP), int32(dungeon.LOC_NE), int32(dungeon.LAMP_DARK))
		g.put(int32(dungeon.ROD), int32(dungeon.LOC_NE), int32(STATE_FOUND))
		g.put(int32(dungeon.DWARF), int32(dungeon.LOC_NE), int32(STATE_FOUND))
		g.Loc = int32(dungeon.LOC_NE)
		g.Oldloc = int32(dungeon.LOC_NE)
		g.Newloc = int32(dungeon.LOC_NE)
		/*  Leave the grate with normal (non-negative) property.
		 *  Reuse sign. */

		g.move(int32(dungeon.GRATE), int32(dungeon.LOC_SW))
		g.move(int32(dungeon.SIGN), int32(dungeon.LOC_SW))

		g.put(int32(dungeon.SNAKE), int32(dungeon.LOC_SW), int32(dungeon.SNAKE_CHASED))
		g.put(int32(dungeon.BIRD), int32(dungeon.LOC_SW), int32(dungeon.BIRD_CAGED))
		g.put(int32(dungeon.CAGE), int32(dungeon.LOC_SW), int32(STATE_FOUND))
		g.put(int32(dungeon.ROD2), int32(dungeon.LOC_SW), int32(STATE_FOUND))
		g.put(int32(dungeon.PILLOW), int32(dungeon.LOC_SW), int32(STATE_FOUND))
		g.put(int32(dungeon.MIRROR), int32(dungeon.LOC_NE), int32(STATE_FOUND))
		g.Objects[int32(dungeon.MIRROR)].Fixed = int32(dungeon.LOC_SW)

		for i := 1; i < dungeon.NOBJECTS; i++ {
			if g.toting(i) {
				g.destroy(int32(i))
			}
		}

		g.rspeak(int32(dungeon.CAVE_CLOSED))
		g.Closed = true

		return g.Closed
	}

	g.lampCheck()
	return false
}

func (g *Game) lampCheck() {
	/* Check game limit and lamp timers */
	if g.Objects[dungeon.LAMP].Prop == dungeon.LAMP_BRIGHT {
		g.Limit--
	}

	/*  Another way we can force an end to things is by having the
	 *  lamp give out.  When it gets close, we come here to warn him.
	 *  First following arm checks if the lamp and fresh batteries are
	 *  here, in which case we replace the batteries and continue.
	 *  Second is for other cases of lamp dying.  Even after it goes
	 *  out, he can explore outside for a while if desired. */

	if g.Limit <= WARNTIME {
		if g.here(dungeon.BATTERY) &&
			g.Objects[dungeon.BATTERY].Prop == dungeon.FRESH_BATTERIES &&
			g.here(dungeon.LAMP) {

			g.rspeak(int32(dungeon.REPLACE_BATTERIES))
			g.Objects[dungeon.BATTERY].Prop = dungeon.DEAD_BATTERIES
			g.Limit += BATTERYLIFE
			g.Lmwarn = false

		} else if !g.Lmwarn && g.here(dungeon.LAMP) {
			g.Lmwarn = true
			if g.Objects[dungeon.BATTERY].Prop == dungeon.DEAD_BATTERIES {
				g.rspeak(int32(dungeon.MISSING_BATTERIES))
			} else if g.Objects[dungeon.BATTERY].Place == int32(dungeon.LOC_NOWHERE) {
				g.rspeak(int32(dungeon.LAMP_DIM))
			} else {
				g.rspeak(int32(dungeon.GET_BATTERIES))
			}
		}
	}

	if g.Limit == 0 {
		g.Limit = -1
		g.Objects[dungeon.LAMP].Prop = dungeon.LAMP_DARK

		if g.here(dungeon.LAMP) {
			g.rspeak(int32(dungeon.LAMP_OUT))
		}
	}
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
