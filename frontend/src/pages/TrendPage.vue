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
                  <span class="inline-block w-1.5 h-2.5 rounded-sm" style="background: var(--chart-primary-count)"></span>
                  次数
                </span>
              </span>
          </h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <template v-if="days === 30">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
              <g v-for="(t, ti) in feedingMlScatter.ticks" :key="'fl'+ti">
                <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
                <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" fill="#6b7280">{{ t.label }}</text>
              </g>
              <g v-for="(t, ti) in feedingCountScatter.ticks" :key="'fr'+ti">
                <text :x="axis.rightX + 5" :y="t.y + 3" text-anchor="start" font-size="9" fill="#9ca3af">{{ t.label }}</text>
              </g>
              <g v-for="(pt, i) in feedingMlScatter.points" :key="'dv'+i">
                <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
              </g>
              <line v-if="feedingMlScatter.trend" :x1="feedingMlScatter.trend.x1" :y1="feedingMlScatter.trend.y1" :x2="feedingMlScatter.trend.x2" :y2="feedingMlScatter.trend.y2" stroke="var(--chart-primary)" stroke-width="1.5" stroke-dasharray="6,3" opacity="0.85"/>
              <line v-if="feedingCountScatter.trend" :x1="feedingCountScatter.trend.x1" :y1="feedingCountScatter.trend.y1" :x2="feedingCountScatter.trend.x2" :y2="feedingCountScatter.trend.y2" stroke="var(--chart-primary-count)" stroke-width="1.5" stroke-dasharray="6,3" opacity="0.85"/>
              <g v-for="(pt, i) in feedingMlScatter.points" :key="'dp'+i">
                <circle :cx="pt.x" :cy="pt.y" r="2.5" fill="var(--chart-primary)"/>
              </g>
              <g v-for="(pt, i) in feedingCountScatter.points" :key="'cp'+i">
                <circle :cx="pt.x" :cy="pt.y" r="2.5" fill="var(--chart-primary-count)"/>
              </g>
            </template>
            <template v-else>
              <g v-for="(b, i) in feedingMl.items" :key="'bm'+i">
                <rect :x="feedingRects(i).mlX" :y="b.y" :width="w2" :height="b.h" rx="2" fill="var(--chart-primary)" opacity="0.85"/>
                <text v-if="b.h > 0" :x="feedingRects(i).mlX + w2 / 2" :y="b.y - 3" text-anchor="middle" font-size="8" fill="#6b7280">{{ b.label }}</text>
              </g>
              <g v-for="(b, i) in feedingCount.items" :key="'bc'+i">
                <rect :x="feedingRects(i).countX" :y="b.y" :width="w2" :height="b.h" rx="2" fill="var(--chart-primary-count)" opacity="0.85"/>
                <text v-if="b.h > 0" :x="feedingRects(i).countX + w2 / 2" :y="b.y - 3" text-anchor="middle" font-size="8" fill="#6b7280">{{ b.label }}</text>
              </g>
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            </template>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <line :x1="axis.rightX" :x2="axis.rightX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">ml</text>
            <text :x="axis.rightX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">次</text>
            <template v-for="(d, i) in trendData" :key="'fx'+i">
              <text v-if="dateLabels[i]?.show" :x="xPos(i)" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ dateLabels[i]?.label }}</text>
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
            <template v-if="days === 30">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
              <g v-for="(t, ti) in diaperScatter.ticks" :key="'dl'+ti">
                <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
                <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" fill="#6b7280">{{ t.label }}</text>
              </g>
              <g v-for="(pt, i) in diaperScatter.points" :key="'dv'+i">
                <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
              </g>
              <line v-if="diaperScatter.trend" :x1="diaperScatter.trend.x1" :y1="diaperScatter.trend.y1" :x2="diaperScatter.trend.x2" :y2="diaperScatter.trend.y2" stroke="var(--chart-diaper)" stroke-width="1.5" stroke-dasharray="6,3" opacity="0.85"/>
              <g v-for="(pt, i) in diaperScatter.points" :key="'dp'+i">
                <circle :cx="pt.x" :cy="pt.y" r="2.5" fill="var(--chart-diaper)"/>
              </g>
            </template>
            <template v-else>
              <g v-for="(b, i) in diaper.items" :key="'db'+i">
                <rect :x="singleRects(i).gl" :y="b.y" :width="barW" :height="b.h" rx="2" fill="var(--chart-diaper)" opacity="0.85"/>
                <text v-if="b.h > 0" :x="singleRects(i).gl + barW / 2" :y="b.y - 3" text-anchor="middle" font-size="8" fill="#6b7280">{{ b.label }}</text>
              </g>
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            </template>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">次</text>
            <template v-for="(d, i) in trendData" :key="'dx'+i">
              <text v-if="dateLabels[i]?.show" :x="xPos(i)" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ dateLabels[i]?.label }}</text>
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
            <template v-if="days === 30">
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
              <g v-for="(t, ti) in sleepScatter.ticks" :key="'sl'+ti">
                <line :x1="axis.leftX" :x2="axis.rightX" :y1="t.y" :y2="t.y" class="chart-grid"/>
                <text :x="axis.leftX - 5" :y="t.y + 3" text-anchor="end" font-size="9" fill="#6b7280">{{ t.label }}</text>
              </g>
              <g v-for="(pt, i) in sleepScatter.points" :key="'sv'+i">
                <line :x1="pt.x" :y1="axis.topY" :x2="pt.x" :y2="axis.baseY" class="chart-guide"/>
              </g>
              <line v-if="sleepScatter.trend" :x1="sleepScatter.trend.x1" :y1="sleepScatter.trend.y1" :x2="sleepScatter.trend.x2" :y2="sleepScatter.trend.y2" stroke="var(--chart-sleep)" stroke-width="1.5" stroke-dasharray="6,3" opacity="0.85"/>
              <g v-for="(pt, i) in sleepScatter.points" :key="'sp'+i">
                <circle :cx="pt.x" :cy="pt.y" r="2.5" fill="var(--chart-sleep)"/>
              </g>
            </template>
            <template v-else>
              <g v-for="(b, i) in sleep.items" :key="'sb'+i">
                <rect :x="singleRects(i).gl" :y="b.y" :width="barW" :height="b.h" rx="2" fill="var(--chart-sleep)" opacity="0.85"/>
                <text v-if="b.h > 0" :x="singleRects(i).gl + barW / 2" :y="b.y - 3" text-anchor="middle" font-size="8" fill="#6b7280">{{ b.label }}</text>
              </g>
              <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            </template>
            <line :x1="axis.leftX" :x2="axis.leftX" :y1="axis.topY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
            <text :x="axis.leftX" :y="axis.topY - 5" text-anchor="middle" font-size="9" fill="#9ca3af">小时</text>
            <template v-for="(d, i) in trendData" :key="'sx'+i">
              <text v-if="dateLabels[i]?.show" :x="xPos(i)" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ dateLabels[i]?.label }}</text>
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
            <line :x1="axis.leftX" :x2="axis.rightX" :y1="axis.baseY" :y2="axis.baseY" stroke="#d1d5db" stroke-width="1"/>
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

