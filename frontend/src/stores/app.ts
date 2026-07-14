import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { babyAPI } from '@/api'
import { useAuthStore } from './auth'

export interface Baby {
  id: number
  user_id: number
  name: string
  birth_date: string
  gender: string
  avatar_color: string
  created_at: string
}

export interface ToastMessage {
  id: number
  message: string
  type: 'success' | 'error' | 'info'
}

export const useAppStore = defineStore('app', () => {
  const babies = ref<Baby[]>([])
  const currentBabyId = ref<number | null>(Number(localStorage.getItem('currentBabyId')) || null)
  const toasts = ref<ToastMessage[]>([])
  const wsConnected = ref(false)
  let toastCounter = 0
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectAttempts = 0

  const currentBaby = computed(() => babies.value.find(b => b.id === currentBabyId.value) || babies.value[0])

  const theme = computed(() => {
    const g = currentBaby.value?.gender
    if (g === 'female') return 'female'
    if (g === 'male') return 'male'
    return 'neutral'
  })

  function defaultAvatarColor(gender: string): string {
    if (gender === 'female') return '#FF7EB3'
    if (gender === 'male') return '#4D9DFD'
    return '#7C6CFF'
  }

  async function loadBabies() {
    try {
      const res = await babyAPI.list()
      babies.value = res.data
      if (babies.value.length > 0 && !currentBabyId.value) {
        setCurrentBaby(babies.value[0].id)
      }
    } catch {
      console.error('加载宝宝列表失败')
    }
  }

  function setCurrentBaby(id: number) {
    currentBabyId.value = id
    localStorage.setItem('currentBabyId', String(id))
  }

  function showToast(message: string, type: 'success' | 'error' | 'info' = 'success') {
    const id = ++toastCounter
    toasts.value.push({ id, message, type })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, 2500)
  }

  function connectWebSocket() {
    const auth = useAuthStore()
    if (!auth.token || ws) return
    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    ws = new WebSocket(`${protocol}//${location.host}/ws?token=${auth.token}`)
    ws.onopen = () => {
      wsConnected.value = true
      reconnectAttempts = 0
    }
    ws.onclose = () => {
      wsConnected.value = false
      ws = null
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000)
      reconnectAttempts++
      reconnectTimer = setTimeout(connectWebSocket, delay)
    }
    ws.onmessage = async (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'record_created') {
          window.dispatchEvent(new CustomEvent('record-created', { detail: msg.payload }))
        } else if (msg.type === 'record_deleted') {
          window.dispatchEvent(new CustomEvent('record-deleted', { detail: msg.payload }))
        }
      } catch (e) {
        console.error('WebSocket 消息解析失败:', e)
      }
    }
  }

  function disconnectWebSocket() {
    if (reconnectTimer !== null) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    reconnectAttempts = 0
    ws?.close()
    ws = null
  }

  return {
    babies, currentBabyId, toasts, wsConnected, theme,
    currentBaby, loadBabies, setCurrentBaby, showToast,
    connectWebSocket, disconnectWebSocket, defaultAvatarColor,
  }
})
