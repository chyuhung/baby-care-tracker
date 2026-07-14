<template>
  <div v-if="record.record_type === 'feeding'" class="bg-white rounded-2xl p-4 shadow-card flex items-start gap-3 cursor-pointer active:scale-[0.99] transition-transform" @click="$emit('edit')">
    <div class="w-1.5 h-12 rounded-full bg-primary flex-shrink-0"></div>
    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between gap-2">
        <span class="text-sm font-semibold text-text-primary">
          {{ feedingTypeLabel }}
        </span>
        <span class="text-xs text-text-secondary font-num">{{ timeAgo }}</span>
      </div>
      <div class="text-xs text-text-secondary mt-1 flex flex-wrap gap-2">
        <span v-if="f.type !== 'breast' && f.amount_ml > 0" class="bg-primary/10 text-primary px-2 py-0.5 rounded-full font-num">{{ f.amount_ml }}ml</span>
        <span v-if="f.type === 'breast' && f.duration_minutes > 0" class="bg-primary/10 text-primary px-2 py-0.5 rounded-full">{{ f.duration_minutes }}分钟</span>
        <span v-if="f.type === 'breast' && f.side" class="bg-primary/10 text-primary px-2 py-0.5 rounded-full">{{ sideLabel }}</span>
        <span v-if="f.brand" class="bg-gray-100 text-text-secondary px-2 py-0.5 rounded-full">{{ f.brand }}</span>
      </div>
      <div v-if="f.note" class="text-xs text-text-secondary mt-1.5 truncate">{{ f.note }}</div>
    </div>
    <button @click.stop="$emit('delete')" class="p-1 text-text-secondary/50 hover:text-red-400 btn-press">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
    </button>
  </div>

  <div v-else class="bg-white rounded-2xl p-4 shadow-card flex items-start gap-3 cursor-pointer active:scale-[0.99] transition-transform" @click="$emit('edit')">
    <div class="w-1.5 h-12 rounded-full bg-diaper flex-shrink-0"></div>
    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between gap-2">
        <span class="text-sm font-semibold text-text-primary">{{ diaperTypeLabel }}</span>
        <span class="text-xs text-text-secondary font-num">{{ timeAgo }}</span>
      </div>
      <div v-if="d.note" class="text-xs text-text-secondary mt-1 truncate">{{ d.note }}</div>
    </div>
    <button @click.stop="$emit('delete')" class="p-1 text-text-secondary/50 hover:text-red-400 btn-press">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{ record: any; showDate?: boolean }>(), { showDate: true })
defineEmits(['edit', 'delete'])

const f = computed(() => props.record.data || {})
const d = computed(() => props.record.data || {})

const feedingTypeMap: Record<string, string> = { breast: '🤱 母乳亲喂', bottle: '🍼 母乳瓶喂', formula: '🍼 配方奶' }
const diaperTypeMap: Record<string, string> = { pee: '💧 小便', poop: '💩 大便', mixed: '💥 混合' }
const sideMap: Record<string, string> = { left: '左侧', right: '右侧', both: '双边' }

const feedingTypeLabel = computed(() => feedingTypeMap[f.value.type] || f.value.type)
const diaperTypeLabel = computed(() => diaperTypeMap[d.value.type] || d.value.type)
const sideLabel = computed(() => sideMap[f.value.side] || f.value.side)

const timeAgo = computed(() => {
  const d = new Date(props.record.occurred_at)
  const pad = (n: number) => String(n).padStart(2, '0')
  const hhmm = `${pad(d.getHours())}:${pad(d.getMinutes())}`
  if (!props.showDate) return hhmm
  const now = new Date()
  const isToday = d.toDateString() === now.toDateString()
  const isYesterday = d.toDateString() === new Date(now.getTime() - 86400000).toDateString()
  if (isToday) return `今天 ${hhmm}`
  if (isYesterday) return `昨天 ${hhmm}`
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${hhmm}`
})
</script>
