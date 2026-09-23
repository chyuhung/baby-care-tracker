<template>
  <!-- 整块页头随内容上滑滚走，不吸顶 -->
  <header ref="rootRef" class="pt-safe">
    <div class="flex items-end justify-between gap-3 px-4 pt-2 pb-1">
      <h1 v-if="large"
        class="min-w-0 text-[34px] font-bold leading-tight tracking-tight text-text-primary truncate">{{ title }}</h1>
      <span v-else class="min-w-0 text-[17px] font-semibold leading-tight text-text-primary truncate">{{ inlineTitleText }}</span>
      <div v-if="$slots.actions" class="flex flex-shrink-0 items-center gap-2 pb-1">
        <slot name="actions" />
      </div>
    </div>
    <div v-if="large && $slots.sub" class="px-4 pb-1">
      <slot name="sub" />
    </div>
    <div v-if="$slots.filters" class="px-4 pb-2">
      <slot name="filters" />
    </div>
  </header>

  <!-- 页头滚出屏幕后浮出的毛玻璃小标题条（仅展示，不拦截指针事件） -->
  <Teleport to="body">
    <div class="pointer-events-none fixed top-0 left-1/2 z-40 w-full max-w-[480px] -translate-x-1/2 glass-surface hairline-bottom pt-safe transition-opacity duration-200"
      :class="collapsed ? 'opacity-100' : 'opacity-0'">
      <div class="flex h-11 items-center px-4">
        <span class="truncate text-[17px] font-semibold text-text-primary">{{ inlineTitleText }}</span>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  title: string
  inlineTitle?: string
  /** 滚动偏移：页头滚出视野即浮出小标题 */
  scrollTop?: number
  /** 该页是否使用 Large Title（列表型页面用；表单页传 false 用 inline） */
  large?: boolean
}>(), { inlineTitle: '', scrollTop: 0, large: true })

// 实测页头高度，作为小标题条浮出阈值（带滞回避免抖动）
const rootRef = ref<HTMLElement | null>(null)
const headerH = ref(0)
let ro: ResizeObserver | undefined

onMounted(() => {
  if (!rootRef.value) return
  headerH.value = rootRef.value.offsetHeight
  if (typeof ResizeObserver !== 'undefined') {
    ro = new ResizeObserver(() => {
      if (rootRef.value) headerH.value = rootRef.value.offsetHeight
    })
    ro.observe(rootRef.value)
  }
})
onBeforeUnmount(() => ro?.disconnect())

const collapsed = ref(false)
watch(() => props.scrollTop, (v) => {
  if (v >= headerH.value + 8) collapsed.value = true
  else if (v <= headerH.value - 8) collapsed.value = false
}, { immediate: true })

const inlineTitleText = computed(() => props.inlineTitle || props.title)
</script>