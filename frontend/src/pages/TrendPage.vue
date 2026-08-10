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
            <path :d="feedingAreaPath" class="chart-fill-primary" opacity="0.08"/>
            <path :d="feedingPath" class="chart-line-primary" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path :d="feedingCountPath" class="chart-line-primary-count" fill="none" stroke-width="2" stroke-dasharray="5,4" stroke-linecap="round" stroke-linejoin="round"/>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <line :x1="axis.rightX" :x2="axis.rightX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">ml</text>
            <text :x="axis.rightX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">次</text>
            <template v-for="(pt, i) in feedingPoints" :key="'fl'+i">
              <g v-if="feedingLabelVis[i]">
                <line :x1="pt.x" :y1="pt.y" :x2="axis.leftX" :y2="pt.y" class="chart-guide"/>
                <text :x="axis.leftX - 4" :y="pt.y + 3" text-anchor="end" font-size="9" fill="#9ca3af">{{ pt.value }}</text>
              </g>
            </template>
            <template v-for="(pt, i) in feedingCountPoints" :key="'fc'+i">
              <g v-if="feedingCountLabelVis[i]">
                <line :x1="pt.x" :y1="pt.y" :x2="axis.rightX" :y2="pt.y" class="chart-guide"/>
                <text :x="axis.rightX + 4" :y="pt.y + 3" text-anchor="start" font-size="9" fill="#9ca3af">{{ pt.value }}</text>
              </g>
            </template>
            <g v-for="(pt, i) in feedingPoints" :key="'fp'+i">
              <circle :cx="pt.x" :cy="pt.y" r="3" fill="white" class="chart-line-primary" stroke-width="2"/>
            </g>
            <g v-for="(pt, i) in feedingCountPoints" :key="'fcp'+i">
              <circle :cx="pt.x" :cy="pt.y" r="2.5" fill="white" class="chart-line-primary-count" stroke-width="2"/>
            </g>
            <template v-for="(d, i) in trendData" :key="'fx'+i">
              <text v-if="dateLabels[i]?.show" :x="feedingPoints[i]?.x" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div>
          <h4 class="text-sm font-semibold text-text-secondary mb-2">
            <span class="flex items-center gap-2">
              🩲 每日尿布
              <span class="text-[10px] font-normal">(次)</span>
            </span>
          </h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <path :d="diaperAreaPath" class="chart-fill-diaper" opacity="0.08"/>
            <path :d="diaperPath" class="chart-line-diaper" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">次</text>
            <template v-for="(pt, i) in diaperPoints" :key="'dl'+i">
              <g v-if="diaperLabelVis[i]">
                <line :x1="pt.x" :y1="pt.y" :x2="axis.leftX" :y2="pt.y" class="chart-guide"/>
                <text :x="axis.leftX - 4" :y="pt.y + 3" text-anchor="end" font-size="9" fill="#9ca3af">{{ pt.value }}</text>
              </g>
            </template>
            <g v-for="(pt, i) in diaperPoints" :key="'dp'+i">
              <circle :cx="pt.x" :cy="pt.y" r="3" fill="white" class="chart-line-diaper" stroke-width="2"/>
            </g>
            <template v-for="(d, i) in trendData" :key="'dx'+i">
              <text v-if="dateLabels[i]?.show" :x="diaperPoints[i]?.x" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div>
          <h4 class="text-sm font-semibold text-text-secondary mb-2">
            <span class="flex items-center gap-2">
              😴 每日睡眠
              <span class="text-[10px] font-normal">(小时)</span>
            </span>
          </h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <path :d="sleepAreaPath" class="chart-fill-sleep" opacity="0.08"/>
            <path :d="sleepPath" class="chart-line-sleep" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">小时</text>
            <template v-for="(pt, i) in sleepPoints" :key="'sl'+i">
              <g v-if="sleepLabelVis[i]">
                <line :x1="pt.x" :y1="pt.y" :x2="axis.leftX" :y2="pt.y" class="chart-guide"/>
                <text :x="axis.leftX - 4" :y="pt.y + 3" text-anchor="end" font-size="9" fill="#9ca3af">{{ sleepLabels[i] }}</text>
              </g>
            </template>
            <g v-for="(pt, i) in sleepPoints" :key="'sp'+i">
              <circle :cx="pt.x" :cy="pt.y" r="3" fill="white" class="chart-line-sleep" stroke-width="2"/>
            </g>
            <template v-for="(d, i) in trendData" :key="'sx'+i">
              <text v-if="dateLabels[i]?.show" :x="sleepPoints[i]?.x" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ dateLabels[i]?.label }}</text>
            </template>
          </svg>
        </div>
        <div>
          <h4 class="text-sm font-semibold text-text-secondary mb-2">
            <span class="flex items-center gap-2">
              🌡️ 每日最高体温
              <span class="text-[10px] font-normal">(°C)</span>
            </span>
          </h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <path :d="tempAreaPath" class="chart-fill-temperature" opacity="0.08"/>
            <path :d="tempPath" class="chart-line-temperature" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <line :x1="axis.leftX" :x2="axis.rightX" :y1="feverLineY" :y2="feverLineY" stroke="#ef4444" stroke-width="1" stroke-dasharray="4,3" opacity="0.5"/>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">°C</text>
            <template v-for="(pt, i) in tempPoints" :key="'tl'+i">
              <g v-if="tempLabelVis[i]">
                <line :x1="pt.x" :y1="pt.y" :x2="axis.leftX" :y2="pt.y" class="chart-guide"/>
                <text :x="axis.leftX - 4" :y="pt.y + 3" text-anchor="end" font-size="9" fill="#9ca3af">{{ pt.value }}</text>
              </g>
            </template>
            <g v-for="(pt, i) in tempPoints" :key="'tp'+i">
              <circle :cx="pt.x" :cy="pt.y" r="3" fill="white" class="chart-line-temperature" stroke-width="2"/>
            </g>
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

