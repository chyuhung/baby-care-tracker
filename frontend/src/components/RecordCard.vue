<template>
  <!-- 记录卡片：左 emoji 色块 + 左列（标题 17px + 值区「时长/量等」+ 发热/备注）+ 右列时间（上下垂直居中）。
       类型由色块弱着色 + emoji 区分；点按=编辑，删除走长按 ContextMenu（编辑/删除→确认），无常显按钮、无滑动删除 -->
  <div role="button" tabindex="0" keydown.enter.prevent="$emit('edit')"
    class="bg-surface rounded-2xl p-4 shadow-card flex items-start gap-3 cursor-pointer press-card"
    @touchstart.passive="lp.onTouchStart" @touchmove="lp.onTouchMove" @touchend="lp.onTouchEnd" @touchcancel="lp.onTouchCancel" @click="onCardClick">
    <div class="w-9 h-9 shrink-0 rounded-xl flex items-center justify-center text-lg leading-none" :class="tintClass">{{ emoji }}</div>
    <div class="flex flex-1 min-w-0 items-center justify-between gap-3">
      <div class="min-w-0">
        <span class="block text-[17px] font-semibold text-text-primary truncate">{{ title }}</span>
        <div v-if="valueText" class="mt-1 text-xs text-text-secondary font-num truncate">{{ valueText }}</div>
        <div v-if="fever && record.record_type === 'temperature'" class="mt-1.5 inline-flex items-center gap-1 text-[11px] font-medium text-danger">
          <span class="inline-block h-1.5 w-1.5 rounded-full bg-danger"></span>发热
        </div>
        <div v-if="rd.note" class="text-xs text-text-secondary mt-1.5 truncate">{{ rd.note }}</div>
      </div>
      <div class="shrink-0 text-right text-xs text-text-secondary font-num">{{ timeLabel }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatDurationCompact, formatTimeRangeDay, formatDayTime } from '@/utils'
import { useLongPress } from '@/composables/useLongPress'
import { clockTick } from '@/composables/useClock'

const props = withDefaults(defineProps<{ record: any; showDate?: boolean }>(), { showDate: true })
const emit = defineEmits(['edit', 'context'])

// 长按 → iOS 上下文菜单（编辑 / 删除）
const lp = useLongPress(() => emit('context', props.record))

// 长按后紧随的合成 click 需要被吞掉，否则会误触发「编辑」
function onCardClick() {
  if (lp.consumeClick()) return
  emit('edit')
}

// 五类记录的 data 字段统一为一个别名
const rd = computed(() => props.record.data || {})
const type = computed(() => props.record.record_type)

const feedEmoji: Record<string, string> = { breast: '🤱', bottle: '🍼', formula: '🍼' }
const feedTitle: Record<string, string> = { breast: '母乳亲喂', bottle: '母乳瓶喂', formula: '配方奶' }
const diaperEmoji: Record<string, string> = { pee: '💧', poop: '💩', mixed: '🌪️' }
const diaperTitle: Record<string, string> = { pee: '小便', poop: '大便', mixed: '混合' }
const sideMap: Record<string, string> = { left: '左侧', right: '右侧', both: '双边' }

const emoji = computed(() => {
  switch (type.value) {
    case 'feeding': return feedEmoji[rd.value.type] || '🍼'
    case 'diaper': return diaperEmoji[rd.value.type] || '💧'
    case 'sleep': return '😴'
    case 'temperature': return '🌡️'
    case 'supplement': return '💊'
    default: return '🌳'
  }
})

const title = computed(() => {
  switch (type.value) {
    case 'feeding': return feedTitle[rd.value.type] || rd.value.type
    case 'diaper': return diaperTitle[rd.value.type] || rd.value.type
    case 'sleep': return '睡眠'
    case 'temperature': return '体温'
    case 'supplement': return rd.value.name
    default: return '户外活动'
  }
})

/* 类型弱着色（单色源，收敛到 emoji 色块，替代原左侧色条） */
const tintClass = computed(() => {
  switch (type.value) {
    case 'feeding': return 'bg-primary/10'
    case 'diaper': return 'bg-diaper/10'
    case 'sleep': return 'bg-sleep/10'
    case 'temperature': return 'bg-temperature/10'
    case 'supplement': return 'bg-supplement/10'
    default: return 'bg-outdoor/10'
  }
})

function rangeMinutes(startedAt: string, endedAt?: string | null) {
  if (!endedAt) return null
  return Math.round((new Date(endedAt).getTime() - new Date(startedAt).getTime()) / 60000)
}

const sleepTimeLabel = computed(() => formatTimeRangeDay(rd.value.started_at, rd.value.ended_at, props.showDate))
const outdoorTimeLabel = computed(() => formatTimeRangeDay(rd.value.started_at, rd.value.ended_at, props.showDate))

// 跨度类记录（睡眠/户外）的值区文案：已结束显示实际时长，进行中显示「已持续时长 · 进行中」。
//
// 此前进行中只标状态、不给数字，理由是本组件没有计时器，写死的「已睡 1h30m」会越来越不准。
// 现由 clockTick 提供 10s 实时刷新，两个信息可以同时给：既看得出已经进行多久，
// 也一眼看得出还在持续。取整用 floor，与后端 int(d.Minutes()) 及首页 sleepMinutes 同口径，
// 也避免「59 分 30 秒就显示 1h」的提前量。
function spanValueText() {
  const { started_at: s, ended_at: e } = rd.value
  if (!s) return '进行中'
  if (e) return formatDurationCompact(rangeMinutes(s, e) ?? 0)
  clockTick.value
  const mins = Math.max(0, Math.floor((Date.now() - new Date(s).getTime()) / 60000))
  return `${formatDurationCompact(mins)} · 进行中`
}

const timeAgo = computed(() => formatDayTime(props.record.occurred_at, props.showDate, false))

const timeLabel = computed(() => {
  if (props.record.record_type === 'sleep') return sleepTimeLabel.value
  if (props.record.record_type === 'outdoor') return outdoorTimeLabel.value
  return timeAgo.value
})

/* 值区一行纯文本（时长/量/温度/补剂等），时间不入此列——时间固定显示在卡片右端垂直居中 */
const valueText = computed(() => {
  const parts: string[] = []
  switch (type.value) {
    case 'feeding':
      if (rd.value.amount_ml > 0) parts.push(`${rd.value.amount_ml}ml`)
      if (rd.value.duration_minutes > 0) parts.push(`${rd.value.duration_minutes}分钟`)
      if (rd.value.type === 'breast' && rd.value.side) parts.push(sideMap[rd.value.side] || rd.value.side)
      if (rd.value.brand) parts.push(rd.value.brand)
      break
    case 'sleep':
      parts.push(spanValueText())
      break
    case 'temperature':
      if (rd.value.temperature) parts.push(`${rd.value.temperature}°C`)
      if (rd.value.location) parts.push(rd.value.location)
      break
    case 'supplement':
      if (rd.value.dosage_value > 0) parts.push(`${rd.value.dosage_value}${rd.value.dosage_unit || ''}`)
      break
    case 'outdoor':
      parts.push(spanValueText())
      break
  }
  return parts.join(' · ')
})

const fever = computed(() => type.value === 'temperature' && rd.value.temperature >= 37.5)
</script>