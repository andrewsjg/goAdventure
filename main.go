package main

import (
	"fmt"

	"github.com/andrewsjg/goAdventure/advent"
)

var ADVENT_AUTOSAVE = true

func main() {

	// TODO: Implement autosave logic

	// TODO: Implement command line argument handling

	game := advent.NewGame()

	fmt.Println(string(game.Zzword[:]))

}
