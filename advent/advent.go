package advent

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"

	"github.com/andrewsjg/goAdventure/dungeon"
)

// TODO: Implement settings
// TODO: Split this up?

func NewGame(seed int, restoreFileName string, autoSaveFileName string, logFileName string, debug bool, oldStyle bool, autoSave bool, scripts []string) Game {

	game := Game{}

	game.Settings.LogFileName = logFileName
	game.Settings.AutoSaveFileName = autoSaveFileName
	game.Settings.RestoreFileName = restoreFileName
	game.Settings.EnableDebug = debug
	game.Settings.OldStyle = oldStyle
	game.Settings.Autosave = autoSave
	game.Settings.Scripts = scripts

	game.Loc = int32(dungeon.LOC_START)
	game.Newloc = int32(dungeon.LOC_START)
	game.Chloc = int32(dungeon.LOC_MAZEEND12)
	game.Oldlc2 = int32(dungeon.LOC_DEADEND13)
	game.Clock1 = WARNTIME
	game.Clock2 = FLASHTIME
	game.Limit = GAMELIMIT
	game.Abbnum = 5
	game.Foobar = WORD_EMPTY

	// Initial Welcom
	game.Output = dungeon.Arbitrary_Messages[dungeon.WELCOME_YOU]

	if debug {
		fmt.Println("Debug mode enabled")
	}

	// TODP: Decide if we want this or not
	if oldStyle {
		fmt.Println("Oldstyle mode enable. Does nothing at the moment.")
	}

	if logFileName != "" {

		if game.Settings.EnableDebug {
			fmt.Printf("Log file: %s\n", logFileName)
		}
		// Open the log file for writing
		logFile, err := os.Create(logFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: can't open logfile %s for write\n", logFileName)
			os.Exit(1)
		}
		defer logFile.Close()
	}

	if restoreFileName != "" {
		// Load the game from the restore file
		err := loadStructFromFile(restoreFileName, &game)

		if err != nil {
			fmt.Println("Error loading game:", err)
		}

	} else {

		// Not restoring a game so this must be a new game
		game.Settings.NewGame = true

		seedval := seed

		if seedval == 0 {
			seedval = rand.Int()
		}

		game.LcgX = int32(seedval) % LCG_M

		if game.LcgX < 0 {
			game.LcgX += LCG_M + game.LcgX
		}

		// Generate the Z'ZZZ word
		for i := 0; i < 5; i++ {
			game.LcgX = getNextLCGValue(game.LcgX)

			if game.LcgX < 0 {
				game.LcgX = -game.LcgX
			}

			game.Zzword[i] = byte('A' + (26 * game.LcgX / LCG_M))
		}

		// Make the second character an apostrophe
		game.Zzword[1] = '\''
		game.Zzword[5] = '\x00'

		for i := 1; i < dungeon.NDWARVES; i++ {
			game.Dwarves[i].Loc = int32(dungeon.DwarfLocs[i-1])

		}

		for i := 1; i < dungeon.NOBJECTS; i++ {
			game.Objects[i].Place = int32(dungeon.LOC_NOWHERE)
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
				game.drop(int32(i+dungeon.NOBJECTS), int32(dungeon.Objects[i].Fixd))
				game.drop(int32(i), int32(dungeon.Objects[i].Plac))
			}
		}

		for i := 1; i < dungeon.NOBJECTS; i++ {
			k := dungeon.NOBJECTS + 1 - i
			game.Objects[k].Fixed = int32(dungeon.Objects[k].Fixd)
			if dungeon.Objects[k].Plac != 0 && dungeon.Objects[k].Fixd <= 0 {
				game.drop(int32(k), int32(dungeon.Objects[k].Plac))
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
				game.Tally++

				if dungeon.Objects[obj].Inventory != "" {
					game.Objects[obj].Prop = STATE_NOTFOUND
				}
			} else {
				game.Objects[obj].Prop = STATE_FOUND
			}

			game.Conds = setBit(dungeon.COND_HBASE)
		}

		game.Seedval = seedval

		// If autosave is enabled immediately save the game
		if game.Settings.Autosave && game.Settings.AutoSaveFileName == "" {
			// Create a new save file
			game.Settings.AutoSaveFileName = "advent.save"

			if fileExists(game.Settings.AutoSaveFileName) {
				if game.Settings.EnableDebug {
					fmt.Println("Autosave file already exists")
				}

				if !game.Settings.EnableDebug {
					// For now just move the the existing save file to a new file with a random string appended
					newAutoSaveFileName := "advent_" + generateRandomString(5) + ".save"
					err := os.Rename(game.Settings.AutoSaveFileName, newAutoSaveFileName)

					if err != nil {
						fmt.Println("Error renaming autosave file:", err)
					}
				}

			}

			err := saveStructToFile(game.Settings.AutoSaveFileName, game)
			if err != nil {
				fmt.Println("Error saving game:", err)
			}
		}
	}

	return game
}

