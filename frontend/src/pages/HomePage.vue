<template>
  <div class="flex flex-col min-h-dvh">
    <!-- Header -->
    <header class="app-header pt-safe px-4 pb-3 border-b border-border-color">
      <div class="flex items-center justify-between gap-2">
        <div class="min-w-0">
          <h1 class="text-lg font-bold text-text-primary truncate">
            {{ app.currentBaby?.name ? `${app.currentBaby?.name} 的记录` : '宝宝护理' }}
          </h1>
          <p v-if="app.currentBaby?.birth_date" class="text-xs text-text-secondary mt-0.5 truncate">
            {{ ageText }} · {{ todayDateText }}
          </p>
        </div>
        <div class="flex items-center gap-2 flex-shrink-0">
          <span v-if="app.wsConnected" class="text-xs text-success flex items-center gap-1">
            <span class="w-2 h-2 bg-success rounded-full inline-block"></span>同步
          </span>
          <span v-else class="text-xs text-text-secondary">离线</span>
        </div>
      </div>

      <!-- 宝宝切换 -->
      <div v-if="app.currentBaby" class="mt-3 flex items-center gap-2">
        <div class="relative flex-1">
          <select v-model="selectedBabyId" @change="switchBaby"
            class="w-full px-3 py-2 bg-white border border-border-color rounded-xl text-sm text-text-primary appearance-none cursor-pointer focus:border-primary focus:outline-none transition-colors pr-8">
            <option v-for="b in app.babies" :key="b.id" :value="b.id">{{ b.name }}</option>
          </select>
          <svg class="w-4 h-4 text-text-secondary pointer-events-none absolute right-3 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
        </div>
      </div>
    </header>

    <!-- Content -->
    <PullRefresh class="flex-1 min-h-0" content-class="px-4 py-4 space-y-4 pb-[calc(5rem+env(safe-area-inset-bottom))]"
      :refresh="loadData" :load-more="loadMoreFromPull">
      <!-- 空状态：无宝宝 -->
      <div v-if="app.babies.length === 0" class="text-center py-16">
        <div class="text-5xl mb-4">👶</div>
        <p class="text-text-secondary mb-4">还没有添加宝宝</p>
        <router-link to="/baby/new"
          class="inline-flex items-center gap-2 px-5 py-2.5 bg-primary-deep text-white rounded-xl font-medium text-sm btn-press shadow-card">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
          添加宝宝
        </router-link>
      </div>

      <!-- 主内容 -->
      <template v-else>
        <!-- 统计卡片（可点击跳转） -->
        <div class="grid grid-cols-2 gap-3">
          <!-- 喂奶卡片 -->
          <div @click="goToTimeline('feeding')" class="bg-white rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日喂奶</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-0.5">
                <span class="text-3xl font-bold text-primary-deep font-num">{{ stats.total_ml_today }}<sup v-if="stats.feeding_count > 0" class="text-[0.55em] font-bold text-primary-deep font-num leading-none">{{ stats.feeding_count }}</sup></span>
                <span class="text-sm text-text-secondary">ml</span>
              </div>
              <div class="text-3xl">🍼</div>
            </div>
            <div v-if="lastFeedingAgo" class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastFeedingAgo.isLong ? 'text-warning' : 'text-text-secondary'">
                {{ lastFeedingAgo.text }}
              </span>
            </div>
            <div v-if="feedingAvgInterval" class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">平均间隔</span>
              <span class="text-xs font-medium text-text-secondary">{{ feedingAvgInterval }}</span>
            </div>
            <!-- 新增喂奶入口 -->
            <button @click.stop="goToAddFeeding"
              class="mt-3 w-full py-2 bg-primary/10 text-primary-deep text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1">
              <span class="text-base">＋</span> 喂奶
            </button>
          </div>

          <!-- 尿布卡片 -->
          <div @click="goToTimeline('diaper')" class="bg-white rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日尿布</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-1">
                <span class="text-3xl font-bold font-num text-diaper-deep">{{ stats.diaper_count }}</span>
                <span class="text-sm text-text-secondary">次</span>
              </div>
              <div class="text-3xl">🩲</div>
            </div>
            <div v-if="lastDiaperAgo" class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastDiaperAgo.isLong ? 'text-warning' : 'text-text-secondary'">
                {{ lastDiaperAgo.text }}
              </span>
            </div>
            <div v-if="diaperAvgInterval" class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">平均间隔</span>
              <span class="text-xs font-medium text-text-secondary">{{ diaperAvgInterval }}</span>
            </div>
            <!-- 新增尿布入口 -->
            <button @click.stop="goToAddDiaper"
              class="mt-3 w-full py-2 bg-diaper/10 text-diaper-deep text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1">
              <span class="text-base">＋</span> 尿布
            </button>
          </div>

          <!-- 睡眠卡片 -->
          <div @click="goToTimeline('sleep')" class="bg-white rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日睡眠</div>
            <div class="flex items-end justify-between">
              <div v-if="currentSleep" class="flex items-baseline gap-1 min-w-0">
                <span class="text-base font-bold text-sleep-deep truncate">已睡 {{ elapsedSleepCompact }}</span>
              </div>
              <div v-else class="flex items-baseline gap-0.5">
                <template v-for="(part, pi) in sleepDurationParts" :key="pi">
                  <span class="text-3xl font-bold font-num text-sleep">{{ part.val }}</span>
                  <span v-if="part.unit" class="text-sm text-text-secondary">{{ part.unit }}</span>
                </template>
                
              </div>
              <div class="text-3xl">😴</div>
            </div>
            <div v-if="lastSleepAgo" class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastSleepAgo.isLong ? 'text-warning' : 'text-text-secondary'">{{ lastSleepAgo.text }}</span>
            </div>
            <div v-if="sleepAvgDuration" class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">平均时长</span>
              <span class="text-xs font-medium text-text-secondary">{{ sleepAvgDuration }}</span>
            </div>
            <button v-if="currentSleep" @click.stop="stopSleep" :disabled="loadingAction === 'stop-sleep'"
              class="mt-3 w-full py-2 bg-danger text-white text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1 disabled:opacity-50">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M6 6h12v12H6z"/></svg>
              {{ loadingAction === 'stop-sleep' ? '处理中...' : '结束' }}
            </button>
            <button v-else @click.stop="startSleep" :disabled="loadingAction === 'start-sleep'"
              class="mt-3 w-full py-2 bg-sleep/10 text-sleep-deep text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1 disabled:opacity-50">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
              {{ loadingAction === 'start-sleep' ? '处理中...' : '开始' }}
            </button>
          </div>

          <!-- 体温卡片 -->
          <div @click="goToTimeline('temperature')" class="bg-white rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日体温</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-1">
                <span v-if="stats.latest_temperature > 0" class="text-3xl font-bold font-num" :class="stats.latest_temperature >= 37.5 ? 'text-danger' : 'text-temperature'">{{ stats.latest_temperature }}</span>
                <span class="text-sm text-text-secondary">°C</span>
              </div>
              <div class="text-3xl">🌡️</div>
            </div>
            <div v-if="lastTempAgo" class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastTempAgo.isLong ? 'text-warning' : 'text-text-secondary'">{{ lastTempAgo.text }}</span>
            </div>
            <div v-if="tempHighValue" class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">最高体温</span>
              <span class="text-xs font-medium" :class="+tempHighValue >= 37.5 ? 'text-danger' : 'text-text-secondary'">{{ tempHighValue }}°C</span>
            </div>
            <button @click.stop="goToAddTemperature"
              class="mt-3 w-full py-2 bg-temperature/10 text-temperature-deep text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1">
              <span class="text-base">＋</span> 测温
            </button>
          </div>

          <!-- 户外活动卡片 -->
          <div @click="goToTimeline('outdoor')" class="col-span-2 bg-white rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日户外活动</div>
            <div class="flex items-end justify-between">
              <div v-if="currentOutdoor" class="flex items-baseline gap-1 min-w-0">
                <span class="text-base font-bold text-outdoor-deep truncate">已活动 {{ elapsedOutdoorCompact }}</span>
              </div>
              <div v-else class="flex items-baseline gap-0.5">
                <template v-for="(part, pi) in outdoorDurationParts" :key="pi">
                  <span class="text-3xl font-bold font-num text-outdoor-deep">{{ part.val }}</span>
                  <span v-if="part.unit" class="text-sm text-text-secondary">{{ part.unit }}</span>
                </template>
              </div>
              <div class="text-3xl">🌳</div>
            </div>
            <div v-if="lastOutdoorAgo" class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastOutdoorAgo.isLong ? 'text-warning' : 'text-text-secondary'">{{ lastOutdoorAgo.text }}</span>
            </div>
            <div v-if="avgOutdoorDuration > 0" class="mt-1 flex items-center justify-between">
              <span class="text-xs text-text-secondary">平均时长</span>
              <span class="text-xs font-medium text-text-secondary">{{ formatAvgOutdoor }}</span>
            </div>
            <button v-if="currentOutdoor" @click.stop="stopOutdoor" :disabled="loadingAction === 'stop-outdoor'"
              class="mt-3 w-full py-2 bg-danger text-white text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1 disabled:opacity-50">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M6 6h12v12H6z"/></svg>
              {{ loadingAction === 'stop-outdoor' ? '处理中...' : '结束' }}
            </button>
            <button v-else @click.stop="startOutdoor" :disabled="loadingAction === 'start-outdoor'"
              class="mt-3 w-full py-2 bg-outdoor/10 text-outdoor-deep text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1 disabled:opacity-50">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
              {{ loadingAction === 'start-outdoor' ? '处理中...' : '开始' }}
            </button>
          </div>
        </div>

        <!-- 最近记录 -->
        <div class="space-y-2">
          <h2 class="text-sm font-semibold text-text-secondary uppercase tracking-wide">最近记录</h2>
          <div v-if="displayRecords.length === 0" class="bg-white rounded-2xl p-6 text-center shadow-card">
            <img src="/icon-192.png" alt="" class="w-12 h-12 mx-auto block mb-2" />
            <p class="text-text-secondary text-sm">还没有记录</p>
          </div>
          <RecordCard v-for="(r, i) in displayRecords" :key="r.record_type + '-' + r.id" :record="r"
            :style="{ animationDelay: `${i * 60}ms` }" class="card-in"
            @edit="editRecord(r)" @delete="deleteRecord(r)" />

          <!-- 加载更多 -->
          <button v-if="!showAllRecords && allRecords.length > displayRecords.length"
            @click="showAllRecords = true"
            class="w-full py-3 bg-white text-primary-deep text-sm font-medium rounded-xl shadow-card btn-press mt-2">
            加载更多 ({{ allRecords.length - displayRecords.length }})
          </button>
        </div>
      </template>
    </PullRefresh>

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/30 flex items-end z-50" @click.self="showDeleteConfirm = false">
      <div class="bg-white w-full rounded-t-2xl p-6 space-y-4 pb-safe animate-slide-up">
        <p class="text-text-secondary text-sm text-center">确定要删除这条记录吗？</p>
        <div class="flex gap-3">
          <button @click="showDeleteConfirm = false" class="flex-1 py-3 bg-muted text-text-primary rounded-xl font-medium btn-press">取消</button>
          <button @click="confirmDelete" :disabled="deleting" class="flex-1 py-3 bg-danger text-white rounded-xl font-medium btn-press disabled:opacity-50">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { babyAPI, recordAPI } from '@/api'
