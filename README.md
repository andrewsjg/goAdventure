# goAdventure
Go implementation of the classic [Colossal Cave adventure](https://en.wikipedia.org/wiki/Colossal_Cave_Adventure) by Will Crowther and Don Woods

Based off Eric S. Raymond's C Port: https://gitlab.com/esr/open-adventure

This implementation of Colossal Cave Adventure was created to explore the original source code (or at least the portable version of it) by way of reimplementing it in Go.

This Go implementation stays faithful to the original but adds some 'modern' affordance's, like a terminal user interface (TUI) for those that want it.  Additionally, for a bit of fun, the game includes [OpenTelemetry Tracing](https://opentelemetry.io/docs/concepts/signals/traces/) which allows the player to 'trace' their journey through the game.

There is an accompanying [Blog Post](https://jgandrews.com/posts/colossal-cave/) about what this is and how this was created.

![Adventure TUI](images/advent-tui.png)

## Installation

The latest binaries can be found on the [releases page](https://github.com/andrewsjg/goAdventure/releases)

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

### Usage


Launching the game with just ```./goAdventure``` and no command line options will launch the TUI interface. The following options are available:

**General Options**

- ```-notui``` launch the game in 'classic' terminal mode
- ```-r <save file>``` Restore a game from a save file
- ```-a <autosave file>``` Specify a file to use for autosave
- ```-l <log file>``` Specify a file to log debug information
- ```-script <script file>``` Specify a walkthrough script to run. See example in this repo.

**Tracing Options** 

- ```-trace```  this will cause the game to emit [OpenTelemetry Traces](https://opentelemetry.io/docs/concepts/signals/traces/) as you progress through the game. The easiest way to see these is to use the [Jaeger All-in-one](https://www.jaegertracing.io/docs/1.76/getting-started/) docker container, which launches a collector and the Jaeger trace platform to view them. Launch it with:

`   docker run -d --name jaeger                                                                       
    -p 16686:16686
    -p 4318:4318
    jaegertracing/all-in-one:latest `

- ```-trace-endpoint <OTEL Collector endpoint>``` Specify an OTEL collector tracing endpoint. The default is `localhost:4318` which is what the Jaeger All-In-One collector will listen on


**AI Options**

- ```ai``` Watch a local LLM attempt to play the game. Relies on having [ollama](https://ollama.com) installed and a local model available. The default model is ``` qwen2.5:7b``` (Qwen 2.5 with 7 billion parameters).
- ```-model <model name>``` Specify a different model to use

- ```-ai-thinking``` Watch AI play with visible thinking/state machine phase (not working properly)
- ```-ai-delay``` AI with slower moves (2 second delay)
- ```-ai-temp <temperature>``` Tweak the AI temperature value. Between 0 and 1. Higher values result in more 'creative' responses. Lower values are better for game play
 - ```-ollama-url <url>:<port>``` Connect to a remote ollama server


The browse to ```http:\\localhost:16686``` to see the spans emitted by the game as you progress.



## Update Log

- 04/2025 - Early scaffold. Nothing really works yet
- 06/2025 - Working game Engine
- 10/2025 - Basic TUI
- 12/2025 - Revamped and working TUI
- 01/2026 - Tracing, UI Polish, GitHub Repo cleanup
