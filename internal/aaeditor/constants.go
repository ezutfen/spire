package aaeditor

// AaCategories maps aa_ability.category to a human readable label. These mirror
// the EQEmu client AA category groupings.
var AaCategories = map[int]string{
	1: "General",
	2: "Archetype",
	3: "Class",
	4: "Special",
	5: "Expansion",
	6: "Prestige",
}

// AaTypes maps aa_ability.type.
var AaTypes = map[int]string{
	0: "Passive",
	1: "Activated",
	2: "Combat",
	3: "Triggered",
}

// AaSpellTypes maps aa_ranks.spell_type.
var AaSpellTypes = map[int]string{
	0: "None",
	1: "Buff",
	2: "Detrimental",
	3: "Discipline",
	4: "Song",
}

// AaStatuses maps aa_ability.status.
var AaStatuses = map[int]string{
	0:  "Available",
	-1: "Hidden",
	1:  "Special",
}

// AaExpansions maps aa_ranks.expansion bitmask values.
var AaExpansions = map[int]string{
	0:  "Classic",
	1:  "Kunark",
	2:  "Velious",
	3:  "Luclin",
	4:  "Planes of Power",
	5:  "Lost Dungeons of Norrath",
	6:  "Gates of Discord",
	7:  "Omens of War",
	8:  "Dragons of Norrath",
	9:  "Depths of Darkhollow",
	10: "Prophecy of Ro",
	11: "The Serpent's Spine",
	12: "The Buried Sea",
	13: "Secrets of Faydwer",
	14: "Seeds of Destruction",
	15: "Underfoot",
	16: "House of Thule",
	17: "Veil of Alaris",
	18: "Rain of Fear",
	19: "Call of the Forsaken",
	20: "The Darkend Sea",
	21: "Empires of Kunark",
	22: "Ring of Scale",
	23: "The Burning Lands",
	24: "Torment of Velious",
	25: "Claws of Veeshan",
	26: "Terror of Luclin",
	27: "Night of Shadows",
	28: "Laurion's Song",
}
