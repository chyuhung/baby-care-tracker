<template>
  <div class="flex flex-col min-h-dvh bg-bg-main">
    <header class="pt-safe px-4 py-3 flex items-center gap-3">
      <button aria-label="返回" @click="router.back()" class="p-2 -ml-2 flex items-center justify-center min-w-[44px] min-h-[44px] btn-press">
        <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
      </button>
      <h1 class="text-lg font-bold text-text-primary">{{ isEdit ? '编辑体温' : '记录体温' }}</h1>
    </header>

    <PullRefresh class="flex-1 min-h-0"
      content-class="px-4 py-4 space-y-5 pb-[calc(6.5rem+env(safe-area-inset-bottom))]"
      :refresh="refreshAll">
      <!-- 时间 -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">测量时间</label>
        <input v-model="form.occurred_at" type="datetime-local"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
      </div>

      <!-- 体温 -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">体温（°C）</label>
        <input v-model.number="form.temperature" type="number" step="0.1" min="30" max="45" inputmode="decimal"
          placeholder="36.5"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-3xl text-center font-num font-bold focus:border-primary focus:outline-none transition-colors" />
      </div>

      <!-- 测量位置（iOS 分段控件，单行 5 段） -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">测量位置</label>
        <Segmented :model-value="form.location" :options="locationOptions" compact
          @update:model-value="(v: string) => form.location = v" />
      </div>

      <!-- 备注 -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">备注</label>
        <textarea v-model="form.note" rows="3" placeholder="如：吃奶后、哭闹等"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary resize-none focus:border-primary focus:outline-none transition-colors"></textarea>
      </div>

      <!-- 发烧提示 -->
      <div v-if="form.temperature && form.temperature >= 37.5"
        class="bg-danger-light text-danger text-sm px-4 py-3 rounded-xl flex items-center gap-2">
        <span>🔥</span>
        <span>体温偏高，请注意观察并考虑就医</span>
      </div>

      <div v-if="error" class="bg-danger-light text-danger text-sm px-4 py-2 rounded-xl text-center">
        {{ error }}
      </div>

      <!-- 删除（编辑态，留在内容区） -->
      <button v-if="isEdit" type="button" @click="showDelete = true"
        class="btn-press w-full py-3 bg-white text-danger font-medium rounded-xl border border-danger/25 min-h-[44px]">
        删除此记录
      </button>
    </PullRefresh>

    <!-- 固定底部操作栏 -->
    <FormBar>
      <button type="button" @click="save" :disabled="saving"
        class="btn-press w-full py-3.5 bg-primary-deep text-white font-semibold rounded-xl shadow-card disabled:opacity-50 flex items-center justify-center gap-2">
        <ActivityIndicator v-if="saving" :size="20" class="text-white" />
        <span>{{ saving ? '保存中...' : (isEdit ? '更新记录' : '记录') }}</span>
      </button>
    </FormBar>

    <ConfirmSheet :open="showDelete" :loading="deleting" message="确定要删除这条体温记录吗？删除后无法恢复。"
      @confirm="doDelete" @cancel="showDelete = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { recordAPI } from '@/api'
import { nowLocalDatetime, toLocalDatetime } from '@/utils'
import Segmented from '@/components/Segmented.vue'
import FormBar from '@/components/FormBar.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'
import ActivityIndicator from '@/components/ActivityIndicator.vue'
import PullRefresh from '@/components/PullRefresh.vue'

const router = useRouter()
const route = useRoute()
const app = useAppStore()

const isEdit = computed(() => !!route.params.id)
const saving = ref(false)
const deleting = ref(false)
const showDelete = ref(false)
const error = ref('')

const locationOptions = [
  { value: '腋下', label: '腋下' },
  { value: '口腔', label: '口腔' },
  { value: '耳温', label: '耳温' },
  { value: '额温', label: '额温' },
  { value: '肛门', label: '肛门' },
]

const form = reactive({
  occurred_at: nowLocalDatetime(),
  temperature: null as number | null,
  location: '腋下',
  note: '',
})

async function loadLastTemperature() {
  const baby = app.currentBaby
  if (!baby) return
  try {
    const res = await recordAPI.list(baby.id, 'temperature', 30)
    const records = res.data as any[]
    if (records.length > 0) {
      const latest = records[0]
      form.location = latest.data.location || '腋下'
      if (latest.data.note) form.note = latest.data.note
    }
  } catch {
    // ignore
  }
}

async function loadRecord() {
  if (!isEdit.value) return
  const baby = app.currentBaby
  if (!baby) return
  try {
    const res = await recordAPI.list(baby.id, 'temperature', 90)
    const record = (res.data as any[]).find(r => r.id === Number(route.params.id))
    if (record) {
      form.occurred_at = toLocalDatetime(record.occurred_at)
      form.temperature = record.data.temperature
      form.location = record.data.location || '腋下'
      form.note = record.data.note || ''
    }
  } catch {
    app.showToast('加载失败', 'error')
    router.back()
  }
}

async function refreshAll() {
  if (isEdit.value) await loadRecord()
  else await loadLastTemperature()
}

async function save() {
  error.value = ''
  if (!form.temperature || form.temperature < 30 || form.temperature > 45) {
    error.value = '请输入正确的体温（30-45°C）'
    return
  }
  if (!form.occurred_at) { error.value = '请选择时间'; return }
  const baby = app.currentBaby
  if (!baby) { error.value = '请先添加宝宝'; return }

  saving.value = true
  try {
    const occurredAt = new Date(form.occurred_at).toISOString()
    const payload = {
      temperature: form.temperature!,
      location: form.location,
      note: form.note,
      occurred_at: occurredAt,
    }
    if (isEdit.value) {
      await recordAPI.update(Number(route.params.id), 'temperature', payload)
    } else {
      await recordAPI.createTemperature(baby.id, payload)
      try { localStorage.setItem('temp_last_location', form.location) } catch { /* ignore */ }
    }
    window.dispatchEvent(new CustomEvent('record-created', { detail: null }))
    app.showToast(isEdit.value ? '已保存' : '体温已记录', 'success')
    router.back()
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '保存失败', 'error')
  } finally {
    saving.value = false
  }
}

async function doDelete() {
  if (!isEdit.value || deleting.value) return
  deleting.value = true
  try {
    await recordAPI.delete(Number(route.params.id), 'temperature')
    window.dispatchEvent(new CustomEvent('record-deleted', { detail: { id: Number(route.params.id), type: 'temperature' } }))
    app.showToast('已删除', 'success')
    router.back()
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '删除失败', 'error')
    showDelete.value = false
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  if (isEdit.value) loadRecord()
  else {
    try {
      const savedLocation = localStorage.getItem('temp_last_location')
      if (savedLocation) form.location = savedLocation
    } catch { /* ignore */ }
    loadLastTemperature()
  }
})
</script>
