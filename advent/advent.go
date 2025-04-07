package advent

import (
	"encoding/json"
	"fmt"
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
			game.Dwarves[i].Loc = dungeon.DwarfLocs[i-1]

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
				game.Drop(int32(i+dungeon.NOBJECTS), int32(dungeon.Objects[i].Fixd))
				game.Drop(int32(i), int32(dungeon.Objects[i].Plac))
			}
		}

		for i := 1; i < dungeon.NOBJECTS; i++ {
			k := dungeon.NOBJECTS + 1 - i
			game.Objects[k].Fixed = int32(dungeon.Objects[k].Fixd)
			if dungeon.Objects[k].Plac != 0 && dungeon.Objects[k].Fixd <= 0 {
				game.Drop(int32(k), int32(dungeon.Objects[k].Plac))
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

func (g *Game) ProcessCommand(command string) string {

	output := ""

	output = fmt.Sprintf("CMD: %s LOC: %d", command, g.Loc)

	cmd := strings.ToUpper(command)

	// Game start condition
	// If this is the start of a new game and the command is yes
	// then the player has asked for instructions

	if g.Settings.NewGame && strings.Contains(cmd, "Y") {
		output = dungeon.Arbitrary_Messages[dungeon.CAVE_NEARBY]
		g.Novice = true
		g.Limit = NOVICELIMIT // Numner of turns allowed for a novice player

		// Reset new game flag since the game has now progressed
		g.Settings.NewGame = false

	} else {

		// Game in progress

		// Descrobe location

	}

	return output
}

// Utility Functions

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