func (g *Game) CheckHints() {

	if dungeon.Conditions[g.Loc] >= g.Conds {
		for hint := 0; hint < dungeon.NHINTS; hint++ {
			{
				if g.Hints[hint].Used {
					continue
				}

				if !condbit(g.Loc, int32(hint+1+dungeon.COND_HBASE)) {
					g.Hints[hint].Lc = -1
				}
				g.Hints[hint].Lc++

				/*  Come here if the player has been int enough at required loc(s)
				 * for some unused hint. */

				if g.Hints[hint].Lc >= int32(dungeon.Hints[hint].Turns) {

					switch hint {

					case 0:
						// Cave
						if g.Objects[dungeon.GRATE].Prop == dungeon.GRATE_CLOSED && !g.here(dungeon.KEYS) {
							break
						}
						g.Hints[hint].Lc = 0
						return
					case 1:
						// bird
						if g.Objects[dungeon.BIRD].Place == g.Loc && g.toting(dungeon.ROD) &&
							g.Oldobj == int32(dungeon.BIRD) {

							break

						}
						return
					case 2:
						// Snake
						if g.here(dungeon.SNAKE) && !g.here(dungeon.BIRD) {
							break
						}

						g.Hints[hint].Lc = 0
						return
					case 3:
						// Maze

						if g.Locs[g.Loc].Atloc == int32(dungeon.NO_OBJECT) &&
							g.Locs[g.Oldloc].Atloc == int32(dungeon.NO_OBJECT) &&
							g.Locs[g.Oldlc2].Atloc == int32(dungeon.NO_OBJECT) &&
							g.Holdng > 1 {
							break
						}
						g.Hints[hint].Lc = 0
						return

					case 4:
						// Dark
						if !g.objectIsNotFound(dungeon.EMERALD) && g.objectIsNotFound(dungeon.PYRAMID) {
							break
						}
						g.Hints[hint].Lc = 0
						return
					case 5:
						// Witt
						break

					case 6:
						// Urn
						if g.Dflag == 0 {
							break
						}
						g.Hints[hint].Lc = 0
						return

					case 7:
						// Woods
						if g.Locs[g.Loc].Atloc == int32(dungeon.NO_OBJECT) &&
							g.Locs[g.Oldloc].Atloc == int32(dungeon.NO_OBJECT) &&
							g.Locs[g.Oldlc2].Atloc == int32(dungeon.NO_OBJECT) {

							break
						}
						return
					case 8:
						// Ogre

						i := g.atDwrf(g.Loc)
						if i < 0 {
							g.Hints[hint].Lc = 0
							return
						}

						if g.here(dungeon.OGRE) && i == 0 {
							break
						}
						return

					case 9:
						// Jade

						if g.Tally == 1 && g.objectIsStashedOrUnseen(dungeon.JADE) {
							break
						}
						g.Hints[hint].Lc = 0
						return

					default:
						// This should never happen
						// TODO: Pring some error here

					}

					/* Fall through to hint display */

					g.Hints[hint].Lc = 0

					wantHint := func(response string, game *Game) string {
						if !strings.Contains(strings.ToUpper(response), "Y") {
							err := game.speak(dungeon.Arbitrary_Messages[dungeon.OK_MAN])

							if err != nil {
								fmt.Println("Error: ", err.Error())
								return fmt.Sprintf("Error: %s", err.Error())
							}

							return ""
						}

						game.speak(dungeon.Hints[hint].Hint)

						if game.Hints[hint].Used && game.Limit > WARNTIME {
							game.Limit += int32(WARNTIME * dungeon.Hints[1].Penalty)
						}

						return dungeon.Hints[hint].Hint
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

						game.rspeak(int32(dungeon.HINT_COST), dungeon.Hints[hint].Penalty)

						game.Output = game.Output + "\n\n" + dungeon.Arbitrary_Messages[dungeon.WANT_HINT]

						game.AskQuestion(game.Output, wantHint)

						return response
					}

					g.AskQuestion(dungeon.Hints[hint].Question, hintQuestion)

				}
			}
		}
	}
}

