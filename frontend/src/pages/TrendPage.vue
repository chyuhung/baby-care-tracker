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
          <h4 class="text-sm font-semibold text-text-secondary mb-2">🍼 每日奶量 (ml)</h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <polygon :points="feedingAreaPoints" class="chart-fill-primary" opacity="0.08"/>
            <polyline :points="feedingLinePoints" class="chart-line-primary" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <g v-for="(pt, i) in feedingPoints" :key="'fp'+i">
              <circle :cx="pt.x" :cy="pt.y" r="3" fill="white" class="chart-line-primary" stroke-width="2"/>
              <text :x="pt.x" :y="pt.y - 8" text-anchor="middle" font-size="10" fill="#6b7280">{{ pt.value }}</text>
            </g>
            <text v-for="(d, i) in trendData" :key="'fx'+i" :x="feedingPoints[i]?.x" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ d.date.slice(5) }}</text>
          </svg>
        </div>
        <div>
          <h4 class="text-sm font-semibold text-text-secondary mb-2">🩲 每日尿布次数</h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <polygon :points="diaperAreaPoints" class="chart-fill-diaper" opacity="0.08"/>
            <polyline :points="diaperLinePoints" class="chart-line-diaper" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <g v-for="(pt, i) in diaperPoints" :key="'dp'+i">
              <circle :cx="pt.x" :cy="pt.y" r="3" fill="white" class="chart-line-diaper" stroke-width="2"/>
              <text :x="pt.x" :y="pt.y - 8" text-anchor="middle" font-size="10" fill="#6b7280">{{ pt.value }}</text>
            </g>
            <text v-for="(d, i) in trendData" :key="'dx'+i" :x="diaperPoints[i]?.x" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ d.date.slice(5) }}</text>
          </svg>
        </div>
        <div>
          <h4 class="text-sm font-semibold text-text-secondary mb-2">😴 每日睡眠 (分钟)</h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <polygon :points="sleepAreaPoints" class="chart-fill-primary" opacity="0.08"/>
            <polyline :points="sleepLinePoints" class="chart-line-primary" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <g v-for="(pt, i) in sleepPoints" :key="'sp'+i">
              <circle :cx="pt.x" :cy="pt.y" r="3" fill="white" class="chart-line-primary" stroke-width="2"/>
              <text :x="pt.x" :y="pt.y - 8" text-anchor="middle" font-size="10" fill="#6b7280">{{ pt.value }}</text>
            </g>
            <text v-for="(d, i) in trendData" :key="'sx'+i" :x="sleepPoints[i]?.x" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ d.date.slice(5) }}</text>
          </svg>
        </div>
        <div>
          <h4 class="text-sm font-semibold text-text-secondary mb-2">🌡️ 每日体温 (°C)</h4>
          <svg viewBox="0 0 340 170" class="w-full block">
            <polygon :points="tempAreaPoints" class="chart-fill-temperature" opacity="0.08"/>
            <polyline :points="tempLinePoints" class="chart-line-temperature" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <line x1="20" :y1="feverLineY" x2="320" :y2="feverLineY" stroke="#ef4444" stroke-width="1" stroke-dasharray="4,3" opacity="0.5"/>
            <g v-for="(pt, i) in tempPoints" :key="'tp'+i">
              <circle :cx="pt.x" :cy="pt.y" r="3" fill="white" class="chart-line-temperature" stroke-width="2"/>
              <text :x="pt.x" :y="pt.y - 8" text-anchor="middle" font-size="10" fill="#6b7280">{{ pt.value }}</text>
            </g>
            <text v-for="(d, i) in trendData" :key="'tx'+i" :x="tempPoints[i]?.x" y="158" text-anchor="middle" font-size="9" fill="#9ca3af">{{ d.date.slice(5) }}</text>
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

const dayOptions = [
  { label: '7天', value: 7 },
  { label: '30天', value: 30 },
]

const CHART = { padL: 20, padR: 20, padT: 25, padB: 35, svgW: 340, svgH: 170 }

function buildLineChart(getValue: (d: any) => number) {
  const data = trendData.value
  if (!data.length) return { points: [], line: '', area: '' }
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
  const line = points.map(p => `${p.x},${p.y}`).join(' ')
  const bY = padT + chartH
  const area = [`${points[0].x},${bY}`, ...points.map(p => `${p.x},${p.y}`), `${points[points.length - 1].x},${bY}`].join(' ')
  return { points, line, area }
}

const feedingChart = computed(() => buildLineChart(d => d.total_ml || 0))
const feedingPoints = computed(() => feedingChart.value.points)
const feedingLinePoints = computed(() => feedingChart.value.line)
const feedingAreaPoints = computed(() => feedingChart.value.area)

const diaperChart = computed(() => buildLineChart(d => d.diaper_count || 0))
const diaperPoints = computed(() => diaperChart.value.points)
const diaperLinePoints = computed(() => diaperChart.value.line)
const diaperAreaPoints = computed(() => diaperChart.value.area)

const sleepChart = computed(() => buildLineChart(d => d.sleep_duration_minutes || 0))
const sleepPoints = computed(() => sleepChart.value.points)
const sleepLinePoints = computed(() => sleepChart.value.line)
const sleepAreaPoints = computed(() => sleepChart.value.area)

const tempChart = computed(() => buildLineChart(d => d.temperature_avg || 0))
const tempPoints = computed(() => tempChart.value.points)
const tempLinePoints = computed(() => tempChart.value.line)
const tempAreaPoints = computed(() => tempChart.value.area)

const feverLineY = computed(() => {
  const { padT, padB, svgH } = CHART
  const chartH = svgH - padT - padB
  const values = trendData.value.map(d => d.temperature_avg || 0)
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
