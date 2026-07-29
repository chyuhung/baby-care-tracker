<template>
  <div class="min-h-screen bg-bg-main">
    <header class="pt-safe bg-white px-4 py-3 border-b border-border-color flex items-center gap-3">
      <button @click="router.back()" class="p-1 -ml-1 btn-press">
        <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
      </button>
      <h1 class="text-lg font-bold text-text-primary">{{ isEdit ? '编辑记录' : '🌡️ 记录体温' }}</h1>
    </header>

    <main class="px-4 py-6 space-y-5">
      <!-- 编辑模式 -->
      <template v-if="isEdit">
        <div>
          <label class="text-sm text-text-secondary block mb-2">体温 (°C)</label>
          <input v-model="editForm.temperature" type="number" step="0.1" min="35" max="42" class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-sm text-text-primary focus:border-primary focus:outline-none transition-colors" />
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">测量位置</label>
          <div class="grid grid-cols-3 gap-3">
            <button v-for="loc in locations" :key="loc" @click="editForm.location = loc"
              :class="['py-4 rounded-xl text-sm font-medium transition-colors btn-press', editForm.location === loc ? 'bg-primary text-white' : 'bg-white border border-border-color text-text-secondary']">{{ loc }}</button>
          </div>
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">时间</label>
          <input v-model="editForm.occurred_at" type="datetime-local" class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-sm text-text-primary focus:border-primary focus:outline-none transition-colors" />
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">备注</label>
          <textarea v-model="editForm.note" rows="2" placeholder="可选" class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary resize-none focus:border-primary focus:outline-none transition-colors" />
        </div>
        <button @click="saveEdit" class="w-full py-3 bg-primary text-white rounded-xl font-semibold shadow-card btn-press">更新记录</button>
        <button @click="deleteRecord" class="w-full py-3 bg-white text-red-500 font-medium rounded-xl border border-red-200 btn-press">删除此记录</button>
      </template>

      <!-- 非编辑模式 -->
      <template v-else>
        <div>
          <label class="text-sm text-text-secondary block mb-2">发生时间</label>
          <input v-model="form.occurred_at" type="datetime-local"
            class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">体温</label>
          <input v-model="form.temperature" type="number" step="0.1" min="35" max="42" placeholder="37.0"
            class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-sm text-text-primary font-num focus:border-primary focus:outline-none transition-colors" />
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-3">测量位置</label>
          <div class="grid grid-cols-3 gap-3">
            <button v-for="loc in locations" :key="loc" @click="form.location = loc"
              :class="['py-4 rounded-xl text-sm font-medium transition-colors btn-press', form.location === loc ? 'bg-primary text-white' : 'bg-white border border-border-color text-text-secondary']">{{ loc }}</button>
          </div>
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">备注</label>
          <textarea v-model="form.note" rows="2" placeholder="可选" class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors resize-none"></textarea>
        </div>
      </template>

      <button @click="submitTemperature" :disabled="!form.temperature"
        class="w-full py-3 bg-primary text-white rounded-xl font-semibold shadow-card btn-press disabled:opacity-50">
        记录
      </button>
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

const locations = ['腋下', '口腔', '耳温', '额温', '肛门']

function nowDatetime() {
  const d = new Date()
  const y = d.getFullYear()
  const M = String(d.getMonth() + 1).padStart(2, '0')
  const D = String(d.getDate()).padStart(2, '0')
  const h = String(d.getHours()).padStart(2, '0')
  const m = String(d.getMinutes()).padStart(2, '0')
  return `${y}-${M}-${D}T${h}:${m}`
}

const form = ref({ temperature: 0, location: '', note: '', occurred_at: nowDatetime() })
const editForm = ref({ temperature: 0, location: '腋下', occurred_at: '', note: '' })

function toLocalDatetime(iso: string) {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function loadData() {
  const baby = app.currentBaby
  if (!baby) return
  try {
    if (isEdit.value) {
      const res = await recordAPI.list(baby.id)
      const temps = (res.data as any[]).filter(r => r.record_type === 'temperature')
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
    const occurredAt = new Date(form.value.occurred_at).toISOString()
    const res = await recordAPI.createTemperature(baby.id, {
      temperature: form.value.temperature,
      location: form.value.location,
      note: form.value.note,
      occurred_at: occurredAt,
    })
    window.dispatchEvent(new CustomEvent('record-created', { detail: res.data }))
    app.showToast('✅ 体温已记录', 'success')
    router.back()
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

onMounted(() => { loadData() })
</script>
