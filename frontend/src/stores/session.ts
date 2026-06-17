import { defineStore } from 'pinia'

export const useSessionStore = defineStore('session', {
  state: () => ({
    accessToken: '',
    isAuthenticated: false,
  }),
  actions: {
    setAccessToken(token: string) {
      this.accessToken = token
      this.isAuthenticated = token.trim().length > 0
    },
    clear() {
      this.accessToken = ''
      this.isAuthenticated = false
    },
  },
})
