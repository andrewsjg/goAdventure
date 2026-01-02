package advent

import "github.com/andrewsjg/goAdventure/dungeon"

type Game struct {
	QueryFlag       bool
	QueryResponse   string
	OnQueryResponse func(response string, game *Game) string `json:"-"`
	Output          string
	OutputType      int // 0 = Regular, 1 == temporary message
	LcgX            int32
	Abbnum          int32
	Bonus           ScoreBonus
	Chloc           int32
	Chloc2          int32
	Clock1          int32
	Clock2          int32
	Clshnt          bool
	Closed          bool
	Closing         bool
	Lmwarn          bool
	Novice          bool
	Panic           bool
	Wzdark          bool
	Blooded         bool
	Conds           int32
	Detail          int32
	Dflag           int32
	Dkill           int32
	Dtotal          int32
	Foobar          int32
	Holdng          int32
	Igo             int32
	Iwest           int32
	Knfloc          int32
	Limit           int32
	Loc             int32
	Newloc          int32
	Numdie          int32
	Oldloc          int32
	Oldlc2          int32
	Oldobj          int32
	Saved           int32
	Tally           int32
	Thresh          int32
	Seenbigwords    bool
	Trnluz          int32
	Turns           int32
	Seedval         int
	Zzword          [5 + 1]byte
	Locs            [dungeon.NLOCATIONS + 1]struct {
		Abbrev int32
		Atloc  int32
	}
	Dwarves [dungeon.NDWARVES + 1]struct {
		Seen   bool
		Loc    int32
		Oldloc int32
	}

	//Objects []dungeon.Object_t

	Objects [dungeon.NOBJECTS + 1]struct {
		Found bool
		Fixed int32
		Prop  int32
		Place int32
	}

	Hints [dungeon.NHINTS]struct {
		Used bool
		Lc   int32
	}
	Link [dungeon.NOBJECTS*2 + 1]int32

	Settings Settings
}

const (
	LCG_A = 1093
	LCG_C = 221587
	LCG_M = 1048576

	CARRIED        = -1
	STATE_NOTFOUND = -1
	STATE_FOUND    = 0

	ADVENT_MAGIC = "goAdventure\n"
	SAVE_VERSION = 1

	NOVICELIMIT = 1000

	PANICTIME = 15

	DALTLC = dungeon.LOC_NUGGET // alternate dwarf location

	PIRATE = dungeon.NDWARVES

	IS_FIXED      = -1
	IS_FREE       = 0
	PIT_KILL_PROB = 35
	WARNTIME      = 30
	FLASHTIME     = 50
	BATTERYLIFE   = 2500
	GAMELIMIT     = 330

	WORD_EMPTY = 0

	MSG_TEMP = 1
	MSG_REG  = 0

	TOKLEN         = 5
	WORD_NOT_FOUND = -1
)

type CmdState int

const (
	EMPTY CmdState = iota
	RAW
	TOKENISED
	GIVEN
	PREPROCESSED
	PROCESSING
	EXECUTED
)

type SpeechPart int

const (
	Unknown SpeechPart = iota
	Intransitive
	Transitive
)

type WordType int

const (
	NO_WORD_TYPE WordType = iota
	MOTION
	OBJECT
	ACTION
	NUMERIC
)

type Command_Word struct {
	Raw      string
	ID       int
	WordType WordType
}

type Command struct {
	Part     SpeechPart
	Word     []Command_Word
	Verb     int
	Obj      int
	CmdState CmdState
}

type Save struct {
	Magic   string
	Version int
	Canary  int
	Game    Game
}

type Settings struct {
	Autosave         bool
	NewGame          bool
	LogFileName      string
	OldStyle         bool
	AutoSaveFileName string
	RestoreFileName  string
	EnableDebug      bool
	Scripts          []string
}

type Travel struct {
	DestType  string
	DestVal   int
	NoDwarves bool
	Stop      bool
}

type Termination int

const (
	EndGame Termination = iota
	QuitGame
	ScoreGame
)

type ScoreBonus int

const (
	None ScoreBonus = iota
	Splatter
	Defeat
	Victory
)

type SpeakType int

const (
	Touch SpeakType = iota
	Look
	Hear
	Study
	Change
)
