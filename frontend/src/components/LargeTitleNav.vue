<template>
  <!-- 吸顶紧凑栏：inline 标题随滚动淡入，操作按钮与筛选控件固定在顶部 -->
  <header class="sticky top-0 z-30 glass-surface hairline-bottom pt-safe">
    <div class="flex items-center justify-between gap-3 px-4 h-11">
      <div class="flex-1 min-w-0 text-[17px] font-semibold text-text-primary truncate transition-opacity duration-200"
        :style="{ opacity: collapsed ? 1 : 0 }" :aria-hidden="large ? 'true' : null">{{ inlineTitleText }}</div>
      <div class="flex items-center gap-2 flex-shrink-0">
        <slot name="actions" />
      </div>
    </div>
    <!-- 筛选控件（吸顶，始终可操作） -->
    <div v-if="$slots.filters" class="px-4 pb-2">
      <slot name="filters" />
    </div>
  </header>

  <!-- 流动大标题：常规文档流，随内容上滑、划入吸顶栏下方消失（iOS 原生行为） -->
  <div v-if="large || $slots.sub" class="px-4 pt-1.5 pb-1">
    <h1 v-if="large"
      class="text-[34px] font-bold leading-tight tracking-tight text-text-primary truncate">{{ title }}</h1>
    <div v-if="$slots.sub"><slot name="sub" /></div>
  </div>
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

const COLLAPSE_AT = 30
const EXPAND_AT = 8

// 本地状态 + 滞回，避免在阈值附近抖动
const folded = ref(false)
watch(() => props.scrollTop, (v) => {
  if (!props.large) return
  if (!folded.value && v > COLLAPSE_AT) folded.value = true
  else if (folded.value && v < EXPAND_AT) folded.value = false
}, { immediate: true })

const collapsed = computed(() => !props.large || folded.value)
const inlineTitleText = computed(() => props.inlineTitle || props.title)
</script>
