<template>
  <div ref="rootRef" :class="rootClass">
    <!-- 顶部悬浮区（标题栏），渲染在内容上、不套内容边距 -->
    <slot name="header" />

    <!-- 内容区：下拉时整体下移，露出顶栏与内容之间的间隙 -->
    <div :class="contentClass" :style="contentStyle">
      <slot />
    </div>

    <!-- iOS 风顶部活动指示器（Teleport 到 body，浮层不受内容位移影响） -->
    <Teleport to="body">
      <div v-show="indicatorVisible"
        class="fixed inset-x-0 z-20 pointer-events-none flex justify-center"
        :style="{ top: indicatorTop + 'px' }">
        <!-- 下拉过程：圆弧随手指旋转、随距离淡入 -->
        <svg v-if="!refreshing" width="26" height="26" viewBox="0 0 24 24" fill="none"
          class="text-text-secondary"
          :style="{ opacity: pullOpacity, transform: `rotate(${pulling * 2.4}deg)` }">
          <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2.4" stroke-linecap="round"
            stroke-dasharray="40 56.55" transform="rotate(-90 12 12)" />
        </svg>
        <!-- 刷新中：持续旋转的活动指示器 -->
        <ActivityIndicator v-else :size="26" class="text-text-secondary" />
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, useAttrs, onMounted, onUnmounted } from 'vue'
import ActivityIndicator from './ActivityIndicator.vue'

const props = withDefaults(defineProps<{
  refresh: () => Promise<unknown> | unknown
  contentClass?: string
}>(), { contentClass: '' })

const attrs = useAttrs()
const rootRef = ref<HTMLElement | null>(null)
const rootClass = computed(() =>
  `relative flex-1 min-h-0 overflow-y-auto overscroll-none ${(attrs.class as string) || ''}`
)

/* 手势阈值 */
const SLOP = 10        // 点击死区：位移小于它不算手势，点击零干扰
const THRESHOLD = 60   // 越过该距离松手即触发刷新
const MAX_PULL = 110   // 内容最大行程
const RESIST = 0.45    // 阻尼系数
const HOLD = 56        // 刷新中内容顶住的高度
const SPIN = 26        // 指示器尺寸

/* 状态 */
const pulling = ref(0)
const refreshing = ref(false)
const armed = ref(false)
const animating = ref(false)
const headerH = ref(0)

let startY = 0
let tracking = false
let engaged = false    // 是否已进入"下拉接管"
let anchorY = 0       // 列表触顶那一刻的触点 Y（从此往上算下拉行程）
let lastY = 0         // 上一帧触点 Y（用于判定本次移动方向）
let performed = false

const indicatorVisible = computed(() => refreshing.value || pulling.value > 0)
const pullOpacity = computed(() => {
  if (refreshing.value) return 1
  return Math.min(pulling.value / THRESHOLD, 1)
})
/* 指示器垂直位置：顶栏底部 + 露出间隙的中央（iOS 观感） */
const indicatorTop = computed(() => {
  const gap = refreshing.value ? HOLD : pulling.value
  return Math.round(headerH.value + gap / 2 - SPIN / 2)
})
const contentStyle = computed(() => {
  const y = refreshing.value ? HOLD : pulling.value
  const transition = animating.value ? 'transform 0.3s cubic-bezier(0.16, 1, 0.3, 1)' : 'none'
  return { transform: y > 0.5 ? `translateY(${y}px)` : '', transition }
})

/* 顶部判定：实时读取滚动容器自身 scrollTop */
function atTop() {
  const el = rootRef.value
  return el ? el.scrollTop <= 0 : true
}
/* 顶栏高度：用于把指示器定位在顶栏下方的露出间隙中央 */
function measureHeader() {
  const el = rootRef.value
  const h = el ? el.querySelector('header') : null
  headerH.value = h ? Math.round(h.getBoundingClientRect().height) : 0
}

/* ========== 下拉状态机（等价 iOS UIRefreshControl） ==========
 * 连续模型：一次手势里，手指向下拖时先让原生滚动把列表带回顶部，
 * 一旦列表触顶（scrollTop<=0），就锚定此刻触点 Y，把之后的向下行程
 * 全部转成「整页下拉」位移，并 preventDefault 压掉原生 overscroll。
 * 因此「先上滑再回顶下拉」与「直接下拉」表现完全一致，
 * 且顶部始终是整页下拉刷新，永不露出底色（--bg-main 粉/白）或滚动条。 */

function onTouchStart(e: TouchEvent) {
  if (refreshing.value || e.touches.length !== 1) return
  measureHeader()
  startY = e.touches[0].clientY
  lastY = startY
  tracking = true
  engaged = false
  performed = false
  animating.value = false
  // 已在顶部：立刻锚定，保证从第一个像素就按整页下拉计算
  anchorY = atTop() ? startY : 0
}

function onTouchMove(e: TouchEvent) {
  if (!tracking || refreshing.value) return
  const el = rootRef.value
  if (!el) return
  const y = e.touches[0].clientY
  const dy = y - startY
  const dyMove = y - lastY
  lastY = y

  // 死区：位移太小视为点击
  if (!engaged && Math.abs(dy) < SLOP) return

  if (el.scrollTop > 0) {
    // 仍有原生滚动余量：交给浏览器滚动，重置锚点（回弹不参与下拉）
    anchorY = 0
    if (pulling.value) pulling.value = 0
    armed.value = false
    return
  }

  // 列表已在顶部
  if (!engaged) {
    if (dyMove <= 0) return // 本次在向上拖（=向下滚动）：交给原生滚动
    engaged = true
    anchorY = y
  }

  // 进入/维持「整页下拉」：接管触摸，压掉原生回弹 / 滚动条
  if (e.cancelable) e.preventDefault()
  pulling.value = Math.max(0, Math.min((y - anchorY) * RESIST, MAX_PULL))
  armed.value = pulling.value >= THRESHOLD
  if (armed.value) performed = true
}

function onTouchEnd() {
  finish(true)
}

function onCancel() {
  finish(false)
}

function finish(trigger: boolean) {
  if (!tracking) return
  tracking = false
  const shouldRefresh = trigger && armed.value
  animating.value = true
  if (shouldRefresh) {
    refreshing.value = true
    pulling.value = 0
    armed.value = false
    Promise.resolve()
      .then(() => props.refresh())
      .catch(() => {})
      .finally(() => {
        refreshing.value = false
        window.setTimeout(() => { animating.value = false }, 340)
      })
  } else {
    pulling.value = 0
    armed.value = false
    window.setTimeout(() => { animating.value = false }, 340)
  }
  if (performed) suppressNextClick()
  engaged = false
  anchorY = 0
  lastY = 0
}

/* 下拉手势结束后，吞掉紧随的合成 click（防幽灵点击） */
function suppressNextClick() {
  performed = false
  document.addEventListener('click', (e) => {
    e.stopPropagation()
    e.preventDefault()
  }, { capture: true, once: true })
}

onMounted(() => {
  const el = rootRef.value
  if (!el) return
  el.addEventListener('touchstart', onTouchStart, { passive: true })
  el.addEventListener('touchmove', onTouchMove, { passive: false }) // 非被动：可 preventDefault
  el.addEventListener('touchend', onTouchEnd)
  el.addEventListener('touchcancel', onCancel)
})

onUnmounted(() => {
  const el = rootRef.value
  if (!el) return
  el.removeEventListener('touchstart', onTouchStart)
  el.removeEventListener('touchmove', onTouchMove)
  el.removeEventListener('touchend', onTouchEnd)
  el.removeEventListener('touchcancel', onCancel)
})
</script>
