<template>
  <div class="flex flex-col h-dvh bg-bg-main">
    <PullRefresh class="flex-1 min-h-0" content-class="px-4 py-4 space-y-4 pb-[calc(6.5rem+env(safe-area-inset-bottom))]"
      :refresh="refresh" @scroll="navScroll = $event">
      <template #header>
        <LargeTitleNav title="成长记录" inline-title="成长记录" :large="false">
          <template #actions>
            <button type="button" @click="openForm"
              class="inline-flex items-center gap-1 h-9 px-3.5 rounded-full bg-primary-fill text-white text-sm font-semibold btn-press">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M12 5v14m7-7H5"/></svg>
              记录
            </button>
          </template>
        </LargeTitleNav>
      </template>

      <div v-if="loading && !stats" class="py-16 flex justify-center">
        <ActivityIndicator :size="28" class="text-text-secondary" />
      </div>

      <template v-else-if="stats && !stats.empty">
        <!-- 最新测量 · WHO 百分位 -->
        <div class="bg-surface rounded-2xl p-4 shadow-card">
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-sm font-semibold text-text-secondary">最新测量</h2>
            <span class="text-xs text-text-secondary">{{ stats.age_months }} 月龄 · {{ stats.gender === 'male' ? '男宝' : '女宝' }}</span>
          </div>
          <div class="grid grid-cols-3 gap-2.5">
            <div v-for="m in metrics" :key="m.key" class="bg-bg-main rounded-xl p-3 text-center">
              <div class="text-xs text-text-secondary">{{ m.label }}</div>
              <div class="font-num text-lg font-bold text-text-primary mt-0.5">{{ m.value }}</div>
              <div class="mt-1.5 inline-flex items-center gap-1 text-[11px] font-semibold px-2 py-0.5 rounded-full"
                :class="pctClass(m.pct)">
                {{ m.pct > 0 ? 'P' + m.pct.toFixed(0) : '--' }}
              </div>
            </div>
          </div>
          <p class="text-[11px] text-text-secondary mt-3 leading-relaxed">
            百分位依据 WHO 儿童生长标准（0–24 月龄）计算，仅供参考，不能替代儿科医生评估。
          </p>
        </div>

        <!-- 趋势曲线 -->
        <div class="bg-surface rounded-2xl px-3 pt-3 pb-3 shadow-card">
          <div class="flex items-center justify-between px-1 mb-1">
            <h2 class="text-sm font-semibold text-text-secondary">成长曲线</h2>
            <Segmented v-model="metric" :options="metricOptions" compact />
          </div>
          <div v-if="chartSeries.length === 0" class="py-10">
            <EmptyState size="sm" icon="chart" title="暂无数据" subtitle="记录几次测量后即可看到趋势" />
          </div>
          <svg v-else :viewBox="`0 0 ${W} ${H}`" class="w-full" role="img" aria-label="成长曲线图">
            <!-- 网格 -->
            <line v-for="(t, i) in yTicks" :key="'g' + i" :x1="PAD_L" :x2="W - PAD_R" :y1="t.y" :y2="t.y"
              class="chart-grid" />
            <text v-for="(t, i) in yTicks" :key="'gt' + i" :x="PAD_L - 4" :y="t.y + 3" text-anchor="end"
              class="chart-axis-label" font-size="9">{{ t.label }}</text>
            <!-- 折线 -->
            <path :d="linePath" fill="none" :stroke="strokeColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
            <!-- 数据点 -->
            <circle v-for="(p, i) in chartSeries" :key="'p' + i" :cx="p.x" :cy="p.y" r="2.8" :fill="strokeColor" />
            <!-- X 轴标签 -->
            <text v-for="(p, i) in xLabels" :key="'x' + i" :x="p.x" :y="H - 4" text-anchor="middle"
              class="chart-axis-label" font-size="9">{{ p.label }}</text>
          </svg>
        </div>

        <!-- 历史记录 -->
        <div class="bg-surface rounded-2xl shadow-card overflow-hidden">
          <h2 class="text-sm font-semibold text-text-secondary px-4 pt-3 pb-1">历史记录</h2>
          <div v-for="g in listDesc" :key="g.id" class="px-4 py-3 border-t border-border-color/60 flex items-center gap-3">
            <div class="flex-1 min-w-0">
              <div class="text-sm font-medium text-text-primary">{{ g.measured_at }}</div>
              <div class="text-xs text-text-secondary mt-0.5">{{ detailOf(g) }}</div>
            </div>
            <button type="button" aria-label="删除此记录" @click="askDelete(g)"
              class="w-11 h-11 flex items-center justify-center text-danger/70 btn-press shrink-0">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M19 7l-.9 12a1.5 1.5 0 01-1.5 1.4H7.4A1.5 1.5 0 015.9 19L5 7m5 0V4.5A1.5 1.5 0 0111.5 3h1A1.5 1.5 0 0114 4.5V7m-9 0h14"/></svg>
            </button>
          </div>
        </div>
      </template>

      <EmptyState v-else icon="chart" title="还没有成长记录"
        subtitle="记录身高、体重、头围，自动生成 WHO 百分位与成长曲线">
        <button type="button" @click="openForm"
          class="inline-flex items-center gap-1.5 px-5 py-2.5 bg-primary-fill text-white text-sm font-semibold rounded-xl btn-press">
          添加测量
        </button>
      </EmptyState>
    </PullRefresh>

    <!-- 新增测量弹层 -->
    <Teleport to="body">
      <transition name="sheet-mask">
        <div v-if="formOpen" class="fixed inset-0 z-[90] flex items-end justify-center bg-black/35" @click.self="formOpen = false">
          <transition name="sheet-panel" appear>
            <div v-if="formOpen" ref="panelRef" tabindex="-1" role="dialog" aria-modal="true" aria-label="新增测量"
              class="w-full max-w-[480px] bg-surface rounded-t-2xl px-4 pt-3 pb-[calc(1rem+env(safe-area-inset-bottom))] outline-none">
              <div class="flex justify-center mb-2"><span class="w-9 h-1 rounded-full bg-border-color"></span></div>
              <h2 class="text-center text-[17px] font-semibold text-text-primary mb-3">新增测量</h2>

              <div class="space-y-3">
                <div>
                  <label class="text-sm text-text-secondary block mb-1.5">测量日期</label>
                  <input v-model="form.measured_at" type="date"
                    class="w-full px-4 py-3 bg-bg-secondary border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none" />
                </div>
                <div class="grid grid-cols-3 gap-2.5">
                  <div>
                    <label class="text-xs text-text-secondary block mb-1.5">体重 kg</label>
                    <input v-model.number="form.weight_kg" type="number" inputmode="decimal" step="0.01" min="0"
                      class="w-full px-3 py-2.5 bg-bg-secondary border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none" />
                  </div>
                  <div>
                    <label class="text-xs text-text-secondary block mb-1.5">身高 cm</label>
                    <input v-model.number="form.height_cm" type="number" inputmode="decimal" step="0.1" min="0"
                      class="w-full px-3 py-2.5 bg-bg-secondary border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none" />
                  </div>
                  <div>
                    <label class="text-xs text-text-secondary block mb-1.5">头围 cm</label>
                    <input v-model.number="form.head_cm" type="number" inputmode="decimal" step="0.1" min="0"
                      class="w-full px-3 py-2.5 bg-bg-secondary border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none" />
                  </div>
                </div>
                <div v-if="formError" class="bg-danger-light text-danger text-sm px-4 py-2 rounded-xl text-center">{{ formError }}</div>
              </div>

              <div class="flex gap-2 mt-4">
                <button type="button" @click="formOpen = false"
                  class="flex-1 py-3 bg-bg-secondary text-text-primary font-medium rounded-xl btn-press">取消</button>
                <button type="button" @click="submit" :disabled="submitting"
                  class="flex-1 py-3 bg-primary-fill text-white font-semibold rounded-xl btn-press disabled:opacity-50 flex items-center justify-center gap-2">
                  <ActivityIndicator v-if="submitting" :size="18" class="text-white" />
                  <span>{{ submitting ? '保存中...' : '保存' }}</span>
                </button>
              </div>
            </div>
          </transition>
        </div>
      </transition>
    </Teleport>

    <ConfirmSheet :open="!!toDelete" title="删除测量记录" message="确定要删除这条成长记录吗？"
      confirm-text="删除" :loading="deleting" @confirm="doDelete" @cancel="toDelete = null" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { babyAPI, GrowthRecord, GrowthStats } from '@/api'
