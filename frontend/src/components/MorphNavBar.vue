<template>
  <!-- iOS 大标题导航：吸顶，标题随滚动进度原地 morph（34px→17px），玻璃/hairline 随进度淡入（顶部严格 flat）。
       与首页「记录」标题栏同规格；actions 区不触发回顶 -->
  <header ref="chromeRef" class="sticky top-0 z-30 hairline-bottom" role="button" aria-label="返回顶部"
    @click="scrollToTop" :style="chromeStyle">
    <div class="pt-safe">
      <div ref="titleRowRef" class="flex h-11 items-center justify-between gap-3 px-4">
        <h1 class="min-w-0 truncate font-bold text-text-primary"
          :style="{ fontSize: `${34 - 17 * morphP}px`, lineHeight: '1', letterSpacing: '-0.02em' }">{{ title }}</h1>
        <div v-if="$slots.actions" class="flex flex-shrink-0 items-center gap-2" @click.stop>
          <slot name="actions" />
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'

const props = defineProps<{ title: string; scrollTop?: number }>()

// ── iOS 大标题 morph：滚动进度驱动标题从 34px 缩到 17px，玻璃随进度淡入 ──
const chromeRef = ref<HTMLElement | null>(null)
const titleRowRef = ref<HTMLElement | null>(null)
const morphEnd = ref(0)
let ro: ResizeObserver | undefined

function updateMorphEnd() {
  const chrome = chromeRef.value
  const row = titleRowRef.value
  if (!chrome || !row) return
  const cb = chrome.getBoundingClientRect()
  const rb = row.getBoundingClientRect()
  morphEnd.value = Math.max(1, rb.top - cb.top + rb.height)
}

onMounted(() => {
  updateMorphEnd()
  if (typeof ResizeObserver !== 'undefined' && chromeRef.value) {
    ro = new ResizeObserver(updateMorphEnd)
    ro.observe(chromeRef.value)
  }
})
onUnmounted(() => ro?.disconnect())
watch(() => props.title, updateMorphEnd)

const morphP = computed(() => {
  const e = morphEnd.value
  return e > 0 ? Math.min(1, Math.max(0, props.scrollTop ?? 0) / e) : 0
})

// 玻璃背景与 hairline 透明度随滚动进度渐变（初始 0 = 严格 flat），文字层不受影响
const chromeStyle = computed(() => {
  const p = morphP.value
  return {
    background: `rgb(var(--surface) / ${0.72 * p})`,
    '--hairline-alpha': `${0.26 * p}`,
    ...(p > 0
      ? {
          backdropFilter: 'saturate(180%) blur(20px)',
          WebkitBackdropFilter: 'saturate(180%) blur(20px)',
        }
      : {}),
  }
})

// 单击标题栏 → 各页 PullRefresh 容器回顶
function scrollToTop() {
  window.dispatchEvent(new CustomEvent('app:scroll-to-top'))
}
</script>