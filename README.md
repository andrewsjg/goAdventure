# goAdventure
Go implementation of the classic [Colossal Cave adventure](https://en.wikipedia.org/wiki/Colossal_Cave_Adventure) by Will Crowther and Don Woods

Based off Eric S. Raymond's C Port: https://gitlab.com/esr/open-adventure

This implementation of Colossal Cave Adventure was created to explore the original source code (or at least the portable version of it) by way of reimplementing it in Go.

This Go implementation stays faithful to the original but adds some 'modern' affordance's, like a terminal user interface (TUI) for those that want it.  Additionally, for a bit of fun, the game includes [OpenTelemetry Tracing](https://opentelemetry.io/docs/concepts/signals/traces/) which allows the player to 'trace' their journey through the game.

There is an accompanying Blog Post (coming soon) about wh and how this was created.

## Installation

### Homebrew

## Building from source

### Requirements

- Python3 - To make the dungeon (see below)
- Go 1.2x

### Building

1. **Clone the repo**
    
    Run: ``` git clone https://github.com/andrewsjg/goAdventure.git ```

2. **Make the Dungeon**

    The game generate the dungeon using the defintion found in ```dungeon.yaml```. There is an included python script that  generates the Go source code for the specified dungeon. Use either ```make_dungeon.sh``` to setup the required python virtual environment, dependencies then build the dungeon. Or run ```make_dungeon.py``` directly

3. **Compile the binary**
    
    Run ```go build``` in the root of the repository. This will produce the ```goAdventure``` binary in the root of the repository.

## Update Log

- 04/2025 - Early scaffold. Nothing really works yet
- 06/2025 - Working game Engine
- 10/2025 - Basic TUI
- 12/2025 - Revamped and working TUI
- 01/2025 - Tracing, UI Polish