func (g *Game) DescribeLocation() {

	msg := dungeon.Locations[g.Loc].Description.Small

	if (math.Mod(float64(g.Locs[g.Loc].Abbrev), float64(g.Abbnum)) == 0) || msg == "" {
		msg = dungeon.Locations[g.Loc].Description.Big
	}

	if !forced(g.Loc) && g.dark() {
		msg = dungeon.Arbitrary_Messages[dungeon.PITCH_DARK]
	}

	if g.toting(dungeon.BEAR) {
		g.rspeak(int32(dungeon.TAME_BEAR))
	}

	g.speak(msg)

	if g.Loc == int32(dungeon.LOC_Y2) && !g.Closing {
		g.rspeak(int32(dungeon.SAYS_PLUGH))
	}

	g.Output = msg
}

func (g *Game) DoMove() bool {

	// Can't leave cave once it's closing (except by main office)
	if outside(g.Newloc) && g.Newloc != 0 && g.Closing {
		g.rspeak(int32(dungeon.EXIT_CLOSED))
		g.Newloc = g.Loc

		if !g.Panic {
			g.Clock2 = PANICTIME
		}
		g.Panic = true
	}

	/*  See if a dwarf has seen the player and has come from where they
	 *  want to go.  If so, the dwarf's blocking the player's way.  If
	 *  coming from place forbidden to pirate (dwarves rooted in
	 *  place) let the player get out (and attacked). */

	if g.Newloc != g.Loc && !forced(g.Loc) && !condbit(g.Loc, dungeon.COND_NOARRR) {

		for i := 1; i <= dungeon.NDWARVES; i++ {
			if (g.Dwarves[i].Oldloc == g.Newloc) && (!g.Dwarves[i].Seen) {
				g.Newloc = g.Loc
				g.rspeak(int32(dungeon.DWARF_BLOCK))
				break
			}
		}
	}

	g.Loc = g.Newloc

	if !g.dwarfmove() {
		g.croak()
		return false
	}

	/* The easiest way to get killed is to fall into a pit in
	 * pitch darkness. */
	if g.Loc == int32(dungeon.LOC_NOWHERE) {
		g.croak()
		return false
	}

	if !forced(g.Loc) && g.dark() && g.Wzdark && pct(PIT_KILL_PROB) {
		g.rspeak(int32(dungeon.PIT_FALL))
		g.Oldlc2 = g.Loc
		g.croak()
		return false
	}

	return true

}

// Miscellaneous game functions

// SPEAK functions

func (g *Game) sspeak(msg int, args ...any) {
	g.speak(dungeon.Arbitrary_Messages[msg], args...)
}

func (g *Game) rspeak(vocab int32, args ...any) error {
	msg, err := g.vspeak(dungeon.Arbitrary_Messages[vocab], false, args...)

	if err != nil {
		return err
	} else {

		g.Output = msg

		return nil
	}
}

// Speak a temnporary message
func (g *Game) tspeak(vocab int32, args ...any) error {
	msg, err := g.vspeak(dungeon.Arbitrary_Messages[vocab], false, args...)

	if err != nil {
		return err
	} else {

		g.OutputType = 1
		g.Output = msg

		return nil
	}
}

// Speak a specified string
func (g *Game) speak(msg string, args ...any) error {
	msg, err := g.vspeak(msg, false, args...)

	if err != nil {
		return err

	} else {

		g.Output = msg
		return nil
	}
}

// TODO: Refactor pSepak. In fact refactor all the speak routines
func (g *Game) pSpeak(msg int32, mode SpeakType, blank bool, skip int32, args ...any) {

	var err error
	var output string

	switch mode {
	case Touch:
		output, err = g.vspeak(dungeon.Objects[msg].Inventory, blank, args...)

	case Look:
		output, err = g.vspeak(dungeon.Objects[msg].Descriptions[skip], blank, args...)

	case Hear:
		output, err = g.vspeak(dungeon.Objects[msg].Sounds[skip], blank, args...)

	case Study:
		output, err = g.vspeak(dungeon.Objects[msg].Texts[skip], blank, args...)

	case Change:
		output, err = g.vspeak(dungeon.Objects[msg].Changes[skip], blank, args...)

	}

	if err != nil {
		g.Output = fmt.Sprintf("Error: %s", err.Error())
	} else {
		g.Output = g.Output + "\n\n" + output
	}

}

