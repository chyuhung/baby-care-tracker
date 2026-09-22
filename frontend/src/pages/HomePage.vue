<template>
  <div class="flex flex-col h-dvh">
    <PullRefresh class="flex-1 min-h-0" content-class="px-4 py-4 space-y-4 pb-[calc(6.5rem+env(safe-area-inset-bottom))]"
      :refresh="loadData" @scroll="navScroll = $event">
      <!-- Header -->
      <template #header>
      <LargeTitleNav :title="app.currentBaby?.name ? `${app.currentBaby?.name} 的记录` : '宝宝护理'"
        inline-title="记录" :scroll-top="navScroll">
        <template #actions>
          <span v-if="app.wsConnected" class="text-xs text-success flex items-center gap-1">
            <span class="w-2 h-2 bg-success rounded-full inline-block"></span>同步
          </span>
          <span v-else class="text-xs text-text-secondary">离线</span>
        </template>
        <template #sub>
          <p v-if="app.currentBaby?.birth_date" class="text-xs text-text-secondary truncate mt-0.5">
            {{ ageText }} · {{ todayDateText }}
          </p>
        </template>
        <template #filters>
          <div v-if="app.currentBaby" class="flex items-center gap-2 mt-2">
            <div class="relative flex-1">
              <select v-model="selectedBabyId" @change="switchBaby"
                class="w-full min-h-[44px] px-3 py-2.5 bg-surface border border-border-color rounded-xl text-base text-text-primary appearance-none cursor-pointer focus:border-primary focus:outline-none transition-colors pr-8">
                <option v-for="b in app.babies" :key="b.id" :value="b.id">{{ b.name }}</option>
              </select>
              <svg class="w-4 h-4 text-text-secondary pointer-events-none absolute right-3 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
            </div>
          </div>
        </template>
      </LargeTitleNav>
    </template>

      <!-- 空状态：无宝宝 -->
      <EmptyState v-if="app.babies.length === 0" title="还没有添加宝宝" icon="folder"
        subtitle="添加宝宝档案后即可开始记录护理数据">
        <router-link to="/baby/new"
          class="inline-flex items-center gap-2 px-5 py-2.5 bg-primary-fill text-white rounded-xl font-medium text-sm btn-press shadow-card">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
          添加宝宝
        </router-link>
      </EmptyState>

      <!-- 主内容 -->
      <template v-else>
        <!-- 统计卡片（可点击跳转） -->
        <div class="grid grid-cols-2 gap-3">
          <!-- 喂奶卡片 -->
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('feeding')" @click="goToTimeline('feeding')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日喂奶</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-0.5">
                <span class="text-3xl font-bold text-text-primary font-num">{{ stats.total_ml_today }}<sup v-if="stats.feeding_count > 0" class="text-[0.55em] font-bold text-text-secondary font-num leading-none">{{ stats.feeding_count }}</sup></span>
                <span :class="UNIT_CLASS">ml</span>
              </div>
              <div class="text-3xl">🍼</div>
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
              <span class="text-base">＋</span> 喂奶
            </button>
          </div>

          <!-- 尿布卡片 -->
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('diaper')" @click="goToTimeline('diaper')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日尿布</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-1">
                <span class="text-3xl font-bold font-num text-text-primary">{{ stats.diaper_count }}</span>
                <span class="text-sm text-text-secondary">次</span>
              </div>
              <div class="text-3xl">🩲</div>
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
              <span class="text-base">＋</span> 尿布
            </button>
          </div>

          <!-- 睡眠卡片 -->
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('sleep')" @click="goToTimeline('sleep')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer btn-press">
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
              <div class="text-3xl">😴</div>
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
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('temperature')" @click="goToTimeline('temperature')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日体温</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-1">
                <span v-if="todayTemp" class="text-3xl font-bold font-num" :class="todayTemp >= 37.5 ? 'text-danger' : 'text-text-primary'">{{ todayTemp }}</span>
                <span v-else class="text-3xl font-bold font-num text-text-secondary">--</span>
                <span class="text-sm text-text-secondary">°C</span>
              </div>
              <div class="text-3xl">🌡️</div>
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
              <span class="text-base">＋</span> 测温
            </button>
          </div>

          <!-- 户外活动卡片 -->
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('outdoor')" @click="goToTimeline('outdoor')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer btn-press">
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
              <div class="text-3xl">🌳</div>
            </div>
            <div class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastOutdoorAgo && lastOutdoorAgo.isLong ? 'text-warning' : 'text-text-secondary'">{{ lastOutdoorAgo ? lastOutdoorAgo.text : '--' }}</span>
            </div>
            <div class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">平均时长</span>
              <span class="text-xs font-medium text-text-secondary">{{ avgOutdoorDuration > 0 ? formatAvgOutdoor : '--' }}</span>
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
          <div role="button" tabindex="0" @keydown.enter.prevent="goToTimeline('supplement')" @click="goToTimeline('supplement')" class="bg-surface rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日补剂</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-1">
                <span class="text-3xl font-bold font-num text-text-primary">{{ stats.supplement_count }}</span>
                <span class="text-sm text-text-secondary">次</span>
              </div>
              <div class="text-3xl">💊</div>
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
              <span class="text-base">＋</span> 补剂
            </button>
          </div>
        </div>

        <!-- 最近记录 -->
        <div class="space-y-2">
          <h2 class="text-sm font-semibold text-text-secondary">最近记录</h2>
          <div v-if="displayRecords.length === 0" class="bg-surface rounded-2xl shadow-card">
            <EmptyState title="还没有记录" subtitle="从上方卡片快速记录喂奶、睡眠等" size="sm" icon="clock" />
          </div>
          <SwipeToDelete v-for="(r, i) in displayRecords" :key="r.record_type + '-' + r.id"
            :style="{ animationDelay: `${i * 60}ms` }" class="card-in" @delete="softDelete(r)">
            <RecordCard :record="r" @edit="editRecord(r)" @delete="deleteRecord(r)" @context="openContext" />
          </SwipeToDelete>

          <!-- 展开全部记录（iOS 朴素文字行） -->
          <button v-if="!showAllRecords && allRecords.length > displayRecords.length"
            @click="showAllRecords = true"
            class="w-full py-3 text-primary-deep text-sm font-medium btn-press mt-1">
            展开全部记录（{{ allRecords.length - displayRecords.length }}）
          </button>
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
import { babyAPI, recordAPI } from '@/api'
import type { BabyStats, SleepRecord, OutdoorRecord } from '@/api'
import RecordCard from '@/components/RecordCard.vue'
import SwipeToDelete from '@/components/SwipeToDelete.vue'
import { useUndoDelete } from '@/composables/useUndoDelete'
import PullRefresh from '@/components/PullRefresh.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'
import ContextMenu from '@/components/ContextMenu.vue'
import { recordDisplay, CONTEXT_ICONS } from '@/utils/recordDisplay'
import EmptyState from '@/components/EmptyState.vue'
import LargeTitleNav from '@/components/LargeTitleNav.vue'
import { durationCompactParts, formatDurationCN, WEEKDAY_SHORT } from '@/utils'

