<template>
  <div class="min-h-screen bg-bg-main">
    <header class="pt-safe bg-white px-4 py-3 border-b border-border-color flex items-center gap-3">
      <button @click="router.back()" class="p-1 -ml-1 btn-press">
        <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
      </button>
      <h1 class="text-lg font-bold text-text-primary">{{ isEdit ? '编辑体温' : '体温记录' }}</h1>
    </header>

    <main class="px-4 py-6 space-y-5">
      <!-- 编辑模式 -->
      <template v-if="isEdit">
        <div>
          <label class="text-sm text-text-secondary block mb-2">体温 (°C)</label>
          <input v-model="editForm.temperature" type="number" step="0.1" min="35" max="42" class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-sm text-text-primary focus:border-primary focus:outline-none" />
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">测量位置</label>
          <div class="grid grid-cols-3 gap-2">
            <button v-for="loc in locations" :key="loc" @click="editForm.location = loc"
              :class="['py-2 rounded-xl text-sm font-medium btn-press', editForm.location === loc ? 'bg-temperature text-white' : 'bg-white border border-border-color text-text-secondary']">{{ loc }}</button>
          </div>
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">时间</label>
          <input v-model="editForm.occurred_at" type="datetime-local" class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-sm text-text-primary focus:border-primary focus:outline-none" />
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">备注</label>
          <textarea v-model="editForm.note" rows="3" placeholder="添加备注..." class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-sm text-text-primary resize-none focus:border-primary focus:outline-none" />
        </div>
        <div class="flex gap-3">
          <button @click="saveEdit" class="flex-1 py-3 bg-primary text-white rounded-xl font-medium btn-press">保存</button>
          <button @click="deleteRecord" class="py-3 px-6 bg-red-500 text-white rounded-xl font-medium btn-press">删除</button>
        </div>
      </template>

      <!-- 非编辑模式 -->
      <template v-else>
        <!-- 最新体温 -->
        <div class="bg-white rounded-2xl shadow-card p-5 text-center">
          <div class="text-xs text-text-secondary mb-2">最新体温</div>
          <div v-if="latestTemp" class="flex items-center justify-center gap-2">
            <span class="text-5xl font-bold font-num" :class="latestTemp.temperature >= 37.5 ? 'text-red-500' : 'text-temperature'">{{ latestTemp.temperature }}</span>
            <span class="text-xl text-text-secondary">°C</span>
            <span v-if="latestTemp.temperature >= 37.5" class="text-red-500 text-lg">🔥</span>
          </div>
          <div v-else class="text-text-secondary text-sm py-4">暂无体温记录</div>
          <div v-if="latestTemp" class="text-xs text-text-secondary mt-2">
            {{ formatTime(latestTemp.occurred_at) }}
            <span v-if="latestTemp.location"> · {{ latestTemp.location }}</span>
          </div>
        </div>

        <!-- 快速记录 -->
        <div class="bg-white rounded-2xl shadow-card p-5 space-y-4">
          <h3 class="text-sm font-semibold text-text-secondary">快速记录</h3>
          <div>
            <label class="text-xs text-text-secondary block mb-1">体温</label>
            <input v-model="form.temperature" type="number" step="0.1" min="35" max="42" placeholder="37.0"
              class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-sm text-text-primary font-num focus:border-primary focus:outline-none" />
          </div>
          <div>
            <label class="text-xs text-text-secondary block mb-1">测量位置</label>
            <div class="grid grid-cols-3 gap-2">
              <button v-for="loc in locations" :key="loc" @click="form.location = loc"
                :class="['py-2 rounded-xl text-sm font-medium btn-press', form.location === loc ? 'bg-temperature text-white' : 'bg-white border border-border-color text-text-secondary']">{{ loc }}</button>
            </div>
          </div>
          <div>
            <label class="text-xs text-text-secondary block mb-1">备注</label>
            <input v-model="form.note" placeholder="可选备注" class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-sm text-text-primary focus:border-primary focus:outline-none" />
          </div>
          <button @click="submitTemperature" :disabled="!form.temperature"
            class="w-full py-3 bg-temperature text-white rounded-xl font-medium btn-press flex items-center justify-center gap-1 disabled:opacity-50">
            <span>🌡️</span> 记录体温
          </button>
        </div>

        <!-- 最近体温记录 -->
        <div class="space-y-2">
          <h3 class="text-sm font-semibold text-text-secondary">最近记录</h3>
          <div v-if="allTemps.length === 0" class="bg-white rounded-2xl p-6 text-center shadow-card">
            <p class="text-text-secondary text-sm">还没有体温记录</p>
          </div>
          <div v-for="t in allTemps" :key="t.id" @click="editTemp(t)" :data="t.data"
            class="bg-white rounded-2xl p-4 shadow-card flex items-start gap-3 cursor-pointer btn-press">
            <div class="w-1.5 h-12 rounded-full bg-temperature flex-shrink-0"></div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between">
                <span class="text-sm font-semibold text-text-primary">
                  🌡️ {{ t.data?.temperature }}°C
                  <span v-if="t.data?.temperature >= 37.5" class="text-red-500">🔥</span>
                </span>
                <span class="text-xs text-text-secondary font-num">{{ formatTime(t.data?.occurred_at || t.occurred_at) }}</span>
              </div>
              <div class="text-xs text-text-secondary mt-1 flex gap-2">
                <span v-if="t.data?.location" class="bg-gray-100 text-text-secondary px-2 py-0.5 rounded-full">{{ t.data?.location }}</span>
                <span v-if="t.data?.note" class="truncate">{{ t.data?.note }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { recordAPI } from '@/api'

const router = useRouter()
const route = useRoute()
const app = useAppStore()

const isEdit = computed(() => !!route.params.id)
const allTemps = ref<any[]>([])
const latestTemp = ref<any>(null)

const locations = ['腋下', '口腔', '耳温', '额温', '肛门']

const form = ref({ temperature: 0, location: '', note: '' })
const editForm = ref({ temperature: 0, location: '腋下', occurred_at: '', note: '' })

function formatTime(iso: string) {
  if (!iso) return ''
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  const now = new Date()
  const isToday = d.toDateString() === now.toDateString()
  const hhmm = `${pad(d.getHours())}:${pad(d.getMinutes())}`
  if (isToday) return `今天 ${hhmm}`
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${hhmm}`
}

function toLocalDatetime(iso: string) {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function loadData() {
  const baby = app.currentBaby
  if (!baby) return
  try {
    const res = await recordAPI.list(baby.id)
    const temps = (res.data as any[]).filter(r => r.record_type === 'temperature')
    allTemps.value = temps
    if (temps.length > 0) {
      latestTemp.value = { ...temps[0].data, id: temps[0].id }
    }

    if (isEdit.value) {
      const record = temps.find((r: any) => r.id === Number(route.params.id))
      if (record) {
        const d = record.data
        editForm.value = {
          temperature: d.temperature,
          location: d.location || '腋下',
          occurred_at: toLocalDatetime(d.occurred_at),
          note: d.note || '',
        }
      }
    }
  } catch {
    app.showToast('加载失败', 'error')
  }
}

async function submitTemperature() {
  const baby = app.currentBaby
  if (!baby || !form.value.temperature) return
  try {
    const now = new Date().toISOString()
    const res = await recordAPI.createTemperature(baby.id, {
      temperature: form.value.temperature,
      location: form.value.location,
      note: form.value.note,
      occurred_at: now,
    })
    window.dispatchEvent(new CustomEvent('record-created', { detail: res.data }))
    app.showToast('✅ 体温已记录', 'success')
    form.value = { temperature: 0, location: '', note: '' }
    loadData()
  } catch {
    app.showToast('记录体温失败', 'error')
  }
}

async function saveEdit() {
  if (!route.params.id) return
  try {
    const occurredAt = new Date(editForm.value.occurred_at).toISOString()
    await recordAPI.update(Number(route.params.id), 'temperature', {
      temperature: editForm.value.temperature,
      location: editForm.value.location,
      note: editForm.value.note,
      occurred_at: occurredAt,
    })
    window.dispatchEvent(new CustomEvent('record-created', { detail: null }))
    app.showToast('✅ 已保存', 'success')
    router.back()
  } catch {
    app.showToast('保存失败', 'error')
  }
}

async function deleteRecord() {
  if (!route.params.id) return
  try {
    await recordAPI.delete(Number(route.params.id), 'temperature')
    window.dispatchEvent(new CustomEvent('record-deleted', { detail: { id: Number(route.params.id), type: 'temperature' } }))
    app.showToast('✅ 已删除', 'success')
    router.back()
  } catch {
    app.showToast('删除失败', 'error')
  }
}

function editTemp(t: any) {
  router.push(`/temperature/${t.id}/edit`)
}

onMounted(() => { loadData() })
</script>
