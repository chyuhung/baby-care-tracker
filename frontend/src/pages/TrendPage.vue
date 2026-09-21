<template>
  <div class="flex flex-col h-dvh">
    <PullRefresh class="flex-1 min-h-0" content-class="px-4 py-4 pb-[calc(6.5rem+env(safe-area-inset-bottom))] space-y-6"
      :refresh="() => loadTrend(true)">
    <template #header>
    <header class="sticky top-0 z-30 glass-surface hairline-bottom pt-safe px-4 py-3">
      <h1 class="text-lg font-bold text-text-primary">趋势</h1>
      <!-- 类别（iOS 下拉菜单）+ 时间范围（分段控件）：单行排列，零横向滚动 -->
      <div class="flex items-center gap-2 mt-2">
        <MenuSelect v-model="category" :options="categoryOptions" title="选择类别" aria-label="选择类别" />
        <div class="flex-1 min-w-0">
          <Segmented :model-value="String(days)" :options="dayOptions" compact
            @update:model-value="(v: string) => { days = Number(v); loadTrend() }" />
        </div>
      </div>
    </header>
    </template>

      <div v-if="loading" class="flex justify-center py-20">
        <ActivityIndicator :size="28" class="text-text-secondary" />
      </div>
      <EmptyState v-else-if="trendData.length === 0" title="暂无趋势数据"
        subtitle="记录几天数据后，这里会生成图表趋势" />
      <template v-else>
        <div v-if="category === 'feeding'" class="bg-surface rounded-2xl shadow-card px-3 pt-2.5 pb-2">
          <div class="flex items-center justify-center gap-4 mb-1 text-[10px] text-text-secondary font-normal">
            <span class="flex items-center gap-1"><span class="inline-block w-4" style="border-top: 2px solid var(--chart-primary)"></span>奶量 ml</span>
            <span class="flex items-center gap-1"><span class="inline-block w-2.5 h-2.5 rounded-sm" style="background: var(--chart-primary-count)"></span>次数</span>
          </div>
          <svg viewBox="0 0 340 228" class="w-full block">
            <template v-if="days === 30">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
              <g v-for="(t, ti) in feedingMlScatter.ticks" :key="'fl'+ti">
                <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
                <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" class="chart-value-label">{{ t.label }}</text>
              </g>
              <g v-for="(t, ti) in feedingCountScatter.ticks" :key="'fr'+ti">
                <text :x="axis.rightX + 5" :y="t.y + 3" text-anchor="start" font-size="9" class="chart-axis-label">{{ t.label }}</text>
              </g>
              <g v-for="(pt, i) in feedingMlScatter.points" :key="'dv'+i">
                <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
              </g>
              <line v-if="feedingMlScatter.trend" :x1="feedingMlScatter.trend.x1" :y1="feedingMlScatter.trend.y1" :x2="feedingMlScatter.trend.x2" :y2="feedingMlScatter.trend.y2" stroke="var(--chart-primary)" stroke-width="1.5" stroke-dasharray="6,3" opacity="0.85"/>
              <line v-if="feedingCountScatter.trend" :x1="feedingCountScatter.trend.x1" :y1="feedingCountScatter.trend.y1" :x2="feedingCountScatter.trend.x2" :y2="feedingCountScatter.trend.y2" stroke="var(--chart-primary-count)" stroke-width="1.5" stroke-dasharray="6,3" opacity="0.85"/>
              <g v-for="(pt, i) in feedingMlScatter.points" :key="'dp'+i">
                <circle :cx="pt.x" :cy="pt.y" r="2.5" fill="var(--chart-primary)" :opacity="nodeOpacity(i)"/>
              </g>
              <g v-for="(pt, i) in feedingCountScatter.points" :key="'cp'+i">
                <circle :cx="pt.x" :cy="pt.y" r="2.5" fill="var(--chart-primary-count)" :opacity="nodeOpacity(i)"/>
              </g>
            </template>
            <template v-else>
              <g v-for="(b, i) in feedingMl.items" :key="'bm'+i">
                <rect :x="feedingRects(i).mlX" :y="b.y" :width="w2" :height="b.h" rx="2" fill="var(--chart-primary)" :opacity="barOpacity(i)"/>
                <text v-if="b.h > 0" :x="feedingRects(i).mlX + w2 / 2" :y="b.y - 3" text-anchor="middle" font-size="8" :font-weight="labelWeight(i)" class="chart-value-label">{{ b.label }}</text>
              </g>
              <g v-for="(b, i) in feedingCount.items" :key="'bc'+i">
                <rect :x="feedingRects(i).countX" :y="b.y" :width="w2" :height="b.h" rx="2" fill="var(--chart-primary-count)" :opacity="barOpacity(i)"/>
                <text v-if="b.h > 0" :x="feedingRects(i).countX + w2 / 2" :y="b.y - 3" text-anchor="middle" font-size="8" :font-weight="labelWeight(i)" class="chart-value-label">{{ b.label }}</text>
              </g>
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            </template>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            <line :x1="axis.rightX" :x2="axis.rightX" :y1="axis.topY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" class="chart-axis-label">ml</text>
            <text :x="axis.rightX" :y="axis.topY - 5" text-anchor="middle" font-size="9" class="chart-axis-label">次</text>
            <template v-for="(d, i) in trendData" :key="'fx'+i">
              <text v-if="dateLabels[i]?.show" :x="dateX(i)" :y="DATE_LABEL_Y" text-anchor="middle" font-size="9" class="chart-axis-label" :font-weight="dateWeight(i)" :fill="dateFill(i)">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div v-if="category === 'diaper'" class="bg-surface rounded-2xl shadow-card px-3 pt-2.5 pb-2">
          <svg viewBox="0 0 340 228" class="w-full block">
            <template v-if="days === 30">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
              <g v-for="(t, ti) in diaperScatter.ticks" :key="'dl'+ti">
                <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
                <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" class="chart-value-label">{{ t.label }}</text>
              </g>
              <g v-for="(pt, i) in diaperScatter.points" :key="'dv'+i">
                <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
              </g>
              <line v-if="diaperScatter.trend" :x1="diaperScatter.trend.x1" :y1="diaperScatter.trend.y1" :x2="diaperScatter.trend.x2" :y2="diaperScatter.trend.y2" stroke="var(--chart-diaper)" stroke-width="1.5" stroke-dasharray="6,3" opacity="0.85"/>
              <g v-for="(pt, i) in diaperScatter.points" :key="'dp'+i">
                <circle :cx="pt.x" :cy="pt.y" r="2.5" fill="var(--chart-diaper)" :opacity="nodeOpacity(i)"/>
              </g>
            </template>
            <template v-else>
              <g v-for="(b, i) in diaper.items" :key="'db'+i">
                <rect :x="singleRects(i).gl" :y="b.y" :width="barW" :height="b.h" rx="2" fill="var(--chart-diaper)" :opacity="barOpacity(i)"/>
                <text v-if="b.h > 0" :x="singleRects(i).gl + barW / 2" :y="b.y - 3" text-anchor="middle" font-size="8" :font-weight="labelWeight(i)" class="chart-value-label">{{ b.label }}</text>
              </g>
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            </template>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" class="chart-axis-label">次</text>
            <template v-for="(d, i) in trendData" :key="'dx'+i">
              <text v-if="dateLabels[i]?.show" :x="dateX(i)" :y="DATE_LABEL_Y" text-anchor="middle" font-size="9" class="chart-axis-label" :font-weight="dateWeight(i)" :fill="dateFill(i)">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div v-if="category === 'sleep'" class="bg-surface rounded-2xl shadow-card px-3 pt-2.5 pb-2">
          <svg viewBox="0 0 340 228" class="w-full block">
            <template v-if="days === 30">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
              <g v-for="(t, ti) in sleepScatter.ticks" :key="'sl'+ti">
                <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
                <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" class="chart-value-label">{{ t.label }}</text>
              </g>
              <g v-for="(pt, i) in sleepScatter.points" :key="'sv'+i">
                <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
              </g>
              <line v-if="sleepScatter.trend" :x1="sleepScatter.trend.x1" :y1="sleepScatter.trend.y1" :x2="sleepScatter.trend.x2" :y2="sleepScatter.trend.y2" stroke="var(--chart-sleep)" stroke-width="1.5" stroke-dasharray="6,3" opacity="0.85"/>
              <g v-for="(pt, i) in sleepScatter.points" :key="'sp'+i">
                <circle :cx="pt.x" :cy="pt.y" r="2.5" fill="var(--chart-sleep)" :opacity="nodeOpacity(i)"/>
              </g>
            </template>
            <template v-else>
              <g v-for="(b, i) in sleep.items" :key="'sb'+i">
                <rect :x="singleRects(i).gl" :y="b.y" :width="barW" :height="b.h" rx="2" fill="var(--chart-sleep)" :opacity="barOpacity(i)"/>
                <text v-if="b.h > 0" :x="singleRects(i).gl + barW / 2" :y="b.y - 3" text-anchor="middle" font-size="8" :font-weight="labelWeight(i)" class="chart-value-label">{{ b.label }}</text>
              </g>
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            </template>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" class="chart-axis-label">小时</text>
            <template v-for="(d, i) in trendData" :key="'sx'+i">
              <text v-if="dateLabels[i]?.show" :x="dateX(i)" :y="DATE_LABEL_Y" text-anchor="middle" font-size="9" class="chart-axis-label" :font-weight="dateWeight(i)" :fill="dateFill(i)">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div v-if="category === 'outdoor'" class="bg-surface rounded-2xl shadow-card px-3 pt-2.5 pb-2">
          <svg viewBox="0 0 340 228" class="w-full block">
            <template v-if="days === 30">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
              <g v-for="(t, ti) in outdoorScatter.ticks" :key="'ol'+ti">
                <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
                <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" class="chart-value-label">{{ t.label }}</text>
              </g>
              <g v-for="(pt, i) in outdoorScatter.points" :key="'ov'+i">
                <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
              </g>
              <line v-if="outdoorScatter.trend" :x1="outdoorScatter.trend.x1" :y1="outdoorScatter.trend.y1" :x2="outdoorScatter.trend.x2" :y2="outdoorScatter.trend.y2" stroke="var(--chart-outdoor)" stroke-width="1.5" stroke-dasharray="6,3" opacity="0.85"/>
              <g v-for="(pt, i) in outdoorScatter.points" :key="'op'+i">
                <circle :cx="pt.x" :cy="pt.y" r="2.5" fill="var(--chart-outdoor)" :opacity="nodeOpacity(i)"/>
              </g>
            </template>
            <template v-else>
              <g v-for="(b, i) in outdoor.items" :key="'ob'+i">
                <rect :x="singleRects(i).gl" :y="b.y" :width="barW" :height="b.h" rx="2" fill="var(--chart-outdoor)" :opacity="barOpacity(i)"/>
                <text v-if="b.h > 0" :x="singleRects(i).gl + barW / 2" :y="b.y - 3" text-anchor="middle" font-size="8" :font-weight="labelWeight(i)" class="chart-value-label">{{ b.label }}</text>
              </g>
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            </template>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" class="chart-axis-label">小时</text>
            <template v-for="(d, i) in trendData" :key="'ox'+i">
              <text v-if="dateLabels[i]?.show" :x="dateX(i)" :y="DATE_LABEL_Y" text-anchor="middle" font-size="9" class="chart-axis-label" :font-weight="dateWeight(i)" :fill="dateFill(i)">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div v-if="category === 'supplement'" class="bg-surface rounded-2xl shadow-card px-3 pt-2.5 pb-2">
          <svg viewBox="0 0 340 228" class="w-full block">
            <template v-if="days === 30">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
              <g v-for="(t, ti) in supplementScatter.ticks" :key="'sul'+ti">
                <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
                <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" class="chart-value-label">{{ t.label }}</text>
              </g>
              <g v-for="(pt, i) in supplementScatter.points" :key="'suv'+i">
                <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
              </g>
              <line v-if="supplementScatter.trend" :x1="supplementScatter.trend.x1" :y1="supplementScatter.trend.y1" :x2="supplementScatter.trend.x2" :y2="supplementScatter.trend.y2" stroke="var(--chart-supplement)" stroke-width="1.5" stroke-dasharray="6,3" opacity="0.85"/>
              <g v-for="(pt, i) in supplementScatter.points" :key="'sup'+i">
                <circle :cx="pt.x" :cy="pt.y" r="2.5" fill="var(--chart-supplement)" :opacity="nodeOpacity(i)"/>
              </g>
            </template>
            <template v-else>
              <g v-for="(b, i) in supplement.items" :key="'sub'+i">
                <rect :x="singleRects(i).gl" :y="b.y" :width="barW" :height="b.h" rx="2" fill="var(--chart-supplement)" :opacity="barOpacity(i)"/>
                <text v-if="b.h > 0" :x="singleRects(i).gl + barW / 2" :y="b.y - 3" text-anchor="middle" font-size="8" :font-weight="labelWeight(i)" class="chart-value-label">{{ b.label }}</text>
              </g>
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            </template>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" class="chart-axis-label">次</text>
            <template v-for="(d, i) in trendData" :key="'sux'+i">
              <text v-if="dateLabels[i]?.show" :x="dateX(i)" :y="DATE_LABEL_Y" text-anchor="middle" font-size="9" class="chart-axis-label" :font-weight="dateWeight(i)" :fill="dateFill(i)">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div v-if="category === 'temperature'" class="bg-surface rounded-2xl shadow-card px-3 pt-2.5 pb-2">
          <svg viewBox="0 0 340 228" class="w-full block">
            <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            <g v-for="(t, ti) in tempTicks" :key="'tl'+ti">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
              <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" class="chart-value-label">{{ t.label }}</text>
            </g>
            <g v-for="(pt, i) in tempPoints" :key="'tv'+i">
              <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
            </g>
            <line :x1="axis.leftX" :x2="axis.rightX" :y1="feverLineY" :y2="feverLineY" stroke="rgb(var(--danger-deep))" stroke-width="1" stroke-dasharray="4,3" opacity="0.5"/>
            <path :d="tempPastPath" fill="none" stroke="var(--chart-temperature)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" opacity="0.3"/>
            <path v-if="tempLastSeg" :d="tempLastSeg" fill="none" stroke="var(--chart-temperature)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            <circle v-if="tempLastPoint" :cx="tempLastPoint.x" :cy="tempLastPoint.y" r="3.5" fill="var(--chart-temperature)"/>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="var(--chart-line)" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" class="chart-axis-label">°C</text>
            <template v-for="(d, i) in trendData" :key="'tx'+i">
              <text v-if="dateLabels[i]?.show" :x="dateX(i)" :y="DATE_LABEL_Y" text-anchor="middle" font-size="9" class="chart-axis-label" :font-weight="dateWeight(i)" :fill="dateFill(i)">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div v-if="trendData.length" class="bg-surface rounded-2xl shadow-card p-4 space-y-3">
          <div class="flex items-center justify-between gap-2">
            <h4 class="text-sm font-semibold text-text-secondary shrink-0">📊 期间对比</h4>
            <span class="text-[11px] text-text-secondary truncate">不含今日 · 当前 vs 上一周期</span>
          </div>
          <div v-if="periodLabel && !summary.empty" class="text-[11px] text-text-secondary">{{ periodLabel }}</div>
          <div v-if="summary.empty" class="bg-bg-main rounded-xl">
            <EmptyState size="sm" :title="`近 ${days} 天暂无记录`" subtitle="记录几天后，这里会生成周期对比">
              <router-link to="/"
                class="inline-flex items-center gap-1.5 px-5 py-2.5 bg-primary-fill text-white rounded-xl font-medium text-sm btn-press shadow-card">
                去记录
              </router-link>
            </EmptyState>
          </div>
          <div v-else class="grid grid-cols-2 gap-2.5">
            <div v-for="c in summary.cards" :key="c.label" class="bg-bg-main rounded-xl p-3">
              <div class="text-xs text-text-secondary">{{ c.label }}</div>
              <div class="flex items-baseline gap-1 mt-1">
                <span class="text-xl font-bold font-num text-text-primary">{{ c.value }}</span>
                <span v-if="c.unit" class="text-xs text-text-secondary">{{ c.unit }}</span>
              </div>
              <div class="flex items-center gap-1 mt-1">
                <span v-if="c.delta !== null" class="text-[11px] font-num font-medium px-1.5 py-0.5 rounded-md" :class="deltaClass(c.delta)">{{ deltaArrow(c.delta) }}{{ Math.abs(c.delta) }}%</span>
                <span class="text-[11px] text-text-secondary truncate">{{ c.prevText }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </PullRefresh>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { babyAPI } from '@/api'
import PullRefresh from '@/components/PullRefresh.vue'
import ActivityIndicator from '@/components/ActivityIndicator.vue'
import EmptyState from '@/components/EmptyState.vue'
import Segmented from '@/components/Segmented.vue'
import MenuSelect from '@/components/MenuSelect.vue'

const app = useAppStore()
const trendData = ref<any[]>([])
const trendCur = ref<any[]>([])
const trendPrev = ref<any[]>([])
const loading = ref(false)
const days = ref(7)
const category = ref('feeding')

const categoryOptions = [
  { label: '喂奶', emoji: '🍼', value: 'feeding' },
  { label: '尿布', emoji: '🩲', value: 'diaper' },
  { label: '睡眠', emoji: '😴', value: 'sleep' },
  { label: '体温', emoji: '🌡️', value: 'temperature' },
  { label: '户外', emoji: '🌳', value: 'outdoor' },
  { label: '补剂', emoji: '💊', value: 'supplement' },
]

const dateLabels = computed(() => {
  return trendData.value.map((d, i) => {
    const parts = d.date.split('-')
    const label = `${parseInt(parts[1])}/${parseInt(parts[2])}`
    const show = days.value < 30 || i % 5 === 0
    return { label, show }
  })
})

const dayOptions = [
  { label: '7天', value: '7' },
  { label: '30天', value: '30' },
]

const summary = computed(() => {
  // 当日数据尚未完善，不参与对比：周期自前一天起算，cur/prev 由 loadTrend 切好
  const data = trendCur.value
  const prevRows = trendPrev.value
  const cards: { label: string, unit: string, value: string, prevText: string, delta: number | null }[] = []
  let noData = false
  if (!data.length) return { cards, empty: true }

  const P = data.length
  const sumOf = (rows: any[], f: (d: any) => number) => rows.reduce((a, d) => a + (f(d) || 0), 0)
  const maxOf = (rows: any[], f: (d: any) => number) => rows.length ? Math.max(...rows.map(d => f(d) || 0)) : 0
  const h1 = (v: number) => (v > 0 && v < 0.1) ? v.toFixed(2) : v.toFixed(1)
  const i0 = (v: number) => String(Math.round(v))
  const pv = prevRows.length ? prevRows : null

  const push = (label: string, unit: string, curVal: number | null, prevVal: number | null, fmt: (v: number) => string) => {
    let delta: number | null = null
    if (curVal !== null && prevVal !== null) {
      if (prevVal > 0) delta = Math.round((curVal - prevVal) / prevVal * 100)
      else if (curVal === 0) delta = 0 // 两期均为 0：无变化，同样给出「—0%」徽章，避免徽章位空缺
    }
    cards.push({
      label, unit,
      value: curVal === null ? '--' : fmt(curVal),
      prevText: prevVal === null ? '—' : `上期 ${fmt(prevVal)}${unit}`,
      delta,
    })
  }

  // 上周期无可比数据时返回 null（全零不算数据）
  const prevAgg = (f: (d: any) => number, agg: (rows: any[]) => number): number | null => {
    if (!pv) return null
    if (!pv.some(d => (f(d) || 0) > 0)) return null
    return agg(pv)
  }
  const sumF = (f: (d: any) => number) => (rows: any[]) => sumOf(rows, f)
  const maxF = (f: (d: any) => number) => (rows: any[]) => maxOf(rows, f)
  const minPosF = (f: (d: any) => number) => (rows: any[]) => { const v = minPositiveDays(rows, f); return v.length ? Math.min(...v) : 0 }
  const avgPosF = (f: (d: any) => number) => (rows: any[]) => avgOver(minPositiveDays(rows, f))
  const avgF = (f: (d: any) => number) => (rows: any[]) => rows.length ? sumOf(rows, f) / rows.length : 0
  const countPosF = (f: (d: any) => number) => (rows: any[]) => rows.filter(d => (f(d) || 0) > 0).length

  // 本周期同口径（取有记录的天，避免无数据日拉低）；无记录天返回 null → 显示 --
  const curMinPos = (f: (d: any) => number) => sumOf(data, f) > 0 ? minPosF(f)(data) : null
  const curAvgPos = (f: (d: any) => number) => sumOf(data, f) > 0 ? avgPosF(f)(data) : null

  const feeding = () => {
    const has = (rows: any[]) => rows.some(d => (d.total_ml || 0) > 0 || (d.feeding_count || 0) > 0)
    if (!has(data) && !(pv && has(pv))) return emptyCards()
    const curHas = has(data)
    const curMl = sumOf(data, d => d.total_ml || 0)
    const curCnt = sumOf(data, d => d.feeding_count || 0)
    const pMl = prevAgg(d => d.total_ml || 0, sumF(d => d.total_ml || 0))
    const pCnt = prevAgg(d => d.feeding_count || 0, sumF(d => d.feeding_count || 0))
    push('日均奶量', 'ml', curHas ? curMl / P : null, pMl !== null ? pMl / P : null, i0)
    push('日均喂养次数', '次', curHas ? curCnt / P : null, pCnt !== null ? pCnt / P : null, h1)
    push('单次平均奶量', 'ml', curCnt ? curMl / curCnt : null, pCnt ? pMl! / pCnt : null, i0)
    push('单日最高奶量', 'ml', curHas ? maxOf(data, d => d.total_ml || 0) : null, prevAgg(d => d.total_ml || 0, maxF(d => d.total_ml || 0)), i0)
  }

  const diaper = () => {
    const f = (d: any) => d.diaper_count || 0
    const has = (rows: any[]) => rows.some(d => f(d) > 0)
    if (!has(data) && !(pv && has(pv))) return emptyCards()
    const curHas = has(data)
    push('日均尿布次数', '次', curHas ? sumOf(data, f) / P : null, prevAgg(f, avgF(f)), h1)
    push('单日最多', '次', curHas ? maxOf(data, f) : null, prevAgg(f, maxF(f)), i0)
    push('单日最少', '次', curHas ? curMinPos(f) : null, prevAgg(f, minPosF(f)), i0)
    push('期间总次数', '次', curHas ? sumOf(data, f) : null, prevAgg(f, sumF(f)), i0)
  }

  const sleep = () => {
    const f = (d: any) => d.sleep_duration_minutes || 0
    const has = (rows: any[]) => rows.some(d => f(d) > 0)
    if (!has(data) && !(pv && has(pv))) return emptyCards()
    const curHas = has(data)
    push('日均睡眠', '小时', curHas ? sumOf(data, f) / P / 60 : null, prevAgg(f, avgF(f)) !== null ? prevAgg(f, avgF(f))! / 60 : null, h1)
    push('单日最长', '小时', curHas ? maxOf(data, f) / 60 : null, prevAgg(f, maxF(f)) !== null ? prevAgg(f, maxF(f))! / 60 : null, h1)
    push('有记录日均', '小时', curAvgPos(f) !== null ? curAvgPos(f)! / 60 : null, prevAgg(f, avgPosF(f)) !== null ? prevAgg(f, avgPosF(f))! / 60 : null, h1)
    push('期间总时长', '小时', curHas ? sumOf(data, f) / 60 : null, prevAgg(f, sumF(f)) !== null ? prevAgg(f, sumF(f))! / 60 : null, h1)
  }

  const outdoor = () => {
    const f = (d: any) => d.outdoor_duration_minutes || 0
    const has = (rows: any[]) => rows.some(d => f(d) > 0)
    if (!has(data) && !(pv && has(pv))) return emptyCards()
    const curHas = has(data)
    push('日均户外', '小时', curHas ? sumOf(data, f) / P / 60 : null, prevAgg(f, avgF(f)) !== null ? prevAgg(f, avgF(f))! / 60 : null, h1)
    push('单日最长', '小时', curHas ? maxOf(data, f) / 60 : null, prevAgg(f, maxF(f)) !== null ? prevAgg(f, maxF(f))! / 60 : null, h1)
    push('有记录日均', '小时', curAvgPos(f) !== null ? curAvgPos(f)! / 60 : null, prevAgg(f, avgPosF(f)) !== null ? prevAgg(f, avgPosF(f))! / 60 : null, h1)
    push('期间总时长', '小时', curHas ? sumOf(data, f) / 60 : null, prevAgg(f, sumF(f)) !== null ? prevAgg(f, sumF(f))! / 60 : null, h1)
  }

  const temperature = () => {
    const f = (d: any) => d.temperature_high || 0
    const meas = (rows: any[]) => rows.filter(d => f(d) > 0)
    const curM = meas(data), pvM = pv ? meas(pv) : []
    if (!curM.length && !pvM.length) return emptyCards()
    const avgTemp = (rows: any[]) => rows.length ? sumOf(rows, d => d.temperature_avg || 0) / rows.length : 0
    const fever = (rows: any[]) => rows.filter(d => f(d) >= 37.5).length
    const pvMeas = pvM.length ? pvM : null
    push('平均体温', '°C', curM.length ? avgTemp(curM) : null, pvMeas ? avgTemp(pvMeas) : null, v => v.toFixed(1))
    push('期间最高', '°C', curM.length ? maxOf(curM, f) : null, pvMeas ? maxOf(pvMeas, f) : null, v => v.toFixed(1))
    push('发烧天数', '天', curM.length ? fever(curM) : null, pvMeas ? fever(pvMeas) : null, i0)
    push('测温天数', '天', curM.length, pvMeas ? pvMeas.length : null, i0)
  }

  const supplement = () => {
    const f = (d: any) => d.supplement_count || 0
    const has = (rows: any[]) => rows.some(d => f(d) > 0)
    if (!has(data) && !(pv && has(pv))) return emptyCards()
    const curHas = has(data)
    push('日均补剂次数', '次', curHas ? sumOf(data, f) / P : null, prevAgg(f, avgF(f)), h1)
    push('单日最多', '次', curHas ? maxOf(data, f) : null, prevAgg(f, maxF(f)), i0)
    push('补剂天数', '天', curHas ? countPosF(f)(data) : null, prevAgg(f, countPosF(f)), i0)
    push('期间总次数', '次', curHas ? sumOf(data, f) : null, prevAgg(f, sumF(f)), i0)
  }

  function minPositiveDays(rows: any[], f: (d: any) => number) {
    return rows.map(d => f(d) || 0).filter(x => x > 0)
  }
  function avgOver(v: number[]) { return v.length ? v.reduce((a, b) => a + b, 0) / v.length : 0 }

  function emptyCards() {
    noData = true
  }

  const builders: Record<string, () => void> = { feeding, diaper, sleep, outdoor, temperature, supplement }
  builders[category.value]?.()

  // 上周期在「当前类别」下是否有可比数据（用于周期行是否标注「上周期无记录」）
  const prevHas = (() => {
    if (!pv) return false
    const c = category.value
    if (c === 'feeding') return pv.some((d: any) => (d.total_ml || 0) > 0 || (d.feeding_count || 0) > 0)
    if (c === 'diaper') return pv.some((d: any) => (d.diaper_count || 0) > 0)
    if (c === 'sleep') return pv.some((d: any) => (d.sleep_duration_minutes || 0) > 0)
    if (c === 'outdoor') return pv.some((d: any) => (d.outdoor_duration_minutes || 0) > 0)
    if (c === 'temperature') return pv.some((d: any) => (d.temperature_high || 0) > 0)
    if (c === 'supplement') return pv.some((d: any) => (d.supplement_count || 0) > 0)
    return false
  })()
  return { cards, empty: noData, prevHas }
})

// 本周期 / 上周期 日期范围文案（自前一天起算，不含今日）
const periodLabel = computed(() => {
  const f = (d: any) => `${parseInt(d.date.split('-')[1])}/${parseInt(d.date.split('-')[2])}`
  const cur = trendCur.value, pv = trendPrev.value
  if (!cur.length) return ''
  const curTxt = `${f(cur[0])} – ${f(cur[cur.length - 1])}`
  const pvTxt = pv.length ? `${f(pv[0])} – ${f(pv[pv.length - 1])}` : '无上期数据'
  // 上周期在当前类别下无记录时，提示语只在此处打印一次，避免四个 tile 重复同一句
  const tail = pv.length && !summary.value.prevHas ? ' · 上周期无记录' : ''
  return `本周期 ${curTxt} · 上周期 ${pvTxt}${tail}`
})

function deltaArrow(d: number): string { return d > 0 ? '↑' : d < 0 ? '↓' : '—' }
// 中性配色：只表达方向，不对“多/少”做价值判断
function deltaClass(_d: number): string { return 'bg-muted text-text-secondary' }

const CHART = { padL: 26, padR: 24, padT: 12, padB: 32, svgW: 340, svgH: 228 }
const DATE_LABEL_Y = CHART.svgH - 11
const MAX_TICKS = Math.max(5, Math.min(9, Math.floor((CHART.svgH - CHART.padT - CHART.padB) / 13)))

const axis = computed(() => {
  const { padL, padR, padT, padB, svgW, svgH } = CHART
  const chartH = svgH - padT - padB
  return {
    leftX: padL,
    rightX: svgW - padR,
    topY: padT,
    baseY: padT + chartH,
  }
})

// 当日列标（图表最右一列 = 今日，鲜亮；其余为过去，压暗）
const lastIdx = computed(() => trendData.value.length - 1)
function barOpacity(i: number) { return i === lastIdx.value ? 1 : 0.3 }
function nodeOpacity(i: number) { return i === lastIdx.value ? 1 : 0.35 }
function labelWeight(i: number) { return i === lastIdx.value ? 700 : 400 }
function dateWeight(i: number) { return i === lastIdx.value ? 700 : 400 }
function dateFill(i: number) { return i === lastIdx.value ? 'rgb(var(--text-primary))' : 'rgb(var(--text-secondary))' }

function xPos(i: number) {
  const n = trendData.value.length
  const { leftX, rightX } = axis.value
  if (n < 2) return leftX + (rightX - leftX) / 2
  return leftX + i * ((rightX - leftX) / (n - 1))
}

function dateX(i: number) {
  return days.value < 30 ? barCenter(i) : xPos(i)
}

const barSlot = computed(() => {
  const n = trendData.value.length
  if (!n) return 0
  return (axis.value.rightX - axis.value.leftX) / n
})

function barCenter(i: number) {
  return axis.value.leftX + barSlot.value / 2 + i * barSlot.value
}

const barW = computed(() => Math.max(4, Math.min(26, barSlot.value * 0.6)))

const groupW = computed(() => Math.max(6, Math.min(26, barSlot.value * 0.72)))

const w2 = computed(() => Math.max(3, Math.floor(groupW.value / 2) - 1))

function clampSpan(cx: number, w: number) {
  let gl = cx - w / 2
  let gr = cx + w / 2
  const { leftX, rightX } = axis.value
  if (gl < leftX) { gl = leftX; gr = gl + w }
  else if (gr > rightX) { gr = rightX; gl = gr - w }
  return { gl }
}

function feedingRects(i: number) {
  const w = w2.value
  const total = 2 * w + 1
  const { gl } = clampSpan(barCenter(i), total)
  return { mlX: gl, countX: gl + w + 1, w }
}

function singleRects(i: number) {
  const w = barW.value
  return { gl: clampSpan(barCenter(i), w).gl, w }
}

function buildMonotonePath(pts: { x: number, y: number }[]): string {
  const n = pts.length
  if (n === 0) return ''
  if (n === 1) return `M ${pts[0].x},${pts[0].y}`
  const h: number[] = []
  const s: number[] = []
  for (let i = 0; i < n - 1; i++) {
    const dx = pts[i + 1].x - pts[i].x
    h.push(dx)
    s.push(dx === 0 ? 0 : (pts[i + 1].y - pts[i].y) / dx)
  }
  const m: number[] = new Array(n).fill(0)
  m[0] = s[0]
  m[n - 1] = s[n - 2]
  for (let i = 1; i < n - 1; i++) {
    if (s[i - 1] * s[i] <= 0) {
      m[i] = 0
    } else {
      const p = (s[i - 1] * h[i] + s[i] * h[i - 1]) / (h[i - 1] + h[i])
      m[i] = (Math.sign(s[i - 1]) + Math.sign(s[i])) * Math.min(Math.abs(s[i - 1]), Math.abs(s[i]), 0.5 * Math.abs(p))
    }
  }
  let d = `M ${pts[0].x},${pts[0].y}`
  for (let i = 0; i < n - 1; i++) {
    const p0 = pts[i]
    const p1 = pts[i + 1]
    const dx = h[i]
    const c1x = p0.x + dx / 3
    const c1y = p0.y + m[i] * dx / 3
    const c2x = p1.x - dx / 3
    const c2y = p1.y - m[i + 1] * dx / 3
    d += ` C ${c1x.toFixed(2)},${c1y.toFixed(2)} ${c2x.toFixed(2)},${c2y.toFixed(2)} ${p1.x},${p1.y}`
  }
  return d
}

function niceStep(raw: number) {
  if (raw <= 0) return 1
  const mag = Math.pow(10, Math.floor(Math.log10(raw)))
  const norm = raw / mag
  const step = norm >= 5 ? 10 : norm >= 2 ? 5 : norm >= 1 ? 2 : 1
  return step * mag
}

function niceTicks(min: number, max: number, maxCount = 5): number[] {
  const span = (max - min) || 1
  const step = niceStep(span / maxCount)
  const start = Math.ceil(min / step) * step
  const ticks: number[] = []
  for (let v = start; v <= max; v += step) {
    ticks.push(v)
  }
  if (ticks.length && ticks[ticks.length - 1] < max) {
    ticks.push(ticks[ticks.length - 1] + step)
  }
  return ticks
}

function niceTicksInt(min: number, max: number, maxCount = 5): number[] {
  const span = (max - min) || 1
  const step = Math.max(1, niceStep(span / maxCount))
  const start = Math.ceil(min / step) * step
  const ticks: number[] = []
  for (let v = start; v <= max; v += step) {
    ticks.push(v)
  }
  if (ticks.length && ticks[ticks.length - 1] < max) {
    ticks.push(ticks[ticks.length - 1] + step)
  }
  return ticks
}

function formatTick(v: number, step: number) {
  let dp = 0
  if (step < 1) dp = Math.min(2, Math.ceil(-Math.log10(step)))
  return v.toFixed(dp)
}

function leastSquaresLine(values: number[]) {
  const n = values.length
  if (n < 2) return null
  let sx = 0, sy = 0, sxx = 0, sxy = 0
  for (let i = 0; i < n; i++) {
    sx += i
    sy += values[i]
    sxx += i * i
    sxy += i * values[i]
  }
  const denom = n * sxx - sx * sx
  if (denom === 0) return null
  const slope = (n * sxy - sx * sy) / denom
  const intercept = (sy - slope * sx) / n
  return { slope, intercept }
}

function buildBars(getValue: (d: any) => number, opts: { decimals?: number } = {}) {
  const data = trendData.value
  const empty = { items: [] as { x: number, y: number, h: number, value: number, label: string }[], yMax: 1 }
  if (!data.length) return empty
  const { padL, padR, padT, padB, svgW, svgH } = CHART
  const chartW = svgW - padL - padR
  const chartH = svgH - padT - padB
  const values = data.map(getValue)
  const rawMax = Math.max(...values, 0)
  const step = Math.max(1, niceStep(Math.max(rawMax, 1) / 5))
  const yMax = Math.ceil(Math.max(rawMax * 1.2, step) / step) * step
  const xstepN = data.length > 1 ? chartW / (data.length - 1) : chartW
  const items = values.map((v, i) => {
    const x = data.length > 1 ? padL + i * xstepN : padL + chartW / 2
    const y = padT + chartH - (v / yMax) * chartH
    const h = Math.max(0, padT + chartH - y)
    return { x, y, h, value: v, label: opts.decimals != null ? v.toFixed(opts.decimals) : String(Math.round(v)) }
  })
  return { items, yMax }
}

function buildLineChart(getValue: (d: any) => number, opts: { integerTicks?: boolean, forceZero?: boolean, withPath?: boolean } = {}) {
  const data = trendData.value
  const empty = { points: [] as { x: number, y: number, value: number }[], path: '', ticks: [] as { y: number, label: string }[], yMin: 0, yRange: 1, trend: null as { x1: number, y1: number, x2: number, y2: number } | null }
  if (!data.length) return empty
  const { padL, padR, padT, padB, svgW, svgH } = CHART
  const chartW = svgW - padL - padR
  const chartH = svgH - padT - padB
  const { integerTicks = false, forceZero = false, withPath = false } = opts
  const maxTicks = integerTicks ? Math.max(4, Math.min(6, Math.floor(chartH / 20))) : MAX_TICKS

  const values = data.map(getValue)
  const rawMin = Math.min(...values)
  const rawMax = Math.max(...values)
  const range = rawMax - rawMin
  const positiveValues = values.filter(v => v > 0)
  const effectiveMin = (rawMin === 0 && positiveValues.length > 0) ? Math.min(...positiveValues) : rawMin
  const yMin = forceZero ? 0 : (rawMin === 0 ? 0 : (range === 0 ? effectiveMin * 0.5 : effectiveMin - range * 0.15))

  const tickVals = integerTicks ? niceTicksInt(rawMin, rawMax, maxTicks) : niceTicks(rawMin, rawMax, maxTicks)
  let yMax = range === 0 ? effectiveMin * 1.5 : rawMax + range * 0.15
  if (tickVals.length) {
    if (tickVals[0] > rawMin) {
      const lower = tickVals[0] - (tickVals[1] - tickVals[0])
      if (lower >= yMin) tickVals.unshift(lower)
    }
    const lastTick = tickVals[tickVals.length - 1]
    if (lastTick >= yMax) yMax = lastTick + ((tickVals[1] - tickVals[0]) || 1)
  }
  const yRange = yMax - yMin || 1
  const tickStep = integerTicks ? Math.max(1, niceStep((range || 1) / maxTicks)) : niceStep((range || 1) / maxTicks)

  const xstep = data.length > 1 ? chartW / (data.length - 1) : 0
  const points = data.map((d, i) => {
    const val = getValue(d)
    const y = padT + chartH - (Math.max(val, yMin) - yMin) / yRange * chartH
    return { x: data.length > 1 ? padL + i * xstep : padL + chartW / 2, y, value: val }
  })

  const ticks = tickVals.map(v => ({
    y: padT + chartH - (v - yMin) / yRange * chartH,
    label: formatTick(v, tickStep),
  }))

  const path = withPath ? buildMonotonePath(points) : ''

  const reg = leastSquaresLine(values)
  let trend = null as { x1: number, y1: number, x2: number, y2: number } | null
  if (reg) {
    const mapY = (v: number) => {
      const c = Math.max(yMin, Math.min(v, yMax))
      return padT + chartH - (c - yMin) / yRange * chartH
    }
    const x0 = points.length > 1 ? padL : padL + chartW / 2
    const x1 = points.length > 1 ? padL + chartW : x0
    trend = {
      x1: x0,
      y1: mapY(reg.intercept),
      x2: x1,
      y2: mapY(reg.slope * (values.length - 1) + reg.intercept),
    }
  }

  return { points, path, ticks, yMin, yRange, trend }
}

const feedingMl = computed(() => buildBars(d => d.total_ml || 0))
const feedingCount = computed(() => buildBars(d => d.feeding_count || 0))
const diaper = computed(() => buildBars(d => d.diaper_count || 0))
const sleep = computed(() => buildBars(d => (d.sleep_duration_minutes || 0) / 60, { decimals: 1 }))
const outdoor = computed(() => buildBars(d => (d.outdoor_duration_minutes || 0) / 60, { decimals: 1 }))
const supplement = computed(() => buildBars(d => d.supplement_count || 0))

const feedingMlScatter = computed(() => buildLineChart(d => d.total_ml || 0, { integerTicks: true }))
const feedingCountScatter = computed(() => buildLineChart(d => d.feeding_count || 0, { integerTicks: true, forceZero: true }))
const diaperScatter = computed(() => buildLineChart(d => d.diaper_count || 0, { integerTicks: true }))
const sleepScatter = computed(() => buildLineChart(d => (d.sleep_duration_minutes || 0) / 60, { integerTicks: true }))
const outdoorScatter = computed(() => buildLineChart(d => (d.outdoor_duration_minutes || 0) / 60, { integerTicks: true }))
const supplementScatter = computed(() => buildLineChart(d => d.supplement_count || 0, { integerTicks: true }))

// 体温折线：仅「有测量」的日子连点（无测量日留空，不落 0），y 轴按实测范围自适应
// （体温恒定在 35–42℃，不能用 0 基线；并保证 37.5 发烧参考线始终在范围内）
const tempSeries = computed(() => {
  const data = trendData.value
  const { padL, padR, padT, padB, svgW, svgH } = CHART
  const chartW = svgW - padL - padR
  const chartH = svgH - padT - padB
  const n = data.length
  const step = n > 1 ? chartW / (n - 1) : 0
  const xOf = (i: number) => (n > 1 ? padL + i * step : padL + chartW / 2)

  const vals = data.map(d => d.temperature_high || 0)
  const pos = vals.filter(v => v > 0)
  if (!pos.length) return { points: [] as { x: number, y: number, value: number }[], last: null as { x: number, y: number, value: number } | null, ticks: [] as { y: number, label: string }[], yMin: 36, yRange: 1 }

  const rawMin = Math.min(...pos, 37.5)
  const rawMax = Math.max(...pos, 37.5)
  const span = Math.max(rawMax - rawMin, 0.5)
  const yMin = Math.floor((rawMin - span * 0.3) * 10) / 10
  const yMax = Math.ceil((rawMax + span * 0.3) * 10) / 10
  const yRange = (yMax - yMin) || 1
  const yOf = (v: number) => padT + chartH - (v - yMin) / yRange * chartH

  const points = data
    .map((d, i) => ({ i, v: d.temperature_high || 0 }))
    .filter(p => p.v > 0)
    .map(p => ({ x: xOf(p.i), y: yOf(p.v), value: p.v }))

  // 刻度取 0.5℃ 的整数倍，最多 6 条
  const ticks: { y: number, label: string }[] = []
  const tStep = Math.max(0.5, Math.ceil((yRange / 5) * 2) / 2)
  for (let v = Math.ceil(yMin / tStep) * tStep; v <= yMax + 1e-9; v += tStep) {
    ticks.push({ y: yOf(v), label: v.toFixed(1) })
  }

  return { points, last: points.length ? points[points.length - 1] : null, ticks, yMin, yRange }
})
const tempTicks = computed(() => tempSeries.value.ticks)
const tempPoints = computed(() => tempSeries.value.points)
// 温度线：过去段压暗、今日点鲜亮（仅连接有测量的日子）
const tempPastPath = computed(() => {
  const p = tempSeries.value.points
  if (!p.length) return ''
  return buildMonotonePath(p.length > 1 ? p.slice(0, -1) : p)
})
const tempLastSeg = computed(() => {
  const p = tempSeries.value.points
  if (p.length < 2) return ''
  return buildMonotonePath(p.slice(-2))
})
const tempLastPoint = computed(() => tempSeries.value.last)

const feverLineY = computed(() => {
  const { padT, padB, svgH } = CHART
  const chartH = svgH - padT - padB
  const s = tempSeries.value
  return padT + chartH - (37.5 - s.yMin) / s.yRange * chartH
})

async function loadTrend(silent: boolean = false) {
  const baby = app.currentBaby
  if (!baby) return
  if (!silent) loading.value = true
  try {
    const res = await babyAPI.trend(baby.id, days.value * 3 + 1)
    const all = res.data || []
    // 图表含当日（今日 + 往前 days 天 = days+1 列）；对比周期不含今日
    // cur = 前 1~N 天；prev = 再往前 N 天
    const d = days.value
    trendData.value = all.slice(-(d + 1))
    trendCur.value = all.slice(-(d + 1), -1)
    trendPrev.value = all.slice(-(2 * d + 1), -(d + 1))
  } catch {
    trendData.value = []
    trendCur.value = []
    trendPrev.value = []
    app.showToast('趋势数据加载失败', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(() => { if (app.currentBaby) loadTrend() })

// 冷启动/切换宝宝时 currentBaby 可能晚于本页挂载就绪（同 P0）：监听其就绪后补载
watch(() => app.currentBaby?.id, (id) => {
  if (id) loadTrend()
})
</script>
