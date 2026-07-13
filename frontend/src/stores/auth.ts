import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI } from '@/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<{ id: number; username: string; family_id?: number | null } | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  async function restoreSession() {
    if (!token.value) return
    try {
      const res = await authAPI.getMe()
      user.value = res.data
    } catch {
      logout()
    }
  }

  async function login(username: string, password: string) {
    const res = await authAPI.login(username, password)
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
  }

  async function register(username: string, password: string) {
    const res = await authAPI.register(username, password)
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  return { token, user, isLoggedIn, restoreSession, login, register, logout }
})
