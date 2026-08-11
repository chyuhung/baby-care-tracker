<template>
  <div class="flex flex-col min-h-screen">
    <header class="app-header pt-safe px-4 py-3 border-b border-border-color">
      <h1 class="text-lg font-bold text-text-primary">趋势</h1>
      <div class="flex gap-2 mt-2">
        <button v-for="d in dayOptions" :key="d.value"
          @click="days = d.value; loadTrend()"
          :class="['px-3 py-1 rounded-full text-xs font-medium transition-colors btn-press',
            days === d.value ? 'bg-primary text-white' : 'bg-gray-100 text-text-secondary']">
          {{ d.label }}
        </button>
      </div>
    </header>

    <main class="flex-1 min-h-0 px-4 py-4 overflow-y-auto pb-20 space-y-6">
      <div v-if="loading" class="text-center py-16 text-text-secondary">加载中...</div>
      <div v-else-if="trendData.length === 0" class="text-center py-16">
        <div class="text-5xl mb-4">📊</div>
        <p class="text-text-secondary">暂无趋势数据</p>
      </div>
      <template v-else>
        <div>
          <h4 class="text-sm font-semibold text-text-secondary mb-2">
            <span class="flex items-center gap-2">
              🍼 每日奶量
              <span class="flex items-center gap-1 text-[10px] font-normal">
                <span class="inline-block w-4" style="border-top: 2px solid var(--chart-primary)"></span>
                奶量
                <span class="inline-block w-4" style="border-top: 2px dashed var(--chart-primary-count)"></span>
                次数
              </span>
            </span>
          </h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <g v-for="(t, ti) in feedingTicks" :key="'fl'+ti">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
              <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" fill="#6b7280">{{ t.label }}</text>
            </g>
            <g v-for="(t, ti) in feedingCountTicks" :key="'fr'+ti">
              <text :x="axis.rightX + 5" :y="t.y + 3" text-anchor="start" font-size="9" fill="#9ca3af">{{ t.label }}</text>
            </g>
            <g v-for="(pt, i) in feedingPoints" :key="'dv'+i">
              <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
            </g>
            <path :d="feedingPath" class="chart-line-primary" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path :d="feedingCountPath" class="chart-line-primary-count" fill="none" stroke-width="2" stroke-dasharray="5,4" stroke-linecap="round" stroke-linejoin="round"/>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <line :x1="axis.rightX" :x2="axis.rightX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">ml</text>
            <text :x="axis.rightX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">次</text>
            <template v-for="(d, i) in trendData" :key="'fx'+i">
              <text v-if="dateLabels[i]?.show" :x="feedingPoints[i]?.x" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div>
          <h4 class="text-sm font-semibold text-text-secondary mb-2">
            <span class="flex items-center gap-2">
              🩲 每日尿布
            </span>
          </h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <g v-for="(t, ti) in diaperTicks" :key="'dl'+ti">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
              <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" fill="#6b7280">{{ t.label }}</text>
            </g>
            <g v-for="(pt, i) in diaperPoints" :key="'dv'+i">
              <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
            </g>
            <path :d="diaperPath" class="chart-line-diaper" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">次</text>
            <template v-for="(d, i) in trendData" :key="'dx'+i">
              <text v-if="dateLabels[i]?.show" :x="diaperPoints[i]?.x" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div>
          <h4 class="text-sm font-semibold text-text-secondary mb-2">
            <span class="flex items-center gap-2">
              😴 每日睡眠
            </span>
          </h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <g v-for="(t, ti) in sleepTicks" :key="'sl'+ti">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
              <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" fill="#6b7280">{{ t.label }}</text>
            </g>
            <g v-for="(pt, i) in sleepPoints" :key="'sv'+i">
              <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
            </g>
            <path :d="sleepPath" class="chart-line-sleep" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">小时</text>
            <template v-for="(d, i) in trendData" :key="'sx'+i">
              <text v-if="dateLabels[i]?.show" :x="sleepPoints[i]?.x" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div>
          <h4 class="text-sm font-semibold text-text-secondary mb-2">
            <span class="flex items-center gap-2">
              🌡️ 每日最高体温
            </span>
          </h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <g v-for="(t, ti) in tempTicks" :key="'tl'+ti">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
              <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" fill="#6b7280">{{ t.label }}</text>
            </g>
            <g v-for="(pt, i) in tempPoints" :key="'tv'+i">
              <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
            </g>
            <line :x1="axis.leftX" :x2="axis.rightX" :y1="feverLineY" :y2="feverLineY" stroke="#ef4444" stroke-width="1" stroke-dasharray="4,3" opacity="0.5"/>
            <path :d="tempPath" class="chart-line-temperature" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">°C</text>
            <template v-for="(d, i) in trendData" :key="'tx'+i">
              <text v-if="dateLabels[i]?.show" :x="tempPoints[i]?.x" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { babyAPI } from '@/api'

const app = useAppStore()
const trendData = ref<any[]>([])
const loading = ref(false)
const days = ref(7)

const dateLabels = computed(() => {
  return trendData.value.map((d, i) => {
    const parts = d.date.split('-')
    const label = `${parseInt(parts[1])}/${parseInt(parts[2])}`
    const show = days.value < 30 || i % 5 === 0
    return { label, show }
  })
})

