package main

import (
	"fmt"

	"github.com/andrewsjg/goAdventure/dungeon"
)

func main() {

	for _, msg := range dungeon.Arbitrary_Messages {
		fmt.Printf("MSG: %s\n\n", msg)
	}

}
