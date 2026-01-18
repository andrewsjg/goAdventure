package advent

import "github.com/andrewsjg/goAdventure/dungeon"

// InventoryDescriptions returns the rendered inventory strings for carried items.
func (g *Game) InventoryDescriptions() []string {
	if g == nil {
		return nil
	}

	items := make([]string, 0)

	for i := 1; i <= dungeon.NOBJECTS; i++ {
		if i == dungeon.BEAR || !g.toting(i) {
			continue
		}

		text := dungeon.Objects[i].Inventory
		if text == "" {
			continue
		}

		rendered, err := g.vspeak(text, false)
		if err != nil || rendered == "" {
			continue
		}

		items = append(items, rendered)
	}

	if g.toting(dungeon.BEAR) {
		rendered, err := g.vspeak(dungeon.Arbitrary_Messages[dungeon.TAME_BEAR], false)
		if err == nil && rendered != "" {
			items = append(items, rendered)
		}
	}

	return items
}