import type { BabyStats, SleepRecord, OutdoorRecord } from '@/api'
import RecordCard from '@/components/RecordCard.vue'
import PullRefresh from '@/components/PullRefresh.vue'
import { durationParts, formatDurationCompact } from '@/utils'

const tick = ref(0)
let tickTimer: number | null = null
const router = useRouter()
const app = useAppStore()
const stats = ref<BabyStats>({ feeding_count: 0, diaper_count: 0, total_ml_today: 0, last_feeding: '', last_diaper: '', sleep_count: 0, sleep_duration: 0, last_sleep_end: '', temperature_count: 0, latest_temperature: 0, last_temperature: '', outdoor_count: 0, outdoor_duration: 0, last_outdoor_end: '' })
const allRecords = ref<any[]>([])
const showAllRecords = ref(false)
const showDeleteConfirm = ref(false)
const recordToDelete = ref<any>(null)
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
  const m = baby.birth_date.match(/^(\d{4})-(\d{2})-(\d{2})(?:T(\d{2}):(\d{2}))?/)
  if (!m) return ''
  const birthYear = +m[1]
  const birthMonth = +m[2] - 1
  const birthDay = +m[3]
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

const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
const todayDateText = computed(() => {
  const d = new Date()
  return `${d.getMonth() + 1}月${d.getDate()}日 ${weekDays[d.getDay()]}`
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
  if (diffDays > 0) text = `${diffDays}天${diffHours % 24}小时前`
  else if (diffHours > 0) text = `${diffHours}小时${diffMins % 60}分钟前`
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

function formatInterval(mins: number) {
  if (mins < 60) return `${mins}分钟`
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return m > 0 ? `${h}小时${m}分` : `${h}小时`
}

const feedingAvgInterval = computed(() => {
  const m = avgIntervalMinutes(allRecords.value, 'feeding')
  return m == null ? null : formatInterval(m)
})

const diaperAvgInterval = computed(() => {
  const m = avgIntervalMinutes(allRecords.value, 'diaper')
  return m == null ? null : formatInterval(m)
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
  return formatInterval(Math.round(recs.reduce((sum, x) => sum + x.t, 0) / recs.length))
})

const tempHighValue = computed(() => {
  const recs = allRecords.value
    .filter(r => r.record_type === 'temperature' && r.data?.temperature > 0)
    .map(r => ({ v: r.data.temperature, occurred: new Date(r.occurred_at).getTime() }))
    .sort((a, b) => b.occurred - a.occurred)
    .slice(0, 3)
  if (!recs.length) return null
  return Math.max(...recs.map(x => x.v)).toFixed(1)
})

const lastFeedingAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_feeding) })
const lastDiaperAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_diaper) })
const lastSleepAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_sleep_end) })
const lastTempAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_temperature) })
const lastOutdoorAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_outdoor_end) })

