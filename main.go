package main

import (
	"flag"
	"fmt"

	"github.com/andrewsjg/goAdventure/advent"
	"github.com/andrewsjg/goAdventure/tui"
)

func main() {

	// TODO: Logs

	logFileName := ""
	autoSaveFileName := ""
	restoreFileName := ""
	debug := true
	oldStyle := false
	autoSave := true

	flag.StringVar(&logFileName, "l", "", "Create a log file of your game named as specified")
	flag.BoolVar(&oldStyle, "o", false, "'Oldstyle' mode (no prompt, no command editing, displays 'Initialising...')")
	flag.StringVar(&autoSaveFileName, "a", "", "Automatic save/restore from specified saved game file")
	flag.StringVar(&restoreFileName, "r", "", "Restore from specified saved game file")
	flag.BoolVar(&debug, "d", false, "Enable debug mode")

	// Parse the command-line flags
	flag.Parse()

	// The remaining arguments are scripts
	scripts := flag.Args()

	// Initialize the game
	if debug {
		fmt.Println("Initializing game...")
	}

	game := advent.NewGame(0, restoreFileName, autoSaveFileName, logFileName, debug, oldStyle, autoSave, scripts)

	if game.Settings.EnableDebug {
		fmt.Println("Starting game...")
		fmt.Printf("ZZWORD: %s\n", string(game.Zzword[:]))
		fmt.Printf("Seedval: %d\n", game.Seedval)
	}

	// Create a new game
	adventure := tui.NewAdventure(game)

	// Start the game
	if _, err := adventure.Run(); err != nil {
		fmt.Println("Error running the adventure:", err)
		return
	}
}