function xPos(i: number) {
  const n = trendData.value.length
  const { leftX, rightX } = axis.value
  if (n < 2) return leftX + (rightX - leftX) / 2
  return leftX + i * ((rightX - leftX) / (n - 1))
}

const xstep = computed(() => {
  const n = trendData.value.length
  if (n < 2) return 0
  return (axis.value.rightX - axis.value.leftX) / (n - 1)
})

const barW = computed(() => Math.max(4, Math.min(26, xstep.value * 0.6)))

const groupW = computed(() => Math.max(6, Math.min(26, xstep.value * 0.72)))

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
  const { gl } = clampSpan(xPos(i), total)
  return { mlX: gl, countX: gl + w + 1, w }
}

function singleRects(i: number) {
  const w = barW.value
  return { gl: clampSpan(xPos(i), w).gl, w }
}

function buildSmoothPath(pts: { x: number, y: number }[]): string {
  if (pts.length === 0) return ''
  if (pts.length === 1) return `M ${pts[0].x},${pts[0].y}`
  let d = `M ${pts[0].x},${pts[0].y}`
  for (let i = 0; i < pts.length - 1; i++) {
    const p0 = pts[i - 1] || pts[i]
    const p1 = pts[i]
    const p2 = pts[i + 1]
    const p3 = pts[i + 2] || p2
    const c1x = p1.x + (p2.x - p0.x) / 3
    const c1y = p1.y + (p2.y - p0.y) / 3
    const c2x = p2.x - (p3.x - p1.x) / 3
    const c2y = p2.y - (p3.y - p1.y) / 3
    d += ` C ${c1x.toFixed(2)},${c1y.toFixed(2)} ${c2x.toFixed(2)},${c2y.toFixed(2)} ${p2.x},${p2.y}`
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

function buildLineChart(getValue: (d: any) => number, opts: { integerTicks?: boolean, forceZero?: boolean } = {}) {
  const data = trendData.value
  const empty = { points: [] as { x: number, y: number, value: number }[], path: '', ticks: [] as { y: number, label: string }[], yMin: 0, yRange: 1, trend: null as { x1: number, y1: number, x2: number, y2: number } | null }
  if (!data.length) return empty
  const { padL, padR, padT, padB, svgW, svgH } = CHART
  const chartW = svgW - padL - padR
  const chartH = svgH - padT - padB
  const { integerTicks = false, forceZero = false } = opts
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

  const path = buildSmoothPath(points)

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

const feedingMlScatter = computed(() => buildLineChart(d => d.total_ml || 0, { integerTicks: true }))
const feedingCountScatter = computed(() => buildLineChart(d => d.feeding_count || 0, { integerTicks: true, forceZero: true }))
const diaperScatter = computed(() => buildLineChart(d => d.diaper_count || 0, { integerTicks: true }))
const sleepScatter = computed(() => buildLineChart(d => (d.sleep_duration_minutes || 0) / 60, { integerTicks: true }))

const tempChart = computed(() => buildLineChart(d => d.temperature_high || 0))
const tempTicks = computed(() => tempChart.value.ticks)
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
