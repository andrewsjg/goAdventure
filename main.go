package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/andrewsjg/goAdventure/advent"
	"github.com/andrewsjg/goAdventure/ollama"
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

	// AI player flags
	aiMode := false
	aiModel := "qwen2.5:7b"
	ollamaURL := "http://localhost:11434"
	aiThinking := false
	aiDelay := 1000
	aiTimeout := 120
	aiTemp := 0.1 // Low temperature for more deterministic responses

	flag.StringVar(&logFileName, "l", "", "Create a log file of your game named as specified")
	flag.BoolVar(&oldStyle, "o", false, "'Oldstyle' mode (no prompt, no command editing, displays 'Initialising...')")
	flag.StringVar(&autoSaveFileName, "a", "", "Automatic save/restore from specified saved game file")
	flag.StringVar(&restoreFileName, "r", "", "Restore from specified saved game file")
	flag.StringVar(&scriptFileName, "script", "", "Execute commands from script file (one command per line, # for comments)")
	flag.BoolVar(&debug, "d", false, "Enable debug mode")
	flag.BoolVar(&noTUI, "notui", false, "Run without TUI (classic terminal mode)")
	flag.BoolVar(&enableTracing, "trace", false, "Enable OpenTelemetry tracing (sends to localhost:4318 by default)")
	flag.StringVar(&tracingEndpoint, "trace-endpoint", "", "OpenTelemetry OTLP endpoint (e.g., localhost:4318)")

	// AI player flags
	flag.BoolVar(&aiMode, "ai", false, "Enable AI player mode (uses Ollama)")
	flag.StringVar(&aiModel, "model", "qwen2.5:7b", "Ollama model to use for AI player")
	flag.StringVar(&ollamaURL, "ollama-url", "http://localhost:11434", "Ollama API URL")
	flag.BoolVar(&aiThinking, "ai-thinking", false, "Show AI reasoning/thinking")
	flag.IntVar(&aiDelay, "ai-delay", 1000, "Delay between AI moves in milliseconds")
	flag.IntVar(&aiTimeout, "ai-timeout", 120, "Timeout for AI requests in seconds")
	flag.Float64Var(&aiTemp, "ai-temp", 0.1, "AI temperature (0.0=deterministic, 1.0=creative)")

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

	// Create AI player and reward tracker if enabled
	var aiPlayer *ollama.Player
	var rewardTracker *ollama.RewardTracker
	if aiMode {
		client := ollama.NewClient(ollamaURL, aiModel, time.Duration(aiTimeout)*time.Second, aiTemp)
		aiPlayer = ollama.NewPlayer(client, aiThinking)
		rewardTracker = ollama.NewRewardTracker()
		if debug {
			fmt.Printf("AI player enabled using model: %s at %s (timeout: %ds, temp: %.2f)\n", aiModel, ollamaURL, aiTimeout, aiTemp)
		}
	}

	if noTUI {
		// Run in classic terminal mode
		runClassicMode(&game, aiPlayer, rewardTracker, aiThinking, aiDelay)
	} else {
		// Create a new game with TUI
		adventure := tui.NewAdventure(&game, aiPlayer, rewardTracker, aiThinking, time.Duration(aiDelay)*time.Millisecond)

		// Start the game
		if _, err := adventure.Run(); err != nil {
			fmt.Println("Error running the adventure:", err)
			return
		}
	}
}

// buildGameContext creates a GameContext from the current game state.
func buildGameContext(game *advent.Game, rewardTracker *ollama.RewardTracker) *ollama.GameContext {
	ctx := &ollama.GameContext{
		GameOutput:      game.Output,
		LocationDesc:    game.GetLocationDescription(),
		VisibleObjects:  game.GetVisibleObjects(),
		Inventory:       game.InventoryDescriptions(),
		Score:           game.GetScore(),
		Turns:           int(game.Turns),
		Hints:           game.GenerateHints(),
		ValidActions:    game.GetAllVerbs(),
		ValidDirections: game.GetAllDirections(),
		ValidObjects:    game.GetInteractableObjects(),
	}
	if rewardTracker != nil {
		ctx.RewardFeedback = rewardTracker.GetFeedback()
	}
	return ctx
}

func runClassicMode(game *advent.Game, aiPlayer *ollama.Player, rewardTracker *ollama.RewardTracker, showThinking bool, aiDelay int) {
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
			} else if aiPlayer != nil {
				// AI handles query response
				ctx := buildGameContext(game, rewardTracker)
				cmd, thinking, err := aiPlayer.GetCommand(ctx)
				if err != nil {
					fmt.Printf("\nAI Error: %v\n", err)
					return
				}
				if thinking != "" && showThinking {
					fmt.Printf("\n[AI Thinking: %s]\n", thinking)
				}
				response = cmd
				fmt.Printf("\n> %s\n", response)
				time.Sleep(time.Duration(aiDelay) * time.Millisecond)
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

		// Get command from script, AI, or user
		var input string
		var scoreBefore int
		if cmd, ok := game.NextScriptCommand(); ok {
			input = cmd
			fmt.Printf("> %s\n", input) // Echo the script command
		} else if aiPlayer != nil {
			// Track score before AI action for reward feedback
			scoreBefore = game.GetScore()

			// AI generates command based on game context
			ctx := buildGameContext(game, rewardTracker)
			cmd, thinking, err := aiPlayer.GetCommand(ctx)
			if err != nil {
				fmt.Printf("AI Error: %v\n", err)
				return
			}
			if thinking != "" && showThinking {
				fmt.Printf("[AI Thinking: %s]\n", thinking)
			}
			fmt.Printf("> %s\n", cmd) // Show AI's command
			input = cmd
			time.Sleep(time.Duration(aiDelay) * time.Millisecond)
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

		// Handle exit (skip for AI - it doesn't quit)
		if aiPlayer == nil && (strings.ToLower(input) == "quit" || strings.ToLower(input) == "exit") {
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

		// Record reward for AI actions
		if aiPlayer != nil && rewardTracker != nil {
			scoreAfter := game.GetScore()
			died := game.GameOver // Simplified death detection
			rewardTracker.RecordAction(input, scoreBefore, scoreAfter, died)
		}
	}
}
