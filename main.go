package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/andrewsjg/goAdventure/advent"
	"github.com/andrewsjg/goAdventure/telemetry"
	"github.com/andrewsjg/goAdventure/tui"
)

func main() {

	// TODO: Logs

	logFileName := ""
	autoSaveFileName := ""
	restoreFileName := ""
	scriptFileName := ""
	debug := false
	oldStyle := false
	autoSave := true
	noTUI := false
	enableTracing := false
	tracingEndpoint := ""

	flag.StringVar(&logFileName, "l", "", "Create a log file of your game named as specified")
	flag.BoolVar(&oldStyle, "o", false, "'Oldstyle' mode (no prompt, no command editing, displays 'Initialising...')")
	flag.StringVar(&autoSaveFileName, "a", "", "Automatic save/restore from specified saved game file")
	flag.StringVar(&restoreFileName, "r", "", "Restore from specified saved game file")
	flag.StringVar(&scriptFileName, "script", "", "Execute commands from script file (one command per line, # for comments)")
	flag.BoolVar(&debug, "d", false, "Enable debug mode")
	flag.BoolVar(&noTUI, "notui", false, "Run without TUI (classic terminal mode)")
	flag.BoolVar(&enableTracing, "trace", false, "Enable OpenTelemetry tracing (sends to localhost:4318 by default)")
	flag.StringVar(&tracingEndpoint, "trace-endpoint", "", "OpenTelemetry OTLP endpoint (e.g., localhost:4318)")

	// Parse the command-line flags
	flag.Parse()

	// The remaining arguments are scripts
	scripts := flag.Args()

	// Initialize tracing
	ctx := context.Background()
	tracingCfg := telemetry.Config{
		Enabled:      enableTracing,
		OTLPEndpoint: tracingEndpoint,
	}
	shutdownTracing, err := telemetry.InitTracing(ctx, tracingCfg)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize tracing: %v\n", err)
	}
	defer func() {
		if shutdownTracing != nil {
			_ = shutdownTracing(ctx)
		}
	}()

	if enableTracing && debug {
		fmt.Println("OpenTelemetry tracing enabled")
	}

	// Initialize the game
	if debug {
		fmt.Println("Initializing game...")
	}

	game := advent.NewGame(0, restoreFileName, autoSaveFileName, logFileName, debug, oldStyle, autoSave, scripts)

	// Load script file if specified
	if scriptFileName != "" {
		if err := game.LoadScript(scriptFileName); err != nil {
			fmt.Printf("Error loading script: %v\n", err)
			return
		}
		if debug {
			fmt.Printf("Loaded %d commands from script\n", len(game.ScriptCommands))
		}
	}

	// Set the tracing context on the game
	game.Ctx = ctx

	// Start initial location span if tracing is enabled
	if enableTracing {
		game.StartLocationSpan(game.Loc)
		defer game.EndLocationSpan()
	}

	if game.Settings.EnableDebug {
		fmt.Println("Starting game...")
		fmt.Printf("ZZWORD: %s\n", string(game.Zzword[:]))
		fmt.Printf("Seedval: %d\n", game.Seedval)
	}

	if noTUI {
		// Run in classic terminal mode
		runClassicMode(&game)
	} else {
		// Create a new game with TUI
		adventure := tui.NewAdventure(&game)

		// Start the game
		if _, err := adventure.Run(); err != nil {
			fmt.Println("Error running the adventure:", err)
			return
		}
	}
}

func runClassicMode(game *advent.Game) {
	reader := bufio.NewReader(os.Stdin)

	// Print initial welcome message
	fmt.Println(game.Output)

	// Main game loop
	for {
		// Check if game is over
		if game.GameOver {
			if game.Output != "" {
				fmt.Println(game.Output)
			}
			fmt.Println("\nThanks for playing!")
			return
		}

		// Check for queries
		if game.QueryFlag {
			fmt.Print(game.Output)
			var response string
			if cmd, ok := game.NextScriptCommand(); ok {
				response = cmd
				fmt.Printf("\n> %s\n", response) // Echo the script command
			} else {
				fmt.Print("\n> ")
				var err error
				response, err = reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading input:", err)
					return
				}
				response = strings.TrimSpace(response)
			}

			game.QueryResponse = response
			game.QueryFlag = false

			if game.OnQueryResponse != nil {
				game.OnQueryResponse(game.QueryResponse, game)
				fmt.Println(game.Output)
			}
			continue
		}

		// Do movement if needed
		if game.Newloc != game.Loc {
			game.DoMove()
			game.DescribeLocation()
			game.ListObjects()
			fmt.Println(game.Output)
		}

		// Check for forced moves
		if game.LocForced() {
			game.MoveHere()
			continue
		}

		// Get command from script or user
		var input string
		if cmd, ok := game.NextScriptCommand(); ok {
			input = cmd
			fmt.Printf("> %s\n", input) // Echo the script command
		} else {
			fmt.Print("> ")
			var err error
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			input = strings.TrimSpace(input)
		}

		// Handle exit
		if strings.ToLower(input) == "quit" || strings.ToLower(input) == "exit" {
			// Autosave if enabled
			if err := game.AutoSave(); err != nil && game.Settings.EnableDebug {
				fmt.Printf("DEBUG: Autosave failed: %s\n", err.Error())
			}
			fmt.Println("Thanks for playing!")
			return
		}

		// Process command
		if err := game.ProcessCommand(input); err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println(game.Output)
		}
	}
}
