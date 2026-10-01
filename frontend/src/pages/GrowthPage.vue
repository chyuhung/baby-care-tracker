<template>
  <div class="flex flex-col h-dvh bg-bg-main">
    <PullRefresh class="flex-1 min-h-0" content-class="px-4 py-4 space-y-4 pb-[calc(6.5rem+env(safe-area-inset-bottom))]"
      :refresh="refresh">
      <template #header>
        <NavBar title="成长记录">
          <template #actions>
            <button type="button" @click="openForm" aria-label="记录测量"
              class="-mr-1 flex h-11 w-11 items-center justify-center rounded-lg text-primary-deep btn-press">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14m7-7H5"/></svg>
            </button>
          </template>
        </NavBar>
      </template>

      <div v-if="loading && !stats" class="py-16 flex justify-center">
        <ActivityIndicator :size="28" class="text-text-secondary" />
      </div>

      <template v-else-if="stats && !stats.empty">
        <!-- 最新测量 · 标准百分位 -->
        <div class="bg-surface rounded-2xl p-4 shadow-card">
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-sm font-semibold text-text-secondary">最新测量</h2>
            <span class="text-xs text-text-secondary">{{ latestSubtitle }}</span>
          </div>
          <div class="grid grid-cols-3 divide-x divide-border-color/60">
            <div v-for="m in metrics" :key="m.key" class="px-3 py-1 text-center">
              <div class="text-xs text-text-secondary">{{ m.label }}</div>
              <div class="font-num text-xl font-bold text-text-primary mt-0.5">{{ m.value }}</div>
              <template v-if="m.meta">
                <div class="mt-1.5 inline-flex items-center text-[11px] font-semibold px-2 py-0.5 rounded-full"
                  :class="m.meta.pill">
                  {{ m.meta.label }}
                </div>
                <div class="mt-0.5 text-[10px] text-text-secondary font-num">{{ m.meta.range }}</div>
              </template>
              <div v-else class="mt-1.5 text-[11px] text-text-secondary">--</div>
            </div>
          </div>
          <p class="text-[11px] text-text-secondary mt-3 leading-relaxed">
            分档依据《7 岁以下儿童生长标准》(WS/T 423-2022) 表1 五级评价：≥P97 为上，P75–P97 为中上，
            P25–P75 为中，P3–P25 为中下，&lt;P3 为下；正常范围 P3–P97。仅供参考，不能替代儿科医生评估。
          </p>
        </div>

        <!-- 趋势曲线 -->
        <div class="bg-surface rounded-2xl px-3 pt-3 pb-3 shadow-card">
          <div class="flex items-center justify-between px-1 mb-1">
            <h2 class="text-sm font-semibold text-text-secondary">成长曲线</h2>
            <Segmented v-model="metric" :options="metricOptions" compact />
          </div>
          <!-- 参考带图例：图内只填 P3–P97 三段（正常范围），上/下两档以红点表示实测点落在范围外 -->
          <div class="flex flex-wrap items-center justify-center gap-x-3 gap-y-1 px-1 mb-1.5 text-[11px] text-text-secondary">
            <span class="inline-flex items-center gap-1">
              <i class="w-2.5 h-2.5 rounded-[3px]" style="background: rgb(var(--success-deep) / 0.1)"></i>中下 P3–P25
            </span>
            <span class="inline-flex items-center gap-1">
              <i class="w-2.5 h-2.5 rounded-[3px]" style="background: rgb(var(--success-deep) / 0.22)"></i>中 P25–P75
            </span>
            <span class="inline-flex items-center gap-1">
              <i class="w-2.5 h-2.5 rounded-[3px]" style="background: rgb(var(--success-deep) / 0.1)"></i>中上 P75–P97
            </span>
            <span class="inline-flex items-center gap-1">
              <i class="w-2 h-2 rounded-full" style="background: rgb(var(--danger-deep) / 0.9)"></i>超出范围
            </span>
            <span class="inline-flex items-center gap-1">
              <i class="w-3 h-0 border-t border-dashed" style="border-color: rgb(var(--text-secondary) / 0.7)"></i>P50 中位
            </span>
            <span class="inline-flex items-center gap-1">
              <i class="w-2 h-2 rounded-full" :style="{ background: strokeColor }"></i>实测
            </span>
          </div>
          <div v-if="!refMetric && chartSeries.length === 0" class="py-10">
            <EmptyState size="sm" icon="chart" title="暂无数据" subtitle="记录几次测量后即可看到趋势" />
          </div>
          <svg v-else :viewBox="`0 0 ${W} ${H}`" class="w-full" role="img" aria-label="成长曲线图">
            <!-- 正常范围 P3–P97 的三段（同色系由浅到深，中最深）。P3 以下 / P97 以上不填色：
                 线性数值轴上该区域占满 65-90% 画面，填红会严重误导；改由实测点变红表达。 -->
            <template v-if="zonePaths">
              <path :d="zonePaths.midLow" style="fill: rgb(var(--success-deep) / 0.1)" />
              <path :d="zonePaths.mid" style="fill: rgb(var(--success-deep) / 0.22)" />
              <path :d="zonePaths.midHigh" style="fill: rgb(var(--success-deep) / 0.1)" />
            </template>
            <!-- 网格（横向实线 + 纵向辅助虚线） -->
            <line v-for="(t, i) in yTicks" :key="'g' + i" :x1="PAD_L" :x2="W - PAD_R" :y1="t.y" :y2="t.y"
              class="chart-grid" />
            <line v-for="(t, i) in xTicks" :key="'gv' + i" :x1="t.x" :x2="t.x" :y1="PAD_T" :y2="H - PAD_B"
              stroke-dasharray="2,3" style="stroke: rgb(var(--text-secondary) / 0.18)" />
            <!-- 轴边框（对齐其他图表） -->
            <line :x1="PAD_L" :x2="PAD_L" :y1="PAD_T" :y2="H - PAD_B" stroke="var(--chart-line)" stroke-width="1" />
            <line :x1="W - PAD_R" :x2="W - PAD_R" :y1="PAD_T" :y2="H - PAD_B" stroke="var(--chart-line)" stroke-width="1" />
            <line :x1="PAD_L" :x2="W - PAD_R" :y1="H - PAD_B" :y2="H - PAD_B" stroke="var(--chart-line)" stroke-width="1" />
            <text v-for="(t, i) in yTicks" :key="'gt' + i" :x="PAD_L - 4" :y="t.y + 3" text-anchor="end"
              class="chart-axis-label" font-size="9">{{ t.label }}</text>
            <!-- 参考中位线 -->
            <path v-if="refPaths" :d="refPaths.p50" fill="none" style="stroke: rgb(var(--text-secondary) / 0.7)"
              stroke-width="1.3" stroke-dasharray="4,3" stroke-linecap="round" />
            <!-- 实测折线 -->
            <path :d="linePath" fill="none" :stroke="strokeColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
            <!-- 数据点：落在 P3–P97 之外时标红（上/下 两档） -->
            <circle v-for="(p, i) in chartSeries" :key="'p' + i" :cx="p.x" :cy="p.y" r="2.8"
              :fill="p.outOfRange ? 'rgb(var(--danger-deep))' : strokeColor" />
            <!-- X 轴标签（月龄） -->
            <text v-for="(t, i) in xTicks" :key="'x' + i" :x="t.x" :y="H - 4" text-anchor="middle"
              class="chart-axis-label" font-size="9">{{ t.label }}</text>
          </svg>
          <p class="text-[11px] text-text-secondary mt-2 leading-relaxed px-1">
            绿带为正常范围 P3–P97（约 94% 儿童），按表1 五级评价分为中下 P3–P25、中 P25–P75、中上 P75–P97；
            &lt;P3 为「下」、≥P97 为「上」，两档落在正常范围之外，图上以红点标出。横轴为月龄，纵轴为实际数值可直接读数。
            2 岁前为身长、2 岁后为身高，头围参考至 3 岁。仅供参考，不能替代儿科医生评估。
          </p>
        </div>

        <!-- 历史记录（整行 tap=编辑，长按=上下文菜单删除，与记录卡同手势口径） -->
        <div class="bg-surface rounded-2xl shadow-card overflow-hidden">
          <h2 class="text-sm font-semibold text-text-secondary px-4 pt-3 pb-1">历史记录</h2>
          <div v-for="g in listDesc" :key="g.id" role="button" tabindex="0" aria-label="编辑此记录，长按可删除"
            class="px-4 py-3 border-t border-border-color/60 btn-press"
            @click="onRowClick(g)" @keydown.enter="openEdit(g)"
            @touchstart.passive="rowTouchStart(g, $event)" @touchmove="rowTouchMove($event)"
            @touchend="rowTouchEnd" @touchcancel="rowTouchEnd">
            <div class="flex items-center gap-2">
              <div class="min-w-0 flex-1">
                <div class="text-sm font-medium text-text-primary">{{ dateLabelOf(g) }}</div>
                <div class="text-xs text-text-secondary mt-0.5 truncate">{{ detailOf(g) }}</div>
              </div>
              <svg class="w-4 h-4 shrink-0 text-text-secondary/50" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
            </div>
          </div>
        </div>
      </template>

      <EmptyState v-else icon="chart" title="还没有成长记录"
        subtitle="记录身高、体重、头围，自动生成百分位与成长曲线">
        <button type="button" @click="openForm"
          class="inline-flex items-center gap-1.5 px-5 py-2.5 bg-primary/10 text-primary-deep text-sm font-semibold rounded-xl btn-press">
          添加测量
        </button>
      </EmptyState>
    </PullRefresh>

    <!-- 新增测量弹层 -->
    <Teleport to="body">
      <transition name="sheet-mask">
        <div v-if="formOpen" class="fixed inset-0 z-[90] flex items-end justify-center bg-black/35" @click.self="closeForm">
          <transition name="sheet-panel" appear>
            <div v-if="formOpen" ref="panelRef" tabindex="-1" role="dialog" aria-modal="true" aria-label="新增测量"
              class="w-full max-w-[480px] bg-surface rounded-t-2xl px-4 pt-3 pb-[calc(1rem+env(safe-area-inset-bottom))] outline-none">
              <div class="flex justify-center mb-2"><span class="w-9 h-1 rounded-full bg-border-color"></span></div>
              <h2 class="text-center text-[17px] font-semibold text-text-primary mb-3">{{ editingId ? '编辑测量' : '新增测量' }}</h2>

              <div class="space-y-3">
                <div>
                  <label class="text-sm text-text-secondary block mb-1.5">测量日期</label>
                  <DateTimeField v-model="form.measured_at" title="测量日期" aria-label="选择测量日期" date-only />
                </div>
                <div class="grid grid-cols-3 gap-2.5">
                  <div>
                    <label class="text-xs text-text-secondary block mb-1.5">身高 cm</label>
                    <input v-model.number="form.height_cm" type="number" inputmode="decimal" step="0.1" min="0"
                      class="w-full px-3 py-2.5 bg-muted border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none" />
                  </div>
                  <div>
                    <label class="text-xs text-text-secondary block mb-1.5">体重 kg</label>
                    <input v-model.number="form.weight_kg" type="number" inputmode="decimal" step="0.01" min="0"
                      class="w-full px-3 py-2.5 bg-muted border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none" />
                  </div>
                  <div>
                    <label class="text-xs text-text-secondary block mb-1.5">头围 cm</label>
                    <input v-model.number="form.head_cm" type="number" inputmode="decimal" step="0.1" min="0"
                      class="w-full px-3 py-2.5 bg-muted border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none" />
                  </div>
                </div>
                <div v-if="formError" class="bg-danger-light text-danger text-sm px-4 py-2 rounded-xl text-center">{{ formError }}</div>
              </div>

              <div class="flex gap-2 mt-4">
                <button type="button" @click="closeForm"
                  class="flex-1 py-3 bg-muted text-text-primary font-medium rounded-xl btn-press">取消</button>
                <button type="button" @click="submit" :disabled="submitting"
                  class="flex-1 py-3 bg-primary/10 text-primary-deep font-semibold rounded-xl btn-press disabled:opacity-50 flex items-center justify-center gap-2">
                  <ActivityIndicator v-if="submitting" :size="18" class="text-primary-deep" />
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

    <!-- 长按上下文菜单（与记录卡一致：编辑 / 删除） -->
    <ContextMenu :open="contextOpen" :title="contextGrowth ? dateLabelOf(contextGrowth) : ''"
      :subtitle="contextGrowth ? detailOf(contextGrowth) : ''" emoji="📏" :actions="contextActions"
      @update:open="contextOpen = $event" @select="onContextSelect" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { babyAPI, GrowthRecord, GrowthStats, GrowthReference, writeErrorMessage } from '@/api'
