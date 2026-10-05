<template>
  <div class="flex flex-col h-dvh">
    <PullRefresh class="flex-1 min-h-0" content-class="px-4 py-4 space-y-4 pb-[calc(6.5rem+env(safe-area-inset-bottom))]"
      :refresh="loadData">
      <template #header>
        <NavBar :title="app.currentBaby?.name || '记录'" />
      </template>

<!-- 空状态：宝宝列表加载失败（后端不可达）与「确实还没有宝宝」必须区分 -->
      <EmptyState v-if="babiesLoadFailed" title="无法加载宝宝列表" icon="wifi"
        subtitle="请检查网络或后端服务后重试">
        <button @click="retryLoadBabies"
          class="inline-flex items-center gap-2 px-5 py-2.5 bg-primary/10 text-primary-deep rounded-xl font-medium text-sm btn-press">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v6h6M20 20v-6h-6M20 9a9 9 0 0 0-15.5-3M4 15a9 9 0 0 0 15.5 3"/></svg>
          重新加载
        </button>
      </EmptyState>

      <!-- 空状态：无宝宝 -->
      <EmptyState v-else-if="app.babies.length === 0" title="还没有添加宝宝" icon="folder"
        subtitle="添加宝宝档案后即可开始记录护理数据">
        <router-link to="/baby/new"
          class="inline-flex items-center gap-2 px-5 py-2.5 bg-primary/10 text-primary-deep rounded-xl font-medium text-sm btn-press">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
          添加宝宝
        </router-link>
      </EmptyState>

      <!-- 主内容 -->
      <template v-else>
        <!-- 统计卡片（可点击跳转） -->
        <div class="grid grid-cols-2 gap-3">
          <!-- 喂奶卡片 -->
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('feeding')" @click="goToTimeline('feeding')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer press-card">
            <div class="text-xs text-text-secondary mb-1">今日喂奶</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-0.5">
                <span class="text-3xl font-bold text-text-primary font-num">{{ stats.total_ml_today }}<sup v-if="stats.feeding_count > 0" class="text-[0.55em] font-bold text-text-secondary font-num leading-none">{{ stats.feeding_count }}</sup></span>
                <span :class="UNIT_CLASS">ml</span>
              </div>
              <div class="w-9 h-9 shrink-0 flex items-center justify-center rounded-xl bg-primary/10 text-lg leading-none">🍼</div>
            </div>
            <div class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastFeedingAgo && lastFeedingAgo.isLong ? 'text-warning' : 'text-text-secondary'">
                {{ lastFeedingAgo ? lastFeedingAgo.text : '--' }}
              </span>
            </div>
            <div class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">平均间隔</span>
              <span class="text-xs font-medium text-text-secondary">{{ feedingAvgInterval || '--' }}</span>
            </div>
            <!-- 新增喂奶入口 -->
            <button @click.stop="goToAddFeeding"
              class="mt-3 w-full min-h-[44px] py-2 bg-primary/10 text-primary-deep text-sm font-medium rounded-xl btn-press flex items-center justify-center gap-1">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14m7-7H5"/></svg>
            喂奶
            </button>
          </div>

          <!-- 尿布卡片 -->
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('diaper')" @click="goToTimeline('diaper')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer press-card">
            <div class="text-xs text-text-secondary mb-1">今日尿布</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-1">
                <span class="text-3xl font-bold font-num text-text-primary">{{ stats.diaper_count }}</span>
                <span class="text-sm text-text-secondary">次</span>
              </div>
              <div class="w-9 h-9 shrink-0 flex items-center justify-center rounded-xl bg-diaper/10 text-lg leading-none">🩲</div>
            </div>
            <div class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastDiaperAgo && lastDiaperAgo.isLong ? 'text-warning' : 'text-text-secondary'">
                {{ lastDiaperAgo ? lastDiaperAgo.text : '--' }}
              </span>
            </div>
            <div class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">平均间隔</span>
              <span class="text-xs font-medium text-text-secondary">{{ diaperAvgInterval || '--' }}</span>
            </div>
            <!-- 新增尿布入口 -->
            <button @click.stop="goToAddDiaper"
              class="mt-3 w-full min-h-[44px] py-2 bg-diaper/10 text-diaper-deep text-sm font-medium rounded-xl btn-press flex items-center justify-center gap-1">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14m7-7H5"/></svg>
            尿布
            </button>
          </div>

          <!-- 睡眠卡片 -->
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('sleep')" @click="goToTimeline('sleep')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer press-card">
            <div class="text-xs text-text-secondary mb-1">今日睡眠</div>
            <div class="flex items-end justify-between">
              <div class="flex items-center gap-1 min-w-0">
                <span v-if="currentSleep" class="w-1.5 h-1.5 rounded-full bg-sleep-deep animate-pulse shrink-0"></span>
                <div class="flex items-baseline gap-px min-w-0">
                  <template v-for="(part, pi) in sleepParts" :key="pi">
                    <span class="text-3xl font-bold font-num text-text-primary leading-none">{{ part.val }}</span>
                    <span :class="UNIT_CLASS">{{ part.unit }}</span>
                  </template>
                </div>
              </div>
              <div class="w-9 h-9 shrink-0 flex items-center justify-center rounded-xl bg-sleep/10 text-lg leading-none">😴</div>
            </div>
            <div class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastSleepAgo && lastSleepAgo.isLong ? 'text-warning' : 'text-text-secondary'">{{ lastSleepAgo ? lastSleepAgo.text : '--' }}</span>
            </div>
            <div class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">平均时长</span>
              <span class="text-xs font-medium text-text-secondary">{{ sleepAvgDuration || '--' }}</span>
            </div>
            <button v-if="currentSleep" @click.stop="stopSleep" :disabled="loadingAction === 'stop-sleep'"
              class="mt-3 w-full min-h-[44px] py-2 bg-danger-fill text-white text-sm font-medium rounded-xl btn-press flex items-center justify-center gap-1 disabled:opacity-50">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M6 6h12v12H6z"/></svg>
              {{ loadingAction === 'stop-sleep' ? '处理中...' : '结束' }}
            </button>
            <button v-else @click.stop="startSleep" :disabled="loadingAction === 'start-sleep'"
              class="mt-3 w-full min-h-[44px] py-2 bg-sleep/10 text-sleep-deep text-sm font-medium rounded-xl btn-press flex items-center justify-center gap-1 disabled:opacity-50">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
              {{ loadingAction === 'start-sleep' ? '处理中...' : '开始' }}
            </button>
          </div>

          <!-- 体温卡片 -->
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('temperature')" @click="goToTimeline('temperature')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer press-card">
            <div class="text-xs text-text-secondary mb-1">今日体温</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-1">
                <span v-if="todayTemp" class="text-3xl font-bold font-num" :class="todayTemp >= 37.5 ? 'text-danger' : 'text-text-primary'">{{ todayTemp }}</span>
                <span v-else class="text-3xl font-bold font-num text-text-secondary">--</span>
                <span class="text-sm text-text-secondary">°C</span>
              </div>
              <div class="w-9 h-9 shrink-0 flex items-center justify-center rounded-xl bg-temperature/10 text-lg leading-none">🌡️</div>
            </div>
            <div class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastTempAgo && lastTempAgo.isLong ? 'text-warning' : 'text-text-secondary'">{{ lastTempAgo ? lastTempAgo.text : '--' }}</span>
            </div>
            <div class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">今日最高</span>
              <span class="text-xs font-medium font-num" :class="todayTemp && (todayTempHigh || 0) >= 37.5 ? 'text-danger' : 'text-text-secondary'">{{ todayTemp ? `${todayTempHigh?.toFixed(1)}°C` : '--' }}</span>
            </div>
            <button @click.stop="goToAddTemperature"
              class="mt-3 w-full min-h-[44px] py-2 bg-temperature/10 text-temperature-deep text-sm font-medium rounded-xl btn-press flex items-center justify-center gap-1">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14m7-7H5"/></svg>
            测温
            </button>
          </div>

          <!-- 户外活动卡片 -->
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('outdoor')" @click="goToTimeline('outdoor')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer press-card">
            <div class="text-xs text-text-secondary mb-1">今日户外活动</div>
            <div class="flex items-end justify-between">
              <div class="flex items-center gap-1 min-w-0">
                <span v-if="currentOutdoor" class="w-1.5 h-1.5 rounded-full bg-outdoor-deep animate-pulse shrink-0"></span>
                <div class="flex items-baseline gap-px min-w-0">
                  <template v-for="(part, pi) in outdoorParts" :key="pi">
                    <span class="text-3xl font-bold font-num text-text-primary leading-none">{{ part.val }}</span>
                    <span :class="UNIT_CLASS">{{ part.unit }}</span>
                  </template>
                </div>
              </div>
              <div class="w-9 h-9 shrink-0 flex items-center justify-center rounded-xl bg-outdoor/10 text-lg leading-none">🌳</div>
            </div>
            <div class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastOutdoorAgo && lastOutdoorAgo.isLong ? 'text-warning' : 'text-text-secondary'">{{ lastOutdoorAgo ? lastOutdoorAgo.text : '--' }}</span>
            </div>
            <div class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">平均时长</span>
              <span class="text-xs font-medium text-text-secondary">{{ avgOutdoorDuration || '--' }}</span>
            </div>
            <button v-if="currentOutdoor" @click.stop="stopOutdoor" :disabled="loadingAction === 'stop-outdoor'"
              class="mt-3 w-full min-h-[44px] py-2 bg-danger-fill text-white text-sm font-medium rounded-xl btn-press flex items-center justify-center gap-1 disabled:opacity-50">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M6 6h12v12H6z"/></svg>
              {{ loadingAction === 'stop-outdoor' ? '处理中...' : '结束' }}
            </button>
            <button v-else @click.stop="startOutdoor" :disabled="loadingAction === 'start-outdoor'"
              class="mt-3 w-full min-h-[44px] py-2 bg-outdoor/10 text-outdoor-deep text-sm font-medium rounded-xl btn-press flex items-center justify-center gap-1 disabled:opacity-50">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
              {{ loadingAction === 'start-outdoor' ? '处理中...' : '开始' }}
            </button>
          </div>

          <!-- 补剂卡片 -->
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('supplement')" @click="goToTimeline('supplement')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer press-card">
            <div class="text-xs text-text-secondary mb-1">今日补剂</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-1">
                <span class="text-3xl font-bold font-num text-text-primary">{{ stats.supplement_count }}</span>
                <span class="text-sm text-text-secondary">次</span>
              </div>
              <div class="w-9 h-9 shrink-0 flex items-center justify-center rounded-xl bg-supplement/10 text-lg leading-none">💊</div>
            </div>
            <div class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastSupplementAgo && lastSupplementAgo.isLong ? 'text-warning' : 'text-text-secondary'">{{ lastSupplementAgo ? lastSupplementAgo.text : '--' }}</span>
            </div>
            <div class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">平均间隔</span>
              <span class="text-xs font-medium text-text-secondary">{{ supplementAvgInterval || '--' }}</span>
            </div>
            <button @click.stop="goToAddSupplement"
              class="mt-3 w-full min-h-[44px] py-2 bg-supplement/10 text-supplement-deep text-sm font-medium rounded-xl btn-press flex items-center justify-center gap-1">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14m7-7H5"/></svg>
            补剂
            </button>
          </div>
        </div>

        <!-- 最近记录 -->
        <div class="space-y-2">
          <h2 class="text-[13px] text-text-secondary">最近记录</h2>
          <div v-if="allRecords.length === 0" class="bg-surface rounded-2xl shadow-card">
            <EmptyState title="还没有记录" subtitle="从上方卡片快速记录喂奶、睡眠等" size="sm" icon="clock" />
          </div>
          <RecordCard v-for="r in allRecords" :key="r.record_type + '-' + r.id" :record="r"
            @edit="editRecord(r)" @context="openContext" />

          <!-- 加载更多（真增量分页）：每次点击一页，避免一次渲染全部卡死 -->
          <div class="h-14 flex items-center justify-center">
            <button v-if="hasMore" @click="loadMore" :disabled="loadingMore"
              class="px-6 py-2.5 rounded-xl shadow-card bg-surface text-sm font-medium text-primary-deep btn-press min-h-[44px] flex items-center gap-2">
              <ActivityIndicator v-if="loadingMore" :size="16" class="text-primary-deep" />
              {{ loadingMore ? '加载中…' : `加载更多（剩余 ${loadMoreRemaining}）` }}
            </button>
            <span v-else class="text-xs text-text-secondary">没有更多了</span>
          </div>
        </div>
      </template>
    </PullRefresh>

    <!-- 删除确认（iOS 底部操作表） -->
    <ConfirmSheet :open="showDeleteConfirm"
      message="确定要删除这条记录吗？删除后可在提示条上撤销。"
      @confirm="confirmDelete" @cancel="showDeleteConfirm = false" />

    <!-- 长按上下文菜单 -->
    <ContextMenu :open="contextOpen" :title="contextRecord?.title" :subtitle="contextRecord?.subtitle"
      :emoji="contextRecord?.emoji" :actions="contextActions"
      @update:open="contextOpen = $event" @select="onContextSelect" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'

