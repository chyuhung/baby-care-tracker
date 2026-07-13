import { defineStore } from 'pinia'
import { ref } from 'vue'
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

  const currentBaby = () => babies.value.find(b => b.id === currentBabyId.value) || babies.value[0]

  async function loadBabies() {
    const res = await babyAPI.list()
    babies.value = res.data
    if (babies.value.length > 0 && !currentBabyId.value) {
      setCurrentBaby(babies.value[0].id)
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
    ws.onopen = () => { wsConnected.value = true }
    ws.onclose = () => {
      wsConnected.value = false
      ws = null
      setTimeout(connectWebSocket, 3000)
    }
    ws.onmessage = async (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'record_created' || msg.type === 'record_deleted') {
          // 触发页面刷新
          window.dispatchEvent(new CustomEvent('app:record-changed'))
        }
      } catch {}
    }
  }

  function disconnectWebSocket() {
    ws?.close()
    ws = null
  }

  return {
    babies, currentBabyId, toasts, wsConnected,
    currentBaby, loadBabies, setCurrentBaby, showToast,
    connectWebSocket, disconnectWebSocket,
  }
})
