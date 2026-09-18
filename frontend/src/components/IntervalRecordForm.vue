<template>
  <div class="flex flex-col min-h-dvh bg-bg-main">
    <header class="sticky top-0 z-30 glass-surface hairline-bottom pt-safe px-4 py-3 flex items-center gap-3">
      <button aria-label="返回" @click="router.back()" class="p-2 -ml-2 flex items-center justify-center min-w-[44px] min-h-[44px] btn-press">
        <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
      </button>
      <h1 class="text-lg font-bold text-text-primary">{{ meta.title }}</h1>
    </header>

    <main class="flex-1 px-4 py-6 space-y-5 pb-[calc(6.5rem+env(safe-area-inset-bottom))]">
      <div v-if="!loaded" class="flex justify-center py-20">
        <ActivityIndicator :size="28" class="text-text-secondary" />
      </div>
      <template v-else>
        <div>
          <label class="text-sm text-text-secondary block mb-2">开始时间</label>
          <input v-model="form.started_at" type="datetime-local"
            class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">结束时间</label>
          <input v-model="form.ended_at" type="datetime-local"
            class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
        </div>
        <div>
          <label class="text-sm text-text-secondary block mb-2">备注</label>
          <textarea v-model="form.note" rows="3" placeholder="可选"
            class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary resize-none focus:border-primary focus:outline-none transition-colors"></textarea>
        </div>

        <div v-if="error" class="bg-danger-light text-danger text-sm px-4 py-2 rounded-xl text-center">{{ error }}</div>

        <button type="button" @click="showDelete = true"
          class="btn-press w-full py-3 bg-white text-danger font-medium rounded-xl border border-danger/25">
          删除此记录
        </button>
      </template>
    </main>

    <FormBar>
      <button type="button" @click="save" :disabled="submitting || !loaded"
        class="btn-press w-full py-3.5 bg-primary-deep text-white font-semibold rounded-xl shadow-card disabled:opacity-50 flex items-center justify-center gap-2">
        <ActivityIndicator v-if="submitting" :size="20" class="text-white" />
        <span>{{ submitting ? '保存中...' : '更新记录' }}</span>
      </button>
    </FormBar>

    <ConfirmSheet :open="showDelete" :loading="deleting" message="确定要删除这条记录吗？删除后无法恢复。"
      @confirm="doDelete" @cancel="showDelete = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { recordAPI, UpdateRecordData } from '@/api'
import { toLocalDatetime } from '@/utils'
import FormBar from './FormBar.vue'
import ConfirmSheet from './ConfirmSheet.vue'
import ActivityIndicator from './ActivityIndicator.vue'

const props = defineProps<{ type: 'sleep' | 'outdoor' }>()

const META = {
  sleep: { title: '编辑睡眠' },
  outdoor: { title: '编辑户外活动' },
} as const

const router = useRouter()
const route = useRoute()
const app = useAppStore()

const meta = computed(() => META[props.type])
const loaded = ref(false)
const submitting = ref(false)
const deleting = ref(false)
const showDelete = ref(false)
const error = ref('')
const form = ref({ started_at: '', ended_at: '', note: '' })

async function load() {
  const baby = app.currentBaby
  if (!baby) { loaded.value = true; return }
  try {
    const res = await recordAPI.list(baby.id)
    const record = (res.data as any[]).find(r => r.id === Number(route.params.id) && r.record_type === props.type)
    if (record) {
      form.value = {
        started_at: toLocalDatetime(record.data.started_at),
        ended_at: record.data.ended_at ? toLocalDatetime(record.data.ended_at) : '',
        note: record.data.note || '',
      }
    }
  } catch {
    app.showToast('加载失败', 'error')
  } finally {
    loaded.value = true
  }
}

async function save() {
  error.value = ''
  if (!form.value.started_at) { error.value = '请选择开始时间'; return }
  if (form.value.ended_at && form.value.ended_at < form.value.started_at) {
    error.value = '结束时间不能早于开始时间'
    return
  }
  submitting.value = true
  try {
    const payload: UpdateRecordData = {
      started_at: new Date(form.value.started_at).toISOString(),
      note: form.value.note,
    }
    if (form.value.ended_at) payload.ended_at = new Date(form.value.ended_at).toISOString()
    await recordAPI.update(Number(route.params.id), props.type, payload)
    window.dispatchEvent(new CustomEvent('record-created', { detail: null }))
    app.showToast('已保存', 'success')
    router.back()
  } catch {
    app.showToast('保存失败', 'error')
  } finally {
    submitting.value = false
  }
}

async function doDelete() {
  deleting.value = true
  try {
    await recordAPI.delete(Number(route.params.id), props.type)
    window.dispatchEvent(new CustomEvent('record-deleted', { detail: { id: Number(route.params.id), type: props.type } }))
    app.showToast('已删除', 'success')
    router.back()
  } catch {
    app.showToast('删除失败', 'error')
    showDelete.value = false
  } finally {
    deleting.value = false
  }
}

onMounted(load)
</script>