import PullRefresh from '@/components/PullRefresh.vue'
import LargeTitleNav from '@/components/LargeTitleNav.vue'
import EmptyState from '@/components/EmptyState.vue'
import Segmented from '@/components/Segmented.vue'
import ActivityIndicator from '@/components/ActivityIndicator.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'
import { hapticSuccess, hapticError } from '@/utils/haptic'

const router = useRouter()
const app = useAppStore()
const navScroll = ref(0)
const loading = ref(false)
const list = ref<GrowthRecord[]>([])
const stats = ref<GrowthStats | null>(null)

const metric = ref<'weight' | 'height' | 'head'>('weight')
const metricOptions = [
  { label: '体重', value: 'weight' },
  { label: '身高', value: 'height' },
  { label: '头围', value: 'head' },
]

const formOpen = ref(false)
const submitting = ref(false)
const formError = ref('')
const panelRef = ref<HTMLElement | null>(null)
const toDelete = ref<GrowthRecord | null>(null)
const deleting = ref(false)
const form = ref({ measured_at: '', weight_kg: 0, height_cm: 0, head_cm: 0 })

const listDesc = computed(() => [...list.value].reverse())

function fmt(v: number) {
  return v > 0 ? String(Number(v.toFixed(2))) : '--'
}

const metrics = computed(() => {
  const s = stats.value
  if (!s) return []
  return [
    { key: 'weight', label: '体重 kg', value: fmt(s.weight_kg || 0), pct: s.weight_pct || 0 },
    { key: 'height', label: '身高 cm', value: fmt(s.height_cm || 0), pct: s.height_pct || 0 },
    { key: 'head', label: '头围 cm', value: fmt(s.head_cm || 0), pct: s.head_pct || 0 },
  ]
})