import PullRefresh from '@/components/PullRefresh.vue'
import NavBar from '@/components/NavBar.vue'
import DateTimeField from '@/components/DateTimeField.vue'
import EmptyState from '@/components/EmptyState.vue'
import Segmented from '@/components/Segmented.vue'
import ActivityIndicator from '@/components/ActivityIndicator.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'
import ContextMenu from '@/components/ContextMenu.vue'
import { parseLocalDate, formatDateCN, measureAgeText } from '@/utils'

const router = useRouter()
const app = useAppStore()
const loading = ref(false)
const list = ref<GrowthRecord[]>([])
const stats = ref<GrowthStats | null>(null)
const reference = ref<GrowthReference | null>(null)

const metric = ref<'weight' | 'height' | 'head'>('height')
const metricOptions = [
  { label: '身高', value: 'height' },
  { label: '体重', value: 'weight' },
  { label: '头围', value: 'head' },
]

const formOpen = ref(false)
const submitting = ref(false)
const formError = ref('')
const editingId = ref<number | null>(null)
const panelRef = ref<HTMLElement | null>(null)
const toDelete = ref<GrowthRecord | null>(null)
const deleting = ref(false)
const form = ref({ measured_at: '', weight_kg: '' as number | '', height_cm: '' as number | '', head_cm: '' as number | '' })