// 显式命名：MainLayout 内层 <keep-alive include="HomePage,..."> 命中缓存
defineOptions({ name: 'HomePage' })

import { babyAPI, recordAPI, writeErrorMessage } from '@/api'
import type { BabyStats, SleepRecord, OutdoorRecord } from '@/api'
import RecordCard from '@/components/RecordCard.vue'
import { useUndoDelete } from '@/composables/useUndoDelete'
import PullRefresh from '@/components/PullRefresh.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'
import ContextMenu from '@/components/ContextMenu.vue'
import ActivityIndicator from '@/components/ActivityIndicator.vue'
import { recordDisplay, CONTEXT_ICONS } from '@/utils/recordDisplay'
import EmptyState from '@/components/EmptyState.vue'
import NavBar from '@/components/NavBar.vue'
import { durationCompactParts, formatDurationCN } from '@/utils'

const tick = ref(0)
let tickTimer: number | null = null
const router = useRouter()
const app = useAppStore()

const UNIT_CLASS = 'text-sm text-text-secondary'
const stats = ref<BabyStats>({ feeding_count: 0, diaper_count: 0, total_ml_today: 0, last_feeding: '', last_diaper: '', sleep_duration: 0, last_sleep_end: '', temperature_count: 0, latest_temperature: 0, last_temperature: '', outdoor_duration: 0, last_outdoor_end: '', supplement_count: 0, last_supplement: '' })
const allRecords = ref<any[]>([])
const PAGE = 20
const nextOffset = ref(0)
const totalCount = ref(0)
const loadedCount = ref(0)
const loadingMore = ref(false)
const hasMore = computed(() => loadedCount.value < totalCount.value)
const loadMoreRemaining = computed(() => Math.max(0, totalCount.value - loadedCount.value))
// 今日测温记录（独立按 days=1 拉取，不依赖分页窗口；新→旧）
const todayTempRecords = ref<any[]>([])
// 各类型最近 10 次记录（独立按 type+limit 拉取，新→旧）。
// 首页主列表是 20 条「6 类混排」的分页窗口，低频类型（睡眠/户外/补剂）在窗口里常常凑不满 10 条，
// 且点「加载更多」还会让均值跳变。独立取数后各项均值只取决于真实数据量：有几次算几次（上限 10 次）。
const AVG_WINDOW = 10
const AVG_TYPES = ['feeding', 'diaper', 'sleep', 'outdoor', 'supplement'] as const
const recentByType = ref<Record<string, any[]>>({})
const showDeleteConfirm = ref(false)
const recordToDelete = ref<any>(null)
const { softDelete } = useUndoDelete(allRecords, { onRestored: () => refreshStatsSoon() })