// TODO: This can probably be refactored to be more go like
func (g *Game) vspeak(msg string, blank bool, args ...any) (string, error) {

	if msg == "" {
		return "", nil
	}

	if blank {
		return fmt.Sprintf("\n"), nil
	}

	pluralise := false

	renderedString := msg

	// If location is outside. Render the string with "ground" instead of "floor"
	if strings.Contains("floor", renderedString) && !inside(g.Loc) {
		renderedString = strings.Replace(renderedString, "floor", "ground", -1)
	}

	if strings.Contains(renderedString, "%d") {

		digits := findOccurrences(renderedString, "%d")

		// if we are rendering digits into a string then the passed args
		// should all be digits.

		if len(args) != len(digits) {
			return "", fmt.Errorf("error: Number of digits in string does not match number of args")
		}

		// replace %d in the string with the args in the order they appear in the
		// passed args
		strVals := make([]string, len(args))
		for i, arg := range args {

			value, ok := arg.(int)
			if ok {
				stringValue := fmt.Sprintf("%d", value)

				// More than one thing, so we should pularise the string
				if value > 1 {
					pluralise = true
				}

				strVals[i] = stringValue
			} else {
				return "", fmt.Errorf("error: Argument %d is not an int", i)
			}

		}

		renderedString = replaceOccurrences(renderedString, "%d", strVals)

	}

	if strings.Contains(renderedString, "%s") {
		strings := findOccurrences("%s", renderedString)

		// if we are rendering strings into a string then the passed args
		// should all be strings.

		if len(args) != len(strings) {
			return "", fmt.Errorf("error: Number of strings in string does not match number of args")
		}

		// replace %s in the string with the args in the order they appear in the
		// passed args
		strVals := make([]string, len(args))
		for i, arg := range args {

			value, ok := arg.(string)
			if ok {
				strVals[i] = value
			} else {
				return "", fmt.Errorf("error: Argument %d is not a string", i)
			}

		}

		renderedString = replaceOccurrences(renderedString, "%s", strVals)
	}

	if strings.Contains(renderedString, "%S") {
		if pluralise {
			renderedString = strings.Replace(renderedString, "%S", "s", -1)
		} else {
			renderedString = strings.Replace(renderedString, "%S", "", -1)
		}
	}
	// g.Output = renderedString

	return renderedString, nil
}

/*
		Dwarf stuff.  See earlier comments for description of
		variables.  Remember sixth dwarf is pirate and is thus
		very different except for motion rules.


		First off, don't let the dwarves follow him into a pit or a
		wall.  Activate the whole mess the first time he gets as far
		as the Hall of Mists (what INDEEP() tests).  If game.newloc
		is forbidden to pirate (in particular, if it's beyond the
		troll bridge), bypass dwarf stuff.  That way pirate can't
		steal return toll, and dwarves can't meet the bear.  Also
		means dwarves won't follow him into dead end in maze, but
		c'est la vie.  They'll wait for him outside the dead end.

		When the dwarves move, return true if the player survives,
	 	false if the player dies
*/

/*  "You're dead, Jim."
 *
 *  If the current loc is zero, it means the clown got himself killed.
 *  We'll allow this maxdie times.  NDEATHS is automatically set based
 *  on the number of snide messages available.  Each death results in
 *  a message (obituaries[n]) which offers reincarnation; if accepted,
 *  this results in message obituaries[0], obituaries[2], etc.  The
 *  last time, if the player wants another chance, they gets a snide
 *  remark as we exit.
 *  When reincarnated, all objects being carried get dropped
 *  at game.oldlc2 (presumably the last place prior to being killed)
 *  without change of props.  The loop runs backwards to assure that
 *  the bird is dropped before the cage.  (This kluge could be changed
 *  once we're sure all references to bird and cage are done by
 *  keywords.)  The lamp is a special case (it wouldn't do to leave it
 *  in the cave). It is turned off and left outside the building (only
 *  if he was carrying it, of course).  He himself is left inside the
 *  building (and heaven help him if he tries to xyzzy back into the
 *  cave without the lamp!).  game.oldloc is zapped so he can't just
 *  "retreat". */

