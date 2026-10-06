import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI } from '@/api'
import { clearScrollMemory } from '@/composables/useScrollMemory'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<{ id: number; username: string; family_id?: number | null } | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  async function restoreSession() {
    if (!token.value) return
    try {
      const res = await authAPI.getMe()
      user.value = res.data
    } catch (e: any) {
      // 只有认证失败（401/403：token 过期/被撤销）才登出。
      // 断网或后端重启时 getMe 也会 reject——此时清掉 token 等于把网络抖一下的
      // 正常用户踢到登录页，本地 token 本来还有效（登录页也进不去反而更糟）。
      const status = e?.response?.status
      if (status === 401 || status === 403) logout()
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
    // 与 keep-alive 的 cacheInclude 登出清空同理：换账号后不应沿用上个账号的滚动位置
    clearScrollMemory()
  }

  return { token, user, isLoggedIn, restoreSession, login, register, logout }
})
