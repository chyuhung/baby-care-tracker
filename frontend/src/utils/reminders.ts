// 提醒（本地通知）：基于 Notification API + localStorage 持久化。
// 说明：PWA 在无 Service Worker push 的情况下只能在前台/拉起时提醒，
// 这里采用「每分钟检查 + 回到前台时补发」的策略，满足日常喂养提醒。

export type ReminderKind = 'feeding' | 'diaper' | 'sleep' | 'temperature' | 'supplement' | 'pump'

export interface Reminder {
  id: string
  kind: ReminderKind
  label: string
  emoji: string
  /** 间隔提醒：每隔 intervalMin 分钟提醒一次（自上次记录起算） */
  intervalMin: number
  enabled: boolean
  /** 上次触发日期，避免同一天重复 */
  lastFired?: string
}

const LS_KEY = 'babytracker.reminders'
const LS_FIRED = 'babytracker.reminderFired'

export const DEFAULT_REMINDERS: Reminder[] = [
  { id: 'feeding', kind: 'feeding', label: '喂奶提醒', emoji: '🍼', intervalMin: 180, enabled: false },
  { id: 'diaper', kind: 'diaper', label: '换尿布提醒', emoji: '🩲', intervalMin: 120, enabled: false },
  { id: 'sleep', kind: 'sleep', label: '睡眠记录提醒', emoji: '😴', intervalMin: 240, enabled: false },
  { id: 'temperature', kind: 'temperature', label: '体温测量提醒', emoji: '🌡️', intervalMin: 360, enabled: false },
  { id: 'supplement', kind: 'supplement', label: '补剂提醒', emoji: '💊', intervalMin: 720, enabled: false },
  { id: 'pump', kind: 'pump', label: '泵奶提醒', emoji: '🥛', intervalMin: 180, enabled: false },
]

export function loadReminders(): Reminder[] {
  try {
    const raw = localStorage.getItem(LS_KEY)
    if (!raw) return DEFAULT_REMINDERS.map(r => ({ ...r }))
    const parsed = JSON.parse(raw) as Reminder[]
    // 与默认项对齐（新增类型自动补齐）
    return DEFAULT_REMINDERS.map(d => {
      const found = parsed.find(p => p.id === d.id)
      return found ? { ...d, ...found } : { ...d }
    })
  } catch {
    return DEFAULT_REMINDERS.map(r => ({ ...r }))
  }
}

export function saveReminders(list: Reminder[]) {
  localStorage.setItem(LS_KEY, JSON.stringify(list))
}

export function notificationsSupported(): boolean {
  return typeof window !== 'undefined' && 'Notification' in window
}

export function notificationPermission(): NotificationPermission | 'unsupported' {
  if (!notificationsSupported()) return 'unsupported'
  return Notification.permission
}

export async function requestNotificationPermission(): Promise<boolean> {
  if (!notificationsSupported()) return false
  if (Notification.permission === 'granted') return true
  if (Notification.permission === 'denied') return false
  try {
    const p = await Notification.requestPermission()
    return p === 'granted'
  } catch {
    return false
  }
}

function todayKey(): string {
  const d = new Date()
  const p2 = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p2(d.getMonth() + 1)}-${p2(d.getDate())}`
}

function firedMap(): Record<string, string> {
  try {
    return JSON.parse(localStorage.getItem(LS_FIRED) || '{}')
  } catch {
    return {}
  }
}

function markFired(id: string) {
  const m = firedMap()
  m[id] = todayKey()
  localStorage.setItem(LS_FIRED, JSON.stringify(m))
}

function alreadyFiredToday(id: string): boolean {
  return firedMap()[id] === todayKey()
}

export function showReminderNotification(r: Reminder) {
  if (!notificationsSupported() || Notification.permission !== 'granted') return
  try {
    const n = new Notification(`${r.emoji} ${r.label}`, {
      body: '该记录啦，点击打开宝宝护理记录',
      tag: 'babytracker-' + r.id,
      icon: '/icon-192.png',
      badge: '/icon-192.png',
    } as NotificationOptions)
    n.onclick = () => {
      window.focus()
      n.close()
    }
  } catch {
    // 某些环境（如 iOS 未安装为 PWA）不支持直接构造 Notification
  }
}

/**
 * 启动提醒调度：每分钟检查一次当前时间是否命中某个提醒的间隔；
 * 回到前台时也会检查一次（补发当天漏掉的提醒）。
 */
export function startReminderScheduler(getList: () => Reminder[]) {
  let timer: number | null = null

  const check = () => {
    const list = getList()
    const now = new Date()
    for (const r of list) {
      if (!r.enabled) continue
      if (alreadyFiredToday(r.id)) continue
      // 命中判定：按 interval 将一天切成若干时点，当前分钟跨越时点即触发
      const minutesToday = now.getHours() * 60 + now.getMinutes()
      if (r.intervalMin > 0 && minutesToday % r.intervalMin === 0) {
        showReminderNotification(r)
        markFired(r.id)
      }
    }
  }

  const start = () => {
    if (timer !== null) return
    check()
    timer = window.setInterval(check, 60_000)
  }
  const stop = () => {
    if (timer !== null) { window.clearInterval(timer); timer = null }
  }
  const onVis = () => {
    if (document.visibilityState === 'visible') check()
  }
  document.addEventListener('visibilitychange', onVis)
  start()

  return () => {
    stop()
    document.removeEventListener('visibilitychange', onVis)
  }
}
