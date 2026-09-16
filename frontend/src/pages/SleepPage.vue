<template>
  <div class="min-h-dvh bg-bg-main">
    <header class="pt-safe glass-surface px-4 py-3 border-b border-border-color flex items-center gap-3">
      <button aria-label="返回" @click="router.back()" class="p-2 -ml-2 flex items-center justify-center min-w-[44px] min-h-[44px] btn-press">
        <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
      </button>
      <h1 class="text-lg font-bold text-text-primary">{{ isEdit ? '编辑记录' : '😴 记录睡眠' }}</h1>
    </header>

    <main class="px-4 py-6 space-y-5">
      <!-- 编辑模式 -->
      <template v-if="isEdit">
        <div>
          <label class="text-sm text-text-secondary block mb-2">开始时间</label>
          <input v-model="editForm.started_at" type="datetime-local" class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-sm text-text-primary focus:border-primary focus:outline-none transition-colors" />
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">结束时间</label>
          <input v-model="editForm.ended_at" type="datetime-local" class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-sm text-text-primary focus:border-primary focus:outline-none transition-colors" />
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">备注</label>
          <textarea v-model="editForm.note" rows="3" placeholder="可选" class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-sm text-text-primary resize-none focus:border-primary focus:outline-none transition-colors" />
        </div>
        <button @click="saveEdit" :disabled="submitting" class="w-full py-3 bg-primary text-white rounded-xl font-semibold shadow-card btn-press disabled:opacity-50">{{ submitting ? '保存中...' : '更新记录' }}</button>
        <button @click="deleteRecord" class="w-full py-3 bg-white text-danger font-medium rounded-xl border border-danger/25 btn-press">删除此记录</button>
      </template>

      <!-- 删除确认弹窗 -->
      <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/30 flex items-end z-50" @click.self="showDeleteConfirm = false">
        <div class="bg-white w-full rounded-t-2xl p-6 space-y-4 pb-safe animate-slide-up">
          <p class="text-text-secondary text-sm text-center">确定要删除这条记录吗？</p>
          <div class="flex gap-3">
            <button @click="showDeleteConfirm = false" class="flex-1 py-3 bg-muted text-text-primary rounded-xl font-medium btn-press">取消</button>
            <button @click="confirmDelete" :disabled="submitting" class="flex-1 py-3 bg-danger text-white rounded-xl font-medium btn-press disabled:opacity-50">确认删除</button>
          </div>
        </div>
      </div>

      <!-- 非编辑模式 -->
      <template v-else>
        <!-- 今日摘要 -->
        <div class="bg-white rounded-2xl shadow-card p-5">
          <div class="text-xs text-text-secondary mb-1">今日睡眠总计</div>
          <div class="flex items-baseline gap-2">
            <span class="text-3xl font-bold text-sleep font-num">{{ formattedDuration }}</span>
            <span class="text-sm text-text-secondary">· {{ todaySleeps.length }}次</span>
          </div>
        </div>

        <!-- 计时器 -->
        <div class="bg-white rounded-2xl shadow-card p-5 text-center">
          <template v-if="currentSleep">
            <div class="text-lg text-text-primary mb-2">😴 正在睡觉</div>
            <div class="text-4xl font-bold text-sleep font-num mb-4">{{ elapsedText }}</div>
            <button @click="stopSleep" :disabled="submitting" class="w-full py-3 bg-danger text-white rounded-xl font-medium shadow-card btn-press flex items-center justify-center gap-2 disabled:opacity-50">
              <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M6 6h12v12H6z"/></svg>
              <span>{{ submitting ? '处理中...' : '结束' }}</span>
            </button>
          </template>
          <template v-else>
            <button @click="startSleep" :disabled="submitting" class="w-full py-3 bg-sleep/10 text-sleep rounded-xl font-medium btn-press flex items-center justify-center gap-2 disabled:opacity-50">
              <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
              开始
            </button>
          </template>
        </div>

        <!-- 今天记录 -->
        <div class="space-y-2">
          <h3 class="text-sm font-semibold text-text-secondary">今天睡眠记录</h3>
          <div v-if="todaySleeps.length === 0" class="bg-white rounded-2xl p-6 text-center shadow-card">
            <div class="text-4xl mb-2">😴</div>
            <p class="text-text-secondary text-sm">今天还没有睡眠记录</p>
          </div>
          <div v-for="s in todaySleeps" :key="s.id" @click="editSleep(s)"
            class="bg-white rounded-2xl p-4 shadow-card flex items-start gap-3 cursor-pointer btn-press">
            <div class="w-1.5 h-12 rounded-full bg-sleep flex-shrink-0"></div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between">
                <span class="text-sm font-semibold text-text-primary">😴 睡眠</span>
                <span class="text-xs text-text-secondary font-num">{{ formatSleepTime(s) }}</span>
              </div>
              <span class="text-xs bg-sleep/10 text-sleep px-2 py-0.5 rounded-full mt-1 inline-block">{{ formatDuration(s) }}</span>
              <div v-if="s.note" class="text-xs text-text-secondary mt-1">{{ s.note }}</div>
            </div>
          </div>
        </div>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { recordAPI } from '@/api'
import type { SleepRecord } from '@/api'
import { toLocalDatetime, formatDuration as fmtDuration } from '@/utils'

const router = useRouter()
const route = useRoute()
const app = useAppStore()