func (g *Game) croak() {
	/*  Okay, player's dead.  Let's get on with it. */

	query := dungeon.Obituaries[g.Numdie].Query
	// Yes_Response := dungeon.Obituaries[g.Numdie].Yes_Response
	croakOutput := ""

	g.Numdie++

	if g.Closing {

		g.rspeak(int32(dungeon.DEATH_CLOSING))
		g.terminate(EndGame) // end game
	}

	// Player has Used up all their lives
	if g.Numdie == dungeon.NDEATHS {
		g.speak(dungeon.Arbitrary_Messages[dungeon.OK_MAN])
		g.terminate(EndGame)
	}

	// Ask if they want to try again
	g.AskQuestion(query, func(response string, game *Game) string {
		if strings.Contains(strings.ToUpper(response), "Y") {

			/* If the player wishes to continue, we empty the liquids in the
			* user's inventory, turn off the lamp, and drop all items
			* where they died. */

			g.Objects[dungeon.WATER].Place = int32(dungeon.LOC_NOWHERE)
			g.Objects[dungeon.OIL].Place = int32(dungeon.LOC_NOWHERE)

			if g.toting(dungeon.LAMP) {
				g.Objects[dungeon.LAMP].Prop = dungeon.LAMP_DARK
			}

			for j := 1; j <= dungeon.NOBJECTS; j++ {
				i := dungeon.NOBJECTS + 1 - j
				if g.toting(i) {
					/* Always leave lamp where it's accessible
					 * aboveground */
					if i == dungeon.LAMP {
						g.drop(int32(i), int32(dungeon.LOC_START))
					} else {
						g.drop(int32(i), g.Oldlc2)
					}
				}
			}

			g.Oldloc = int32(dungeon.LOC_BUILDING)
			g.Newloc = int32(dungeon.LOC_BUILDING)
			g.Loc = int32(dungeon.LOC_BUILDING)

		} else {
			g.speak(dungeon.Arbitrary_Messages[dungeon.OK_MAN])
			g.terminate(EndGame)
		}

		return croakOutput
	})

}

