<template>
  <div v-if="record.record_type === 'feeding'" role="button" tabindex="0" @keydown.enter.prevent="$emit('edit')" class="bg-white rounded-2xl p-4 shadow-card flex items-start gap-3 cursor-pointer btn-press" @click="$emit('edit')">
    <div class="w-1.5 h-12 rounded-full bg-primary flex-shrink-0"></div>
    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between gap-2">
        <span class="text-sm font-semibold text-text-primary">{{ feedingTypeLabel }}</span>
        <span class="text-xs text-text-secondary font-num">{{ timeAgo }}</span>
      </div>
      <div class="text-xs text-text-secondary mt-1 flex flex-wrap gap-2">
        <span v-if="rd.type !== 'breast' && rd.amount_ml > 0" class="bg-primary/10 text-primary-deep px-2 py-0.5 rounded-full font-num">{{ rd.amount_ml }}ml</span>
        <span v-if="rd.type === 'breast' && rd.duration_minutes > 0" class="bg-primary/10 text-primary-deep px-2 py-0.5 rounded-full">{{ rd.duration_minutes }}分钟</span>
        <span v-if="rd.type === 'breast' && rd.side" class="bg-primary/10 text-primary-deep px-2 py-0.5 rounded-full">{{ sideLabel }}</span>
        <span v-if="rd.brand" class="bg-muted text-text-secondary px-2 py-0.5 rounded-full">{{ rd.brand }}</span>
      </div>
      <div v-if="rd.note" class="text-xs text-text-secondary mt-1.5 truncate">{{ rd.note }}</div>
    </div>
    <button aria-label="删除此记录" @click.stop="$emit('delete')" class="p-2 text-text-secondary/50 hover:text-danger/70 btn-press min-w-[44px] min-h-[44px] flex items-center justify-center">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
    </button>
  </div>

  <div v-else-if="record.record_type === 'diaper'" role="button" tabindex="0" @keydown.enter.prevent="$emit('edit')" class="bg-white rounded-2xl p-4 shadow-card flex items-start gap-3 cursor-pointer btn-press" @click="$emit('edit')">
    <div class="w-1.5 h-12 rounded-full bg-diaper flex-shrink-0"></div>
    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between gap-2">
        <span class="text-sm font-semibold text-text-primary">{{ diaperTypeLabel }}</span>
        <span class="text-xs text-text-secondary font-num">{{ timeAgo }}</span>
      </div>
      <div v-if="rd.note" class="text-xs text-text-secondary mt-1 truncate">{{ rd.note }}</div>
    </div>
    <button aria-label="删除此记录" @click.stop="$emit('delete')" class="p-2 text-text-secondary/50 hover:text-danger/70 btn-press min-w-[44px] min-h-[44px] flex items-center justify-center">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
    </button>
  </div>

  <div v-else-if="record.record_type === 'sleep'" role="button" tabindex="0" @keydown.enter.prevent="$emit('edit')" class="bg-white rounded-2xl p-4 shadow-card flex items-start gap-3 cursor-pointer btn-press" @click="$emit('edit')">
    <div class="w-1.5 h-12 rounded-full bg-sleep flex-shrink-0"></div>
    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between gap-2">
        <span class="text-sm font-semibold text-text-primary">😴 睡眠</span>
        <span class="text-xs text-text-secondary font-num">{{ sleepTimeLabel }}</span>
      </div>
      <div class="text-xs text-text-secondary mt-1 flex flex-wrap gap-2">
        <span class="bg-sleep/10 text-sleep-deep px-2 py-0.5 rounded-full font-num">{{ sleepDurationLabel }}</span>
      </div>
      <div v-if="rd.note" class="text-xs text-text-secondary mt-1.5 truncate">{{ rd.note }}</div>
    </div>
    <button aria-label="删除此记录" @click.stop="$emit('delete')" class="p-2 text-text-secondary/50 hover:text-danger/70 btn-press min-w-[44px] min-h-[44px] flex items-center justify-center">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
    </button>
  </div>

  <div v-else-if="record.record_type === 'temperature'" role="button" tabindex="0" @keydown.enter.prevent="$emit('edit')" class="bg-white rounded-2xl p-4 shadow-card flex items-start gap-3 cursor-pointer btn-press" @click="$emit('edit')">
    <div class="w-1.5 h-12 rounded-full bg-temperature flex-shrink-0"></div>
    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between gap-2">
        <span class="text-sm font-semibold text-text-primary">🌡️ 体温</span>
        <span class="text-xs text-text-secondary font-num">{{ timeAgo }}</span>
      </div>
      <div class="text-xs text-text-secondary mt-1 flex flex-wrap gap-2">
        <span v-if="rd.temperature" class="bg-temperature/10 text-temperature-deep px-2 py-0.5 rounded-full font-num">{{ rd.temperature }}°C</span>
        <span v-if="rd.location" class="bg-muted text-text-secondary px-2 py-0.5 rounded-full">{{ rd.location }}</span>
        <span v-if="rd.temperature >= 37.5" class="text-danger px-1">🔥</span>
      </div>
      <div v-if="rd.note" class="text-xs text-text-secondary mt-1.5 truncate">{{ rd.note }}</div>
    </div>
    <button aria-label="删除此记录" @click.stop="$emit('delete')" class="p-2 text-text-secondary/50 hover:text-danger/70 btn-press min-w-[44px] min-h-[44px] flex items-center justify-center">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
    </button>
  </div>

  <div v-else-if="record.record_type === 'supplement'" role="button" tabindex="0" @keydown.enter.prevent="$emit('edit')" class="bg-white rounded-2xl p-4 shadow-card flex items-start gap-3 cursor-pointer btn-press" @click="$emit('edit')">
    <div class="w-1.5 h-12 rounded-full bg-supplement flex-shrink-0"></div>
    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between gap-2">
        <span class="text-sm font-semibold text-text-primary">💊 {{ rd.name }}</span>
        <span class="text-xs text-text-secondary font-num">{{ timeAgo }}</span>
      </div>
      <div class="text-xs text-text-secondary mt-1 flex flex-wrap gap-2">
        <span v-if="rd.dosage_value > 0" class="bg-supplement/10 text-supplement-deep px-2 py-0.5 rounded-full font-num">{{ rd.dosage_value }}{{ rd.dosage_unit }}</span>
      </div>
      <div v-if="rd.note" class="text-xs text-text-secondary mt-1.5 truncate">{{ rd.note }}</div>
    </div>
    <button aria-label="删除此记录" @click.stop="$emit('delete')" class="p-2 text-text-secondary/50 hover:text-danger/70 btn-press min-w-[44px] min-h-[44px] flex items-center justify-center">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
    </button>
  </div>

  <div v-else role="button" tabindex="0" @keydown.enter.prevent="$emit('edit')" class="bg-white rounded-2xl p-4 shadow-card flex items-start gap-3 cursor-pointer btn-press" @click="$emit('edit')">
    <div class="w-1.5 h-12 rounded-full bg-outdoor flex-shrink-0"></div>
    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between gap-2">
        <span class="text-sm font-semibold text-text-primary">🌳 户外活动</span>
        <span class="text-xs text-text-secondary font-num">{{ outdoorTimeLabel }}</span>
      </div>
      <div class="text-xs text-text-secondary mt-1 flex flex-wrap gap-2">
        <span class="bg-outdoor/10 text-outdoor-deep px-2 py-0.5 rounded-full font-num">{{ outdoorDurationLabel }}</span>
      </div>
      <div v-if="rd.note" class="text-xs text-text-secondary mt-1.5 truncate">{{ rd.note }}</div>
    </div>
    <button aria-label="删除此记录" @click.stop="$emit('delete')" class="p-2 text-text-secondary/50 hover:text-danger/70 btn-press min-w-[44px] min-h-[44px] flex items-center justify-center">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatDurationCompact, formatTimeRange, formatDayTime } from '@/utils'