// ── 长按上下文菜单 ─────────────────────────────────────────
const contextOpen = ref(false)
const contextRecord = ref<any>(null)
// 长按只留删除：编辑走卡片点按（与 ContextMenu 无关的独立手势，避免误触）
const contextActions = [
  { key: 'delete', label: '删除', icon: CONTEXT_ICONS.delete, danger: true },
]
function openContext(rec: any) {
  contextRecord.value = { record: rec, ...recordDisplay(rec) }
  contextOpen.value = true
}
function onContextSelect(key: string) {
  const rec = contextRecord.value?.record
  if (!rec) return
  if (key === 'delete') deleteRecord(rec)
}

// 删除/撤销后仅刷新统计（不重拉列表，避免打断撤销窗口内的乐观 UI）
let statsSoonTimer: number | null = null
function refreshStatsSoon() {
  if (statsSoonTimer) clearTimeout(statsSoonTimer)
  statsSoonTimer = window.setTimeout(async () => {
    const baby = app.currentBaby
    if (!baby) return
    try {
      const [statsRes, curSleep, curOutdoor] = await Promise.all([
        babyAPI.stats(baby.id),
        recordAPI.getCurrentSleep(baby.id),
        recordAPI.getCurrentOutdoor(baby.id),
      ])
      stats.value = statsRes.data
      currentSleep.value = (curSleep.data as any)?.id ? (curSleep.data as any) : null
      currentOutdoor.value = (curOutdoor.data as any)?.id ? (curOutdoor.data as any) : null
    } catch { /* 静默 */ }
  }, 120)
}
const currentSleep = ref<SleepRecord | null>(null)
const currentOutdoor = ref<OutdoorRecord | null>(null)
const loadingAction = ref<string | null>(null)
const deleting = ref(false)
let loadGeneration = 0

