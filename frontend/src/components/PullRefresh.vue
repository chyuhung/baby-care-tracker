<template>
  <div ref="scrollerRef" :class="rootClass" @mousedown="onMouseDown">
    <!-- 顶部：下拉刷新胶囊 -->
    <div v-show="topPillVisible"
      class="absolute inset-x-0 top-0 z-10 flex justify-center pointer-events-none transition-opacity duration-200">
      <span class="inline-flex items-center gap-1 bg-white/85 backdrop-blur border border-border-color rounded-full px-4 py-1.5 text-xs font-medium text-text-secondary shadow-card whitespace-nowrap mt-3">
        <span v-if="refreshing" class="icon-spin inline-block">🔄</span>
        <template v-else>{{ pulling >= REFRESH_VISUAL ? '🔄 释放刷新' : '⬇️ 下拉刷新' }}</template>
        <span v-if="refreshing"> 刷新中...</span>
      </span>
    </div>

    <div ref="slotWrap" :class="contentClass" :style="slotStyle">
      <slot />
    </div>

    <!-- 底部：上拉加载更多胶囊 -->
    <div v-show="bottomPillVisible"
      class="absolute inset-x-0 bottom-0 z-10 flex justify-center pointer-events-none transition-opacity duration-200">
      <span class="inline-flex items-center gap-1 bg-white/85 backdrop-blur border border-border-color rounded-full px-4 py-1.5 text-xs font-medium text-text-secondary shadow-card whitespace-nowrap mb-3">
        <span v-if="loadingMore" class="icon-spin inline-block">🔄</span>{{ loadingMore ? ' 加载中...' : '⬆️ 上拉加载更多' }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, useAttrs, onMounted, onUnmounted } from 'vue'

const props = withDefaults(defineProps<{
  refresh: () => Promise<unknown> | unknown
  loadMore?: () => Promise<unknown> | unknown
  contentClass?: string
}>(), { contentClass: '' })

const attrs = useAttrs()
const rootClass = computed(() => `relative overflow-y-auto overscroll-contain ${(attrs.class as string) || ''}`)

const REFRESH_RAW = 70
const REFRESH_VISUAL = 30
const LOAD_MORE_RAW = 30
const LOAD_MORE_VISUAL = 14

const scrollerRef = ref<HTMLElement | null>(null)

const pulling = ref(0)
const bottomPulling = ref(0)
const refreshing = ref(false)
const loadingMore = ref(false)
const animating = ref(false)
const active = ref(false)
const dragMode = ref<'none' | 'top' | 'bottom'>('none')

let startY = 0
let mouseDown = false
let performedGesture = false
let touchMoveHandler: ((e: TouchEvent) => void) | null = null

const topPillVisible = computed(() => refreshing.value || pulling.value > 0)
const bottomPillVisible = computed(() => bottomPulling.value > 0 || loadingMore.value)

const slotStyle = computed(() => {
  const hold = refreshing.value ? REFRESH_VISUAL : pulling.value
  const holdBottom = loadingMore.value ? LOAD_MORE_VISUAL : bottomPulling.value
  const offset = hold - holdBottom
  const transform = offset > 0.5 || offset < -0.5 ? `translateY(${offset}px)` : 'none'
  const transition = animating.value && !active.value
    ? 'transform 0.3s cubic-bezier(0.16, 1, 0.3, 1)'
    : 'none'
  return { transform, transition }
})

function beginDrag(y: number) {
  if (refreshing.value || loadingMore.value) return
  startY = y
  active.value = true
  animating.value = false
  dragMode.value = 'none'
  performedGesture = false
}

function moveDrag(y: number, prevent: () => void) {
  if (!active.value || refreshing.value || loadingMore.value) return
  const s = scrollerRef.value
  if (!s) return
  const dy = y - startY
  const atTop = s.scrollTop <= 0
  const atBottom = s.scrollTop + s.clientHeight >= s.scrollHeight - 4

  if (dragMode.value === 'none') {
    if (dy > 0 && atTop) { dragMode.value = 'top'; prevent() }
    else if (dy < 0 && atBottom && props.loadMore) { dragMode.value = 'bottom'; prevent() }
    else return
  }

  if (dragMode.value === 'top') {
    if (dy > 0) {
      prevent()
      pulling.value = Math.min(dy * 0.45, 110)
      if (pulling.value >= REFRESH_VISUAL) performedGesture = true
    } else {
      pulling.value = 0
      dragMode.value = 'none'
    }
  } else if (dragMode.value === 'bottom') {
    if (dy < 0) {
      prevent()
      bottomPulling.value = Math.min((-dy) * 0.45, 60)
      if (bottomPulling.value >= LOAD_MORE_VISUAL) performedGesture = true
    } else {
      bottomPulling.value = 0
      dragMode.value = 'none'
    }
  }
}

function endDrag() {
  if (!active.value) return
  active.value = false
  animating.value = true
  if (dragMode.value === 'top' && pulling.value >= REFRESH_VISUAL) {
    refreshing.value = true
    Promise.resolve()
      .then(() => props.refresh())
      .catch(() => {})
      .finally(() => {
        refreshing.value = false
        pulling.value = 0
        animating.value = true
      })
  } else if (dragMode.value === 'bottom' && bottomPulling.value >= LOAD_MORE_VISUAL && props.loadMore) {
    loadingMore.value = true
    Promise.resolve()
      .then(() => props.loadMore!())
      .catch(() => {})
      .finally(() => {
        loadingMore.value = false
        bottomPulling.value = 0
        animating.value = true
      })
  } else {
    pulling.value = 0
    bottomPulling.value = 0
  }
  dragMode.value = 'none'
  restoreSelect()
  scheduleClickSuppress()
}

function cancelDrag() {
  if (!active.value) return
  active.value = false
  animating.value = true
  pulling.value = 0
  bottomPulling.value = 0
  dragMode.value = 'none'
  restoreSelect()
  scheduleClickSuppress()
}

/* ---------- 触屏 ---------- */
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

/* ---------- 鼠标拖拽 ---------- */
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
  const el = scrollerRef.value
  if (!el) return
  el.addEventListener('touchstart', onTouchStart, { passive: true })
  touchMoveHandler = onTouchMove
  el.addEventListener('touchmove', touchMoveHandler, { passive: false })
  el.addEventListener('touchend', onTouchEnd, { passive: true })
  el.addEventListener('touchcancel', onTouchCancel, { passive: true })
})

onUnmounted(() => {
  const el = scrollerRef.value
  if (el) {
    el.removeEventListener('touchstart', onTouchStart)
    if (touchMoveHandler) el.removeEventListener('touchmove', touchMoveHandler)
    el.removeEventListener('touchend', onTouchEnd)
    el.removeEventListener('touchcancel', onTouchCancel)
  }
  restoreSelect()
})
</script>