const props = withDefaults(defineProps<{ record: any; showDate?: boolean }>(), { showDate: true })
defineEmits(['edit', 'delete'])

// 五类记录的 data 字段统一为一个别名
const rd = computed(() => props.record.data || {})

const feedingTypeMap: Record<string, string> = { breast: '🤱 母乳亲喂', bottle: '🍼 母乳瓶喂', formula: '🍼 配方奶' }
const diaperTypeMap: Record<string, string> = { pee: '💧 小便', poop: '💩 大便', mixed: '🌪️ 混合' }
const sideMap: Record<string, string> = { left: '左侧', right: '右侧', both: '双边' }

const feedingTypeLabel = computed(() => feedingTypeMap[rd.value.type] || rd.value.type)
const diaperTypeLabel = computed(() => diaperTypeMap[rd.value.type] || rd.value.type)
const sideLabel = computed(() => sideMap[rd.value.side] || rd.value.side)

function rangeMinutes(startedAt: string, endedAt?: string | null) {
  if (!endedAt) return null
  return Math.round((new Date(endedAt).getTime() - new Date(startedAt).getTime()) / 60000)
}

const sleepTimeLabel = computed(() => formatTimeRange(rd.value.started_at, rd.value.ended_at))
const outdoorTimeLabel = computed(() => formatTimeRange(rd.value.started_at, rd.value.ended_at))

const sleepDurationLabel = computed(() => {
  const mins = rangeMinutes(rd.value.started_at, rd.value.ended_at)
  return mins === null ? '进行中' : formatDurationCompact(mins)
})
const outdoorDurationLabel = computed(() => {
  const mins = rangeMinutes(rd.value.started_at, rd.value.ended_at)
  return mins === null ? '进行中' : formatDurationCompact(mins)
})

const timeAgo = computed(() => formatDayTime(props.record.occurred_at, props.showDate))
</script>
