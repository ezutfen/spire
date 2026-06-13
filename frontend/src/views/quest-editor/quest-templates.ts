export const QuestTemplates = {
  blank(relativePath) {
    const name = relativePath.split('/').pop().replace('.lua', '')
    return `-- ${name}\n`
  },

  zoneController(zoneName) {
    return `-- Zone Controller for ${zoneName}\n\nfunction event_enter_zone(e)\nend\n\nfunction event_click_door(e)\nend\n\nfunction event_timer(e)\nend\n`
  },

  npc(npcName) {
    return `-- NPC Script: ${npcName || 'npc'}\n\nfunction event_say(e)\n  if e.message:findi("hail") then\n    e.other:Message(0, "Hello there, traveler.")\n  end\nend\n\nfunction event_trade(e)\n  local item_lib = require("item_lib")\n  item_lib.return_items(e.self, e.other, e.trade)\nend\n`
  },

  item(itemName) {
    return `-- Item Script: ${itemName || 'item'}\n\nfunction event_equip(e)\n  return false\nend\n\nfunction event_unequip(e)\n  return false\nend\n\nfunction event_proc(e)\nend\n`
  },

  global() {
    return `-- Global Quest Script\n\nfunction event_say(e)\nend\n\nfunction event_trade(e)\nend\n`
  },

  plugin(moduleName) {
    const name = moduleName || 'mymodule'
    return `-- Plugin Module: ${name}\n\nlocal ${name} = {}\n\nfunction ${name}.init()\nend\n\nreturn ${name}\n`
  },
}