const listDesc = computed(() => [...list.value].reverse())

function fmt(v: number) {
  return v > 0 ? String(Number(v.toFixed(2))) : '--'
}

const metrics = computed(() => {
  const s = stats.value
  if (!s) return []
  // meta 预先解析好，模板里无需重复调用 gradeMeta（也省去非空断言）
  return [
    { key: 'height', label: '身高 cm', value: fmt(s.height_cm || 0), meta: gradeMeta(s.height_grade || '') },
    { key: 'weight', label: '体重 kg', value: fmt(s.weight_kg || 0), meta: gradeMeta(s.weight_grade || '') },
    { key: 'head', label: '头围 cm', value: fmt(s.head_cm || 0), meta: gradeMeta(s.head_grade || '') },
  ]
})

// 最新测量副标题：月龄为「测量时」的月龄（后端 monthsBetween(birth, measured)），
// 不是当前月龄；档案性别为「保密」时如实显示，不再被静默显示成女宝
const latestSubtitle = computed(() => {
  const s = stats.value
  if (!s) return ''
  const g = s.gender_label === 'male' ? '男宝' : s.gender_label === 'female' ? '女宝' : '保密'
  const gender = s.gender_fallback ? `${g}（暂按女宝标准）` : g
  return `${s.age_months ?? 0} 月龄 · ${gender}`
})