function pctClass(p: number) {
  if (p <= 0) return 'bg-muted text-text-secondary'
  if (p < 3 || p > 97) return 'bg-danger/10 text-danger'
  if (p < 15 || p > 85) return 'bg-warning/15 text-warning-deep'
  return 'bg-success/15 text-success'
}

function detailOf(g: GrowthRecord) {
  const parts: string[] = []
  if (g.weight_kg > 0) parts.push(`体重 ${fmt(g.weight_kg)}kg`)
  if (g.height_cm > 0) parts.push(`身高 ${fmt(g.height_cm)}cm`)
  if (g.head_cm > 0) parts.push(`头围 ${fmt(g.head_cm)}cm`)
  return parts.join(' · ') || '—'
}

// ── 图表 ──
const W = 340, H = 210, PAD_L = 30, PAD_R = 12, PAD_T = 14, PAD_B = 26

const series = computed(() => {
  const key = metric.value === 'weight' ? 'weight_kg' : metric.value === 'height' ? 'height_cm' : 'head_cm'
  return list.value
    .map(g => ({ date: g.measured_at, v: Number((g as any)[key]) || 0 }))
    .filter(p => p.v > 0)
})

const bounds = computed(() => {
  const vs = series.value.map(p => p.v)
  if (!vs.length) return { min: 0, max: 1 }
  let min = Math.min(...vs), max = Math.max(...vs)
  const pad = (max - min) * 0.15 || Math.max(1, max * 0.1)
  min -= pad; max += pad
  if (min < 0) min = 0
  return { min, max }
})

