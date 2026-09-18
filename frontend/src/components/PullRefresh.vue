<template>
  <div ref="rootRef" :class="rootClass"
    @pointerdown="onPointerDown" @pointermove="onPointerMove"
    @pointerup="onPointerUp" @pointercancel="onPointerCancel">
    <!-- 顶部悬浮区（标题栏），渲染在内容上、不套内容边距 -->
    <slot name="header" />
    <div :class="contentClass">
      <slot />
    </div>

    <!-- iOS 风顶部活动指示器（Teleport 到 body，内容原地不动；无胶囊底、无文字） -->
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
import { ref, computed, useAttrs, onUnmounted } from 'vue'
import ActivityIndicator from './ActivityIndicator.vue'

const props = withDefaults(defineProps<{
  refresh: () => Promise<unknown> | unknown
  contentClass?: string
}>(), { contentClass: '' })

const attrs = useAttrs()
const rootClass = computed(() =>
  `relative flex-1 min-h-0 overflow-y-auto overscroll-contain touch-pan-y ${(attrs.class as string) || ''}`
)

/* 手势阈值 */
const SLOP = 10        // 点击死区：位移小于它不算手势，点击零干扰
const THRESHOLD = 60   // 越过该距离松手即触发刷新
const MAX_PULL = 110   // 指示器最大行程
const RESIST = 0.45    // 阻尼系数

/* 状态 */
const pulling = ref(0)
const refreshing = ref(false)
const armed = ref(false)

let active = false
let pointerId: number | null = null
let startY = 0
let performed = false

const indicatorVisible = computed(() => refreshing.value || pulling.value > 0)
const pullOpacity = computed(() => {
  if (refreshing.value) return 1
  return Math.min(pulling.value / THRESHOLD, 1)
})

function container() {
  return rootRef.value
}

/* 顶部判定：实时读取滚动容器自身 scrollTop，不依赖 window 快照 */
function atTop() {
  const el = container()
  return el ? el.scrollTop <= 0 : true
}

/* ========== 经典下拉状态机（等价 iOS UIRefreshControl） ========== */

function onPointerDown(e: PointerEvent) {
  if (refreshing.value) return
  if (active) return // 第二只手指落下：忽略
  active = true
  pointerId = e.pointerId
  startY = e.clientY
  pulling.value = 0
  armed.value = false
  performed = false
  try {
    container()?.setPointerCapture(e.pointerId)
  } catch {
    /* 旧浏览器无捕获也不影响 */
  }
}

function onPointerMove(e: PointerEvent) {
  if (!active || e.pointerId !== pointerId || refreshing.value) return
  const dy = e.clientY - startY

  // 死区：位移太小视为点击，不接管、不 preventDefault
  if (Math.abs(dy) < SLOP) return

  // 方向向上、或当前不在顶部 → 复位，交给原生滚动（pan-y）
  if (dy < 0 || !atTop()) {
    pulling.value = 0
    armed.value = false
    return
  }

  // 真正进入下拉：触屏接管，防止浏览器处理；鼠标拉下时禁用文本选择
  if (e.pointerType === 'touch' && e.cancelable) e.preventDefault()
  if (e.pointerType === 'mouse') document.body.classList.add('select-none')

  pulling.value = Math.min(dy * RESIST, MAX_PULL)
  armed.value = pulling.value >= THRESHOLD
  if (armed.value) performed = true
}

function onPointerUp() {
  finish()
}

function onPointerCancel() {
  // 浏览器接管滚动（pan-y）等场景：直接复位，无残留
  finish(false)
}

function finish(trigger: boolean = true) {
  if (!active) return
  active = false
  pointerId = null
  document.body.classList.remove('select-none')
  const shouldRefresh = trigger && armed.value
  if (shouldRefresh) {
    refreshing.value = true
    pulling.value = 0
    armed.value = false
    Promise.resolve()
      .then(() => props.refresh())
      .catch(() => {})
      .finally(() => {
        refreshing.value = false
      })
  } else {
    pulling.value = 0
    armed.value = false
  }
  if (performed) suppressNextClick()
}

/* 下拉手势结束后，吞掉紧随的合成 click（防幽灵点击） */
function suppressNextClick() {
  performed = false
  document.addEventListener('click', (e) => {
    e.stopPropagation()
    e.preventDefault()
  }, { capture: true, once: true })
}

onUnmounted(() => {
  document.body.classList.remove('select-none')
})
</script>