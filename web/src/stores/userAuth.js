import { defineStore } from 'pinia'

import {
  fetchMe,
  login as loginRequest,
  logout as logoutRequest,
  register as registerRequest,
} from '../api/userAuth'

export const useUserAuthStore = defineStore('user-auth', {
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
    async register(payload) {
      this.pending = true
      try {
        const response = await registerRequest(payload)
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
