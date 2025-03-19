package main

import (
	"fmt"

	"github.com/andrewsjg/goAdventure/advent"
)

var ADVENT_AUTOSAVE = true

func main() {

	// TODO: Implement autosave logic

	// TODO: Implement command line argument handling

	game := advent.NewGame(0)

	fmt.Println("ZZWORD:  " + string(game.Zzword[:]))
	fmt.Printf("Seedval: %d\n", game.Seedval)

}