/**
 * WS/T 423-2022 表1 五级评价。判定完全采用后端返回的 *_grade
 * （直接与 P3/P25/P75/P97 四个边界比较，不经插值），此处只负责配色与文案。
 *
 * 正常范围是 P3–P97（约 94% 儿童），中下/中/中上三档都在正常范围内；
 * 仅「下」(<P3) 与「上」(≥P97) 落在正常范围之外。
 * 对称口径：中下与中上同色，不因偏高/偏低而赋予褒贬色。
 */
interface GradeMeta { label: string; range: string; pill: string }

const GRADE_META: Record<string, GradeMeta> = {
  上: { label: '上', range: '≥P97', pill: 'bg-danger/10 text-danger' },
  中上: { label: '中上', range: 'P75–P97', pill: 'bg-muted text-text-secondary' },
  中: { label: '中', range: 'P25–P75', pill: 'bg-success/10 text-success' },
  中下: { label: '中下', range: 'P3–P25', pill: 'bg-muted text-text-secondary' },
  下: { label: '下', range: '<P3', pill: 'bg-danger/10 text-danger' },
}

function gradeMeta(g: string): GradeMeta | null {
  return GRADE_META[g] || null
}

function detailOf(g: GrowthRecord) {
  const parts: string[] = []
  if (g.height_cm > 0) parts.push(`身高 ${fmt(g.height_cm)}cm`)
  if (g.weight_kg > 0) parts.push(`体重 ${fmt(g.weight_kg)}kg`)
  if (g.head_cm > 0) parts.push(`头围 ${fmt(g.head_cm)}cm`)
  return parts.join(' · ') || '—'
}