const elapsedSleepCompact = computed(() => {
  tick.value
  if (!currentSleep.value?.started_at) return ''
  const start = new Date(currentSleep.value.started_at)
  const mins = Math.round((Date.now() - start.getTime()) / 60000)
  return formatDurationCompact(mins)
})

const sleepDurationParts = computed(() => durationParts(stats.value.sleep_duration))

const elapsedOutdoorCompact = computed(() => {
  tick.value
  if (!currentOutdoor.value?.started_at) return ''
  const start = new Date(currentOutdoor.value.started_at)
  const mins = Math.round((Date.now() - start.getTime()) / 60000)
  return formatDurationCompact(mins)
})

const outdoorDurationParts = computed(() => durationParts(stats.value.outdoor_duration))

const avgOutdoorDuration = computed(() => stats.value.outdoor_count > 0
  ? Math.round(stats.value.outdoor_duration / stats.value.outdoor_count)
  : 0)

const formatAvgOutdoor = computed(() => {
  const parts = durationParts(avgOutdoorDuration.value)
  return parts.map(p => p.val + (p.unit || '')).join(' ')
})

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

function loadMoreFromPull() {
  if (!showAllRecords.value) showAllRecords.value = true
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

async function startSleep() {
  const baby = app.currentBaby
  if (!baby || loadingAction.value) return
  loadingAction.value = 'start-sleep'
  try {
    const now = new Date().toISOString()
    const res = await recordAPI.createSleepStart(baby.id, { started_at: now })
    currentSleep.value = res.data
    window.dispatchEvent(new CustomEvent('record-created', { detail: res.data }))
    app.showToast('😴 开始睡觉', 'success')
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
    app.showToast('✅ 睡眠已结束', 'success')
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
    currentOutdoor.value = res.data
    window.dispatchEvent(new CustomEvent('record-created', { detail: res.data }))
    app.showToast('🌳 开始户外活动', 'success')
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
    app.showToast('✅ 户外活动已结束', 'success')
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
  deleting.value = true
  try {
    const { id, record_type: typ } = recordToDelete.value
    await recordAPI.delete(id, typ)
    window.dispatchEvent(new CustomEvent('record-deleted', { detail: { id, type: typ } }))
    app.showToast('✅ 已删除', 'success')
    showDeleteConfirm.value = false
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '删除失败', 'error')
  } finally {
    deleting.value = false
  }
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
