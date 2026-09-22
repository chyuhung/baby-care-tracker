<template>
  <!-- iOS 风日期时间行：点击弹出滚轮选择器（替代原生 datetime-local） -->
  <button type="button" @click="open = true"
    class="w-full flex items-center justify-between px-4 py-3 bg-surface border border-border-color rounded-xl text-left btn-press focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/40"
    :aria-label="ariaLabel" aria-haspopup="dialog">
    <span class="text-text-primary tabular-nums">{{ displayText }}</span>
    <svg class="w-4 h-4 text-text-secondary shrink-0" viewBox="0 0 24 24" fill="none">
      <path d="M6 9l6 6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
    </svg>
  </button>

  <DateTimeWheel :open="open" :model-value="modelValue" :title="title"
    @update:open="(v: boolean) => open = v" @confirm="onConfirm" />
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import DateTimeWheel from './DateTimeWheel.vue'

const props = withDefaults(defineProps<{
  modelValue: string
  title?: string
  ariaLabel?: string
}>(), {
  title: '选择时间',
  ariaLabel: '选择时间',
})

const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const open = ref(false)

const WEEK = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']

const displayText = computed(() => {
  const m = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})/.exec(props.modelValue || '')
  if (!m) return '选择时间'
  const dt = new Date(+m[1], +m[2] - 1, +m[3])
  const today = new Date(); today.setHours(0, 0, 0, 0)
  const day = new Date(dt); day.setHours(0, 0, 0, 0)
  const diff = Math.round((day.getTime() - today.getTime()) / 86400000)
  const dayTxt = diff === 0 ? '今天' : diff === -1 ? '昨天' : `${+m[2]}月${+m[3]}日 ${WEEK[dt.getDay()]}`
  return `${dayTxt} ${m[4]}:${m[5]}`
})

function onConfirm(v: string) {
  emit('update:modelValue', v)
}
</script>