func (g *Game) dwarfmove() bool {
	var kk, stick, attack int
	var tk [21]int32

	if g.Loc == int32(dungeon.LOC_NOWHERE) || forced(g.Loc) || condbit(g.Loc, dungeon.COND_NOARRR) {
		return true
	}

	/* Dwarf activity level ratchets up */

	if g.Dflag == 0 {
		if indeep(g.Loc) {
			g.Dflag = 1
		}
		return true
	}

	/*  When we encounter the first dwarf, we kill 0, 1, or 2 of
	 *  the 5 dwarves.  If any of the survivors is at game.loc,
	 *  replace them with the alternate. */
	if g.Dflag == 1 {
		if !indeep(g.Loc) || pct(95) && (!condbit(g.Loc, dungeon.COND_NOBACK) || pct(85)) {
			return true
		}

		g.Dflag = 2
		for i := 1; i <= 2; i++ {
			j := 1 + g.randRange(dungeon.NDWARVES-1)
			if pct(50) {
				g.Dwarves[j].Loc = 0
			}
		}

		/* Alternate initial loc for dwarf, in case one of them
		   starts out on top of the adventurer. */

		for i := 1; i <= dungeon.NDWARVES; i++ {
			if int32(g.Dwarves[i].Loc) == g.Loc {
				g.Dwarves[i].Loc = int32(DALTLC)
			}
			g.Dwarves[i].Oldloc = g.Dwarves[i].Loc

		}
		g.rspeak(int32(dungeon.DWARF_RAN))
		g.drop(int32(dungeon.AXE), g.Loc)
		return true

	}

	/*  Things are in full swing.  Move each dwarf at random,
	 *  except if he's seen us he sticks with us.  Dwarves stay
	 *  deep inside.  If wandering at random, they don't back up
	 *  unless there's no alternative.  If they don't have to
	 *  move, they attack.  And, of course, dead dwarves don't do
	 *  much of anything. */

	g.Dtotal = 0
	attack = 0
	stick = 0

	for i := 1; i <= dungeon.NDWARVES; i++ {
		if g.Dwarves[i].Loc == 0 {
			continue
		}

		j := 1

		/*  Fill tk array with all the places this dwarf might go. */
		kk = int(dungeon.TKey[g.Dwarves[i].Loc])
		if kk == 0 {
			for {
				destType := dungeon.Travel[kk].DestType
				g.Newloc = int32(dungeon.Travel[kk].DestVal)

				if destType != dungeon.DestGoto {
					continue
				} else if !indeep(g.Newloc) {
					continue
				} else if g.Newloc == g.Dwarves[i].Oldloc {
					continue
				} else if j > 1 && g.Newloc == tk[j-1] {
					continue
				} else if j >= len(tk)-1 {
					// apparently this can't happen
					continue
				} else if forced(g.Newloc) {
					continue
				} else if i == PIRATE && condbit(g.Newloc, dungeon.COND_NOARRR) {
					continue
				} else if dungeon.Travel[kk].NoDwarves {
					continue
				}
				j++
				tk[j] = g.Newloc

				if dungeon.Travel[kk].Stop {
					break
				}
			}
		}
		tk[j] = g.Dwarves[i].Oldloc

		if j >= 2 {
			j--
		}

		j = 1 + int(g.randRange(int32(j)))
		g.Dwarves[i].Oldloc = g.Dwarves[i].Loc
		g.Dwarves[i].Loc = int32(tk[j])

		id := 1
		if !indeep(g.Loc) {
			id = 0
		}
		g.Dwarves[i].Seen = false

		if (!g.Dwarves[i].Seen && id != 0) || g.Dwarves[i].Loc == g.Loc || g.Dwarves[i].Oldloc == g.Loc {
			g.Dwarves[i].Seen = true
		}

		if g.Dwarves[i].Seen == false {
			continue
		}

		g.Dwarves[i].Loc = g.Loc

		if g.spottedByPirate(i) {
			continue
		}

		g.Dtotal++

		if g.Dwarves[i].Oldloc == g.Dwarves[i].Loc {
			attack++
			if g.Knfloc >= int32(dungeon.LOC_NOWHERE) {
				g.Knfloc = g.Loc
			}

			if g.randRange(1000) < 95*(g.Dflag-2) {
				stick++
			}
		}
	}

	/*  Now we know what's happening.  Let's tell the poor sucker about it.
	 */

	if g.Dtotal == 0 {
		return true
	}

	if g.Dtotal == 1 {
		g.rspeak(int32(dungeon.DWARF_SINGLE), g.Dtotal)
	} else {
		g.rspeak(int32(dungeon.DWARF_PACK), g.Dtotal)
	}

	if attack == 0 {
		return true
	}

	if g.Dflag == 2 {
		g.Dflag = 3
	}

	if attack > 1 {
		g.rspeak(int32(dungeon.THROWN_KNIVES), attack)

		if stick > 1 {
			g.rspeak(int32(dungeon.MULTIPLE_HITS), stick)
		} else if stick == 1 {
			g.rspeak(int32(dungeon.ONE_HIT), stick)
		} else {
			g.rspeak(int32(dungeon.NONE_HIT), stick)
		}
	} else {
		g.rspeak(int32(dungeon.KNIFE_THROWN))

		if stick != 0 {
			g.rspeak(int32(dungeon.GETS_YOU))
		} else {
			g.rspeak(int32(dungeon.MISSES_YOU))
		}
	}

	if stick == 0 {
		return true
	}

	g.Oldlc2 = g.Loc

	return false
}

func (g *Game) getNextLCGValue() int32 {
	oldx := g.LcgX
	g.LcgX = (LCG_A*g.LcgX + LCG_C) % LCG_M
	if g.Settings.EnableDebug {
		fmt.Printf("random %d\n", oldx)
	}

	return oldx
}

func (g *Game) randRange(rndRange int32) int32 {
	// Generate a random number between 0 and range
	return rndRange * g.getNextLCGValue() % LCG_M
}