const isEdit = computed(() => !!route.params.id)
const currentSleep = ref<SleepRecord | null>(null)
const allSleeps = ref<any[]>([])
const tick = ref(0)
let tickTimer: number | null = null
const showDeleteConfirm = ref(false)
const submitting = ref(false)

const editForm = ref({ started_at: '', ended_at: '', note: '' })

const todaySleeps = computed(() => {
  const today = new Date().toDateString()
  return allSleeps.value.filter((s: any) => {
    const d = new Date(s.started_at || s.occurred_at)
    return d.toDateString() === today
  })
})

const formattedDuration = computed(() => {
  const mins = allSleeps.value.reduce((sum: number, s: any) => {
    if (!s.ended_at) return sum
    const start = new Date(s.started_at)
    const end = new Date(s.ended_at)
    return sum + Math.round((end.getTime() - start.getTime()) / 60000)
  }, 0)
  return fmtDuration(mins)
})

const elapsedText = computed(() => {
  tick.value
  if (!currentSleep.value?.started_at) return ''
  const start = new Date(currentSleep.value.started_at)
  const mins = Math.round((Date.now() - start.getTime()) / 60000)
  return fmtDuration(mins)
})

function formatSleepTime(s: any) {
  const pad = (n: number) => String(n).padStart(2, '0')
  const start = new Date(s.started_at)
  const end = s.ended_at ? new Date(s.ended_at) : null
  const hhmm1 = `${pad(start.getHours())}:${pad(start.getMinutes())}`
  if (!end) return hhmm1
  return `${hhmm1}~${pad(end.getHours())}:${pad(end.getMinutes())}`
}

function formatDuration(s: any) {
  if (!s.ended_at) return '进行中'
  const start = new Date(s.started_at)
  const end = new Date(s.ended_at)
  const mins = Math.round((end.getTime() - start.getTime()) / 60000)
  return fmtDuration(mins)
}

async function loadData() {
  const baby = app.currentBaby
  if (!baby) return
  try {
    const [recordsRes, sleepRes] = await Promise.all([
      recordAPI.list(baby.id),
      recordAPI.getCurrentSleep(baby.id),
    ])
    allSleeps.value = (recordsRes.data as any[]).filter(r => r.record_type === 'sleep')
    currentSleep.value = sleepRes.data?.id ? sleepRes.data : null

    if (isEdit.value) {
      const record = allSleeps.value.find((r: any) => r.id === Number(route.params.id))
      if (record) {
        editForm.value = {
          started_at: toLocalDatetime(record.data.started_at),
          ended_at: record.data.ended_at ? toLocalDatetime(record.data.ended_at) : '',
          note: record.data.note || '',
        }
      }
    }
  } catch {
    app.showToast('加载失败', 'error')
  }
}

async function startSleep() {
  const baby = app.currentBaby
  if (!baby || submitting.value) return
  submitting.value = true
  try {
    const now = new Date().toISOString()
    const res = await recordAPI.createSleepStart(baby.id, { started_at: now })
    currentSleep.value = res.data
    window.dispatchEvent(new CustomEvent('record-created', { detail: res.data }))
    app.showToast('😴 开始睡觉', 'success')
  } catch (e: any) {
    console.error('开始睡眠失败:', e?.response?.data?.error || e)
    app.showToast(e?.response?.data?.error || '开始睡眠失败', 'error')
  } finally {
    submitting.value = false
  }
}

async function stopSleep() {
  const baby = app.currentBaby
  if (!baby || !currentSleep.value || submitting.value) return
  submitting.value = true
  try {
    const now = new Date().toISOString()
    await recordAPI.stopSleep(baby.id, currentSleep.value.id, { ended_at: now })
    currentSleep.value = null
    await loadData()
    app.showToast('✅ 睡眠已结束', 'success')
  } catch {
    app.showToast('结束睡眠失败', 'error')
  } finally {
    submitting.value = false
  }
}

async function saveEdit() {
  if (!route.params.id || submitting.value) return
  submitting.value = true
  try {
    const startedAt = new Date(editForm.value.started_at).toISOString()
    const endedAt = editForm.value.ended_at ? new Date(editForm.value.ended_at).toISOString() : ''
    await recordAPI.update(Number(route.params.id), 'sleep', {
      started_at: startedAt,
      ended_at: endedAt,
      note: editForm.value.note,
    })
    window.dispatchEvent(new CustomEvent('record-created', { detail: null }))
    app.showToast('✅ 已保存', 'success')
    router.back()
  } catch {
    app.showToast('保存失败', 'error')
  } finally {
    submitting.value = false
  }
}

function deleteRecord() {
  if (!route.params.id) return
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (!route.params.id || submitting.value) return
  submitting.value = true
  try {
    await recordAPI.delete(Number(route.params.id), 'sleep')
    window.dispatchEvent(new CustomEvent('record-deleted', { detail: { id: Number(route.params.id), type: 'sleep' } }))
    app.showToast('✅ 已删除', 'success')
    showDeleteConfirm.value = false
    router.back()
  } catch {
    app.showToast('删除失败', 'error')
  } finally {
    submitting.value = false
  }
}

function editSleep(s: any) {
  router.push(`/sleep/${s.id}/edit`)
}

onMounted(() => {
  loadData()
  tickTimer = window.setInterval(() => { tick.value++ }, 30000)
})
onUnmounted(() => {
  if (tickTimer !== null) clearInterval(tickTimer)
})
</script>