const dayOptions = [
  { label: '7天', value: 7 },
  { label: '30天', value: 30 },
]

const CHART = { padL: 32, padR: 30, padT: 15, padB: 35, svgW: 340, svgH: 170 }
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

function buildSmoothPath(pts: { x: number, y: number }[]): string {
  if (pts.length === 0) return ''
  if (pts.length === 1) return `M ${pts[0].x},${pts[0].y}`
  let d = `M ${pts[0].x},${pts[0].y}`
  for (let i = 0; i < pts.length - 1; i++) {
    const p0 = pts[i - 1] || pts[i]
    const p1 = pts[i]
    const p2 = pts[i + 1]
    const p3 = pts[i + 2] || p2
    const c1x = p1.x + (p2.x - p0.x) / 6
    const c1y = p1.y + (p2.y - p0.y) / 6
    const c2x = p2.x - (p3.x - p1.x) / 6
    const c2y = p2.y - (p3.y - p1.y) / 6
    d += ` C ${c1x.toFixed(2)},${c1y.toFixed(2)} ${c2x.toFixed(2)},${c2y.toFixed(2)} ${p2.x},${p2.y}`
  }
  return d
}

function buildLineChart(getValue: (d: any) => number, opts: { integerTicks?: boolean } = {}) {
  const data = trendData.value
  const empty = { points: [] as { x: number, y: number, value: number }[], path: '', ticks: [] as { y: number, label: string }[], yMin: 0, yRange: 1 }
  if (!data.length) return empty
  const { padL, padR, padT, padB, svgW, svgH } = CHART
  const chartW = svgW - padL - padR
  const chartH = svgH - padT - padB
  const { integerTicks = false } = opts
  const maxTicks = integerTicks ? Math.max(4, Math.min(6, Math.floor(chartH / 20))) : MAX_TICKS

  const values = data.map(getValue)
  const rawMin = Math.min(...values)
  const rawMax = Math.max(...values)
  const range = rawMax - rawMin
  const positiveValues = values.filter(v => v > 0)
  const effectiveMin = (rawMin === 0 && positiveValues.length > 0) ? Math.min(...positiveValues) : rawMin
  const yMin = rawMin === 0 ? 0 : (range === 0 ? effectiveMin * 0.5 : effectiveMin - range * 0.15)

  const tickVals = integerTicks ? niceTicksInt(rawMin, rawMax, maxTicks) : niceTicks(rawMin, rawMax, maxTicks)
  let yMax = range === 0 ? effectiveMin * 1.5 : rawMax + range * 0.15
  if (tickVals.length) {
    if (tickVals[0] > rawMin) {
      const lower = tickVals[0] - (tickVals[1] - tickVals[0])
      if (lower >= yMin) tickVals.unshift(lower)
    }
    if (tickVals[tickVals.length - 1] > yMax) yMax = tickVals[tickVals.length - 1]
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

  const path = buildSmoothPath(points)
  return { points, path, ticks, yMin, yRange }
}

const feedingChart = computed(() => buildLineChart(d => d.total_ml || 0, { integerTicks: true }))
const feedingPoints = computed(() => feedingChart.value.points)
const feedingPath = computed(() => feedingChart.value.path)

const feedingCountChart = computed(() => buildLineChart(d => d.feeding_count || 0, { integerTicks: true }))
const feedingCountPoints = computed(() => feedingCountChart.value.points)
const feedingCountPath = computed(() => feedingCountChart.value.path)

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

const feedingTicks = computed(() => feedingChart.value.ticks)
const feedingCountTicks = computed(() => feedingCountChart.value.ticks.map(t => ({
  y: t.y,
  label: t.label,
})))
const diaperTicks = computed(() => diaperChart.value.ticks)
const sleepTicks = computed(() => sleepChart.value.ticks)
const tempTicks = computed(() => tempChart.value.ticks)

const diaperChart = computed(() => buildLineChart(d => d.diaper_count || 0))
const diaperPoints = computed(() => diaperChart.value.points)
const diaperPath = computed(() => diaperChart.value.path)

const sleepChart = computed(() => buildLineChart(d => (d.sleep_duration_minutes || 0) / 60))
const sleepPoints = computed(() => sleepChart.value.points)
const sleepPath = computed(() => sleepChart.value.path)

const tempChart = computed(() => buildLineChart(d => d.temperature_high || 0))
const tempPoints = computed(() => tempChart.value.points)
const tempPath = computed(() => tempChart.value.path)

const feverLineY = computed(() => {
  const { padT, padB, svgH } = CHART
  const chartH = svgH - padT - padB
  const sc = tempChart.value
  if (!sc.ticks.length) return padT + chartH / 2
  return padT + chartH - (37.5 - sc.yMin) / sc.yRange * chartH
})

async function loadTrend() {
  const baby = app.currentBaby
  if (!baby) return
  loading.value = true
  try {
    const res = await babyAPI.trend(baby.id, days.value)
    trendData.value = res.data
  } catch {
    trendData.value = []
    app.showToast('趋势数据加载失败', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(() => { if (app.currentBaby) loadTrend() })
</script>