const tick = ref(0)
let tickTimer: number | null = null
const router = useRouter()
const app = useAppStore()
const navScroll = ref(0)
const UNIT_CLASS = 'text-sm text-text-secondary'
const stats = ref<BabyStats>({ feeding_count: 0, diaper_count: 0, total_ml_today: 0, last_feeding: '', last_diaper: '', sleep_count: 0, sleep_duration: 0, last_sleep_end: '', temperature_count: 0, latest_temperature: 0, last_temperature: '', outdoor_count: 0, outdoor_duration: 0, last_outdoor_end: '', supplement_count: 0, last_supplement: '' })
const allRecords = ref<any[]>([])
const showAllRecords = ref(false)
const showDeleteConfirm = ref(false)
const recordToDelete = ref<any>(null)
const { softDelete } = useUndoDelete(allRecords, { onRestored: () => refreshStatsSoon() })

// ── 长按上下文菜单 ─────────────────────────────────────────
const contextOpen = ref(false)
const contextRecord = ref<any>(null)
const contextActions = [
  { key: 'edit', label: '编辑', icon: CONTEXT_ICONS.edit },
  { key: 'delete', label: '删除', icon: CONTEXT_ICONS.delete, danger: true },
]
function openContext(rec: any) {
  contextRecord.value = { record: rec, ...recordDisplay(rec) }
  contextOpen.value = true
}
function onContextSelect(key: string) {
  const rec = contextRecord.value?.record
  if (!rec) return
  if (key === 'edit') editRecord(rec)
  else if (key === 'delete') deleteRecord(rec)
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
const selectedBabyId = ref<number | null>(null)
let loadGeneration = 0


// 只显示今天和昨天
const displayRecords = computed(() => {
  if (showAllRecords.value) return allRecords.value
  const now = new Date()
  const today = now.toDateString()
  const yesterday = new Date(now.getTime() - 86400000).toDateString()
  return allRecords.value.filter(r => {
    const d = new Date(r.occurred_at)
    return d.toDateString() === today || d.toDateString() === yesterday
  })
})

const ageText = computed(() => {
  const baby = app.currentBaby
  if (!baby?.birth_date) return ''
  const bd = new Date(baby.birth_date)
  if (isNaN(bd.getTime())) return ''
  const birthYear = bd.getFullYear()
  const birthMonth = bd.getMonth()
  const birthDay = bd.getDate()
  const now = new Date()
  if (now.getFullYear() < birthYear ||
      (now.getFullYear() === birthYear && (now.getMonth() < birthMonth ||
        (now.getMonth() === birthMonth && now.getDate() < birthDay)))) {
    return '未出生'
  }
  let months = (now.getFullYear() - birthYear) * 12 + now.getMonth() - birthMonth
  const prevMonthDays = new Date(now.getFullYear(), now.getMonth(), 0).getDate()
  const effBirthDay = Math.min(birthDay, prevMonthDays)
  let days: number
  if (now.getDate() >= effBirthDay) {
    days = now.getDate() - effBirthDay
  } else {
    months--
    days = prevMonthDays - effBirthDay + now.getDate()
  }
  if (months > 0 && days === 0) return `${months}个月`
  if (months > 0) return `${months}个月${days}天`
  return `${days}天`
})

const todayDateText = computed(() => {
  const d = new Date()
  return `${d.getMonth() + 1}月${d.getDate()}日 ${WEEKDAY_SHORT[d.getDay()]}`
})

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

function avgIntervalMinutes(records: any[], type: string): number | null {
  const times = records
    .filter(r => r.record_type === type)
    .map(r => new Date(r.occurred_at).getTime())
    .sort((a, b) => a - b)
    .slice(-10)
  if (times.length < 2) return null
  let sum = 0
  for (let i = 1; i < times.length; i++) sum += (times[i] - times[i - 1]) / 60000
  return Math.round(sum / (times.length - 1))
}

const feedingAvgInterval = computed(() => {
  const m = avgIntervalMinutes(allRecords.value, 'feeding')
  return m == null ? null : formatDurationCN(m)
})

const diaperAvgInterval = computed(() => {
  const m = avgIntervalMinutes(allRecords.value, 'diaper')
  return m == null ? null : formatDurationCN(m)
})

const supplementAvgInterval = computed(() => {
  const m = avgIntervalMinutes(allRecords.value, 'supplement')
  return m == null ? null : formatDurationCN(m)
})

const sleepAvgDuration = computed(() => {
  const recs = allRecords.value
    .filter(r => r.record_type === 'sleep' && r.data?.started_at && r.data?.ended_at)
    .map(r => ({
      t: (new Date(r.data.ended_at).getTime() - new Date(r.data.started_at).getTime()) / 60000,
      occurred: new Date(r.occurred_at).getTime(),
    }))
    .filter(x => x.t > 0)
    .sort((a, b) => b.occurred - a.occurred)
    .slice(0, 10)
  if (!recs.length) return null
  return formatDurationCN(Math.round(recs.reduce((sum, x) => sum + x.t, 0) / recs.length))
})

// 今日日期判定（按本地时区）
function isToday(iso?: string | null) {
  if (!iso) return false
  const d = new Date(iso)
  const n = new Date()
  return d.getFullYear() === n.getFullYear() && d.getMonth() === n.getMonth() && d.getDate() === n.getDate()
}

// 今日测温记录（新→旧）；今日未测温时温度为 null，主数字显示 -- 占位符
const todayTempRecords = computed(() =>
  allRecords.value
    .filter(r => r.record_type === 'temperature' && r.data?.temperature > 0 && isToday(r.occurred_at))
    .sort((a, b) => new Date(b.occurred_at).getTime() - new Date(a.occurred_at).getTime()))
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
  const recs = allRecords.value.filter(r => r.record_type === 'outdoor').map(r => r.occurred_at).sort()
  return getTimeAgo(recs.length ? recs[recs.length - 1] : null)
})
const lastSupplementAgo = computed(() => {
  tick.value
  const t = stats.value.last_supplement
  if (t) return getTimeAgo(t)
  const recs = allRecords.value.filter(r => r.record_type === 'supplement').map(r => r.occurred_at).sort()
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

// 户外平均时长：与睡眠同口径，取最近 10 次已结束记录的平均时长
const avgOutdoorDuration = computed(() => {
  const recs = allRecords.value
    .filter(r => r.record_type === 'outdoor' && r.data?.started_at && r.data?.ended_at)
    .map(r => ({
      t: (new Date(r.data.ended_at).getTime() - new Date(r.data.started_at).getTime()) / 60000,
      occurred: new Date(r.occurred_at).getTime(),
    }))
    .filter(x => x.t > 0)
    .sort((a, b) => b.occurred - a.occurred)
    .slice(0, 10)
  if (!recs.length) return 0
  return Math.round(recs.reduce((sum, x) => sum + x.t, 0) / recs.length)
})

const formatAvgOutdoor = computed(() => formatDurationCN(avgOutdoorDuration.value))

async function loadData() {
  if (app.babies.length === 0) {
    await app.loadBabies()
  }
  const baby = app.currentBaby
  if (!baby) return
  selectedBabyId.value = baby.id
  const gen = ++loadGeneration
  try {
    const [statsRes, recordsRes, sleepRes, outdoorRes] = await Promise.all([
      babyAPI.stats(baby.id),
      recordAPI.list(baby.id),
      recordAPI.getCurrentSleep(baby.id),
      recordAPI.getCurrentOutdoor(baby.id),
    ])
    if (gen !== loadGeneration) return
    stats.value = statsRes.data
    allRecords.value = recordsRes.data as any[]
    currentSleep.value = sleepRes.data?.id ? sleepRes.data : null
    currentOutdoor.value = outdoorRes.data?.id ? outdoorRes.data : null
  } catch {
    app.showToast('数据加载失败', 'error')
  }
}

function switchBaby() {
  if (selectedBabyId.value) {
    app.setCurrentBaby(selectedBabyId.value)
    showAllRecords.value = false
    loadData()
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
    app.showToast(e?.response?.data?.error || '开始睡眠失败', 'error')
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
    app.showToast(e?.response?.data?.error || '结束睡眠失败', 'error')
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
    app.showToast(e?.response?.data?.error || '开始户外活动失败', 'error')
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
    app.showToast(e?.response?.data?.error || '结束户外活动失败', 'error')
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
  }
}

function onRecordDeleted(e: Event) {
  const { id, type } = (e as CustomEvent).detail || {}
  allRecords.value = allRecords.value.filter(r => !(r.id === id && r.record_type === (type || r.record_type)))
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