// 宝宝列表加载失败（后端不可达）——与「确实还没有宝宝」区分，避免冷启动离线时误报无宝宝并给出无法保存的添加入口
const babiesLoadFailed = ref(false)
async function retryLoadBabies() {
  babiesLoadFailed.value = false
  try {
    await app.loadBabies()
    if (app.babies.length === 0) return
    await loadData()
  } catch {
    babiesLoadFailed.value = true
  }
}

function getTimeAgo(isoString: string | null) {
  if (!isoString) return null
  const last = new Date(isoString)
  const now = new Date()
  const diffMs = now.getTime() - last.getTime()
  if (diffMs < 0) return null
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)
  let text = ''
  // 「距上次」最长显示 30 天：超过 30 天统一显示「30天前」
  if (diffDays >= 30) text = '30天前'
  else if (diffDays > 0) text = diffHours % 24 > 0 ? `${diffDays}天${diffHours % 24}小时前` : `${diffDays}天前`
  else if (diffHours > 0) text = diffMins % 60 > 0 ? `${diffHours}小时${diffMins % 60}分钟前` : `${diffHours}小时前`
  else if (diffMins > 0) text = `${diffMins}分钟前`
  else text = '刚刚'
  return { text, isLong: diffHours >= 4, minutes: diffMins }
}

