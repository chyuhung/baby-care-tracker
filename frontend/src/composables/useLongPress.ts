/* ============================================================
   useLongPress — 长按手势（iOS 上下文菜单触发）
   ------------------------------------------------------------
   · 默认 480ms 触发
   · 手指移动超过 10px 视为滚动，取消
   · 返回 handlers 供元素 v-on 绑定
   ============================================================ */

export interface LongPressOptions {
  delay?: number
  moveTolerance?: number
}

export function useLongPress(
  onLongPress: (payload: { x: number; y: number; target: EventTarget | null }) => void,
  opts: LongPressOptions = {},
) {
  const delay = opts.delay ?? 480
  const tol = opts.moveTolerance ?? 10

  let timer: number | null = null
  let startX = 0
  let startY = 0
  let fired = false
  let target: EventTarget | null = null

  function clear() {
    if (timer) { clearTimeout(timer); timer = null }
  }

  function onTouchStart(e: TouchEvent) {
    if (e.touches.length !== 1) return
    const t = e.touches[0]
    startX = t.clientX; startY = t.clientY
    fired = false
    target = e.currentTarget
    clear()
    timer = window.setTimeout(() => {
      fired = true
      onLongPress({ x: startX, y: startY, target })
    }, delay)
  }

  function onTouchMove(e: TouchEvent) {
    if (timer == null || fired) return
    const t = e.touches[0]
    if (Math.abs(t.clientX - startX) > tol || Math.abs(t.clientY - startY) > tol) clear()
  }

  function onTouchEnd() { clear() }
  function onTouchCancel() { clear() }

  /** 长按后紧随的 click 需要被吞掉（返回 true 表示本次点击应忽略并复位） */
  function consumeClick() {
    if (fired) { fired = false; return true }
    return false
  }

  return { onTouchStart, onTouchMove, onTouchEnd, onTouchCancel, consumeClick }
}
