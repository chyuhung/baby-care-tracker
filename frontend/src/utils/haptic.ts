/* ============================================================
   触觉反馈（Haptics）— iOS UIImpactFeedbackGenerator 的 Web 近似
   ------------------------------------------------------------
   · 原生桥优先（Capacitor / 宿主注入 window.__haptic）
   · 否则降级 navigator.vibrate（Android Chrome 支持；iOS Safari 忽略）
   · 用户可在「我的」页关闭（localStorage 持久化）
   ============================================================ */

export type HapticStyle =
  | 'light'      // 轻点：tab 切换 / 普通按钮
  | 'medium'     // 中等：主要操作
  | 'heavy'      // 重：破坏性操作
  | 'selection'  // 选择变更：分段控件 / 滚轮
  | 'success'    // 成功
  | 'warning'    // 警告
  | 'error'      // 错误

const PATTERNS: Record<HapticStyle, number | number[]> = {
  light: 10,
  medium: 18,
  heavy: 30,
  selection: 6,
  success: [12, 40, 18],
  warning: [18, 60, 18],
  error: [26, 50, 26, 50, 26],
}

const STORAGE_KEY = 'haptics'

let enabled = true
try {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved !== null) enabled = saved === '1'
} catch { /* SSR / 隐私模式 */ }

export function isHapticsEnabled() {
  return enabled
}

export function setHapticsEnabled(v: boolean) {
  enabled = v
  try { localStorage.setItem(STORAGE_KEY, v ? '1' : '0') } catch { /* ignore */ }
}

/** 触发一次触觉反馈 */
export function haptic(style: HapticStyle = 'light') {
  if (!enabled) return
  const pattern = PATTERNS[style]
  try {
    // 原生桥（Capacitor / 自定义宿主）
    const bridge = (window as any).__haptic
    if (typeof bridge === 'function') { bridge(style); return }
    if (bridge && typeof bridge.impact === 'function') { bridge.impact(style); return }
    const nav = navigator as Navigator & { vibrate?: (p: number | number[]) => boolean }
    if (typeof nav.vibrate === 'function') nav.vibrate(pattern)
  } catch { /* 静默：触觉属增强能力 */ }
}

export const hapticTap = () => haptic('light')
export const hapticMedium = () => haptic('medium')
export const hapticHeavy = () => haptic('heavy')
export const hapticSelection = () => haptic('selection')
export const hapticSuccess = () => haptic('success')
export const hapticWarning = () => haptic('warning')
export const hapticError = () => haptic('error')
export const hapticTick = () => haptic('selection')