// 某个类型最近若干次记录（独立窗口，已按新→旧、已截断）
function recentOf(type: string): any[] { return recentByType.value[type] || [] }

// 相邻间隔均值：升序后取尾部 AVG_WINDOW 个算相邻差值的平均。
// 数据不足时「有几次算几次」：3 条即按 2 个间隔算；仅 1 条时不存在间隔可言，返回 null 显示「--」。
function avgGapMinutes(occurredList: (string | undefined | null)[]): number | null {
  const times = occurredList
    .filter((t): t is string => !!t)
    .map(t => new Date(t).getTime())
    .filter(t => !Number.isNaN(t))
    .sort((a, b) => a - b)
    .slice(-AVG_WINDOW)
  if (times.length < 2) return null
  let sum = 0
  for (let i = 1; i < times.length; i++) sum += (times[i] - times[i - 1]) / 60000
  return Math.round(sum / (times.length - 1))
}

// 平均时长均值：同样「有几次算几次」，0 条返回 null
function avgDurationMinutes(rows: any[]): number | null {
  const mins = rows
    .filter(r => r.data?.started_at && r.data?.ended_at)
    .map(r => (new Date(r.data.ended_at).getTime() - new Date(r.data.started_at).getTime()) / 60000)
    .filter(t => t > 0 && !Number.isNaN(t))
    .slice(0, AVG_WINDOW)
  if (!mins.length) return null
  return Math.round(mins.reduce((sum, t) => sum + t, 0) / mins.length)
}

const fmtGap = (rows: any[]) => {
  const m = avgGapMinutes(rows.map(r => r.occurred_at))
  return m == null ? null : formatDurationCN(m)
}

const feedingAvgInterval = computed(() => fmtGap(recentOf('feeding')))
const diaperAvgInterval = computed(() => fmtGap(recentOf('diaper')))
const supplementAvgInterval = computed(() => fmtGap(recentOf('supplement')))

const sleepAvgDuration = computed(() => {
  const m = avgDurationMinutes(recentOf('sleep'))
  return m == null ? null : formatDurationCN(m)
})

