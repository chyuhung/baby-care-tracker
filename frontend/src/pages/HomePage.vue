<template>
  <div class="flex flex-col min-h-screen">
    <!-- Header -->
    <header class="app-header pt-safe px-4 pb-3 border-b border-border-color">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-lg font-bold text-text-primary">
            {{ app.currentBaby?.name ? `${app.currentBaby?.name} 的记录` : '宝宝护理' }}
          </h1>
          <p v-if="app.currentBaby?.birth_date" class="text-xs text-text-secondary mt-0.5">
            {{ ageText }} · {{ todayDateText }}
          </p>
        </div>
        <div class="flex items-center gap-2">
          <span v-if="app.wsConnected" class="text-xs text-success flex items-center gap-1">
            <span class="w-2 h-2 bg-success rounded-full inline-block"></span>同步
          </span>
          <span v-else class="text-xs text-text-secondary">离线</span>
        </div>
      </div>

      <!-- 宝宝切换 -->
      <div v-if="app.currentBaby" class="mt-3 flex items-center gap-2">
        <select v-model="selectedBabyId" @change="switchBaby"
          class="flex-1 px-3 py-2 bg-white border border-border-color rounded-xl text-sm text-text-primary appearance-none cursor-pointer focus:border-primary focus:outline-none transition-colors">
          <option v-for="b in app.babies" :key="b.id" :value="b.id">{{ b.name }}</option>
        </select>
      </div>
    </header>

    <!-- Content -->
    <main class="flex-1 min-h-0 px-4 py-4 space-y-4 overflow-y-auto pb-20">
      <!-- 空状态：无宝宝 -->
      <div v-if="app.babies.length === 0" class="text-center py-16">
        <div class="text-5xl mb-4">👶</div>
        <p class="text-text-secondary mb-4">还没有添加宝宝</p>
        <router-link to="/baby/new"
          class="inline-flex items-center gap-2 px-5 py-2.5 bg-primary text-white rounded-xl font-medium text-sm btn-press shadow-card">
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
                <span class="text-3xl font-bold text-primary font-num">{{ stats.total_ml_today }}<sup v-if="stats.feeding_count > 0" class="text-[0.55em] font-bold text-primary font-num leading-none">{{ stats.feeding_count }}</sup></span>
                <span class="text-sm text-text-secondary">ml</span>
              </div>
              <div class="text-3xl">🍼</div>
            </div>
            <div v-if="lastFeedingAgo" class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastFeedingAgo.isLong ? 'text-orange-500' : 'text-text-secondary'">
                {{ lastFeedingAgo.text }}
              </span>
            </div>
            <!-- 新增喂奶入口 -->
            <button @click.stop="goToAddFeeding"
              class="mt-3 w-full py-2 bg-primary/10 text-primary text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1">
              <span class="text-base">＋</span> 喂奶
            </button>
          </div>

          <!-- 尿布卡片 -->
          <div @click="goToTimeline('diaper')" class="bg-white rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日尿布</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-1">
                <span class="text-3xl font-bold font-num text-diaper">{{ stats.diaper_count }}</span>
                <span class="text-sm text-text-secondary">次</span>
              </div>
              <div class="text-3xl">🩲</div>
            </div>
            <div v-if="lastDiaperAgo" class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium" :class="lastDiaperAgo.isLong ? 'text-orange-500' : 'text-text-secondary'">
                {{ lastDiaperAgo.text }}
              </span>
            </div>
            <!-- 新增尿布入口 -->
            <button @click.stop="goToAddDiaper"
              class="mt-3 w-full py-2 bg-diaper text-white text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1">
              <span class="text-base">＋</span> 尿布
            </button>
          </div>

          <!-- 睡眠卡片 -->
          <div @click="goToTimeline('sleep')" class="bg-white rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日睡眠</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-1">
                <span v-if="currentSleep" class="text-base font-bold text-primary">😴 正在睡觉</span>
                <template v-else>
                  <span class="text-3xl font-bold font-num text-primary">{{ formattedSleepDuration }}</span>
                  <span class="text-sm text-text-secondary ml-1">· {{ stats.sleep_count }}次</span>
                </template>
              </div>
              <div class="text-3xl">😴</div>
            </div>
            <div v-if="lastSleepAgo" class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium text-text-secondary">{{ lastSleepAgo.text }}</span>
            </div>
            <button v-if="currentSleep" @click.stop="stopSleep"
              class="mt-3 w-full py-2 bg-red-500 text-white text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1">
              <span>■</span> 停止睡觉
            </button>
            <button v-else @click.stop="startSleep"
              class="mt-3 w-full py-2 bg-primary/10 text-primary text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1">
              <span>😴</span> 开始睡觉
            </button>
          </div>

          <!-- 体温卡片 -->
          <div @click="goToTimeline('temperature')" class="bg-white rounded-2xl shadow-card p-4 cursor-pointer btn-press">
            <div class="text-xs text-text-secondary mb-1">今日体温</div>
            <div class="flex items-end justify-between">
              <div class="flex items-baseline gap-1">
                <span v-if="stats.latest_temperature > 0" class="text-3xl font-bold font-num" :class="stats.latest_temperature >= 37.5 ? 'text-red-500' : 'text-temperature'">{{ stats.latest_temperature }}</span>
                <span class="text-sm text-text-secondary">°C</span>
              </div>
              <div class="text-3xl">🌡️</div>
            </div>
            <div v-if="lastTempAgo" class="mt-2 flex items-center justify-between">
              <span class="text-xs text-text-secondary">距上次</span>
              <span class="text-xs font-medium text-text-secondary">{{ lastTempAgo.text }}</span>
            </div>
            <button @click.stop="goToAddTemperature"
              class="mt-3 w-full py-2 bg-temperature/10 text-temperature text-sm font-medium rounded-lg btn-press flex items-center justify-center gap-1">
              <span class="text-base">＋</span> 测温
            </button>
          </div>
        </div>

        <!-- 最近记录 -->
        <div class="space-y-2">
          <h2 class="text-sm font-semibold text-text-secondary uppercase tracking-wide">最近记录</h2>
          <div v-if="displayRecords.length === 0" class="bg-white rounded-2xl p-6 text-center shadow-card">
            <div class="text-4xl mb-2">🍼</div>
            <p class="text-text-secondary text-sm">还没有记录</p>
          </div>
          <RecordCard v-for="(r, i) in displayRecords" :key="r.record_type + '-' + r.id" :record="r"
            :style="{ animationDelay: `${i * 60}ms` }" class="card-in"
            @edit="editRecord(r)" @delete="deleteRecord(r)" />

          <!-- 加载更多 -->
          <button v-if="!showAllRecords && allRecords.length > displayRecords.length"
            @click="showAllRecords = true"
            class="w-full py-3 bg-white text-primary text-sm font-medium rounded-xl shadow-card btn-press mt-2">
            加载更多 ({{ allRecords.length - displayRecords.length }})
          </button>
        </div>
      </template>
    </main>

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/30 flex items-end z-50" @click.self="showDeleteConfirm = false">
      <div class="bg-white w-full rounded-t-2xl p-6 space-y-4 pb-safe animate-slide-up">
        <p class="text-text-secondary text-sm text-center">确定要删除这条记录吗？</p>
        <div class="flex gap-3">
          <button @click="showDeleteConfirm = false" class="flex-1 py-3 bg-gray-100 text-text-primary rounded-xl font-medium btn-press">取消</button>
          <button @click="confirmDelete" class="flex-1 py-3 bg-red-500 text-white rounded-xl font-medium btn-press">确认删除</button>
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
import RecordCard from '@/components/RecordCard.vue'