// 历史记录日期：2026年9月1日（3月5天）；宝宝无出生日期或测量早于出生时只显示日期
function dateLabelOf(g: GrowthRecord) {
  const date = formatDateCN(g.measured_at)
  const age = measureAgeText(app.currentBaby?.birth_date || '', g.measured_at)
  return age ? `${date}（${age}）` : date
}

// ── 图表（月龄轴 + WS/T 423-2022 参考曲线，医院图三色风格）──
const W = 340, H = 210, PAD_L = 30, PAD_R = 12, PAD_T = 14, PAD_B = 26

type PctKey = 'p3' | 'p25' | 'p50' | 'p75' | 'p97'

function monthOf(dateStr: string): number {
  const birth = parseLocalDate(app.currentBaby?.birth_date || '')
  const d = parseLocalDate(dateStr)
  if (!birth || !d) return 0
  const days = (d.getTime() - birth.getTime()) / 86400000
  return Math.max(0, days / 30.4375)
}

const series = computed(() => {
  const key = metric.value === 'weight' ? 'weight_kg' : metric.value === 'height' ? 'height_cm' : 'head_cm'
  return list.value
    .map(g => ({ date: g.measured_at, v: Number((g as any)[key]) || 0, month: monthOf(g.measured_at) }))
    .filter(p => p.v > 0)
})

const refMetric = computed(() => {
  const r = reference.value
  if (!r) return null
  return metric.value === 'weight' ? r.weight : metric.value === 'height' ? r.height : r.head
})

// X 轴域：0..max(最大月龄, 12)（至少 12 月窗口便于观察）
const xMax = computed(() => {
  const maxM = series.value.length ? Math.max(...series.value.map(p => p.month)) : 0
  return Math.max(maxM, 12)
})

// 可见参考点（截到 x 轴域；头围标准仅至 3 岁）
const visibleRef = computed(() => {
  const rm = refMetric.value
  if (!rm) return null
  const pts = rm.points.filter(p => p.month <= xMax.value)
  return pts.length > 1 ? pts : null
})

// 线性 Y 值域：取「可见参考带的 P3/P97」与「实测值」的并集，上下各留 6% 余量。
// 留白的作用是让范围外的实测点（上/下两档）仍留在绘图区内可见，而不是被裁掉。
const bounds = computed(() => {
  const vs = series.value.map(p => p.v)
  const vis = visibleRef.value
  if (vis) for (const p of vis) { vs.push(p.p3, p.p97) }
  if (!vs.length) return { min: 0, max: 1 }
  const lo = Math.min(...vs), hi = Math.max(...vs)
  const pad = (hi - lo) * 0.06 || Math.max(1, hi * 0.06)
  return { min: Math.max(0, lo - pad), max: hi + pad }
})

function xAt(month: number) {
  const plot = W - PAD_L - PAD_R
  return PAD_L + (month / (xMax.value || 1)) * plot
}

