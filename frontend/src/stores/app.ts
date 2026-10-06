import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { babyAPI } from '@/api'
import { useAuthStore } from './auth'
import { backendReachable, offline, initReachability, probeBackendNow, sinceBootMs, BOOT_GRACE_MS } from '@/utils/reachability'

export interface Baby {
  id: number
  user_id: number
  name: string
  birth_date: string
  gender: string
  avatar_color: string
  created_at: string
}

export interface ToastAction {
  label: string
  handler: () => void
}

export interface ToastMessage {
  id: number
  message: string
  type: 'success' | 'error' | 'info'
  action?: ToastAction
}

export const useAppStore = defineStore('app', () => {
  const babies = ref<Baby[]>([])
  const currentBabyId = ref<number | null>(Number(localStorage.getItem('currentBabyId')) || null)
  const toasts = ref<ToastMessage[]>([])
  const wsConnected = ref(false)
  let toastCounter = 0
  const toastTimers = new Map<number, ReturnType<typeof setTimeout>>()
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
    if (gender === 'female') return '#F25C8C'
    if (gender === 'male') return '#348EED'
    return '#F25C8C'
  }

  async function loadBabies() {
    try {
      const res = await babyAPI.list()
      babies.value = res.data
      // 剪枝记忆的 currentBabyId：宝宝可能已在别处删除（软删除后不在列表里），
      // 留着失效 id 会让 currentBaby 静默回落到 babies[0] 而 localStorage 仍记着旧值
      if (babies.value.length === 0) {
        if (currentBabyId.value !== null) {
          currentBabyId.value = null
          localStorage.removeItem('currentBabyId')
        }
      } else if (!babies.value.some(b => b.id === currentBabyId.value)) {
        setCurrentBaby(babies.value[0].id)
      }
      return true
    } catch (e) {
      console.error('加载宝宝列表失败', e)
      return false
    }
  }

  function setCurrentBaby(id: number) {
    currentBabyId.value = id
    localStorage.setItem('currentBabyId', String(id))
  }

  function dismissToast(id: number) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  function showToast(
    message: string,
    type: 'success' | 'error' | 'info' = 'success',
    action?: ToastAction,
    duration?: number,
  ) {
    const id = ++toastCounter
    toasts.value.push({ id, message, type, action })
    // 带「撤销」的 toast 停留更久；普通 toast 2.5s
    const ttl = duration ?? (action ? 5000 : 2500)
    const timer = setTimeout(() => {
      toastTimers.delete(id)
      dismissToast(id)
    }, ttl)
    toastTimers.set(id, timer)
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
      if (document.hidden) return
      // 不断线 toast：WS 是纯接收的同步通道，不参与读写判定，
      // 运营商 NAT 会回收空闲连接——为它报警等于误报。
      // 真正影响使用的是 HTTP，提示统一由 backendReachable 驱动（见 reachability.ts）。
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000)
      reconnectAttempts++
      const jitter = Math.random() * 1000
      reconnectTimer = setTimeout(connectWebSocket, delay + jitter)
    }
    ws.onmessage = async (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'record_created') {
          window.dispatchEvent(new CustomEvent('record-created', { detail: msg.payload }))
        } else if (msg.type === 'record_deleted') {
          window.dispatchEvent(new CustomEvent('record-deleted', { detail: msg.payload }))
        } else if (msg.type === 'record_updated') {
          window.dispatchEvent(new CustomEvent('record-updated', { detail: msg.payload }))
        } else if (msg.type === 'baby_created') {
          // WS 已按家庭过滤，收到即本家庭的宝宝。列表顺序与 GetBabies 的 created_at DESC 一致
          const b = msg.payload as Baby
          if (b?.id && !babies.value.some(x => x.id === b.id)) {
            babies.value = [b, ...babies.value]
            if (!currentBabyId.value) setCurrentBaby(b.id)
          }
        } else if (msg.type === 'baby_updated') {
          const b = msg.payload as Baby
          const i = babies.value.findIndex(x => x.id === b?.id)
          if (i >= 0) babies.value.splice(i, 1, b)
        } else if (msg.type === 'baby_deleted') {
          const id = (msg.payload as { id?: number })?.id
          if (id) removeBabyLocal(id)
        }
      } catch (e) {
        console.error('WebSocket 消息解析失败:', e)
      }
    }
  }

  // 本地移除宝宝（家人删除广播用）：同步剪 currentBabyId，
  // currentBaby 计算属性自动落到下一个宝宝，各页 watch(currentBaby.id) 触发重载
  function removeBabyLocal(id: number) {
    babies.value = babies.value.filter(x => x.id !== id)
    if (currentBabyId.value === id) {
      currentBabyId.value = null
      localStorage.removeItem('currentBabyId')
      if (babies.value.length > 0) setCurrentBaby(babies.value[0].id)
    }
  }

  function disconnectWebSocket() {
    if (reconnectTimer !== null) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    reconnectAttempts = 0
    const sock = ws
    ws = null
    if (sock) {
      // 先摘 onclose 再 close()：close 触发的 onclose 是异步回调，不摘的话 logout 后
      // 残留的 onclose 会按退避时间把刚断开的 socket 重新连上（登出又重新登录出现了旧连接）。
      sock.onclose = null
      sock.close()
    }
  }

  function onVisibilityChange() {
    if (!document.hidden && !ws && useAuthStore().token) {
      reconnectAttempts = 0 // 回前台重连不背后台累积的退避次数
      connectWebSocket()
    }
  }
  if (typeof document !== 'undefined') {
    document.addEventListener('visibilitychange', onVisibilityChange)
  }

  // 离线提示：延迟 5s 再报，期间恢复就撤销。
  // 与「立即禁用提交」故意不同步——提交该早封（发出去也是白费），
  // 而提示不该为一次瞬时抖动惊动用户（这正是启动时那对 toast 的成因）。
  // 冷启动例外：宽限期（10s）内不判离线，判定必然发生在 10s 之后；
  // 此时已等待过整个连接建立窗口，提示按「超出宽限的部分」即时补足（打开后约 10s 出），
  // 不再叠满 5s（否则真离线要到 15s 才提示）。启动窗口之后的离线仍保留完整 5s 防抖。
  const OFFLINE_TOAST_DELAY_MS = 5000
  let offlineToastTimer: ReturnType<typeof setTimeout> | null = null
  let offlineToastShown = false

  watch(backendReachable, (ok) => {
    if (offlineToastTimer !== null) {
      clearTimeout(offlineToastTimer)
      offlineToastTimer = null
    }
    if (!ok) {
      const sinceBoot = sinceBootMs()
      const delay = Math.min(OFFLINE_TOAST_DELAY_MS, Math.max(0, sinceBoot - BOOT_GRACE_MS))
      offlineToastTimer = setTimeout(() => {
        offlineToastTimer = null
        if (backendReachable.value) return
        offlineToastShown = true
        showToast('当前离线，无法保存新记录', 'error')
      }, delay)
    } else if (offlineToastShown) {
      // 只在真的报过离线后才提示恢复，否则又是一次「凭空多出一个 toast」
      offlineToastShown = false
      showToast('已重新连接', 'success')
    }
  })

  initReachability()

  return {
    babies, currentBabyId, toasts, wsConnected, offline, theme,
    currentBaby, loadBabies, setCurrentBaby, showToast, dismissToast,
    connectWebSocket, disconnectWebSocket, defaultAvatarColor,
    retryBackend: probeBackendNow,
  }
})
