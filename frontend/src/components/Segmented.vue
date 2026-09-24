<template>
  <!-- iOS UISegmentedControl：灰底容器内等分段，选中段白底浮起 -->
  <div class="flex bg-muted rounded-xl p-1 gap-1" role="tablist">
    <button v-for="opt in options" :key="String(opt.value)" type="button" role="tab"
      :aria-selected="modelValue === opt.value"
      @click="onPick(opt.value)"
      :class="['flex-1 min-h-[44px] min-w-0 px-1.5 rounded-[10px] flex items-center justify-center gap-1.5 font-medium transition-all btn-press',
        compact ? 'text-xs' : 'text-sm',
        modelValue === opt.value ? 'bg-segment text-text-primary shadow-sm font-semibold' : 'text-text-secondary']">
      <span v-if="opt.emoji" class="text-base leading-none">{{ opt.emoji }}</span>
      <span class="truncate">{{ opt.label }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
export interface SegOption {
  value: string
  label: string
  emoji?: string
}

const props = withDefaults(defineProps<{
  modelValue: string
  options: SegOption[]
  compact?: boolean
}>(), { compact: false })

const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

function onPick(v: string) {
  emit('update:modelValue', v)
}
</script>