const avgOutdoorDuration = computed(() => {
  const m = avgDurationMinutes(recentOf('outdoor'))
  return m == null ? null : formatDurationCN(m)
})

// 今日日期判定（按本地时区）
function isToday(iso?: string | null) {
  if (!iso) return false
  const d = new Date(iso)
  const n = new Date()
  return d.getFullYear() === n.getFullYear() && d.getMonth() === n.getMonth() && d.getDate() === n.getDate()
}

// 今日测温独立拉取（days=1 专用窗口），不随主列表分页；今日未测温时温度为 null，主数字显示 -- 占位符
const todayTemp = computed<number | null>(() => todayTempRecords.value.length ? todayTempRecords.value[0].data.temperature : null)
const todayTempHigh = computed<number | null>(() => todayTempRecords.value.length
  ? Math.max(...todayTempRecords.value.map((r: any) => r.data.temperature))
  : null)

const lastFeedingAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_feeding) })
const lastDiaperAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_diaper) })
const lastSleepAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_sleep_end) })
// 距上次取「全局最近一次」，不受今日是否有记录影响（超过 30 天统一显示 30天前）
const lastTempAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_temperature) })
const lastOutdoorAgo = computed(() => {
  tick.value
  const t = stats.value.last_outdoor_end
  if (t) return getTimeAgo(t)
  const recs = recentOf('outdoor').map(r => r.occurred_at).filter(Boolean).sort()
  return getTimeAgo(recs.length ? recs[recs.length - 1] : null)
})
const lastSupplementAgo = computed(() => {
  tick.value
  const t = stats.value.last_supplement
  if (t) return getTimeAgo(t)
  // stats 缺字段时回落到独立取数的最近若干条（取最旧一条即全局最近一次）
  const recs = recentOf('supplement').map(r => r.occurred_at).filter(Boolean).sort()
  return getTimeAgo(recs.length ? recs[recs.length - 1] : null)
})

// 进行中已持续分钟数
function elapsedMins(startedAt?: string) {
  tick.value
  if (!startedAt) return 0
  return Math.round((Date.now() - new Date(startedAt).getTime()) / 60000)
}

// 睡眠 / 户外主数值：进行中取实时时长，否则取今日总计（统一紧凑 h/m 大数字）
const sleepParts = computed(() =>
  durationCompactParts(currentSleep.value ? elapsedMins(currentSleep.value.started_at) : stats.value.sleep_duration))
const outdoorParts = computed(() =>
  durationCompactParts(currentOutdoor.value ? elapsedMins(currentOutdoor.value.started_at) : stats.value.outdoor_duration))

async function loadData() {
  if (app.babies.length === 0) {
    const loaded = await app.loadBabies()
    if (!loaded) {
      babiesLoadFailed.value = true
      return
    }
  }
  const baby = app.currentBaby
  if (!baby) return
  babiesLoadFailed.value = false
  const gen = ++loadGeneration
  const tasks: Promise<any>[] = [
    babyAPI.stats(baby.id),
    recordAPI.list(baby.id, { offset: 0, limit: PAGE }),
    recordAPI.count(baby.id),
    recordAPI.getCurrentSleep(baby.id),
    recordAPI.getCurrentOutdoor(baby.id),
    recordAPI.list(baby.id, { type: 'temperature', days: 1 }),
    // 各类型最近 AVG_WINDOW 次（后端 type+limit 走窗口分页，返回新→旧），供各项均值使用
    ...AVG_TYPES.map(t => recordAPI.list(baby.id, { type: t, limit: AVG_WINDOW })),
  ]
  // 逐请求独立落地：任一子请求失败不再牵连整页。
  // 旧写法 Promise.all 是全有全无——11 个请求任一 reject 就整批丢弃，首页变空。
  const settled = await Promise.allSettled(tasks)
  if (gen !== loadGeneration) return
  const val = (i: number): any => (settled[i].status === 'fulfilled' ? settled[i].value : undefined)

  const statsRes = val(0)
  if (statsRes) stats.value = statsRes.data

  const recordsRes = val(1)
  if (recordsRes) {
    allRecords.value = recordsRes.data as any[]
    loadedCount.value = allRecords.value.length
    nextOffset.value = allRecords.value.length
  }

  const countRes = val(2)
  if (countRes) totalCount.value = countRes.data.total

  // 进行中的睡眠/户外：失败时保留旧值，避免把「正在睡」误清成无
  const sleepRes = val(3)
  if (sleepRes) currentSleep.value = sleepRes.data?.id ? sleepRes.data : null
  const outdoorRes = val(4)
  if (outdoorRes) currentOutdoor.value = outdoorRes.data?.id ? outdoorRes.data : null

  const tempTodayRes = val(5)
  if (tempTodayRes) {
    todayTempRecords.value = (tempTodayRes.data as any[])
      .filter(r => r.data?.temperature > 0 && isToday(r.occurred_at))
      .sort((a, b) => new Date(b.occurred_at).getTime() - new Date(a.occurred_at).getTime())
  }

  // 均值窗口失败时保留上一轮，避免闪断时统计跳变
  recentByType.value = Object.fromEntries(
    AVG_TYPES.map((t, i) => {
      const r = settled[6 + i]
      const rows = r.status === 'fulfilled' ? ((r.value?.data as any[]) || []) : recentByType.value[t] || []
      return [t, rows]
    }),
  )

  // 仅当主列表失败且当前无数据可展示时才提示，避免可选/后台请求失败反复刷屏
  if (settled[1].status === 'rejected' && allRecords.value.length === 0) {
    app.showToast('数据加载失败，请检查网络或后端服务', 'error')
  }
}

