package dungeon

// NOTE: Remember all the enum blocks below are going to be out by one because of the iota's. Need to fix or accommodate in code

var Arbitrary_Messages = []string {{
{arbitrary_messages}
}};

var  Classes = []Class_t {{
{classes}
}};

var Turn_Thresholds  = []Turn_Threshold_t{{
    {turn_thresholds}
}}

var Locations  = []Location_t{{
    {locations}
}}

var Objects  = []Object_t {{
    {objects}
}}

var Obituaries  = []Obituary_t {{
    {obituaries}
}}

var Hints  = []Hint_t {{
    {hints}
}}

var Conditions  = []int64{{
    {conditions}
}}

var Motions = []Motion_t{{
    {motions}
}}

var Actions  = []Action_t{{
    {actions}
}}

var TKey  = []int64{{{tkeys}}}

var Travel  = []Travelop_t {{
    {travel}
}}

var Ignore string = "{ignore}"


/* Dwarf starting locations */
var DwarfLocs = [NDWARVES]int{{{dwarflocs}}} //location

/* end */