function buildLineChart(getValue: (d: any) => number) {
  const data = trendData.value
  const empty = { points: [] as { x: number, y: number, value: number }[], path: '', areaPath: '' }
  if (!data.length) return empty
  const { padL, padR, padT, padB, svgW, svgH } = CHART
  const chartW = svgW - padL - padR
  const chartH = svgH - padT - padB

  const values = data.map(getValue)
  const rawMin = Math.min(...values)
  const rawMax = Math.max(...values)
  const range = rawMax - rawMin
  const positiveValues = values.filter(v => v > 0)
  const effectiveMin = (rawMin === 0 && positiveValues.length > 0) ? Math.min(...positiveValues) : rawMin
  const yMin = range === 0 ? effectiveMin * 0.5 : effectiveMin - range * 0.15
  const yMax = range === 0 ? effectiveMin * 1.5 : rawMax + range * 0.15
  const yRange = yMax - yMin || 1

  const step = data.length > 1 ? chartW / (data.length - 1) : 0
  const points = data.map((d, i) => {
    const val = getValue(d)
    const y = padT + chartH - (Math.max(val, yMin) - yMin) / yRange * chartH
    return { x: data.length > 1 ? padL + i * step : padL + chartW / 2, y, value: val }
  })

  const path = buildSmoothPath(points)
  const bY = padT + chartH
  let areaPath = ''
  if (points.length >= 2) {
    const first = points[0]
    const last = points[points.length - 1]
    areaPath = `M ${first.x},${bY} ${path.slice(2)} L ${last.x},${bY} Z`
  }
  return { points, path, areaPath }
}

const feedingChart = computed(() => buildLineChart(d => d.total_ml || 0))
const feedingPoints = computed(() => feedingChart.value.points)
const feedingPath = computed(() => feedingChart.value.path)
const feedingAreaPath = computed(() => feedingChart.value.areaPath)

const feedingCountChart = computed(() => buildLineChart(d => d.feeding_count || 0))
const feedingCountPoints = computed(() => feedingCountChart.value.points)
const feedingCountPath = computed(() => feedingCountChart.value.path)

function pickLabels(pts: { y: number }[], minGap = 13): boolean[] {
  const vis = pts.map(() => false)
  let lastY = -Infinity
  for (let i = 0; i < pts.length; i++) {
    if (Math.abs(pts[i].y - lastY) >= minGap) {
      vis[i] = true
      lastY = pts[i].y
    }
  }
  return vis
}

const feedingLabelVis = computed(() => pickLabels(feedingPoints.value))
const feedingCountLabelVis = computed(() => pickLabels(feedingCountPoints.value))
const diaperLabelVis = computed(() => pickLabels(diaperPoints.value))
const sleepLabelVis = computed(() => pickLabels(sleepPoints.value))
const tempLabelVis = computed(() => pickLabels(tempPoints.value))

const diaperChart = computed(() => buildLineChart(d => d.diaper_count || 0))
const diaperPoints = computed(() => diaperChart.value.points)
const diaperPath = computed(() => diaperChart.value.path)
const diaperAreaPath = computed(() => diaperChart.value.areaPath)

const sleepChart = computed(() => buildLineChart(d => d.sleep_duration_minutes || 0))
const sleepPoints = computed(() => sleepChart.value.points)
const sleepPath = computed(() => sleepChart.value.path)
const sleepAreaPath = computed(() => sleepChart.value.areaPath)

function formatSleepLabel(mins: number) {
  if (mins <= 0) return '0'
  return `${Math.round(mins / 60 * 10) / 10}`
}

const sleepLabels = computed(() => trendData.value.map(d => formatSleepLabel(d.sleep_duration_minutes || 0)))

const tempChart = computed(() => buildLineChart(d => d.temperature_high || 0))
const tempPoints = computed(() => tempChart.value.points)
const tempPath = computed(() => tempChart.value.path)
const tempAreaPath = computed(() => tempChart.value.areaPath)

const feverLineY = computed(() => {
  const { padT, padB, svgH } = CHART
  const chartH = svgH - padT - padB
  const values = trendData.value.map(d => d.temperature_high || 0)
  if (!values.length) return padT + chartH / 2
  const rawMin = Math.min(...values)
  const rawMax = Math.max(...values)
  const range = rawMax - rawMin
  const yMin = range === 0 ? rawMin * 0.5 : rawMin - range * 0.15
  const yMax = range === 0 ? rawMin * 1.5 : rawMax + range * 0.15
  const yRange = yMax - yMin || 1
  return padT + chartH - (37.5 - yMin) / yRange * chartH
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