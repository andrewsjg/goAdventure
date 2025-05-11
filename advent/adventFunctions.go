package advent

import "github.com/andrewsjg/goAdventure/dungeon"

// TODO: refactor these perhaps
func (g *Game) at(object int32) bool {
	return g.Objects[object].Place == g.Loc ||
		g.Objects[object].Fixed == g.Loc
}

func (g *Game) here(object int) bool {
	return g.at(int32(object)) || g.toting(object)
}

func (g *Game) toting(object int) bool {
	return g.Objects[object].Place == CARRIED
}

func (g *Game) objectIsNotFound(object int) bool {
	return g.Objects[object].Prop == STATE_NOTFOUND
}

func (g *Game) objectStashed(object int) int32 {
	return g.propStashify(g.Objects[object].Prop)
}

func (g *Game) propStashify(object int32) int32 {
	return (-1 - (object))
}

func (g *Game) dark() bool {

	return !condbit(g.Loc, dungeon.COND_LIT) &&
		(g.Objects[dungeon.LAMP].Prop == dungeon.LAMP_DARK ||
			!g.here(int(dungeon.LAMP)))
}

func (g *Game) objectIsStashed(object int) bool {
	return g.Objects[object].Prop < STATE_NOTFOUND
}

func (g *Game) objectIsStashedOrUnseen(object int) bool {
	return g.Objects[object].Prop < 0
}

func (g *Game) objectSetFound(object int) bool {
	return g.Objects[object].Prop == STATE_FOUND
}

func (g *Game) objectIsFound(object int) bool {
	return g.Objects[object].Prop == STATE_FOUND
}

func (g *Game) LocForced() bool {
	return condbit(g.Loc, dungeon.COND_FORCED)
}

func (g *Game) MoveHere() {
	g.PlayerMove(int32(dungeon.HERE))
}

func (g *Game) LiqLoc() int32 {
	if condbit(g.Loc, dungeon.COND_FLUID) {
		if condbit(g.Loc, dungeon.COND_OILY) {
			return int32(dungeon.OIL)
		}
		return int32(dungeon.WATER)
	}

	return int32(dungeon.NO_OBJECT)
}

/*
 *  DESTROY(N)  = Get rid of an item by putting it in LOC_NOWHERE
 *  MOD(N,M)    = Arithmetic modulus
 *  TOTING(OBJ) = true if the OBJ is being carried
 *  AT(OBJ)     = true if on either side of two-placed object
 *  HERE(OBJ)   = true if the OBJ is at "LOC" (or is being carried)
 *  CNDBIT(L,N) = true if COND(L) has bit n set (bit 0 is units bit)
 *  LIQUID()    = object number of liquid in bottle
 *  LIQLOC(LOC) = object number of liquid (if any) at LOC
 *  FORCED(LOC) = true if LOC moves without asking for input (COND=2)
 *  DARK(LOC)   = true if location "LOC" is dark
 *  PCT(N)      = true N% of the time (N integer from 0 to 100)
 *  GSTONE(OBJ) = true if OBJ is a gemstone
 *  FOREST(LOC) = true if LOC is part of the forest
 *  OUTSID(LOC) = true if location not in the cave
 *  INSIDE(LOC) = true if location is in the cave or the building at the
 * beginning of the game INDEEP(LOC) = true if location is in the Hall of Mists
 * or deeper BUG(X)      = report bug and exit
 */

func forest(location int32) bool {
	return condbit(location, dungeon.COND_FOREST)
}

func outside(loction int32) bool {
	return condbit(loction, dungeon.COND_ABOVE) || forest(loction)
}

func inside(location int32) bool {
	return !outside(location) || location == int32(dungeon.LOC_BUILDING)
}

func tstbit(mask int32, bit int32) bool {

	return (mask & (1 << bit)) != 0
}

func condbit(L int32, N int32) bool {
	return tstbit(dungeon.Conditions[L], N)
}

// TODO: Could refactor to use LocForced as defined above
func forced(location int32) bool {
	return condbit(location, dungeon.COND_FORCED)
}

func indeep(location int32) bool {
	return condbit(location, dungeon.COND_DEEP)
}