async function loadMore() {
  const baby = app.currentBaby
  if (!baby || loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  try {
    const res = await recordAPI.list(baby.id, { offset: nextOffset.value, limit: PAGE })
    const seen = new Set(allRecords.value.map(r => r.record_type + '-' + r.id))
    const added = (res.data as any[]).filter(r => !seen.has(r.record_type + '-' + r.id))
    allRecords.value = allRecords.value.concat(added)
    allRecords.value.sort((a, b) => (b.occurred_at || '').localeCompare(a.occurred_at || ''))
    loadedCount.value += added.length
    nextOffset.value += res.data.length
  } catch {
    app.showToast('数据加载失败', 'error')
  } finally {
    loadingMore.value = false
  }
}

function goToTimeline(filter: string) {
  router.push(`/timeline?filter=${filter}`)
}

function goToAddFeeding() {
  router.push('/record/feeding')
}

function goToAddDiaper() {
  router.push('/record/diaper')
}

function goToAddTemperature() {
  router.push('/temperature')
}

function goToAddSupplement() {
  router.push('/supplement')
}

async function startSleep() {
  const baby = app.currentBaby
  if (!baby || loadingAction.value) return
  loadingAction.value = 'start-sleep'
  try {
    const now = new Date().toISOString()
    const res = await recordAPI.createSleepStart(baby.id, { started_at: now })
    currentSleep.value = (res.data as any).data ?? res.data
    window.dispatchEvent(new CustomEvent('record-created', { detail: res.data }))
    app.showToast('开始睡觉', 'success')
  } catch (e: any) {
    console.error('开始睡眠失败:', e?.response?.data || e)
    app.showToast(writeErrorMessage(e, '开始睡眠失败'), 'error')
  } finally {
    loadingAction.value = null
  }
}

async function stopSleep() {
  const baby = app.currentBaby
  if (!baby || !currentSleep.value || loadingAction.value) return
  loadingAction.value = 'stop-sleep'
  try {
    const now = new Date().toISOString()
    await recordAPI.stopSleep(baby.id, currentSleep.value.id, { ended_at: now })
    currentSleep.value = null
    await loadData()
    app.showToast('睡眠已结束', 'success')
  } catch (e: any) {
    console.error('结束睡眠失败:', e?.response?.data || e)
    app.showToast(writeErrorMessage(e, '结束睡眠失败'), 'error')
  } finally {
    loadingAction.value = null
  }
}

async function startOutdoor() {
  const baby = app.currentBaby
  if (!baby || loadingAction.value) return
  loadingAction.value = 'start-outdoor'
  try {
    const now = new Date().toISOString()
    const res = await recordAPI.createOutdoorStart(baby.id, { started_at: now })
    currentOutdoor.value = (res.data as any).data ?? res.data
    window.dispatchEvent(new CustomEvent('record-created', { detail: res.data }))
    app.showToast('开始户外活动', 'success')
  } catch (e: any) {
    console.error('开始户外活动失败:', e?.response?.data || e)
    app.showToast(writeErrorMessage(e, '开始户外活动失败'), 'error')
  } finally {
    loadingAction.value = null
  }
}

async function stopOutdoor() {
  const baby = app.currentBaby
  if (!baby || !currentOutdoor.value || loadingAction.value) return
  loadingAction.value = 'stop-outdoor'
  try {
    const now = new Date().toISOString()
    await recordAPI.stopOutdoor(baby.id, currentOutdoor.value.id, { ended_at: now })
    currentOutdoor.value = null
    await loadData()
    app.showToast('户外活动已结束', 'success')
  } catch (e: any) {
    console.error('结束户外活动失败:', e?.response?.data || e)
    app.showToast(writeErrorMessage(e, '结束户外活动失败'), 'error')
  } finally {
    loadingAction.value = null
  }
}

function editRecord(r: any) {
  if (r.record_type === 'sleep') {
    router.push(`/sleep/${r.id}/edit`)
  } else if (r.record_type === 'temperature') {
    router.push(`/temperature/${r.id}/edit`)
  } else if (r.record_type === 'outdoor') {
    router.push(`/outdoor/${r.id}/edit`)
  } else if (r.record_type === 'supplement') {
    router.push(`/supplement/${r.id}/edit`)
  } else {
    router.push(`/record/${r.record_type}/${r.id}/edit`)
  }
}

function deleteRecord(r: any) {
  recordToDelete.value = r
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (!recordToDelete.value || deleting.value) return
  const target = recordToDelete.value
  showDeleteConfirm.value = false
  softDelete(target)
  refreshStatsSoon()
}

function onRecordCreated(e: Event) {
  const record = (e as CustomEvent).detail
  if (!record) { loadData(); return }
  if (record.baby_id === app.currentBaby?.id) {
    if (record.record_type === 'sleep' || record.record_type === 'outdoor') {
      if (!record.data?.ended_at) {
        if (record.record_type === 'sleep') currentSleep.value = record.data
        else currentOutdoor.value = record.data
        return
      }
      if (record.record_type === 'sleep' && currentSleep.value?.id === record.id) currentSleep.value = null
      if (record.record_type === 'outdoor' && currentOutdoor.value?.id === record.id) currentOutdoor.value = null
      allRecords.value.unshift(record)
      loadData()
      return
    }
    allRecords.value.unshift(record)
    // 同步进该类型的最近窗口，均值/距上次立即跟上（否则要等下次整页刷新）。
    // 睡眠/户外在上面的分支已 return（结束时会整页 loadData），此处只处理即时记录的喂奶/尿布/补剂。
    const t = record.record_type as (typeof AVG_TYPES)[number]
    if (AVG_TYPES.includes(t)) {
      const rows = [record, ...recentOf(t).filter(r => r.id !== record.id)]
        .sort((a, b) => (b.occurred_at || '').localeCompare(a.occurred_at || ''))
        .slice(0, AVG_WINDOW)
      recentByType.value = { ...recentByType.value, [t]: rows }
    }
  }
}

function onRecordDeleted(e: Event) {
  const { id, type } = (e as CustomEvent).detail || {}
  allRecords.value = allRecords.value.filter(r => !(r.id === id && r.record_type === (type || r.record_type)))
  const targets = type ? [type] : [...AVG_TYPES]
  const next = { ...recentByType.value }
  let changed = false
  for (const t of targets) {
    if (!next[t]) continue
    next[t] = next[t].filter(r => r.id !== id)
    changed = true
  }
  if (changed) recentByType.value = next
}

onMounted(() => {
  loadData()
  window.addEventListener('record-created', onRecordCreated)
  window.addEventListener('record-deleted', onRecordDeleted)
  tickTimer = window.setInterval(() => { tick.value++ }, 10000)
})
onUnmounted(() => {
  window.removeEventListener('record-created', onRecordCreated)
  window.removeEventListener('record-deleted', onRecordDeleted)
  if (tickTimer !== null) clearInterval(tickTimer)
})
</script>
