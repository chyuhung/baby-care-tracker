<template>
  <!-- iOS 风格导航栏：Large Title 与 inline 标题共用一行（不额外占高），随滚动交叉淡入 -->
  <header class="sticky top-0 z-30 glass-surface hairline-bottom pt-safe">
    <!-- 标题行：标题与右侧操作（如「同步」状态）同基线居中，避免大片留白 -->
    <div class="flex items-center justify-between gap-3 px-4 h-11">
      <div class="relative flex-1 min-w-0 h-11">
        <!-- Large Title（滚动后收起、淡出） -->
        <span
          class="absolute inset-0 flex items-center text-[34px] font-bold tracking-tight text-text-primary truncate transition-opacity duration-200"
          :style="{ opacity: collapsed ? 0 : 1, pointerEvents: 'none' }"
        >{{ title }}</span>
        <!-- Inline 标题（滚动后淡入） -->
        <span
          class="absolute inset-0 flex items-center text-[17px] font-semibold text-text-primary truncate transition-opacity duration-200"
          :style="{ opacity: collapsed ? 1 : 0 }"
        >{{ inlineTitleText }}</span>
      </div>
      <div class="flex items-center gap-2 flex-shrink-0">
        <slot name="actions" />
      </div>
    </div>
    <!-- 副标题（紧凑，不额外撑高标题行） -->
    <div v-if="$slots.sub" class="px-4 pb-0.5">
      <slot name="sub" />
    </div>
    <!-- 筛选控件（贴标题行下沿，压缩垂直间距） -->
    <div v-if="$slots.filters" class="px-4 pb-2">
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
  /** 该页是否使用 Large Title（列表型页面用；表单页传 false 用 inline） */
  large?: boolean
}>(), { inlineTitle: '', scrollTop: 0, large: true })

const COLLAPSE_AT = 22
const EXPAND_AT = 6

// 本地状态 + 滞回，避免在阈值附近抖动
const folded = ref(false)
watch(() => props.scrollTop, (v) => {
  if (!props.large) return
  if (!folded.value && v > COLLAPSE_AT) folded.value = true
  else if (folded.value && v < EXPAND_AT) folded.value = false
})

const collapsed = computed(() => !props.large || folded.value)
const inlineTitleText = computed(() => props.inlineTitle || props.title)
</script>
