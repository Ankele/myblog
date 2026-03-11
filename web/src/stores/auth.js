import { defineStore } from 'pinia'

import { fetchMe, login as loginRequest, logout as logoutRequest } from '../api/admin'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    initialized: false,
    pending: false,
  }),
  getters: {
    isAuthenticated(state) {
      return Boolean(state.user)
    },
  },
  actions: {
    async fetchMe() {
      try {
        const response = await fetchMe()
        this.user = response.data
      } catch (error) {
        this.user = null
      } finally {
        this.initialized = true
      }
    },
    async login(payload) {
      this.pending = true
      try {
        const response = await loginRequest(payload)
        this.user = response.data
        this.initialized = true
        return response.data
      } finally {
        this.pending = false
      }
    },
    async logout() {
      await logoutRequest()
      this.user = null
      this.initialized = true
    },
  },
})
