<template>
  <!-- iOS 风格 Large Title 导航栏：大标题随滚动折叠为 inline 标题 -->
  <header class="sticky top-0 z-30 glass-surface hairline-bottom pt-safe">
    <div class="flex items-end justify-between gap-2 px-4 min-h-[44px]">
      <!-- 折叠后的 inline 小标题（滚动后淡入） -->
      <h1 class="text-[17px] font-semibold text-text-primary leading-tight transition-opacity duration-200"
        :style="{ opacity: collapsed ? 1 : 0 }">
        {{ inlineTitleText }}
      </h1>
      <div class="flex items-center gap-2 flex-shrink-0 min-h-[44px]">
        <slot name="actions" />
      </div>
    </div>
    <!-- 大标题区（滚动时向上收起） -->
    <div class="px-4 overflow-hidden transition-[height] duration-200 ease-out"
      :style="{ height: largeHeight + 'px' }">
      <h1 class="text-[34px] font-bold text-text-primary leading-[52px] tracking-tight truncate">{{ title }}</h1>
    </div>
    <div v-if="$slots.sub || $slots.filters" class="px-4 pb-3">
      <slot name="sub" />
      <slot name="filters" />
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  title: string
  inlineTitle?: string
  /** 滚动偏移：>阈值即折叠 */
  scrollTop?: number
  /** 该页是否使用 Large Title（列表型页面用；表单页用 inline） */
  large?: boolean
}>(), { inlineTitle: '', scrollTop: 0, large: true })

const COLLAPSE_AT = 22
const EXPAND_AT = 6
const LARGE_H = 52

// 本地状态 + 滞回，避免在阈值附近抖动
const folded = ref(false)
watch(() => props.scrollTop, (v) => {
  if (!props.large) return
  if (!folded.value && v > COLLAPSE_AT) folded.value = true
  else if (folded.value && v < EXPAND_AT) folded.value = false
})

const collapsed = computed(() => !props.large || folded.value)
const largeHeight = computed(() => (collapsed.value ? 0 : LARGE_H))
const inlineTitleText = computed(() => props.inlineTitle || props.title)
</script>