function xAt(i: number) {
  const n = series.value.length
  if (n <= 1) return PAD_L + (W - PAD_L - PAD_R) / 2
  return PAD_L + (i / (n - 1)) * (W - PAD_L - PAD_R)
}
function yAt(v: number) {
  const { min, max } = bounds.value
  const span = max - min || 1
  return PAD_T + (1 - (v - min) / span) * (H - PAD_T - PAD_B)
}

const chartSeries = computed(() => series.value.map((p, i) => ({ x: xAt(i), y: yAt(p.v) })))

const linePath = computed(() => {
  const pts = chartSeries.value
  if (!pts.length) return ''
  return pts.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(1)} ${p.y.toFixed(1)}`).join(' ')
})

const yTicks = computed(() => {
  const { min, max } = bounds.value
  const n = 4
  const out: { y: number, label: string }[] = []
  for (let i = 0; i <= n; i++) {
    const v = min + (max - min) * (i / n)
    out.push({ y: yAt(v), label: v.toFixed(1) })
  }
  return out
})

const xLabels = computed(() => {
  const n = series.value.length
  if (!n) return []
  const step = Math.max(1, Math.ceil(n / 4))
  const out: { x: number, label: string }[] = []
  for (let i = 0; i < n; i += step) {
    const d = series.value[i].date
    const m = /(\d{4})-(\d{2})-(\d{2})/.exec(d)
    out.push({ x: xAt(i), label: m ? `${+m[2]}/${+m[3]}` : d })
  }
  return out
})

const strokeColor = computed(() => {
  const key = metric.value === 'weight' ? '--chart-primary' : metric.value === 'height' ? '--chart-sleep' : '--chart-outdoor'
  return `var(${key})`
})

// ── 数据 ──
async function refresh() {
  await load()
}

async function load() {
  // 单独直达 /growth 时 app.babies 可能尚未加载，先补齐
  if (app.babies.length === 0) {
    try { await app.loadBabies() } catch { /* ignore */ }
  }
  const baby = app.currentBaby
  if (!baby) { loading.value = false; return }
  loading.value = true
  try {
    const [l, s] = await Promise.all([babyAPI.growth(baby.id), babyAPI.growthStats(baby.id)])
    list.value = l.data || []
    stats.value = s.data || null
  } catch {
    app.showToast('加载失败', 'error')
  } finally {
    loading.value = false
  }
}

function openForm() {
  const d = new Date()
  const p2 = (n: number) => String(n).padStart(2, '0')
  form.value = {
    measured_at: `${d.getFullYear()}-${p2(d.getMonth() + 1)}-${p2(d.getDate())}`,
    weight_kg: 0, height_cm: 0, head_cm: 0,
  }
  formError.value = ''
  formOpen.value = true
  nextTick(() => panelRef.value?.focus())
}

async function submit() {
  formError.value = ''
  if (!form.value.measured_at) { formError.value = '请选择测量日期'; return }
  if (form.value.weight_kg <= 0 && form.value.height_cm <= 0 && form.value.head_cm <= 0) {
    formError.value = '至少填写一项测量值'; return
  }
  const baby = app.currentBaby
  if (!baby) return
  submitting.value = true
  try {
    await babyAPI.createGrowth(baby.id, {
      measured_at: form.value.measured_at,
      weight_kg: form.value.weight_kg || 0,
      height_cm: form.value.height_cm || 0,
      head_cm: form.value.head_cm || 0,
    })
    hapticSuccess()
    formOpen.value = false
    await load()
    app.showToast('已保存', 'success')
  } catch (e: any) {
    hapticError()
    formError.value = e.response?.data?.error || '保存失败'
  } finally {
    submitting.value = false
  }
}

function askDelete(g: GrowthRecord) {
  toDelete.value = g
}

async function doDelete() {
  if (!toDelete.value) return
  deleting.value = true
  try {
    await babyAPI.deleteGrowth(toDelete.value.id)
    toDelete.value = null
    await load()
    app.showToast('已删除', 'success')
  } catch {
    app.showToast('删除失败', 'error')
  } finally {
    deleting.value = false
  }
}

onMounted(load)
</script>