const tick = ref(0)
let tickTimer: number | null = null
const router = useRouter()
const app = useAppStore()
const stats = ref({ feeding_count: 0, diaper_count: 0, total_ml_today: 0, last_feeding: '', last_diaper: '', sleep_count: 0, sleep_duration: 0, last_sleep_end: '', temperature_count: 0, latest_temperature: 0, last_temperature: '' })
const allRecords = ref<any[]>([])
const showAllRecords = ref(false)
const showDeleteConfirm = ref(false)
const recordToDelete = ref<any>(null)
const currentSleep = ref<any>(null)
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
  const birth = new Date(+m[1], +m[2] - 1, +m[3], +(m[4] || 0), +(m[5] || 0))
  const now = new Date()
  const diff = Math.floor((now.getTime() - birth.getTime()) / (1000 * 60 * 60 * 24))
  if (diff < 0) return '未出生'
  const months = Math.floor(diff / 30)
  const days = diff % 30
  if (months > 0) return `${months}个月${days}天`
  return `${diff}天`
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

const lastFeedingAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_feeding) })
const lastDiaperAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_diaper) })
const lastSleepAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_sleep_end) })
const lastTempAgo = computed(() => { tick.value; return getTimeAgo(stats.value.last_temperature) })

const formattedSleepDuration = computed(() => {
  const mins = stats.value.sleep_duration
  if (mins <= 0) return '0'
  if (mins < 60) return `${mins}分钟`
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return m > 0 ? `${h}h${m}m` : `${h}小时`
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
    const [statsRes, recordsRes, sleepRes] = await Promise.all([
      babyAPI.stats(baby.id),
      recordAPI.list(baby.id),
      recordAPI.getCurrentSleep(baby.id),
    ])
    if (gen !== loadGeneration) return
    stats.value = statsRes.data
    allRecords.value = recordsRes.data as any[]
    currentSleep.value = sleepRes.data?.id ? sleepRes.data : null
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

async function startSleep() {
  const baby = app.currentBaby
  if (!baby) return
  try {
    const now = new Date().toISOString()
    const res = await recordAPI.createSleepStart(baby.id, { started_at: now })
    currentSleep.value = res.data
    window.dispatchEvent(new CustomEvent('record-created', { detail: res.data }))
    app.showToast('😴 开始睡觉', 'success')
  } catch {
    app.showToast('开始睡眠失败', 'error')
  }
}

async function stopSleep() {
  const baby = app.currentBaby
  if (!baby || !currentSleep.value) return
  try {
    const now = new Date().toISOString()
    await recordAPI.stopSleep(baby.id, currentSleep.value.id, { ended_at: now })
    currentSleep.value = null
    loadData()
    app.showToast('✅ 睡眠已结束', 'success')
  } catch {
    app.showToast('结束睡眠失败', 'error')
  }
}

function editRecord(r: any) {
  if (r.record_type === 'sleep') {
    router.push(`/sleep/${r.id}/edit`)
  } else if (r.record_type === 'temperature') {
    router.push(`/temperature/${r.id}/edit`)
  } else {
    router.push(`/record/${r.record_type}/${r.id}/edit`)
  }
}

function deleteRecord(r: any) {
  recordToDelete.value = r
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (!recordToDelete.value) return
  try {
    const { id, record_type: typ } = recordToDelete.value
    await recordAPI.delete(id, typ)
    window.dispatchEvent(new CustomEvent('record-deleted', { detail: { id, type: typ } }))
    app.showToast('✅ 已删除', 'success')
    showDeleteConfirm.value = false
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '删除失败', 'error')
  }
}

function onRecordCreated(e: Event) {
  const record = (e as CustomEvent).detail
  if (record) {
    allRecords.value.unshift(record)
    if (record.record_type === 'sleep' && record.data?.ended_at) {
      loadData()
    }
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
