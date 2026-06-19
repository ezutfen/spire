import { defineStore } from 'pinia'

export const useAppStore = defineStore('app', {
  state: () => ({
    appEnvLoaded: false,
    debugEnabled: false,
    spellLegacyIconsEnabled: false,
  }),
  actions: {
    setAppEnvLoaded(value: boolean) {
      this.appEnvLoaded = value
    },
    setDebugEnabled(value: boolean) {
      this.debugEnabled = value
    },
    setSpellLegacyIconsEnabled(value: boolean) {
      this.spellLegacyIconsEnabled = value
    },
  },
})