func (g *Game) spottedByPirate(dwarfNum int) bool {
	if dwarfNum != PIRATE {
		return false
	}

	/*  The pirate's spotted the player.  Pirate leaves the player
	    alone once we've found chest.  K counts if a treasure is here.
		If not, and tally=1 for an unseen chest, let the pirate be spotted.
		Note that game.objexts,place[CHEST] = LOC_NOWHERE might mean that he's
	    thrown it to the troll, but in that case he's seen the chest
	    OBJECT_IS_FOUND(CHEST) == true. */

	if g.Loc == g.Chloc || !g.objectIsNotFound(dungeon.CHEST) {
		return true
	}

	snarfed := 0
	movechest := false
	robplayer := false

	for treasure := 1; treasure <= dungeon.NOBJECTS; treasure++ {
		if !dungeon.Objects[treasure].Is_Treasure {
			continue
		}

		/*  Pirate won't take pyramid from plover room or dark
		 *  room (too easy!). */

		if treasure == dungeon.PYRAMID && g.Loc == int32(dungeon.Objects[dungeon.PYRAMID].Plac) ||
			g.Loc == int32(dungeon.Objects[dungeon.EMERALD].Plac) {
			continue
		}

		if g.toting(treasure) || g.here(treasure) {
			snarfed++
		}

		if g.toting(treasure) {
			movechest = true
			robplayer = true
		}

		/* Force chest placement before player finds last treasure */
		if g.Tally == 1 && snarfed == 0 &&
			g.Objects[dungeon.CHEST].Place == int32(dungeon.LOC_NOWHERE) &&
			g.here(int(dungeon.LAMP)) &&
			g.Objects[dungeon.LAMP].Prop == dungeon.LAMP_BRIGHT {

			g.rspeak(int32(dungeon.PIRATE_SPOTTED))
			movechest = true
		}

		/* Do things in this order (chest move before robbery) so chest is
		* listed last at the maze location. */

		if movechest {
			g.move(int32(dungeon.CHEST), g.Chloc)
			g.move(int32(dungeon.MESSAG), g.Chloc2)

			g.Dwarves[PIRATE].Loc = g.Chloc
			g.Dwarves[PIRATE].Oldloc = g.Chloc
			g.Dwarves[PIRATE].Seen = false
		} else {
			if g.Dwarves[PIRATE].Oldloc != g.Dwarves[PIRATE].Loc && pct(20) {
				g.rspeak(int32(dungeon.PIRATE_RUSTLES))
			}
		}

		if robplayer {
			g.rspeak(int32(dungeon.PIRATE_POUNCES))

			for treasure := 1; treasure <= dungeon.NOBJECTS; treasure++ {

				if !dungeon.Objects[treasure].Is_Treasure {
					continue
				}

				if !(treasure == dungeon.PYRAMID &&
					(g.Loc == g.Objects[dungeon.PYRAMID].Place ||
						g.Loc == g.Objects[dungeon.EMERALD].Place)) {

					if g.at(int32(treasure)) && g.Objects[treasure].Fixed == IS_FREE {
						g.carry(int32(treasure), g.Loc)
					}

					if g.toting(treasure) {
						g.drop(int32(treasure), g.Chloc)
					}
				}
			}
		}

	}

	return true
}

func (g *Game) carry(object, where int32) {
	/*  Start toting an object, removing it from the list of things at its
	 * former location.  Incr holdng unless it was already being toted.  If
	 * object>NOBJECTS (moving "fixed" second loc), don't change game.place
	 * or game.holdng. */

	if object > dungeon.NOBJECTS {
		if g.Objects[object].Place == CARRIED {
			return
		}

		g.Objects[object].Place = CARRIED

		/*
		 * Without this conditional your inventory is overcounted
		 * when you pick up the bird while it's caged. This fixes
		 * a cosmetic bug in the original.
		 *
		 * Possibly this check should be skipped whwn oldstyle is on.
		 */
		if object != int32(dungeon.BIRD) {
			g.Holdng++
		}
	}

	if g.Locs[where].Atloc == object {
		g.Locs[where].Atloc = g.Link[object]
		return
	}
	temp := g.Locs[where].Atloc
	for temp != object {
		temp = g.Link[temp]
	}

	g.Link[temp] = g.Link[object]

}