/**
 * 某月龄处的参考带边界值（用于判断实测点是否落在 P3–P97 之外）。
 * 保留线性插值：参考点密度为 0-11 月逐月、其后每 3 月一行，
 * 实测月龄常落在两行之间，需要插值才能取到该月龄的边界。
 */
function refAt(month: number) {
  const vis = visibleRef.value
  if (!vis || !vis.length) return null
  // 参考点按月递增，找到区间后线性插值；超出末点则沿用末点（此时外侧已是留白）
  if (month <= vis[0].month) return vis[0]
  if (month >= vis[vis.length - 1].month) return vis[vis.length - 1]
  for (let i = 0; i < vis.length - 1; i++) {
    const a = vis[i], b = vis[i + 1]
    if (month >= a.month && month <= b.month) {
      const t = b.month === a.month ? 0 : (month - a.month) / (b.month - a.month)
      const mix = (k: PctKey) => a[k] + (b[k] - a[k]) * t
      return { p3: mix('p3'), p25: mix('p25'), p50: mix('p50'), p75: mix('p75'), p97: mix('p97') }
    }
  }
  return vis[vis.length - 1]
}

/**
 * Y 轴 = 线性真实数值轴（kg / cm），可直接按刻度读数。
 *
 * 曾用「百分位锚点弯曲」（p3/p25/p50/p75/p97 重映射到均分位置）把中档拉宽到 50%，
 * 现已删除：那个做法的唯一目的是让「正常区」变宽，而它之所以显得窄，
 * 是因为当时把正常区误定义为 P25–P75。表1 的正常范围是 P3–P97，
 * 在线性数值轴上已占据参考带全域，色带宽度由数据本身决定，无需拉伸。
 */
function yAt(v: number) {
  const plot = H - PAD_T - PAD_B
  const { min, max } = bounds.value
  const span = max - min || 1
  return PAD_T + (1 - (v - min) / span) * plot
}

/**
 * 实测点是否落在该月龄的 P3–P97 正常范围之外（对应表1 的「下」/「上」两档）。
 * 边界用 `>=` / `<`，与后端 Grade() 完全对齐：P97 本身即判「上」，
 * 若这里写 `> r.p97`，恰好落在 P97 的点会出现「红色胶囊 + 正常色点」的自相矛盾。
 */
function outOfRangeAt(v: number, month: number): boolean {
  const r = refAt(month)
  if (!r) return false
  return v < r.p3 || v >= r.p97
}

const chartSeries = computed(() =>
  series.value.map(p => ({
    x: xAt(p.month),
    y: yAt(p.v),
    outOfRange: outOfRangeAt(p.v, p.month),
  })),
)

