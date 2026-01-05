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

	// Guard against empty command or commands with insufficient words
	if len(command.Word) == 0 {
		command.CmdState = PREPROCESSED
		return true
	}

	if len(command.Word) > 1 && command.Word[0].WordType == MOTION && command.Word[0].ID == dungeon.ENTER &&
		(command.Word[1].ID == dungeon.STREAM || command.Word[1].ID == dungeon.WATER) {

		if g.LiqLoc() == int32(dungeon.WATER) {
			// TODO: t speak or rspeak?
			g.tspeak(int32(dungeon.FEET_WET))

		} else {

			g.rspeak(int32(dungeon.WHERE_QUERY))

		}
	} else {
		if command.Word[0].WordType == OBJECT {
			if len(command.Word) > 1 && command.Word[1].WordType == ACTION {
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

			if len(command.Word) > 1 && (command.Word[0].ID == dungeon.WATER || command.Word[0].ID == dungeon.OIL) &&
				(command.Word[1].ID == dungeon.PLANT || command.Word[1].ID == dungeon.DOOR) {

				if g.at(int32(command.Word[1].ID)) {
					command.Word[1] = command.Word[0]
					command.Word[0].ID = dungeon.POUR
					command.Word[0].WordType = ACTION
					command.Word[0].Raw = "pour"
				}
			}

			if len(command.Word) > 1 && command.Word[0].ID == dungeon.CAGE && command.Word[1].ID == dungeon.BIRD && g.here(dungeon.CAGE) &&
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
	// Clear output from previous command
	g.Output = ""

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
		g.ListObjects()

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
		g.ListObjects()

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

			if len(cmd.Word) > 0 && cmd.Word[0].ID == WORD_NOT_FOUND {
				g.sspeak(dungeon.DONT_KNOW, cmd.Word[0].Raw)
				cmd.Word = []Command_Word{} // Clear the command
				continue
			}

			// If command is empty, skip processing
			if len(cmd.Word) == 0 {
				break
			}

			/* Give user hints of shortcuts */
			if strnCaseCmpEqual(cmd.Word[0].Raw, "WEST", len("WEST")) {

				g.Iwest++
				if g.Iwest == 10 {
					g.rspeak(int32(dungeon.W_IS_WEST))
				}
			}

			if len(cmd.Word) > 1 && strnCaseCmpEqual(cmd.Word[0].Raw, "GO", len("GO")) && cmd.Word[1].ID != WORD_EMPTY {
				g.Igo++
				if g.Igo == 10 {
					g.rspeak(int32(dungeon.GO_UNNEEDED))
				}
			}

			switch cmd.Word[0].WordType {
			case MOTION:
				g.PlayerMove(int32(cmd.Word[0].ID))
				cmd.CmdState = EXECUTED
				continue

			case OBJECT:
				cmd.Part = Unknown
				cmd.Obj = cmd.Word[0].ID
				break
			case ACTION:
				if len(cmd.Word) > 1 && cmd.Word[1].WordType == NUMERIC {
					cmd.Part = Transitive
				} else {
					cmd.Part = Intransitive
				}
				cmd.Verb = cmd.Word[0].ID
				break
			case NUMERIC:
				if !g.Settings.OldStyle {
					g.sspeak(dungeon.DONT_KNOW, cmd.Word[0].Raw)
					cmd = Command{}
					continue
				}
				break
			default:
				// Should not happen
				continue
			}

			// Execute the action
			phaseCode := g.action(&cmd)

			switch phaseCode {
			case GO_TERMINATE:
				cmd.CmdState = EXECUTED
				break
			case GO_MOVE:
				g.PlayerMove(int32(dungeon.NUL))
				cmd.CmdState = EXECUTED
				break
			case GO_WORD2:
				// Get second word for analysis
				if len(cmd.Word) > 1 {
					cmd.Word[0] = cmd.Word[1]
					cmd.Word = cmd.Word[:1]
				}
				cmd.CmdState = PREPROCESSED
				break
			case GO_UNKNOWN:
				// Random intransitive verbs come here
				if len(cmd.Word[0].Raw) > 0 {
					cmd.Word[0].Raw = strings.ToUpper(cmd.Word[0].Raw[:1]) + cmd.Word[0].Raw[1:]
				}
				g.sspeak(dungeon.DO_WHAT, cmd.Word[0].Raw)
				cmd.Obj = NO_OBJECT
				cmd.CmdState = GIVEN
				break
			case GO_CHECKHINT:
				cmd.CmdState = GIVEN
				break
			case GO_DWARFWAKE:
				// Oh dear, he's disturbed the dwarves
				g.rspeak(int32(dungeon.DWARVES_AWAKEN))
				g.terminate(EndGame)
				break
			case GO_CLEAROBJ:
				cmd = Command{}
				break
			case GO_TOP:
				break
			default:
				// Unknown phase code
				continue
			}

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

// Action dispatcher and handlers

func (g *Game) action(command *Command) PhaseCode {
	// Analyze a verb. Remember what it was, go back for object if second
	// word unless verb is "say", which snarfs arbitrary second word.

	// Check if this is a no-action verb (just displays a message)
	if command.Verb < dungeon.NACTIONS && dungeon.Actions[command.Verb].Message != "" {
		// Check if this action requires no further processing
		noAction := dungeon.Actions[command.Verb].Words.N == 0
		if noAction {
			g.speak(dungeon.Actions[command.Verb].Message)
			return GO_CLEAROBJ
		}
	}

	if command.Part == Unknown {
		// Analyse an object word. See if the thing is here, whether
		// we've got a verb yet, and so on. Object must be here
		// unless verb is "find" or "invent(ory)" (and no new verb
		// yet to be analysed). Water and oil are also funny, since
		// they are never actually dropped at any location, but might
		// be here inside the bottle or urn or as a feature of the location.

		if g.here(command.Obj) {
			// Object is here, continue
		} else if command.Obj == dungeon.DWARF && g.atDwrf(g.Loc) > 0 {
			// Dwarf is present
		} else if !g.Closed && ((g.liquid() == int32(command.Obj) && g.here(dungeon.BOTTLE)) ||
			int32(command.Obj) == g.LiqLoc()) {
			// Liquid in bottle or at location
		} else if command.Obj == dungeon.OIL && g.here(dungeon.URN) &&
			g.Objects[dungeon.URN].Prop != dungeon.URN_EMPTY {
			command.Obj = dungeon.URN
		} else if command.Obj == dungeon.PLANT && g.at(int32(dungeon.PLANT2)) &&
			g.Objects[dungeon.PLANT2].Prop != dungeon.PLANT_THIRSTY {
			command.Obj = dungeon.PLANT2
		} else if command.Obj == dungeon.KNIFE && g.Knfloc == g.Loc {
			g.Knfloc = -1
			g.rspeak(int32(dungeon.KNIVES_VANISH))
			return GO_CLEAROBJ
		} else if command.Obj == dungeon.ROD && g.here(dungeon.ROD2) {
			command.Obj = dungeon.ROD2
		} else if (command.Verb == dungeon.FIND || command.Verb == dungeon.INVENTORY) &&
			(len(command.Word) < 2 || command.Word[1].ID == WORD_EMPTY || command.Word[1].ID == WORD_NOT_FOUND) {
			// Find or inventory without object
		} else {
			g.sspeak(dungeon.NO_SEE, command.Word[0].Raw)
			return GO_CLEAROBJ
		}

		if command.Verb != 0 {
			command.Part = Transitive
		}
	}

	switch command.Part {
	case Intransitive:
		// Check if there's a second word and verb is not SAY
		if len(command.Word) > 1 && command.Word[1].Raw != "" && command.Verb != dungeon.SAY {
			return GO_WORD2
		}
		if command.Verb == dungeon.SAY {
			// SAY can take anything as object
			if len(command.Word) > 1 && command.Word[1].Raw != "" {
				command.Obj = dungeon.KEYS // Not special, just not NO_OBJECT or INTRANSITIVE
			} else {
				command.Obj = NO_OBJECT
			}
		}
		if command.Obj == NO_OBJECT || command.Obj == INTRANSITIVE {
			// Analyse an intransitive verb (no object given yet)
			return g.handleIntransitiveAction(command)
		}
		fallthrough

	case Transitive:
		// Analyse a transitive verb
		return g.handleTransitiveAction(command)
	}

	return GO_CLEAROBJ
}

func (g *Game) handleIntransitiveAction(command *Command) PhaseCode {
	switch command.Verb {
	case dungeon.CARRY:
		return g.vcarry(command.Verb, INTRANSITIVE)
	case dungeon.DROP:
		return GO_UNKNOWN
	case dungeon.SAY:
		return GO_UNKNOWN
	case dungeon.UNLOCK:
		return g.lock(command.Verb, INTRANSITIVE)
	case dungeon.NOTHING:
		g.rspeak(int32(dungeon.OK_MAN))
		return GO_CLEAROBJ
	case dungeon.LOCK:
		return g.lock(command.Verb, INTRANSITIVE)
	case dungeon.LIGHT:
		return g.light(command.Verb, INTRANSITIVE)
	case dungeon.EXTINGUISH:
		return g.extinguish(command.Verb, INTRANSITIVE)
	case dungeon.WAVE:
		return GO_UNKNOWN
	case dungeon.TAME:
		return GO_UNKNOWN
	case dungeon.GO:
		g.speak(dungeon.Actions[command.Verb].Message)
		return GO_CLEAROBJ
	case dungeon.ATTACK:
		command.Obj = INTRANSITIVE
		return g.attack(command)
	case dungeon.POUR:
		return g.pour(command.Verb, INTRANSITIVE)
	case dungeon.EAT:
		return g.eat(command.Verb, INTRANSITIVE)
	case dungeon.DRINK:
		return g.drink(command.Verb, INTRANSITIVE)
	case dungeon.RUB:
		return GO_UNKNOWN
	case dungeon.THROW:
		return GO_UNKNOWN
	case dungeon.QUIT:
		return g.quit()
	case dungeon.FIND:
		return GO_UNKNOWN
	case dungeon.INVENTORY:
		return g.inven()
	case dungeon.FEED:
		return GO_UNKNOWN
	case dungeon.FILL:
		return g.fill(command.Verb, INTRANSITIVE)
	case dungeon.BLAST:
		g.blast()
		return GO_CLEAROBJ
	case dungeon.SCORE:
		g.score(ScoreGame)
		return GO_CLEAROBJ
	case dungeon.FEE, dungeon.FIE, dungeon.FOE, dungeon.FOO, dungeon.FUM:
		return g.bigwords(command.Word[0].ID)
	case dungeon.BRIEF:
		return g.brief()
	case dungeon.READ:
		command.Obj = INTRANSITIVE
		return g.read(command)
	case dungeon.BREAK:
		return GO_UNKNOWN
	case dungeon.WAKE:
		return GO_UNKNOWN
	case dungeon.SAVE:
		return g.suspend()
	case dungeon.RESUME:
		return g.resume()
	case dungeon.FLY:
		return g.fly(command.Verb, INTRANSITIVE)
	case dungeon.LISTEN:
		return g.listen()
	case dungeon.PART:
		return g.reservoir()
	default:
		// Unknown intransitive verb
		return GO_UNKNOWN
	}
}

func (g *Game) handleTransitiveAction(command *Command) PhaseCode {
	switch command.Verb {
	case dungeon.CARRY:
		return g.vcarry(command.Verb, command.Obj)
	case dungeon.DROP:
		return g.discard(command.Verb, command.Obj)
	case dungeon.SAY:
		return g.say(command)
	case dungeon.UNLOCK:
		return g.lock(command.Verb, command.Obj)
	case dungeon.NOTHING:
		g.rspeak(int32(dungeon.OK_MAN))
		return GO_CLEAROBJ
	case dungeon.LOCK:
		return g.lock(command.Verb, command.Obj)
	case dungeon.LIGHT:
		return g.light(command.Verb, command.Obj)
	case dungeon.EXTINGUISH:
		return g.extinguish(command.Verb, command.Obj)
	case dungeon.WAVE:
		return g.wave(command.Verb, command.Obj)
	case dungeon.TAME:
		g.speak(dungeon.Actions[command.Verb].Message)
		return GO_CLEAROBJ
	case dungeon.GO:
		g.speak(dungeon.Actions[command.Verb].Message)
		return GO_CLEAROBJ
	case dungeon.ATTACK:
		return g.attack(command)
	case dungeon.POUR:
		return g.pour(command.Verb, command.Obj)
	case dungeon.EAT:
		return g.eat(command.Verb, command.Obj)
	case dungeon.DRINK:
		return g.drink(command.Verb, command.Obj)
	case dungeon.RUB:
		return g.rub(command.Verb, command.Obj)
	case dungeon.THROW:
		return g.throwit(command)
	case dungeon.QUIT:
		g.speak(dungeon.Actions[command.Verb].Message)
		return GO_CLEAROBJ
	case dungeon.FIND:
		return g.find(command.Verb, command.Obj)
	case dungeon.INVENTORY:
		return g.find(command.Verb, command.Obj)
	case dungeon.FEED:
		return g.feed(command.Verb, command.Obj)
	case dungeon.FILL:
		return g.fill(command.Verb, command.Obj)
	case dungeon.BLAST:
		g.blast()
		return GO_CLEAROBJ
	case dungeon.SCORE:
		g.speak(dungeon.Actions[command.Verb].Message)
		return GO_CLEAROBJ
	case dungeon.FEE, dungeon.FIE, dungeon.FOE, dungeon.FOO, dungeon.FUM:
		g.speak(dungeon.Actions[command.Verb].Message)
		return GO_CLEAROBJ
	case dungeon.BRIEF:
		g.speak(dungeon.Actions[command.Verb].Message)
		return GO_CLEAROBJ
	case dungeon.READ:
		return g.read(command)
	case dungeon.BREAK:
		return g.vbreak(command.Verb, command.Obj)
	case dungeon.WAKE:
		return g.wake(command.Verb, command.Obj)
	case dungeon.SAVE:
		g.speak(dungeon.Actions[command.Verb].Message)
		return GO_CLEAROBJ
	case dungeon.RESUME:
		g.speak(dungeon.Actions[command.Verb].Message)
		return GO_CLEAROBJ
	case dungeon.FLY:
		return g.fly(command.Verb, command.Obj)
	case dungeon.LISTEN:
		g.speak(dungeon.Actions[command.Verb].Message)
		return GO_CLEAROBJ
	case dungeon.PART:
		return g.reservoir()
	default:
		return GO_UNKNOWN
	}
}

// Action handler implementations

func (g *Game) vcarry(verb, obj int) PhaseCode {
	// Carry an object. Special cases for bird and cage (if bird in cage,
	// can't take one without the other). Liquids also special, since they
	// depend on status of bottle. Also various side effects, etc.
	if obj == INTRANSITIVE {
		// Carry, no object given yet. OK if only one object present.
		if g.Locs[g.Loc].Atloc == NO_OBJECT ||
			g.Link[g.Locs[g.Loc].Atloc] != 0 ||
			g.atDwrf(g.Loc) > 0 {
			return GO_UNKNOWN
		}
		obj = int(g.Locs[g.Loc].Atloc)
	}

	if g.toting(obj) {
		g.speak(dungeon.Actions[verb].Message)
		return GO_CLEAROBJ
	}

	if obj == dungeon.MESSAG {
		g.rspeak(int32(dungeon.REMOVE_MESSAGE))
		g.destroy(int32(dungeon.MESSAG))
		return GO_CLEAROBJ
	}

	if g.Objects[obj].Fixed != IS_FREE {
		switch obj {
		case dungeon.PLANT:
			// Next guard tests whether plant is tiny or stashed
			if g.Objects[dungeon.PLANT].Prop <= dungeon.PLANT_THIRSTY {
				g.rspeak(int32(dungeon.DEEP_ROOTS))
			} else {
				g.rspeak(int32(dungeon.YOU_JOKING))
			}
		case dungeon.BEAR:
			if g.Objects[dungeon.BEAR].Prop == dungeon.SITTING_BEAR {
				g.rspeak(int32(dungeon.BEAR_CHAINED))
			} else {
				g.rspeak(int32(dungeon.YOU_JOKING))
			}
		case dungeon.CHAIN:
			if g.Objects[dungeon.BEAR].Prop != dungeon.UNTAMED_BEAR {
				g.rspeak(int32(dungeon.STILL_LOCKED))
			} else {
				g.rspeak(int32(dungeon.YOU_JOKING))
			}
		case dungeon.RUG:
			if g.Objects[dungeon.RUG].Prop == dungeon.RUG_HOVER {
				g.rspeak(int32(dungeon.RUG_HOVERS))
			} else {
				g.rspeak(int32(dungeon.YOU_JOKING))
			}
		case dungeon.URN:
			g.rspeak(int32(dungeon.URN_NOBUDGE))
		case dungeon.CAVITY:
			g.rspeak(int32(dungeon.DOUGHNUT_HOLES))
		case dungeon.BLOOD:
			g.rspeak(int32(dungeon.FEW_DROPS))
		case dungeon.SIGN:
			g.rspeak(int32(dungeon.HAND_PASSTHROUGH))
		default:
			g.rspeak(int32(dungeon.YOU_JOKING))
		}
		return GO_CLEAROBJ
	}

	if obj == dungeon.WATER || obj == dungeon.OIL {
		if !g.here(dungeon.BOTTLE) || g.liquid() != int32(obj) {
			if !g.toting(dungeon.BOTTLE) {
				g.rspeak(int32(dungeon.NO_CONTAINER))
				return GO_CLEAROBJ
			}
			if g.Objects[dungeon.BOTTLE].Prop == dungeon.EMPTY_BOTTLE {
				return g.fill(verb, dungeon.BOTTLE)
			} else {
				g.rspeak(int32(dungeon.BOTTLE_FULL))
			}
			return GO_CLEAROBJ
		}
		obj = dungeon.BOTTLE
	}

	if g.Holdng >= INVLIMIT {
		g.rspeak(int32(dungeon.CARRY_LIMIT))
		return GO_CLEAROBJ
	}

	if obj == dungeon.BIRD && g.Objects[dungeon.BIRD].Prop != dungeon.BIRD_CAGED &&
		!g.objectIsStashed(dungeon.BIRD) {
		if g.Objects[dungeon.BIRD].Prop == dungeon.BIRD_FOREST_UNCAGED {
			g.destroy(int32(dungeon.BIRD))
			g.rspeak(int32(dungeon.BIRD_CRAP))
			return GO_CLEAROBJ
		}
		if !g.toting(dungeon.CAGE) {
			g.rspeak(int32(dungeon.CANNOT_CARRY))
			return GO_CLEAROBJ
		}
		if g.toting(dungeon.ROD) {
			g.rspeak(int32(dungeon.BIRD_EVADES))
			return GO_CLEAROBJ
		}
		g.Objects[dungeon.BIRD].Prop = dungeon.BIRD_CAGED
	}
	if (obj == dungeon.BIRD || obj == dungeon.CAGE) &&
		(g.Objects[dungeon.BIRD].Prop == dungeon.BIRD_CAGED ||
			g.objectStashed(dungeon.BIRD) == dungeon.BIRD_CAGED) {
		// Expression maps BIRD to CAGE and CAGE to BIRD
		g.carry(int32(dungeon.BIRD+dungeon.CAGE-obj), g.Loc)
	}

	g.carry(int32(obj), g.Loc)

	if obj == dungeon.BOTTLE && g.liquid() != int32(dungeon.NO_OBJECT) {
		g.Objects[int(g.liquid())].Place = CARRIED
	}

	if gstone(obj) && !g.objectIsFound(obj) {
		g.Objects[obj].Found = true
		g.Objects[dungeon.CAVITY].Prop = dungeon.CAVITY_EMPTY
	}
	g.rspeak(int32(dungeon.OK_MAN))
	return GO_CLEAROBJ
}

func (g *Game) discard(verb, obj int) PhaseCode {
	// Discard object. "Throw" also comes here for most objects. Special
	// cases for bird (might attack snake or dragon) and cage (might contain
	// bird) and vase. Drop coins at vending machine for extra batteries.
	if obj == dungeon.ROD && !g.toting(dungeon.ROD) && g.toting(dungeon.ROD2) {
		obj = dungeon.ROD2
	}

	if !g.toting(obj) {
		g.speak(dungeon.Actions[verb].Message)
		return GO_CLEAROBJ
	}

	if gstone(obj) && g.at(int32(dungeon.CAVITY)) &&
		g.Objects[dungeon.CAVITY].Prop != dungeon.CAVITY_FULL {
		g.rspeak(int32(dungeon.GEM_FITS))
		g.Objects[obj].Prop = 1 // STATE_IN_CAVITY
		g.Objects[dungeon.CAVITY].Prop = dungeon.CAVITY_FULL
		if g.here(dungeon.RUG) &&
			((obj == dungeon.EMERALD && g.Objects[dungeon.RUG].Prop != dungeon.RUG_HOVER) ||
				(obj == dungeon.RUBY && g.Objects[dungeon.RUG].Prop == dungeon.RUG_HOVER)) {
			if obj == dungeon.RUBY {
				g.rspeak(int32(dungeon.RUG_SETTLES))
			} else if g.toting(dungeon.RUG) {
				g.rspeak(int32(dungeon.RUG_WIGGLES))
			} else {
				g.rspeak(int32(dungeon.RUG_RISES))
			}
			if !g.toting(dungeon.RUG) || obj == dungeon.RUBY {
				k := int32(dungeon.RUG_FLOOR)
				if g.Objects[dungeon.RUG].Prop == dungeon.RUG_HOVER {
					k = int32(dungeon.RUG_FLOOR)
				} else {
					k = int32(dungeon.RUG_HOVER)
				}
				g.Objects[dungeon.RUG].Prop = k
				var moveK int32
				if k == dungeon.RUG_HOVER {
					moveK = int32(dungeon.Objects[dungeon.SAPPH].Plac)
				} else {
					moveK = 0
				}
				g.move(int32(dungeon.RUG+dungeon.NOBJECTS), moveK)
			}
		}
		g.drop(int32(obj), g.Loc)
		return GO_CLEAROBJ
	}

	if obj == dungeon.COINS && g.here(dungeon.VEND) {
		g.destroy(int32(dungeon.COINS))
		g.drop(int32(dungeon.BATTERY), g.Loc)
		g.pSpeak(int32(dungeon.BATTERY), Look, true, dungeon.FRESH_BATTERIES)
		return GO_CLEAROBJ
	}

	if g.liquid() == int32(obj) {
		obj = dungeon.BOTTLE
	}
	if obj == dungeon.BOTTLE && g.liquid() != int32(dungeon.NO_OBJECT) {
		g.Objects[int(g.liquid())].Place = int32(dungeon.LOC_NOWHERE)
	}

	if obj == dungeon.BEAR && g.at(int32(dungeon.TROLL)) {
		g.stateChange(dungeon.TROLL, dungeon.TROLL_GONE)
		g.move(int32(dungeon.TROLL), int32(dungeon.LOC_NOWHERE))
		g.move(int32(dungeon.TROLL+dungeon.NOBJECTS), IS_FREE)
		g.move(int32(dungeon.TROLL2), int32(dungeon.Objects[dungeon.TROLL].Plac))
		g.move(int32(dungeon.TROLL2+dungeon.NOBJECTS), g.Objects[dungeon.TROLL].Fixed)
		g.juggle(int32(dungeon.CHASM))
		g.drop(int32(obj), g.Loc)
		return GO_CLEAROBJ
	}

	if obj == dungeon.VASE {
		if g.Loc != int32(dungeon.Objects[dungeon.PILLOW].Plac) {
			newProp := dungeon.VASE_DROPPED
			if g.at(int32(dungeon.PILLOW)) {
				newProp = dungeon.VASE_WHOLE
			}
			g.stateChange(dungeon.VASE, int32(newProp))
			if g.Objects[dungeon.VASE].Prop != dungeon.VASE_WHOLE {
				g.Objects[dungeon.VASE].Fixed = IS_FIXED
			}
			g.drop(int32(obj), g.Loc)
			return GO_CLEAROBJ
		}
	}

	if obj == dungeon.CAGE && g.Objects[dungeon.BIRD].Prop == dungeon.BIRD_CAGED {
		g.drop(int32(dungeon.BIRD), g.Loc)
	}

	if obj == dungeon.BIRD {
		if g.at(int32(dungeon.DRAGON)) && g.Objects[dungeon.DRAGON].Prop == dungeon.DRAGON_BARS {
			g.rspeak(int32(dungeon.BIRD_BURNT))
			g.destroy(int32(dungeon.BIRD))
			return GO_CLEAROBJ
		}
		if g.here(dungeon.SNAKE) {
			g.rspeak(int32(dungeon.BIRD_ATTACKS))
			if g.Closed {
				return GO_DWARFWAKE
			}
			g.destroy(int32(dungeon.SNAKE))
			// Set game.prop for use by travel options
			g.Objects[dungeon.SNAKE].Prop = dungeon.SNAKE_CHASED
		} else {
			g.rspeak(int32(dungeon.OK_MAN))
		}

		newProp := int32(dungeon.BIRD_UNCAGED)
		if forest(g.Loc) {
			newProp = int32(dungeon.BIRD_FOREST_UNCAGED)
		}
		g.Objects[dungeon.BIRD].Prop = newProp
		g.drop(int32(obj), g.Loc)
		return GO_CLEAROBJ
	}

	g.rspeak(int32(dungeon.OK_MAN))
	g.drop(int32(obj), g.Loc)
	return GO_CLEAROBJ
}

func (g *Game) inven() PhaseCode {
	hasItems := false
	for i := 1; i <= dungeon.NOBJECTS; i++ {
		if i == dungeon.BEAR || !g.toting(i) {
			continue
		}
		if !hasItems {
			g.rspeak(int32(dungeon.NOW_HOLDING))
			hasItems = true
		}
		g.pSpeak(int32(i), Touch, false, -1)
	}

	if g.toting(dungeon.BEAR) {
		g.rspeak(int32(dungeon.TAME_BEAR))
	}

	if !hasItems {
		g.rspeak(int32(dungeon.NO_CARRY))
	}
	return GO_CLEAROBJ
}

func (g *Game) quit() PhaseCode {
	g.AskQuestion(dungeon.Arbitrary_Messages[dungeon.REALLY_QUIT], func(response string, game *Game) string {
		if strings.ToUpper(response)[0] == 'Y' {
			game.terminate(QuitGame)
		}
		return ""
	})
	return GO_CLEAROBJ
}

func (g *Game) brief() PhaseCode {
	g.Abbnum = 10000
	g.Detail = 3
	g.rspeak(int32(dungeon.BRIEF_CONFIRM))
	return GO_CLEAROBJ
}

func (g *Game) say(command *Command) PhaseCode {
	if len(command.Word) > 1 && command.Word[1].Raw != "" {
		g.Output = command.Word[1].Raw
	}
	return GO_CLEAROBJ
}

// Full implementations of all action handlers

func (g *Game) lock(verb, obj int) PhaseCode {
	// Lock, unlock, no object given. Assume various things if present.
	if obj == INTRANSITIVE {
		if g.here(dungeon.CLAM) {
			obj = dungeon.CLAM
		}
		if g.here(dungeon.OYSTER) {
			obj = dungeon.OYSTER
		}
		if g.at(int32(dungeon.DOOR)) {
			obj = dungeon.DOOR
		}
		if g.at(int32(dungeon.GRATE)) {
			obj = dungeon.GRATE
		}
		if g.here(dungeon.CHAIN) {
			obj = dungeon.CHAIN
		}
		if obj == INTRANSITIVE {
			g.rspeak(int32(dungeon.NOTHING_LOCKED))
			return GO_CLEAROBJ
		}
	}

	// Lock, unlock object. Special stuff for opening clam/oyster and for chain.
	switch obj {
	case dungeon.CHAIN:
		if g.here(dungeon.KEYS) {
			return g.chain(verb)
		} else {
			g.rspeak(int32(dungeon.NO_KEYS))
		}
	case dungeon.GRATE:
		if g.here(dungeon.KEYS) {
			if g.Closing {
				g.rspeak(int32(dungeon.EXIT_CLOSED))
				if !g.Panic {
					g.Clock2 = PANICTIME
				}
				g.Panic = true
			} else {
				if verb == dungeon.LOCK {
					g.stateChange(dungeon.GRATE, dungeon.GRATE_CLOSED)
				} else {
					g.stateChange(dungeon.GRATE, dungeon.GRATE_OPEN)
				}
			}
		} else {
			g.rspeak(int32(dungeon.NO_KEYS))
		}
	case dungeon.CLAM:
		if verb == dungeon.LOCK {
			g.rspeak(int32(dungeon.HUH_MAN))
		} else if g.toting(dungeon.CLAM) {
			g.rspeak(int32(dungeon.DROP_CLAM))
		} else if !g.toting(dungeon.TRIDENT) {
			g.rspeak(int32(dungeon.CLAM_OPENER))
		} else {
			g.destroy(int32(dungeon.CLAM))
			g.drop(int32(dungeon.OYSTER), g.Loc)
			g.drop(int32(dungeon.PEARL), int32(dungeon.LOC_CULDESAC))
			g.rspeak(int32(dungeon.PEARL_FALLS))
		}
	case dungeon.OYSTER:
		if verb == dungeon.LOCK {
			g.rspeak(int32(dungeon.HUH_MAN))
		} else if g.toting(dungeon.OYSTER) {
			g.rspeak(int32(dungeon.DROP_OYSTER))
		} else if !g.toting(dungeon.TRIDENT) {
			g.rspeak(int32(dungeon.OYSTER_OPENER))
		} else {
			g.rspeak(int32(dungeon.OYSTER_OPENS))
		}
	case dungeon.DOOR:
		if g.Objects[dungeon.DOOR].Prop == dungeon.DOOR_UNRUSTED {
			g.rspeak(int32(dungeon.OK_MAN))
		} else {
			g.rspeak(int32(dungeon.RUSTY_DOOR))
		}
	case dungeon.CAGE:
		g.rspeak(int32(dungeon.NO_LOCK))
	case dungeon.KEYS:
		g.rspeak(int32(dungeon.CANNOT_UNLOCK))
	default:
		g.speak(dungeon.Actions[verb].Message)
	}

	return GO_CLEAROBJ
}

func (g *Game) chain(verb int) PhaseCode {
	// Do something to the bear's chain
	if verb != dungeon.LOCK {
		if g.Objects[dungeon.BEAR].Prop == dungeon.UNTAMED_BEAR {
			g.rspeak(int32(dungeon.BEAR_BLOCKS))
			return GO_CLEAROBJ
		}
		if g.Objects[dungeon.CHAIN].Prop == dungeon.CHAIN_HEAP {
			g.rspeak(int32(dungeon.ALREADY_UNLOCKED))
			return GO_CLEAROBJ
		}
		g.Objects[dungeon.CHAIN].Prop = dungeon.CHAIN_HEAP
		g.Objects[dungeon.CHAIN].Fixed = IS_FREE
		if g.Objects[dungeon.BEAR].Prop != dungeon.BEAR_DEAD {
			g.Objects[dungeon.BEAR].Prop = dungeon.CONTENTED_BEAR
		}

		if g.Objects[dungeon.BEAR].Prop == dungeon.BEAR_DEAD {
			g.Objects[dungeon.BEAR].Fixed = IS_FIXED
		} else {
			g.Objects[dungeon.BEAR].Fixed = IS_FREE
		}
		g.rspeak(int32(dungeon.CHAIN_UNLOCKED))
		return GO_CLEAROBJ
	}

	if g.Objects[dungeon.CHAIN].Prop != dungeon.CHAIN_HEAP {
		g.rspeak(int32(dungeon.ALREADY_LOCKED))
		return GO_CLEAROBJ
	}
	if g.Loc != int32(dungeon.Objects[dungeon.CHAIN].Plac) {
		g.rspeak(int32(dungeon.NO_LOCKSITE))
		return GO_CLEAROBJ
	}

	g.Objects[dungeon.CHAIN].Prop = dungeon.CHAIN_FIXED

	if g.toting(dungeon.CHAIN) {
		g.drop(int32(dungeon.CHAIN), g.Loc)
	}
	g.Objects[dungeon.CHAIN].Fixed = IS_FIXED

	g.rspeak(int32(dungeon.CHAIN_LOCKED))
	return GO_CLEAROBJ
}

func (g *Game) light(verb, obj int) PhaseCode {
	// Light. Applicable only to lamp and urn.
	if obj == INTRANSITIVE {
		selects := 0
		if g.here(dungeon.LAMP) && g.Objects[dungeon.LAMP].Prop == dungeon.LAMP_DARK && g.Limit >= 0 {
			obj = dungeon.LAMP
			selects++
		}
		if g.here(dungeon.URN) && g.Objects[dungeon.URN].Prop == dungeon.URN_DARK {
			obj = dungeon.URN
			selects++
		}
		if selects != 1 {
			return GO_UNKNOWN
		}
	}

	switch obj {
	case dungeon.URN:
		state := int32(dungeon.URN_EMPTY)
		if g.Objects[dungeon.URN].Prop != dungeon.URN_EMPTY {
			state = int32(dungeon.URN_LIT)
		}
		g.stateChange(dungeon.URN, state)
	case dungeon.LAMP:
		if g.Limit < 0 {
			g.rspeak(int32(dungeon.LAMP_OUT))
			break
		}
		g.stateChange(dungeon.LAMP, dungeon.LAMP_BRIGHT)
		if g.Wzdark {
			return GO_TOP
		}
	default:
		g.speak(dungeon.Actions[verb].Message)
	}
	return GO_CLEAROBJ
}

func (g *Game) extinguish(verb, obj int) PhaseCode {
	// Extinguish. Lamp, urn, dragon/volcano (nice try).
	if obj == INTRANSITIVE {
		if g.here(dungeon.LAMP) && g.Objects[dungeon.LAMP].Prop == dungeon.LAMP_BRIGHT {
			obj = dungeon.LAMP
		}
		if g.here(dungeon.URN) && g.Objects[dungeon.URN].Prop == dungeon.URN_LIT {
			obj = dungeon.URN
		}
		if obj == INTRANSITIVE {
			return GO_UNKNOWN
		}
	}

	switch obj {
	case dungeon.URN:
		if g.Objects[dungeon.URN].Prop != dungeon.URN_EMPTY {
			g.stateChange(dungeon.URN, dungeon.URN_DARK)
		} else {
			g.pSpeak(int32(dungeon.URN), Change, true, dungeon.URN_DARK)
		}
	case dungeon.LAMP:
		g.stateChange(dungeon.LAMP, dungeon.LAMP_DARK)
		if g.dark() {
			g.rspeak(int32(dungeon.PITCH_DARK))
		} else {
			g.rspeak(int32(dungeon.NO_MESSAGE))
		}
	case dungeon.DRAGON, dungeon.VOLCANO:
		g.rspeak(int32(dungeon.BEYOND_POWER))
	default:
		g.speak(dungeon.Actions[verb].Message)
	}
	return GO_CLEAROBJ
}

func (g *Game) attack(command *Command) PhaseCode {
	// Attack. Assume target if unambiguous. "Throw" also links here.
	verb := command.Verb
	obj := command.Obj

	if obj == INTRANSITIVE {
		changes := 0
		if g.atDwrf(g.Loc) > 0 {
			obj = dungeon.DWARF
			changes++
		}
		if g.here(dungeon.SNAKE) {
			obj = dungeon.SNAKE
			changes++
		}
		if g.at(int32(dungeon.DRAGON)) && g.Objects[dungeon.DRAGON].Prop == dungeon.DRAGON_BARS {
			obj = dungeon.DRAGON
			changes++
		}
		if g.at(int32(dungeon.TROLL)) {
			obj = dungeon.TROLL
			changes++
		}
		if g.at(int32(dungeon.OGRE)) {
			obj = dungeon.OGRE
			changes++
		}
		if g.here(dungeon.BEAR) && g.Objects[dungeon.BEAR].Prop == dungeon.UNTAMED_BEAR {
			obj = dungeon.BEAR
			changes++
		}
		// Check for low-priority targets
		if obj == INTRANSITIVE {
			if g.here(dungeon.BIRD) && verb != dungeon.THROW {
				obj = dungeon.BIRD
				changes++
			}
			if g.here(dungeon.VEND) && verb != dungeon.THROW {
				obj = dungeon.VEND
				changes++
			}
			if g.here(dungeon.CLAM) || g.here(dungeon.OYSTER) {
				obj = dungeon.CLAM
				changes++
			}
		}
		if changes >= 2 {
			return GO_UNKNOWN
		}
	}

	if obj == dungeon.BIRD {
		if g.Closed {
			g.rspeak(int32(dungeon.UNHAPPY_BIRD))
		} else {
			g.destroy(int32(dungeon.BIRD))
			g.rspeak(int32(dungeon.BIRD_DEAD))
		}
		return GO_CLEAROBJ
	}
	if obj == dungeon.VEND {
		newProp := int32(dungeon.VEND_UNBLOCKS)
		if g.Objects[dungeon.VEND].Prop == dungeon.VEND_BLOCKS {
			newProp = int32(dungeon.VEND_UNBLOCKS)
		} else {
			newProp = int32(dungeon.VEND_BLOCKS)
		}
		g.stateChange(dungeon.VEND, newProp)
		return GO_CLEAROBJ
	}

	if obj == dungeon.BEAR {
		switch g.Objects[dungeon.BEAR].Prop {
		case dungeon.UNTAMED_BEAR:
			g.rspeak(int32(dungeon.BEAR_HANDS))
		case dungeon.SITTING_BEAR, dungeon.CONTENTED_BEAR:
			g.rspeak(int32(dungeon.BEAR_CONFUSED))
		case dungeon.BEAR_DEAD:
			g.rspeak(int32(dungeon.ALREADY_DEAD))
		}
		return GO_CLEAROBJ
	}

	if obj == dungeon.DRAGON && g.Objects[dungeon.DRAGON].Prop == dungeon.DRAGON_BARS {
		g.rspeak(int32(dungeon.BARE_HANDS_QUERY))
		// Simplified yes/no - in real implementation you'd ask the user
		g.stateChange(dungeon.DRAGON, int32(dungeon.DRAGON_DEAD))
		g.Objects[dungeon.RUG].Prop = int32(dungeon.RUG_FLOOR)
		g.move(int32(dungeon.DRAGON+dungeon.NOBJECTS), IS_FIXED)
		g.move(int32(dungeon.RUG+dungeon.NOBJECTS), IS_FREE)
		g.move(int32(dungeon.DRAGON), int32(dungeon.LOC_SECRET5))
		g.move(int32(dungeon.RUG), int32(dungeon.LOC_SECRET5))
		g.drop(int32(dungeon.BLOOD), int32(dungeon.LOC_SECRET5))
		for i := 1; i <= dungeon.NOBJECTS; i++ {
			if g.Objects[i].Place == int32(dungeon.Objects[dungeon.DRAGON].Plac) ||
				g.Objects[i].Place == g.Objects[dungeon.DRAGON].Fixed {
				g.move(int32(i), int32(dungeon.LOC_SECRET5))
			}
		}
		g.Loc = int32(dungeon.LOC_SECRET5)
		return GO_MOVE
	}

	if obj == dungeon.OGRE {
		g.rspeak(int32(dungeon.OGRE_DODGE))
		if g.atDwrf(g.Loc) == 0 {
			return GO_CLEAROBJ
		}
		g.rspeak(int32(dungeon.KNIFE_THROWN))
		g.destroy(int32(dungeon.OGRE))
		dwarves := 0
		for i := 1; i < PIRATE; i++ {
			if g.Dwarves[i].Loc == g.Loc {
				dwarves++
				g.Dwarves[i].Loc = int32(dungeon.LOC_LONGWEST)
				g.Dwarves[i].Seen = false
			}
		}
		if dwarves > 1 {
			g.rspeak(int32(dungeon.OGRE_PANIC1))
		} else {
			g.rspeak(int32(dungeon.OGRE_PANIC2))
		}
		return GO_CLEAROBJ
	}

	switch obj {
	case INTRANSITIVE:
		g.rspeak(int32(dungeon.NO_TARGET))
	case dungeon.CLAM, dungeon.OYSTER:
		g.rspeak(int32(dungeon.SHELL_IMPERVIOUS))
	case dungeon.SNAKE:
		g.rspeak(int32(dungeon.SNAKE_WARNING))
	case dungeon.DWARF:
		if g.Closed {
			return GO_DWARFWAKE
		}
		g.rspeak(int32(dungeon.BARE_HANDS_QUERY))
	case dungeon.DRAGON:
		g.rspeak(int32(dungeon.ALREADY_DEAD))
	case dungeon.TROLL:
		g.rspeak(int32(dungeon.ROCKY_TROLL))
	default:
		g.speak(dungeon.Actions[verb].Message)
	}
	return GO_CLEAROBJ
}

func (g *Game) pour(verb, obj int) PhaseCode {
	// Pour. If no object, or object is bottle, assume contents of bottle.
	if obj == dungeon.BOTTLE || obj == INTRANSITIVE {
		obj = int(g.liquid())
	}
	if obj == NO_OBJECT {
		return GO_UNKNOWN
	}
	if !g.toting(obj) {
		g.speak(dungeon.Actions[verb].Message)
		return GO_CLEAROBJ
	}

	if obj != dungeon.OIL && obj != dungeon.WATER {
		g.rspeak(int32(dungeon.CANT_POUR))
		return GO_CLEAROBJ
	}
	if g.here(dungeon.URN) && g.Objects[dungeon.URN].Prop == dungeon.URN_EMPTY {
		return g.fill(verb, dungeon.URN)
	}
	g.Objects[dungeon.BOTTLE].Prop = dungeon.EMPTY_BOTTLE
	g.Objects[obj].Place = int32(dungeon.LOC_NOWHERE)
	if !(g.at(int32(dungeon.PLANT)) || g.at(int32(dungeon.DOOR))) {
		g.rspeak(int32(dungeon.GROUND_WET))
		return GO_CLEAROBJ
	}
	if !g.at(int32(dungeon.DOOR)) {
		if obj == dungeon.WATER {
			// Cycle through the three plant states
			g.stateChange(dungeon.PLANT, (g.Objects[dungeon.PLANT].Prop+1)%3)
			g.Objects[dungeon.PLANT2].Prop = g.Objects[dungeon.PLANT].Prop
			return GO_MOVE
		} else {
			g.rspeak(int32(dungeon.SHAKING_LEAVES))
			return GO_CLEAROBJ
		}
	} else {
		newState := int32(dungeon.DOOR_RUSTED)
		if obj == dungeon.OIL {
			newState = int32(dungeon.DOOR_UNRUSTED)
		}
		g.stateChange(dungeon.DOOR, newState)
		return GO_CLEAROBJ
	}
}

func (g *Game) eat(verb, obj int) PhaseCode {
	// Eat. Intransitive: assume food if present, else ask what.
	switch obj {
	case INTRANSITIVE:
		if !g.here(dungeon.FOOD) {
			return GO_UNKNOWN
		}
		fallthrough
	case dungeon.FOOD:
		g.destroy(int32(dungeon.FOOD))
		g.rspeak(int32(dungeon.THANKS_DELICIOUS))
	case dungeon.BIRD, dungeon.SNAKE, dungeon.CLAM, dungeon.OYSTER, dungeon.DWARF, dungeon.DRAGON, dungeon.TROLL, dungeon.BEAR, dungeon.OGRE:
		g.rspeak(int32(dungeon.LOST_APPETITE))
	default:
		g.speak(dungeon.Actions[verb].Message)
	}
	return GO_CLEAROBJ
}

func (g *Game) drink(verb, obj int) PhaseCode {
	// Drink. If no object, assume water and look for it here.
	if obj == INTRANSITIVE && g.LiqLoc() != int32(dungeon.WATER) &&
		(g.liquid() != int32(dungeon.WATER) || !g.here(dungeon.BOTTLE)) {
		return GO_UNKNOWN
	}

	if obj == dungeon.BLOOD {
		g.destroy(int32(dungeon.BLOOD))
		g.stateChange(dungeon.DRAGON, dungeon.DRAGON_BLOODLESS)
		g.Blooded = true
		return GO_CLEAROBJ
	}

	if obj != INTRANSITIVE && obj != dungeon.WATER {
		g.rspeak(int32(dungeon.RIDICULOUS_ATTEMPT))
		return GO_CLEAROBJ
	}
	if g.liquid() == int32(dungeon.WATER) && g.here(dungeon.BOTTLE) {
		g.Objects[dungeon.WATER].Place = int32(dungeon.LOC_NOWHERE)
		g.stateChange(dungeon.BOTTLE, dungeon.EMPTY_BOTTLE)
		return GO_CLEAROBJ
	}

	g.speak(dungeon.Actions[verb].Message)
	return GO_CLEAROBJ
}

func (g *Game) rub(verb, obj int) PhaseCode {
	// Rub. Yields various snide remarks except for lit urn.
	if obj == dungeon.URN && g.Objects[dungeon.URN].Prop == dungeon.URN_LIT {
		g.destroy(int32(dungeon.URN))
		g.drop(int32(dungeon.AMBER), g.Loc)
		g.Objects[dungeon.AMBER].Prop = dungeon.AMBER_IN_ROCK
		g.Tally--
		g.drop(int32(dungeon.CAVITY), g.Loc)
		g.rspeak(int32(dungeon.URN_GENIES))
	} else if obj != dungeon.LAMP {
		g.rspeak(int32(dungeon.PECULIAR_NOTHING))
	} else {
		g.speak(dungeon.Actions[verb].Message)
	}
	return GO_CLEAROBJ
}

func (g *Game) throwit(command *Command) PhaseCode {
	// Throw. Same as discard unless axe.
	if !g.toting(command.Obj) {
		g.speak(dungeon.Actions[command.Verb].Message)
		return GO_CLEAROBJ
	}
	if dungeon.Objects[command.Obj].Is_Treasure && g.at(int32(dungeon.TROLL)) {
		// Snarf a treasure for the troll
		g.drop(int32(command.Obj), int32(dungeon.LOC_NOWHERE))
		g.move(int32(dungeon.TROLL), int32(dungeon.LOC_NOWHERE))
		g.move(int32(dungeon.TROLL+dungeon.NOBJECTS), IS_FREE)
		g.drop(int32(dungeon.TROLL2), int32(dungeon.Objects[dungeon.TROLL].Plac))
		g.drop(int32(dungeon.TROLL2+dungeon.NOBJECTS), g.Objects[dungeon.TROLL].Fixed)
		g.juggle(int32(dungeon.CHASM))
		g.rspeak(int32(dungeon.TROLL_SATISFIED))
		return GO_CLEAROBJ
	}
	if command.Obj == dungeon.FOOD && g.here(dungeon.BEAR) {
		// Throwing food is another story
		command.Obj = dungeon.BEAR
		return g.feed(command.Verb, command.Obj)
	}
	if command.Obj != dungeon.AXE {
		return g.discard(command.Verb, command.Obj)
	} else {
		if g.atDwrf(g.Loc) <= 0 {
			if g.at(int32(dungeon.DRAGON)) && g.Objects[dungeon.DRAGON].Prop == dungeon.DRAGON_BARS {
				g.rspeak(int32(dungeon.DRAGON_SCALES))
				g.drop(int32(dungeon.AXE), g.Loc)
				return GO_MOVE
			}
			if g.at(int32(dungeon.TROLL)) {
				g.rspeak(int32(dungeon.TROLL_RETURNS))
				g.drop(int32(dungeon.AXE), g.Loc)
				return GO_MOVE
			}
			if g.at(int32(dungeon.OGRE)) {
				g.rspeak(int32(dungeon.OGRE_DODGE))
				g.drop(int32(dungeon.AXE), g.Loc)
				return GO_MOVE
			}
			if g.here(dungeon.BEAR) && g.Objects[dungeon.BEAR].Prop == dungeon.UNTAMED_BEAR {
				g.drop(int32(dungeon.AXE), g.Loc)
				g.Objects[dungeon.AXE].Fixed = IS_FIXED
				g.juggle(int32(dungeon.BEAR))
				g.stateChange(dungeon.AXE, dungeon.AXE_LOST)
				return GO_CLEAROBJ
			}
			command.Obj = INTRANSITIVE
			return g.attack(command)
		}

		// Throw axe at dwarf
		if g.randRange(dungeon.NDWARVES+1) < g.Dflag {
			g.rspeak(int32(dungeon.DWARF_DODGES))
		} else {
			i := g.atDwrf(g.Loc)
			g.Dwarves[i].Seen = false
			g.Dwarves[i].Loc = int32(dungeon.LOC_NOWHERE)
			g.Dkill++
			if g.Dkill == 1 {
				g.rspeak(int32(dungeon.DWARF_SMOKE))
			} else {
				g.rspeak(int32(dungeon.KILLED_DWARF))
			}
		}
		g.drop(int32(dungeon.AXE), g.Loc)
		return GO_MOVE
	}
}

func (g *Game) find(verb, obj int) PhaseCode {
	// Find. Might be carrying it, or it might be here.
	if g.toting(obj) {
		g.rspeak(int32(dungeon.ALREADY_CARRYING))
		return GO_CLEAROBJ
	}

	if g.Closed {
		g.rspeak(int32(dungeon.NEEDED_NEARBY))
		return GO_CLEAROBJ
	}

	if g.at(int32(obj)) || (g.liquid() == int32(obj) && g.at(int32(dungeon.BOTTLE))) ||
		int32(obj) == g.LiqLoc() || (obj == dungeon.DWARF && g.atDwrf(g.Loc) > 0) {
		g.rspeak(int32(dungeon.YOU_HAVEIT))
		return GO_CLEAROBJ
	}

	g.speak(dungeon.Actions[verb].Message)
	return GO_CLEAROBJ
}

func (g *Game) feed(verb, obj int) PhaseCode {
	// Feed. If bird, no seed. Snake, dragon, troll: quip. If dwarf, make him mad. Bear, special.
	switch obj {
	case dungeon.BIRD:
		g.rspeak(int32(dungeon.BIRD_PINING))
	case dungeon.DRAGON:
		if g.Objects[dungeon.DRAGON].Prop != dungeon.DRAGON_BARS {
			g.rspeak(int32(dungeon.RIDICULOUS_ATTEMPT))
		} else {
			g.rspeak(int32(dungeon.NOTHING_EDIBLE))
		}
	case dungeon.SNAKE:
		if !g.Closed && g.here(dungeon.BIRD) {
			g.destroy(int32(dungeon.BIRD))
			g.rspeak(int32(dungeon.BIRD_DEVOURED))
		} else {
			g.rspeak(int32(dungeon.NOTHING_EDIBLE))
		}
	case dungeon.TROLL:
		g.rspeak(int32(dungeon.TROLL_VICES))
	case dungeon.DWARF:
		if g.here(dungeon.FOOD) {
			g.Dflag += 2
			g.rspeak(int32(dungeon.REALLY_MAD))
		} else {
			g.speak(dungeon.Actions[verb].Message)
		}
	case dungeon.BEAR:
		if g.Objects[dungeon.BEAR].Prop == dungeon.BEAR_DEAD {
			g.rspeak(int32(dungeon.RIDICULOUS_ATTEMPT))
			break
		}
		if g.Objects[dungeon.BEAR].Prop == dungeon.UNTAMED_BEAR {
			if g.here(dungeon.FOOD) {
				g.destroy(int32(dungeon.FOOD))
				g.Objects[dungeon.AXE].Fixed = IS_FREE
				g.Objects[dungeon.AXE].Prop = dungeon.AXE_HERE
				g.stateChange(dungeon.BEAR, dungeon.SITTING_BEAR)
			} else {
				g.rspeak(int32(dungeon.NOTHING_EDIBLE))
			}
			break
		}
		g.speak(dungeon.Actions[verb].Message)
	case dungeon.OGRE:
		if g.here(dungeon.FOOD) {
			g.rspeak(int32(dungeon.OGRE_FULL))
		} else {
			g.speak(dungeon.Actions[verb].Message)
		}
	default:
		g.rspeak(int32(dungeon.AM_GAME))
	}
	return GO_CLEAROBJ
}

func (g *Game) fill(verb, obj int) PhaseCode {
	// Fill. Bottle or urn must be empty, and liquid available.
	if obj == dungeon.VASE {
		if g.LiqLoc() == int32(dungeon.NO_OBJECT) {
			g.rspeak(int32(dungeon.FILL_INVALID))
			return GO_CLEAROBJ
		}
		if !g.toting(dungeon.VASE) {
			g.rspeak(int32(dungeon.ARENT_CARRYING))
			return GO_CLEAROBJ
		}
		g.rspeak(int32(dungeon.SHATTER_VASE))
		g.Objects[dungeon.VASE].Prop = dungeon.VASE_BROKEN
		g.Objects[dungeon.VASE].Fixed = IS_FIXED
		g.drop(int32(dungeon.VASE), g.Loc)
		return GO_CLEAROBJ
	}

	if obj == dungeon.URN {
		if g.Objects[dungeon.URN].Prop != dungeon.URN_EMPTY {
			g.rspeak(int32(dungeon.FULL_URN))
			return GO_CLEAROBJ
		}
		if !g.here(dungeon.BOTTLE) {
			g.rspeak(int32(dungeon.FILL_INVALID))
			return GO_CLEAROBJ
		}
		k := int(g.liquid())
		switch k {
		case dungeon.WATER:
			g.Objects[dungeon.BOTTLE].Prop = dungeon.EMPTY_BOTTLE
			g.rspeak(int32(dungeon.WATER_URN))
		case dungeon.OIL:
			g.Objects[dungeon.URN].Prop = dungeon.URN_DARK
			g.Objects[dungeon.BOTTLE].Prop = dungeon.EMPTY_BOTTLE
			g.rspeak(int32(dungeon.OIL_URN))
		case dungeon.NO_OBJECT:
			fallthrough
		default:
			g.rspeak(int32(dungeon.FILL_INVALID))
			return GO_CLEAROBJ
		}
		g.Objects[k].Place = int32(dungeon.LOC_NOWHERE)
		return GO_CLEAROBJ
	}
	if obj != INTRANSITIVE && obj != dungeon.BOTTLE {
		g.speak(dungeon.Actions[verb].Message)
		return GO_CLEAROBJ
	}
	if obj == INTRANSITIVE && !g.here(dungeon.BOTTLE) {
		return GO_UNKNOWN
	}

	if g.here(dungeon.URN) && g.Objects[dungeon.URN].Prop != dungeon.URN_EMPTY {
		g.rspeak(int32(dungeon.URN_NOPOUR))
		return GO_CLEAROBJ
	}
	if g.liquid() != int32(dungeon.NO_OBJECT) {
		g.rspeak(int32(dungeon.BOTTLE_FULL))
		return GO_CLEAROBJ
	}
	if g.LiqLoc() == int32(dungeon.NO_OBJECT) {
		g.rspeak(int32(dungeon.NO_LIQUID))
		return GO_CLEAROBJ
	}

	newProp := int32(dungeon.WATER_BOTTLE)
	if g.LiqLoc() == int32(dungeon.OIL) {
		newProp = int32(dungeon.OIL_BOTTLE)
	}
	g.stateChange(dungeon.BOTTLE, newProp)
	if g.toting(dungeon.BOTTLE) {
		g.Objects[int(g.liquid())].Place = CARRIED
	}
	return GO_CLEAROBJ
}

func (g *Game) blast() {
	// Blast. No effect unless you've got dynamite, which is a neat trick!
	if g.objectIsNotFound(dungeon.ROD2) || !g.Closed {
		g.rspeak(int32(dungeon.REQUIRES_DYNAMITE))
	} else {
		if g.here(dungeon.ROD2) {
			g.Bonus = Splatter
			g.rspeak(int32(dungeon.SPLATTER_MESSAGE))
		} else if g.Loc == int32(dungeon.LOC_NE) {
			g.Bonus = Defeat
			g.rspeak(int32(dungeon.DEFEAT_MESSAGE))
		} else {
			g.Bonus = Victory
			g.rspeak(int32(dungeon.VICTORY_MESSAGE))
		}
		g.terminate(EndGame)
	}
}

func (g *Game) bigwords(word int) PhaseCode {
	// Only called on FEE FIE FOE FOO (AND FUM). Advance to next state if given in proper order.
	foobar := g.Foobar
	if foobar < 0 {
		foobar = -foobar
	}

	// Only FEE can start a magic-word sequence
	if foobar == WORD_EMPTY && (word == dungeon.FIE || word == dungeon.FOE || word == dungeon.FOO || word == dungeon.FUM) {
		g.rspeak(int32(dungeon.NOTHING_HAPPENS))
		return GO_CLEAROBJ
	}

	if (foobar == WORD_EMPTY && word == dungeon.FEE) ||
		(foobar == int32(dungeon.FEE) && word == dungeon.FIE) ||
		(foobar == int32(dungeon.FIE) && word == dungeon.FOE) ||
		(foobar == int32(dungeon.FOE) && word == dungeon.FOO) {
		g.Foobar = int32(word)
		if word != dungeon.FOO {
			g.rspeak(int32(dungeon.OK_MAN))
			return GO_CLEAROBJ
		}
		g.Foobar = WORD_EMPTY
		if g.Objects[dungeon.EGGS].Place == int32(dungeon.Objects[dungeon.EGGS].Plac) ||
			(g.toting(dungeon.EGGS) && g.Loc == int32(dungeon.Objects[dungeon.EGGS].Plac)) {
			g.rspeak(int32(dungeon.NOTHING_HAPPENS))
			return GO_CLEAROBJ
		} else {
			// Bring back troll if we steal the eggs back from him before crossing
			if g.Objects[dungeon.EGGS].Place == int32(dungeon.LOC_NOWHERE) &&
				g.Objects[dungeon.TROLL].Place == int32(dungeon.LOC_NOWHERE) &&
				g.Objects[dungeon.TROLL].Prop == dungeon.TROLL_UNPAID {
				g.Objects[dungeon.TROLL].Prop = dungeon.TROLL_PAIDONCE
			}
			if g.here(dungeon.EGGS) {
				g.pSpeak(int32(dungeon.EGGS), Look, true, dungeon.EGGS_VANISHED)
			} else if g.Loc == int32(dungeon.Objects[dungeon.EGGS].Plac) {
				g.pSpeak(int32(dungeon.EGGS), Look, true, dungeon.EGGS_HERE)
			} else {
				g.pSpeak(int32(dungeon.EGGS), Look, true, dungeon.EGGS_DONE)
			}
			g.move(int32(dungeon.EGGS), int32(dungeon.Objects[dungeon.EGGS].Plac))
			return GO_CLEAROBJ
		}
	} else {
		// Magic-word sequence was started but is incorrect
		if g.Settings.OldStyle || g.Seenbigwords {
			g.rspeak(int32(dungeon.START_OVER))
		} else {
			g.rspeak(int32(dungeon.WELL_POINTLESS))
		}
		g.Foobar = WORD_EMPTY
		return GO_CLEAROBJ
	}
}

func (g *Game) read(command *Command) PhaseCode {
	// Read. Print stuff based on objtxt. Oyster (?) is special case.
	if command.Obj == INTRANSITIVE {
		command.Obj = NO_OBJECT
		for i := 1; i <= dungeon.NOBJECTS; i++ {
			if g.here(i) && len(dungeon.Objects[i].Texts) > 0 && !g.objectIsStashed(i) {
				command.Obj = command.Obj*dungeon.NOBJECTS + i
			}
		}
		if command.Obj > dungeon.NOBJECTS || command.Obj == NO_OBJECT || g.dark() {
			return GO_UNKNOWN
		}
	}

	if g.dark() {
		g.sspeak(dungeon.NO_SEE, command.Word[0].Raw)
	} else if command.Obj == dungeon.OYSTER {
		if !g.toting(dungeon.OYSTER) || !g.Closed {
			g.rspeak(int32(dungeon.DONT_UNDERSTAND))
		} else if !g.Clshnt {
			// Ask for clue
			g.Clshnt = true
			g.rspeak(int32(dungeon.WAYOUT_CLUE))
		} else {
			g.pSpeak(int32(dungeon.OYSTER), Hear, true, 1)
		}
	} else if len(dungeon.Objects[command.Obj].Texts) == 0 || g.objectIsNotFound(command.Obj) {
		g.speak(dungeon.Actions[command.Verb].Message)
	} else {
		g.pSpeak(int32(command.Obj), Study, true, g.Objects[command.Obj].Prop)
	}
	return GO_CLEAROBJ
}

func (g *Game) vbreak(verb, obj int) PhaseCode {
	// Break. Only works for mirror in repository and, of course, the vase.
	switch obj {
	case dungeon.MIRROR:
		if g.Closed {
			g.stateChange(dungeon.MIRROR, dungeon.MIRROR_BROKEN)
			return GO_DWARFWAKE
		} else {
			g.rspeak(int32(dungeon.TOO_FAR))
		}
	case dungeon.VASE:
		if g.Objects[dungeon.VASE].Prop == dungeon.VASE_WHOLE {
			if g.toting(dungeon.VASE) {
				g.drop(int32(dungeon.VASE), g.Loc)
			}
			g.stateChange(dungeon.VASE, dungeon.VASE_BROKEN)
			g.Objects[dungeon.VASE].Fixed = IS_FIXED
			break
		}
		fallthrough
	default:
		g.speak(dungeon.Actions[verb].Message)
	}
	return GO_CLEAROBJ
}

func (g *Game) wake(verb, obj int) PhaseCode {
	// Wake. Only use is to disturb the dwarves.
	if obj != dungeon.DWARF || !g.Closed {
		g.speak(dungeon.Actions[verb].Message)
		return GO_CLEAROBJ
	} else {
		g.rspeak(int32(dungeon.PROD_DWARF))
		return GO_DWARFWAKE
	}
}

func (g *Game) suspend() PhaseCode {
	// Suspend the game by saving and exiting
	// Warn the player about the penalty
	g.rspeak(int32(dungeon.SUSPEND_WARNING))

	// Set up confirmation query
	g.AskQuestion(
		dungeon.Arbitrary_Messages[dungeon.THIS_ACCEPTABLE],
		func(response string, game *Game) string {
			response = strings.ToUpper(strings.TrimSpace(response))
			if !strings.HasPrefix(response, "Y") {
				game.rspeak(int32(dungeon.OK_MAN))
				return game.Output
			}

			// Charge 5 points for saving
			game.Saved += 5

			// Ask for filename
			game.AskQuestion(
				"File name: ",
				func(filename string, game *Game) string {
					filename = strings.TrimSpace(filename)
					if filename == "" {
						filename = "advent.sav"
					}

					// Save the game
					err := game.SaveToFile(filename)
					if err != nil {
						game.Output = fmt.Sprintf("Failed to save game: %s\nTry again.", err.Error())
						return game.Output
					}

					// Save successful
					game.rspeak(int32(dungeon.RESUME_HELP))
					game.terminate(EndGame)
					return game.Output
				},
			)
			return game.Output
		},
	)

	return GO_CLEAROBJ
}

func (g *Game) resume() PhaseCode {
	// Resume a saved game
	// Check if we're at the start of the game
	if g.Loc != int32(dungeon.LOC_START) || g.Locs[dungeon.LOC_START].Abbrev != 1 {
		// Not at start, ask for confirmation
		g.rspeak(int32(dungeon.RESUME_ABANDON))
		g.AskQuestion(
			dungeon.Arbitrary_Messages[dungeon.THIS_ACCEPTABLE],
			func(response string, game *Game) string {
				response = strings.ToUpper(strings.TrimSpace(response))
				if !strings.HasPrefix(response, "Y") {
					game.rspeak(int32(dungeon.OK_MAN))
					return game.Output
				}

				// Ask for filename
				game.AskQuestion(
					"File name: ",
					func(filename string, game *Game) string {
						filename = strings.TrimSpace(filename)
						if filename == "" {
							filename = "advent.sav"
						}

						// Load the game
						err := game.LoadFromFile(filename)
						if err != nil {
							game.Output = fmt.Sprintf("Failed to load game: %s\nTry again.", err.Error())
							return game.Output
						}

						// Load successful, describe the location
						game.DescribeLocation()
						game.ListObjects()
						return game.Output
					},
				)
				return game.Output
			},
		)
	} else {
		// At start, just ask for filename
		g.AskQuestion(
			"File name: ",
			func(filename string, game *Game) string {
				filename = strings.TrimSpace(filename)
				if filename == "" {
					filename = "advent.sav"
				}

				// Load the game
				err := game.LoadFromFile(filename)
				if err != nil {
					game.Output = fmt.Sprintf("Failed to load game: %s\nTry again.", err.Error())
					return game.Output
				}

				// Load successful, describe the location
				game.DescribeLocation()
				game.ListObjects()
				return game.Output
			},
		)
	}

	return GO_CLEAROBJ
}

func (g *Game) fly(verb, obj int) PhaseCode {
	// Fly. Snide remarks unless hovering rug is here.
	if obj == INTRANSITIVE {
		if !g.here(dungeon.RUG) {
			g.rspeak(int32(dungeon.FLAP_ARMS))
			return GO_CLEAROBJ
		}
		if g.Objects[dungeon.RUG].Prop != dungeon.RUG_HOVER {
			g.rspeak(int32(dungeon.RUG_NOTHING2))
			return GO_CLEAROBJ
		}
		obj = dungeon.RUG
	}

	if obj != dungeon.RUG {
		g.speak(dungeon.Actions[verb].Message)
		return GO_CLEAROBJ
	}
	if g.Objects[dungeon.RUG].Prop != dungeon.RUG_HOVER {
		g.rspeak(int32(dungeon.RUG_NOTHING1))
		return GO_CLEAROBJ
	}

	if g.Loc == int32(dungeon.LOC_CLIFF) {
		g.Oldlc2 = g.Oldloc
		g.Oldloc = g.Loc
		g.Newloc = int32(dungeon.LOC_LEDGE)
		g.rspeak(int32(dungeon.RUG_GOES))
	} else if g.Loc == int32(dungeon.LOC_LEDGE) {
		g.Oldlc2 = g.Oldloc
		g.Oldloc = g.Loc
		g.Newloc = int32(dungeon.LOC_CLIFF)
		g.rspeak(int32(dungeon.RUG_RETURNS))
	} else {
		g.rspeak(int32(dungeon.NOTHING_HAPPENS))
	}
	return GO_TERMINATE
}

func (g *Game) listen() PhaseCode {
	// Listen. Intransitive only. Print stuff based on object sound properties.
	soundlatch := false
	sound := dungeon.Locations[g.Loc].Sound
	if sound != dungeon.SILENT {
		g.rspeak(int32(sound))
		if !dungeon.Locations[g.Loc].Loud {
			g.rspeak(int32(dungeon.NO_MESSAGE))
		}
		soundlatch = true
	}
	for i := 1; i <= dungeon.NOBJECTS; i++ {
		if !g.here(i) || len(dungeon.Objects[i].Sounds) == 0 || g.objectIsStashedOrUnseen(i) {
			continue
		}
		mi := g.Objects[i].Prop
		// Some unpleasant magic on object states here
		if i == dungeon.BIRD {
			if g.Blooded {
				mi += 3
			}
		}
		g.pSpeak(int32(i), Hear, true, mi, string(g.Zzword[:]))
		g.rspeak(int32(dungeon.NO_MESSAGE))
		if i == dungeon.BIRD && mi == dungeon.BIRD_ENDSTATE {
			g.destroy(int32(dungeon.BIRD))
		}
		soundlatch = true
	}
	if !soundlatch {
		g.rspeak(int32(dungeon.ALL_SILENT))
	}
	return GO_CLEAROBJ
}

func (g *Game) reservoir() PhaseCode {
	// Z'ZZZ (word gets recomputed at startup; different each game).
	if !g.at(int32(dungeon.RESER)) && g.Loc != int32(dungeon.LOC_RESBOTTOM) {
		g.rspeak(int32(dungeon.NOTHING_HAPPENS))
		return GO_CLEAROBJ
	} else {
		newState := int32(dungeon.WATERS_UNPARTED)
		if g.Objects[dungeon.RESER].Prop == dungeon.WATERS_PARTED {
			newState = int32(dungeon.WATERS_UNPARTED)
		} else {
			newState = int32(dungeon.WATERS_PARTED)
		}
		g.stateChange(dungeon.RESER, newState)
		if g.at(int32(dungeon.RESER)) {
			return GO_CLEAROBJ
		} else {
			g.Oldlc2 = g.Loc
			g.Newloc = int32(dungeon.LOC_NOWHERE)
			g.rspeak(int32(dungeon.NOT_BRIGHT))
			return GO_TERMINATE
		}
	}
}

func (g *Game) wave(verb, obj int) PhaseCode {
	// Wave. No effect unless waving rod at fissure or at bird.
	if obj != dungeon.ROD || !g.toting(obj) ||
		(!g.here(dungeon.BIRD) && (g.Closing || !g.at(int32(dungeon.FISSURE)))) {
		if !g.toting(obj) && (obj != dungeon.ROD || !g.toting(dungeon.ROD2)) {
			g.speak(dungeon.Arbitrary_Messages[dungeon.ARENT_CARRYING])
		} else {
			g.speak(dungeon.Actions[verb].Message)
		}
		return GO_CLEAROBJ
	}

	if g.Objects[dungeon.BIRD].Prop == dungeon.BIRD_UNCAGED &&
		g.Loc == int32(dungeon.Objects[dungeon.STEPS].Plac) && g.objectIsNotFound(dungeon.JADE) {
		g.drop(int32(dungeon.JADE), g.Loc)
		g.Objects[dungeon.JADE].Found = true
		g.Tally--
		g.rspeak(int32(dungeon.NECKLACE_FLY))
		return GO_CLEAROBJ
	} else {
		if g.Closed {
			if g.Objects[dungeon.BIRD].Prop == dungeon.BIRD_CAGED {
				g.rspeak(int32(dungeon.CAGE_FLY))
			} else {
				g.rspeak(int32(dungeon.FREE_FLY))
			}
			return GO_DWARFWAKE
		}
		if g.Closing || !g.at(int32(dungeon.FISSURE)) {
			if g.Objects[dungeon.BIRD].Prop == dungeon.BIRD_CAGED {
				g.rspeak(int32(dungeon.CAGE_FLY))
			} else {
				g.rspeak(int32(dungeon.FREE_FLY))
			}
			return GO_CLEAROBJ
		}
		if g.here(dungeon.BIRD) {
			if g.Objects[dungeon.BIRD].Prop == dungeon.BIRD_CAGED {
				g.rspeak(int32(dungeon.CAGE_FLY))
			} else {
				g.rspeak(int32(dungeon.FREE_FLY))
			}
		}

		newState := int32(dungeon.UNBRIDGED)
		if g.Objects[dungeon.FISSURE].Prop == dungeon.BRIDGED {
			newState = int32(dungeon.UNBRIDGED)
		} else {
			newState = int32(dungeon.BRIDGED)
		}
		g.stateChange(dungeon.FISSURE, newState)
		return GO_CLEAROBJ
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
