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
  `relative flex-1 min-h-0 overflow-y-auto overscroll-contain ${(attrs.class as string) || ''}`
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
let decided = false   // 是否已判定接管为"下拉"
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

/* ========== 经典下拉状态机（等价 iOS UIRefreshControl） ========== */

function onTouchStart(e: TouchEvent) {
  if (refreshing.value || e.touches.length !== 1) return
  measureHeader()
  startY = e.touches[0].clientY
  tracking = true
  decided = false
  performed = false
  animating.value = false
}

function onTouchMove(e: TouchEvent) {
  if (!tracking || refreshing.value) return
  const dy = e.touches[0].clientY - startY

  // 死区：位移太小视为点击
  if (!decided) {
    if (Math.abs(dy) < SLOP) return
    // 只接管"从顶部向下拉"；其它情况交给原生滚动
    if (dy < 0 || !atTop()) { tracking = false; return }
    decided = true
  }

  // 拖动过程中已不在顶部 → 复位，交给原生滚动
  if (!atTop()) {
    tracking = false
    pulling.value = 0
    armed.value = false
    return
  }

  // 真正进入下拉：接管触摸，压掉原生回弹 / 滚动条
  if (e.cancelable) e.preventDefault()
  pulling.value = Math.min(dy * RESIST, MAX_PULL)
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