const linePath = computed(() => {
  const pts = chartSeries.value
  if (!pts.length) return ''
  return pts.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(1)} ${p.y.toFixed(1)}`).join(' ')
})

// Y 轴刻度：线性数值轴，取「整齐」的整数步长（体重 0.5/1/2 kg，身高 5/10 cm），
// 标签即该处的真实数值，可直接读数
const yTicks = computed(() => {
  const { min, max } = bounds.value
  const span = max - min
  if (!(span > 0)) return []
  // 目标 5-6 条刻度线。步长从 1/2/2.5/5/10 × 10^n 中取「最接近 span/5」的一个——
  // 必须取最接近而非向上取整：身高量程约 52cm 时 span/5=10.4，向上取整会跳到 20，
  // 只剩 2 条刻度线（实测过），取最接近则得 10 → 6 条。
  const rawStep = span / 5
  const mag = Math.pow(10, Math.floor(Math.log10(rawStep)))
  const candidates = [1, 2, 2.5, 5, 10].map(m => m * mag)
  const step = candidates.reduce((best, c) =>
    Math.abs(c - rawStep) < Math.abs(best - rawStep) ? c : best,
  )
  // 2.5 这类步长需要一位小数
  const decimals = step < 1 || step % 1 !== 0 ? 1 : 0
  const out: { y: number; label: string }[] = []
  // 从 min 向上取整到步长整数倍，避免首条刻度贴在轴外
  for (let v = Math.ceil(min / step) * step; v <= max + 1e-9; v += step) {
    out.push({ y: yAt(v), label: v.toFixed(decimals) })
  }
  return out
})

// X 轴刻度（月龄）
const xTicks = computed(() => {
  const step = xMax.value <= 13 ? 3 : xMax.value <= 37 ? 6 : 12
  const out: { x: number, label: string }[] = []
  for (let m = 0; m <= xMax.value + 0.01; m += step) {
    out.push({ x: xAt(m), label: m === 0 ? '0' : m % 12 === 0 ? `${m / 12}岁` : `${m}月` })
  }
  return out
})

// 参考中位线（只留 P50 中位虚线；P3/P25/P75/P97 的边界语义由色带承担）
const refPaths = computed<Record<string, string> | null>(() => {
  const vis = visibleRef.value
  if (!vis) return null
  const out: Record<string, string> = {}
  for (const key of ['p50'] as PctKey[]) {
    out[key] = vis.map((p, i) => `${i === 0 ? 'M' : 'L'}${xAt(p.month).toFixed(1)} ${yAt(p[key]).toFixed(1)}`).join(' ')
  }
  return out
})

/**
 * 正常范围 P3–P97 的三段，对应表1 五级评价中的中下 / 中 / 中上。
 *
 * 三段同色系（浅 → 深 → 浅，中最深）：中下与中上刻意同色，因为标准对偏高与偏低
 * 是对称表述，不应让任何一档看起来更「好」或更「差」。
 * P3 以下与 P97 以上不填色——线性数值轴上这两块占满 65-90% 画面，填红会造成
 * 「大部分区域都异常」的错觉；范围外的实测点改用红点表达（见 chartSeries.outOfRange）。
 */
const zonePaths = computed(() => {
  const vis = visibleRef.value
  if (!vis) return null
  const fwd = (key: PctKey) => vis.map((p, i) => `${i === 0 ? 'M' : 'L'}${xAt(p.month).toFixed(1)} ${yAt(p[key]).toFixed(1)}`).join(' ')
  const back = (key: PctKey) => vis.slice().reverse().map(p => `L${xAt(p.month).toFixed(1)} ${yAt(p[key]).toFixed(1)}`).join(' ')
  return {
    midHigh: `${fwd('p97')} ${back('p75')} Z`,
    mid: `${fwd('p75')} ${back('p25')} Z`,
    midLow: `${fwd('p25')} ${back('p3')} Z`,
  }
})

// 身高用主题色（默认展示项），体重/头围各取一色，切段时颜色稳定不跳
const strokeColor = computed(() => {
  const key = metric.value === 'height' ? '--chart-primary' : metric.value === 'weight' ? '--chart-sleep' : '--chart-outdoor'
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
    const [l, s, r] = await Promise.all([
      babyAPI.growth(baby.id),
      babyAPI.growthStats(baby.id),
      babyAPI.growthReference(baby.id),
    ])
    list.value = l.data || []
    stats.value = s.data || null
    reference.value = r.data || null
  } catch {
    app.showToast('加载失败', 'error')
  } finally {
    loading.value = false
  }
}

// Date → YYYY-MM-DD（本地日历日，补零）
function ymd(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

function openForm() {
  form.value = {
    measured_at: ymd(new Date()),
    weight_kg: '', height_cm: '', head_cm: '',
  }
  editingId.value = null
  formError.value = ''
  formOpen.value = true
  nextTick(() => panelRef.value?.focus())
}

function openEdit(g: GrowthRecord) {
  const d = parseLocalDate(g.measured_at)
  form.value = {
    measured_at: d ? ymd(d) : '',
    weight_kg: g.weight_kg > 0 ? g.weight_kg : '',
    height_cm: g.height_cm > 0 ? g.height_cm : '',
    head_cm: g.head_cm > 0 ? g.head_cm : '',
  }
  editingId.value = g.id
  formError.value = ''
  formOpen.value = true
  nextTick(() => panelRef.value?.focus())
}

function closeForm() {
  formOpen.value = false
  editingId.value = null
}

// 空串/NaN → 0（0 表示该项未测），数字原样返回
function numOf(v: number | ''): number {
  return typeof v === 'number' && isFinite(v) ? v : 0
}

async function submit() {
  formError.value = ''
  if (!form.value.measured_at) { formError.value = '请选择测量日期'; return }
  const w = numOf(form.value.weight_kg)
  const h = numOf(form.value.height_cm)
  const hd = numOf(form.value.head_cm)
  if (w <= 0 && h <= 0 && hd <= 0) {
    formError.value = '至少填写一项测量值'; return
  }
  const data = {
    measured_at: form.value.measured_at,
    weight_kg: w,
    height_cm: h,
    head_cm: hd,
  }
  submitting.value = true
  try {
    if (editingId.value) {
      await babyAPI.updateGrowth(editingId.value, data)
    } else {
      const baby = app.currentBaby
      if (!baby) return
      await babyAPI.createGrowth(baby.id, data)
    }
    formOpen.value = false
    editingId.value = null
    await load()
    app.showToast('已保存', 'success')
  } catch (e: any) {
    formError.value = writeErrorMessage(e, '保存失败')
  } finally {
    submitting.value = false
  }
}

function askDelete(g: GrowthRecord) {
  toDelete.value = g
}

/* ── 长按上下文菜单（v-for 行不适合逐行 useLongPress，用单例状态内联实现，参数与 useLongPress 一致）── */
const contextOpen = ref(false)
const contextGrowth = ref<GrowthRecord | null>(null)
// 长按只留删除：编辑走整行点按 / 回车键
const contextActions = [
  { key: 'delete', label: '删除', icon: 'M19 7l-.9 12a1.5 1.5 0 01-1.5 1.4H7.4A1.5 1.5 0 015.9 19L5 7m5 0V4.5A1.5 1.5 0 0111.5 3h1A1.5 1.5 0 0114 4.5V7m-9 0h14', danger: true },
]
let lpTimer: number | null = null
let lpFired = false
let lpStartX = 0
let lpStartY = 0
function rowTouchStart(g: GrowthRecord, e: TouchEvent) {
  if (e.touches.length !== 1) return
  const t = e.touches[0]
  lpStartX = t.clientX; lpStartY = t.clientY; lpFired = false
  if (lpTimer !== null) { clearTimeout(lpTimer); lpTimer = null }
  lpTimer = window.setTimeout(() => {
    lpFired = true
    contextGrowth.value = g
    contextOpen.value = true
  }, 480)
}
function rowTouchMove(e: TouchEvent) {
  if (lpTimer === null || lpFired) return
  const t = e.touches[0]
  if (Math.abs(t.clientX - lpStartX) > 10 || Math.abs(t.clientY - lpStartY) > 10) { clearTimeout(lpTimer); lpTimer = null }
}
function rowTouchEnd() {
  if (lpTimer !== null) { clearTimeout(lpTimer); lpTimer = null }
}
function onRowClick(g: GrowthRecord) {
  if (lpFired) { lpFired = false; return }
  openEdit(g)
}
function onContextSelect(key: string) {
  const g = contextGrowth.value
  if (!g) return
  if (key === 'delete') askDelete(g)
}

async function doDelete() {
  if (!toDelete.value) return
  deleting.value = true
  try {
    await babyAPI.deleteGrowth(toDelete.value.id)
    toDelete.value = null
    await load()
    app.showToast('已删除', 'success')
  } catch (e: any) {
    app.showToast(writeErrorMessage(e, '删除失败'), 'error')
  } finally {
    deleting.value = false
  }
}

function onGrowthUpdated() {
  load()
}

onMounted(() => {
  load()
  window.addEventListener('record-updated', onGrowthUpdated)
})
onUnmounted(() => {
  window.removeEventListener('record-updated', onGrowthUpdated)
  if (lpTimer !== null) { clearTimeout(lpTimer); lpTimer = null }
})
</script>
