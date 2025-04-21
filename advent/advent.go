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

//TODO: Implement settings

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

func (g *Game) Start() string {

	return dungeon.Arbitrary_Messages[dungeon.WELCOME_YOU]
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

	} else if g.Settings.EnableDebug && cmd == "ZZTEST1" {
		g.croak()

	} else if g.Settings.EnableDebug && cmd == "ZZTEST2" {

		g.Output = "ZZTEST2. Old Output: " + g.Output

	} else if g.QueryFlag {
		// Game has asked a question. The command will be the response
		g.QueryResponse = cmd
		g.QueryFlag = false

	} else {

		g.DescribeLocation()
		// Game in progress

		// Do we need to move?

		// Describe location

	}

	return err
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
			if (g.Dwarves[i].Oldloc == g.Newloc) && (g.Dwarves[i].Seen != 0) {
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

func (g *Game) rspeak(vocab int32, args ...any) error {
	msg, err := g.vspeak(dungeon.Arbitrary_Messages[vocab], false, args...)

	if err != nil {
		return err
	} else {

		g.Output = msg
		return nil
	}
}

// Speak a specified string
func (g *Game) speak(msg string, args ...any) error {
	msg, err := g.vspeak(msg, true, args...)

	if err != nil {
		return err
	} else {

		g.Output = msg
		return nil
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

	if strings.Contains("%d", renderedString) {
		digits := findOccurrences("%d", renderedString)

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

	if strings.Contains("%s", renderedString) {
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

	if strings.Contains("%S", renderedString) {
		if pluralise {
			renderedString = strings.Replace(renderedString, "%S", "s", -1)
		} else {
			renderedString = strings.Replace(renderedString, "%S", "", -1)
		}
	}

	g.Output = renderedString
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
	g.AskQuestion(query, func(response string) string {
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

func (g *Game) AskQuestion(query string, callback func(response string) string) {
	g.QueryFlag = true
	g.QueryResponse = ""

	g.Output = query
	g.OnQueryResponse = callback
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
		g.Dwarves[i].Seen = 0

		if (g.Dwarves[i].Seen != 0 && id != 0) || g.Dwarves[i].Loc == g.Loc || g.Dwarves[i].Oldloc == g.Loc {
			g.Dwarves[i].Seen = 1
		}

		if g.Dwarves[i].Seen == 0 {
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
			g.Dwarves[PIRATE].Seen = 0
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

func (g *Game) move(object int32, where int32) {
	/*  Place any object anywhere by picking it up and dropping it.  May
	 *  already be toting, in which case the carry is a no-op.  Mustn't
	 *  pick up objects which are not at any loc, since carry wants to
	 *  remove objects from game atloc chains. */

	var from int32

	if object > dungeon.NOBJECTS {
		from = g.Objects[object-dungeon.NOBJECTS].Fixed
	} else {
		from = g.Objects[object].Place
	}

	if from != int32(dungeon.LOC_NOWHERE) && from != CARRIED {
		g.carry(object, from)
	}
	g.drop(object, where)
}

// TODO: refactor these perhaps
func (g *Game) at(object int32) bool {
	return g.Objects[object].Place == g.Loc ||
		g.Objects[object].Fixed == g.Loc
}

func (g *Game) here(object int) bool {
	return g.at(int32(object)) || g.toting(object)
}

func (g *Game) toting(object int) bool {
	return g.Objects[object].Place == CARRIED
}

func (g *Game) objectIsNotFound(object int) bool {
	return g.Objects[object].Prop == STATE_NOTFOUND
}

func (g *Game) dark() bool {

	return !condbit(g.Loc, dungeon.COND_LIT) &&
		(g.Objects[dungeon.LAMP].Prop == dungeon.LAMP_DARK ||
			!g.here(int(dungeon.LAMP)))
}

func (g *Game) objectIsStashed(object int) bool {
	return g.Objects[object].Prop < STATE_NOTFOUND
}

func (g *Game) objectIsFound(object int) bool {
	return g.Objects[object].Prop == STATE_FOUND
}

/*
 *  DESTROY(N)  = Get rid of an item by putting it in LOC_NOWHERE
 *  MOD(N,M)    = Arithmetic modulus
 *  TOTING(OBJ) = true if the OBJ is being carried
 *  AT(OBJ)     = true if on either side of two-placed object
 *  HERE(OBJ)   = true if the OBJ is at "LOC" (or is being carried)
 *  CNDBIT(L,N) = true if COND(L) has bit n set (bit 0 is units bit)
 *  LIQUID()    = object number of liquid in bottle
 *  LIQLOC(LOC) = object number of liquid (if any) at LOC
 *  FORCED(LOC) = true if LOC moves without asking for input (COND=2)
 *  DARK(LOC)   = true if location "LOC" is dark
 *  PCT(N)      = true N% of the time (N integer from 0 to 100)
 *  GSTONE(OBJ) = true if OBJ is a gemstone
 *  FOREST(LOC) = true if LOC is part of the forest
 *  OUTSID(LOC) = true if location not in the cave
 *  INSIDE(LOC) = true if location is in the cave or the building at the
 * beginning of the game INDEEP(LOC) = true if location is in the Hall of Mists
 * or deeper BUG(X)      = report bug and exit
 */

func forest(location int32) bool {
	return condbit(location, dungeon.COND_FOREST)
}

func outside(loction int32) bool {
	return condbit(loction, dungeon.COND_ABOVE) || forest(loction)
}

func inside(location int32) bool {
	return !outside(location) || location == int32(dungeon.LOC_BUILDING)
}

func tstbit(mask int32, bit int32) bool {

	return (mask & (1 << bit)) != 0
}

func condbit(L int32, N int32) bool {
	return tstbit(dungeon.Conditions[L], N)
}

func forced(location int32) bool {
	return condbit(location, dungeon.COND_FORCED)
}

func indeep(location int32) bool {
	return condbit(location, dungeon.COND_DEEP)
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
