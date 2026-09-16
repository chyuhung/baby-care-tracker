<template>
  <div :class="rootClass">
    <!-- 顶部固定指示胶囊（整页下拉刷新） -->
    <div v-show="topPillVisible"
      class="fixed inset-x-0 top-0 z-50 flex justify-center pointer-events-none transition-opacity duration-200">
      <span class="inline-flex items-center gap-1 bg-white/85 backdrop-blur border border-border-color rounded-full px-4 py-1.5 text-xs font-medium text-text-secondary shadow-card whitespace-nowrap mt-[calc(env(safe-area-inset-top)+10px)]">
        <span v-if="refreshing" class="icon-spin inline-block">🔄</span>
        <template v-else>{{ pulling >= REFRESH_VISUAL ? '🔄 释放刷新' : '⬇️ 下拉刷新' }}</template>
        <span v-if="refreshing"> 刷新中...</span>
      </span>
    </div>

    <div :class="contentClass" :style="slotStyle">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, useAttrs, onMounted, onUnmounted } from 'vue'

const props = withDefaults(defineProps<{
  refresh: () => Promise<unknown> | unknown
  contentClass?: string
}>(), { contentClass: '' })

const attrs = useAttrs()
const rootClass = computed(() => `relative ${(attrs.class as string) || ''}`)

const REFRESH_VISUAL = 30

const pulling = ref(0)
const refreshing = ref(false)
const animating = ref(false)
const active = ref(false)
const armed = ref(false)

let startY = 0
let mouseDown = false
let performedGesture = false
let touchMoveHandler: ((e: TouchEvent) => void) | null = null

const topPillVisible = computed(() => refreshing.value || pulling.value > 0)

const slotStyle = computed(() => {
  const hold = refreshing.value ? REFRESH_VISUAL : pulling.value
  const transform = hold > 0.5 ? `translateY(${hold}px)` : 'none'
  const transition = animating.value && !active.value
    ? 'transform 0.3s cubic-bezier(0.16, 1, 0.3, 1)'
    : 'none'
  return { transform, transition }
})

// 整页滚动即文档滚动：警顶判断用 window.scrollY
function atPageTop() {
  return window.scrollY <= 0
}

function beginDrag(y: number) {
  if (refreshing.value) return
  startY = y
  active.value = true
  animating.value = false
  armed.value = false
  performedGesture = false
  pulling.value = 0
}

function moveDrag(y: number, prevent: () => void) {
  if (!active.value || refreshing.value) return
  const dy = y - startY
  if (!atPageTop() || dy <= 0) {
    armed.value = false
    pulling.value = 0
    return
  }
  prevent()
  pulling.value = Math.min(dy * 0.45, 110)
  if (pulling.value >= REFRESH_VISUAL) performedGesture = true
  armed.value = pulling.value >= REFRESH_VISUAL
}

function endDrag() {
  if (!active.value) return
  active.value = false
  animating.value = true
  if (armed.value) {
    refreshing.value = true
    Promise.resolve()
      .then(() => props.refresh())
      .catch(() => {})
      .finally(() => {
        refreshing.value = false
        pulling.value = 0
        armed.value = false
        animating.value = true
      })
  } else {
    pulling.value = 0
    armed.value = false
  }
  restoreSelect()
  scheduleClickSuppress()
}

function cancelDrag() {
  if (!active.value) return
  active.value = false
  animating.value = true
  pulling.value = 0
  armed.value = false
  restoreSelect()
  scheduleClickSuppress()
}

/* ---------- 触屏（window 级，覆盖整页含头部） ---------- */
function onTouchStart(e: TouchEvent) {
  if (e.touches.length !== 1) return
  beginDrag(e.touches[0].clientY)
}

function onTouchMove(e: TouchEvent) {
  const prevent = () => { if (e.cancelable) e.preventDefault() }
  moveDrag(e.touches[0].clientY, prevent)
}

function onTouchEnd() { endDrag() }
function onTouchCancel() { cancelDrag() }

/* ---------- 鼠标拖拽（window 级） ---------- */
function onMouseDown(e: MouseEvent) {
  if (e.button !== 0 || mouseDown) return
  const target = e.target as HTMLElement | null
  if (target?.closest('input, select, textarea, a, [contenteditable="true"]')) return
  beginDrag(e.clientY)
  if (!active.value) return
  mouseDown = true
  window.addEventListener('mousemove', onMouseMove)
  window.addEventListener('mouseup', onMouseUp)
  document.body.style.userSelect = 'none'
  document.body.style.webkitUserSelect = 'none'
}

function onMouseMove(e: MouseEvent) {
  if (!mouseDown) return
  moveDrag(e.clientY, () => {})
}

function onMouseUp() {
  if (!mouseDown) return
  mouseDown = false
  window.removeEventListener('mousemove', onMouseMove)
  window.removeEventListener('mouseup', onMouseUp)
  endDrag()
}

function restoreSelect() {
  if (mouseDown) {
    mouseDown = false
    window.removeEventListener('mousemove', onMouseMove)
    window.removeEventListener('mouseup', onMouseUp)
  }
  document.body.style.userSelect = ''
  document.body.style.webkitUserSelect = ''
}

function suppressClick(e: Event) {
  e.stopPropagation()
  e.preventDefault()
}

function scheduleClickSuppress() {
  if (performedGesture) {
    window.addEventListener('click', suppressClick, { capture: true, once: true })
    performedGesture = false
  }
}

onMounted(() => {
  window.addEventListener('touchstart', onTouchStart, { passive: true })
  touchMoveHandler = onTouchMove
  window.addEventListener('touchmove', touchMoveHandler, { passive: false })
  window.addEventListener('touchend', onTouchEnd, { passive: true })
  window.addEventListener('touchcancel', onTouchCancel, { passive: true })
  window.addEventListener('mousedown', onMouseDown, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('touchstart', onTouchStart)
  if (touchMoveHandler) window.removeEventListener('touchmove', touchMoveHandler)
  window.removeEventListener('touchend', onTouchEnd)
  window.removeEventListener('touchcancel', onTouchCancel)
  window.removeEventListener('mousedown', onMouseDown)
  restoreSelect()
})
</script>