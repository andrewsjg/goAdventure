package dungeon

const SILENT =	-1	/* no sound */

/* Symbols for cond bits */
const COND_LIT	= 0	/* Light */
const COND_OILY	= 1	/* If bit 2 is on: on for oil, off for water */
const COND_FLUID	= 2	/* Liquid asset, see bit 1 */
const COND_NOARRR	= 3	/* Pirate doesn't go here unless following */
const COND_NOBACK =	4	/* Cannot use "back" to move away */
const COND_ABOVE	= 5	/* Aboveground, but not in forest */
const COND_DEEP	= 6	/* Deep - e.g where dwarves are active */
const COND_FOREST	= 7	/* In the forest */
const COND_FORCED= 8	/* Only one way in or out of here */
const COND_ALLDIFFERENT	= 9	/* Room is in maze all different */
const COND_ALLALIKE	= 10	/* Room is in maze all alike */
const past = 11 /* indicate areas of interest to "hint" routines */
const COND_HBASE	= 11	/* Base for location hint bits */
const COND_HCAVE	= 12	/* Trying to get into cave */
const COND_HBIRD	= 13	/* Trying to catch bird */
const COND_HSNAKE	= 14	/* Trying to deal with snake */
const COND_HMAZE	= 15	/* Lost in maze */
const COND_HDARK	= 16	/* Pondering dark room */
const COND_HWITT	= 17	/* At Witt's End */
const COND_HCLIFF	= 18	/* Cliff with urn */
const COND_HWOODS	= 19	/* Lost in forest */
const COND_HOGRE	= 20	/* Trying to deal with ogre */
const COND_HJADE	= 21	/* Found all treasures except jade */
const NDWARVES    =  {ndwarflocs}          // number of dwarves

type Refs int

type String_Group_t struct {{
  Strs []string
  N int
}}

type Object_t struct {{
  Words String_Group_t
  Inventory string
  Plac Refs // Location
  Fixd Refs // Location
  Is_Treasure bool
  Descriptions []string
  Sounds []string
  Texts []string
  Changes []string
}}

type Descriptions_t struct {{
  Small string
  Big string
}}

type Location_t struct {{
  Description Descriptions_t
  Sound Refs // arbitrary_messages
  Loud bool
}}

type Obituary_t struct {{ 
  Query string
  Yes_Response string
}}

type Turn_Threshold_t struct {{
  Threshold int
  Point_loss int
  Message string
}}

type Class_t struct {{
  Threshold int
  Message string
}}

type Hint_t struct {{
  Number int
  Turns int
  Penalty int
  Question string
  Hint string
}}

type Motion_t struct {{
  Words String_Group_t
}}

type Action_t struct {{
  Words String_Group_t 
  Message string
  NoAction bool
}}

type CondType int
const (
    CondGoto CondType = iota
    CondPct
    CondCarry
    CondWith
    CondNot
)

type DestType int
const (
    DestGoto DestType = iota
    DestSpecial
    DestSpeak
)

type Travelop_t struct {{
  Motion Refs
  CondType CondType
  CondArg1 Refs
  CondArg2 int64
  DestType DestType
  DestVal  interface{{}} //Location_Refs
  NoDwarves bool
  Stop bool
}}

const NLOCATIONS = {num_locations}
const NOBJECTS =	{num_objects}
const NHINTS	=	{num_hints}
const NCLASSES =	{num_classes}
const NDEATHS	 =	{num_deaths}
const NTHRESHOLDS =	{num_thresholds}
const NMOTIONS   = {num_motions}
const NACTIONS  =	{num_actions}
const NTRAVEL	=	{num_travel}
const NKEYS	= 	{num_keys}

const BIRD_ENDSTATE = {bird_endstate}

const (
   {arbitrary_messages}
)

const (
  {locations}
)


const (
  {objects}
)

const (
  {motions}
)


const (
  {actions}
)

/* State definitions */

{state_definitions}
