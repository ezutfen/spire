package aaeditor

// AaCategories maps aa_ability.category to the AA category labels used by
// PEQEditor / EQEmu AA data.
var AaCategories = map[int]string{
	0: "None",
	1: "Passive",
	2: "Progression",
	3: "Shroud Passive",
	4: "Shroud Active",
	5: "Veteran Reward",
	6: "Tradeskill",
	7: "Expendable",
	8: "Racial Innate",
	9: "Everquest",
}

// AaTypes maps aa_ability.type to the AA family / era labels used by
// PEQEditor / EQEmu AA data.
var AaTypes = map[int]string{
	0:  "Not Applicable",
	1:  "General",
	2:  "Archetype",
	3:  "Class",
	4:  "PoP Advanced",
	5:  "PoP Abilities",
	6:  "Gates of Discord",
	7:  "Omens of War",
	8:  "Veteran",
	9:  "Dragons of Norrath",
	10: "Depths of Darkhollow",
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