func (g *Game) drop(object, where int32) {

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

func (g *Game) PlayerMove(loc int32) {
	//TODO: Implement PlayerMove. Does this need to be exported?
}

func (g *Game) ListObjects() {
	if !g.dark() {

		// TODO: Figure out how to handle this better
		//g.Locs[g.Loc].Abbrev++

		for i := g.Locs[g.Loc].Atloc; i != 0; i = g.Link[i] {
			obj := i

			if obj > dungeon.NOBJECTS {
				obj -= dungeon.NOBJECTS
			}

			if obj == int32(dungeon.STEPS) && g.toting(int(dungeon.NUGGET)) {
				continue
			}

			/* (ESR) Warning: it looks like you could get away with
			 * running this code only on objects with the treasure
			 * property set. Nope.  There is mystery here.
			 */

			if g.objectIsStashedOrUnseen(int(obj)) {
				if g.Closed {
					continue
				}

				g.objectSetFound(int(obj))

				if obj == int32(dungeon.RUG) {
					g.Objects[dungeon.RUG].Prop = dungeon.RUG_DRAGON
				}

				if obj == int32(dungeon.CHAIN) {
					g.Objects[dungeon.CHAIN].Prop = dungeon.CHAINING_BEAR
				}

				if obj == int32(dungeon.EGGS) {
					g.Seenbigwords = true
				}

				g.Tally--

				/*  Note: There used to be a test here to see
				 * whether the player had blown it so badly that
				 * he could never ever see the remaining
				 * treasures, and if so the lamp was zapped to
				 *  35 turns.  But the tests were too
				 * simple-minded; things like killing the bird
				 * before the snake was gone (can never see
				 * jewelry), and doing it "right" was hopeless.
				 * E.G., could cross troll bridge several times,
				 * using up all available treasures, breaking
				 * vase, using coins to buy batteries, etc., and
				 * eventually never be able to get across again.
				 * If bottle were left on far side, could then
				 *  never get eggs or trident, and the effects
				 * propagate.  So the whole thing was flushed.
				 * anyone who makes such a gross blunder isn't
				 * likely to find everything else anyway (so
				 * goes the rationalisation). */

			}

			kk := g.Objects[obj].Prop

			if obj == int32(dungeon.STEPS) {
				if g.Loc == g.Objects[int32(dungeon.STEPS)].Fixed {
					kk = dungeon.STEPS_UP
				} else {
					kk = dungeon.STEPS_DOWN
				}
			}

			g.pSpeak(obj, Look, true, kk)

		}
	}
}

func (g *Game) atDwrf(where int32) int {
	/*  Return the index of first dwarf at the given location, zero if no
	 * dwarf is there (or if dwarves not active yet), -1 if all dwarves are
	 * dead.  Ignore the pirate (6th dwarf). */

	at := 0

	if g.Dflag < 2 {
		return at
	}

	at = -1
	for i := 1; i <= dungeon.NDWARVES-1; i++ {
		if g.Dwarves[i].Loc == where {
			return i
		}

		if g.Dwarves[i].Loc != 0 {
			at = 0
		}
	}

	return at
}

/*  Print message X, wait for yes/no answer.  If yes, print Y and return
 * true; if no, print Z and return false. */

// Utility Functions

// Checks if a randomly generated number between 0 and 99 is less than N
func pct(n int32) bool {
	return (rand.Int31n(100) < n)
}

// SaveStructToFile saves the struct to a file in JSON format.
func saveStructToFile(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode struct to JSON: %w", err)
	}
	return nil
}

// LoadStructFromFile loads the struct from a JSON file.
func loadStructFromFile(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}
	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		// File exists
		return true
	}
	if os.IsNotExist(err) {
		// File does not exist
		return false
	}
	// Some other error occurred
	fmt.Println("Error checking file:", err)
	return false
}

func generateRandomString(length int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	//rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func setBit(bit int32) int32 {
	return 1 << bit
}

func getNextLCGValue(lcgX int32) int32 {
	return (LCG_A*lcgX + LCG_C) % LCG_M
}

func findOccurrences(haystack, needle string) []int {
	var indices []int
	offset := 0

	for {
		// Find the index of the substring starting from the current offset
		index := strings.Index(haystack[offset:], needle)
		if index == -1 {
			break // No more occurrences
		}

		// Add the absolute index to the result
		absoluteIndex := offset + index
		indices = append(indices, absoluteIndex)

		// Move the offset forward to continue searching
		offset = absoluteIndex + len(needle)
	}

	return indices
}

func replaceOccurrences(haystack, needle string, replacements []string) string {
	var result strings.Builder
	offset := 0
	replacementIndex := 0

	for {
		// Find the next occurrence of the substring
		index := strings.Index(haystack[offset:], needle)
		if index == -1 {
			// No more occurrences, append the rest of the string
			result.WriteString(haystack[offset:])
			break
		}

		// Append the part of the string before the match
		result.WriteString(haystack[offset : offset+index])

		// Append the replacement string
		if replacementIndex < len(replacements) {
			result.WriteString(replacements[replacementIndex])
			replacementIndex++
		} else {
			// If no more replacements are available, append the original substring
			result.WriteString(needle)
		}

		// Move the offset past the current match
		offset += index + len(needle)
	}

	return result.String()
}

func replaceAtIndex(haystack string, index int, length int, replacement string) (string, error) {
	// Ensure the index and length are within bounds
	if index < 0 || index+length > len(haystack) {
		return "", fmt.Errorf("index or length out of bounds")
	}

	// Split the string into three parts
	before := haystack[:index]             // Part before the substring
	after := haystack[index+length:]       // Part after the substring
	result := before + replacement + after // Concatenate with replacement

	return result, nil
}
