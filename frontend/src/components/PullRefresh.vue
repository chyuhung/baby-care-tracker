<template>
  <div ref="rootRef" :class="rootClass">
    <div :class="contentClass">
      <slot />
    </div>

    <!-- iOS 风顶部活动指示器（Teleport 到 body，不随整页位移；无胶囊底、无文字） -->
    <Teleport to="body">
      <div v-show="indicatorVisible"
        class="fixed inset-x-0 top-0 z-50 pointer-events-none flex justify-center"
        :style="{ paddingTop: 'calc(env(safe-area-inset-top) + 10px)' }">
        <!-- 下拉过程：圆弧随手指旋转、随距离淡入 -->
        <svg v-if="!refreshing" width="28" height="28" viewBox="0 0 24 24" fill="none"
          class="text-text-secondary"
          :style="{ opacity: pullOpacity, transform: `rotate(${pulling * 2.4}deg)` }">
          <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2.4" stroke-linecap="round"
            stroke-dasharray="40 56.55" transform="rotate(-90 12 12)" />
        </svg>
        <!-- 刷新中：持续旋转的活动指示器 -->
        <ActivityIndicator v-else :size="28" class="text-text-secondary" />
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, useAttrs, onMounted, onUnmounted } from 'vue'
import ActivityIndicator from './ActivityIndicator.vue'

const props = withDefaults(defineProps<{
  refresh: () => Promise<unknown> | unknown
  contentClass?: string
}>(), { contentClass: '' })

const attrs = useAttrs()
const rootClass = computed(() => `relative ${(attrs.class as string) || ''}`)

const REFRESH_VISUAL = 30
const TAP_SLOP = 10

const rootRef = ref<HTMLElement | null>(null)
let pageEl: HTMLElement | null = null

const pulling = ref(0)
const refreshing = ref(false)
const animating = ref(false)
const active = ref(false)
const armed = ref(false)
const startAtTop = ref(false)

let startY = 0
let mouseDown = false
let performedGesture = false
let touchMoveHandler: ((e: TouchEvent) => void) | null = null
let lockedUp = false

const indicatorVisible = computed(() => refreshing.value || pulling.value > 0)
const pullOpacity = computed(() => {
  if (refreshing.value) return 1
  return Math.min(pulling.value / REFRESH_VISUAL, 1)
})

// 整页滚动即文档滚动
function atPageTop() {
  return window.scrollY <= 0
}

// 位移整个页面（含头部），实现「整页下拉」
function applyTransform() {
  if (!pageEl) return
  const hold = refreshing.value ? REFRESH_VISUAL : pulling.value
  if (hold > 0.5) {
    pageEl.style.transform = `translateY(${hold}px)`
    pageEl.style.transition = animating.value && !active.value
      ? 'transform 0.3s cubic-bezier(0.16, 1, 0.3, 1)'
      : 'none'
  } else {
    pageEl.style.transform = ''
    pageEl.style.transition = ''
  }
}

watch([pulling, refreshing, animating, active], applyTransform)

function beginDrag(y: number) {
  if (refreshing.value) return
  startY = y
  active.value = true
  animating.value = false
  armed.value = false
  performedGesture = false
  pulling.value = 0
  lockedUp = false
  startAtTop.value = atPageTop()
}

function moveDrag(y: number, prevent: () => void) {
  if (!active.value || refreshing.value) return
  const dy = y - startY
  if (dy < 0) lockedUp = true
  // 点击死区：手指微抖（<10px）不接管手势，避免 preventDefault 吞掉正常点击
  if (Math.abs(dy) < TAP_SLOP) {
    armed.value = false
    pulling.value = 0
    return
  }
  // 仅「页面整体」在顶部起始的下拉才触发刷新；中部产生的下拉只用于滚动
  if (!startAtTop.value || lockedUp) {
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
    startAtTop.value = false
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
  pageEl = rootRef.value?.parentElement ?? null
  window.addEventListener('touchstart', onTouchStart, { passive: true })
  touchMoveHandler = onTouchMove
  window.addEventListener('touchmove', touchMoveHandler, { passive: false })
  window.addEventListener('touchend', onTouchEnd, { passive: true })
  window.addEventListener('touchcancel', onTouchCancel, { passive: true })
  window.addEventListener('mousedown', onMouseDown)
})

onUnmounted(() => {
  window.removeEventListener('touchstart', onTouchStart)
  if (touchMoveHandler) window.removeEventListener('touchmove', touchMoveHandler)
  window.removeEventListener('touchend', onTouchEnd)
  window.removeEventListener('touchcancel', onTouchCancel)
  window.removeEventListener('mousedown', onMouseDown)
  if (pageEl) {
    pageEl.style.transform = ''
    pageEl.style.transition = ''
  }
  restoreSelect()
})
</script>
