<template>
  <div class="flex flex-col min-h-dvh bg-bg-main">
    <NavBar :title="meta.title">
      <template #left>
        <button aria-label="返回" @click="router.back()" class="-ml-2 flex min-h-[44px] min-w-[44px] items-center justify-center btn-press">
          <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
        </button>
      </template>
    </NavBar>

    <main class="flex-1 px-4 py-6 space-y-5 mx-auto w-full max-w-[480px] pb-[calc(6.5rem+env(safe-area-inset-bottom))]">
      <div v-if="!loaded" class="flex justify-center py-20">
        <ActivityIndicator :size="28" class="text-text-secondary" />
      </div>
      <template v-else>
        <div>
          <label class="text-[13px] text-text-secondary block mb-2">开始时间</label>
          <DateTimeField v-model="form.started_at" title="开始时间" aria-label="选择开始时间" />
        </div>
        <div>
          <label class="text-[13px] text-text-secondary block mb-2">结束时间</label>
          <DateTimeField v-model="form.ended_at" title="结束时间" aria-label="选择结束时间" />
        </div>
        <div>
          <label class="text-[13px] text-text-secondary block mb-2">备注</label>
          <textarea v-model="form.note" rows="3" placeholder="可选"
            class="w-full px-4 py-3 bg-surface border border-border-color rounded-xl text-text-primary resize-none focus:border-primary focus:outline-none transition-colors"></textarea>
        </div>

        <div v-if="error" class="bg-danger-light text-danger text-sm px-4 py-2 rounded-xl text-center">{{ error }}</div>

        <button type="button" @click="showDelete = true"
          class="btn-press w-full py-3 bg-surface text-danger font-medium rounded-xl border border-danger/25">
          删除此记录
        </button>
      </template>
    </main>

    <FormBar>
      <button type="button" @click="save" :disabled="submitting || !loaded || app.offline"
        class="btn-press w-full py-3.5 bg-primary/10 text-primary-deep font-semibold rounded-xl disabled:opacity-50 flex items-center justify-center gap-2">
        <ActivityIndicator v-if="submitting" :size="20" class="text-primary-deep" />
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
import { recordAPI, writeErrorMessage, UpdateRecordData } from '@/api'
import { toLocalDatetime } from '@/utils'
import FormBar from './FormBar.vue'
import NavBar from './NavBar.vue'
import ConfirmSheet from './ConfirmSheet.vue'
import ActivityIndicator from './ActivityIndicator.vue'
import DateTimeField from './DateTimeField.vue'

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
// 取数失败标记：加载失败后禁止保存，避免空表单覆盖真实记录
const loadFailed = ref(false)
const form = ref({ started_at: '', ended_at: '', note: '' })

async function load() {
  try {
    const res = await recordAPI.get(Number(route.params.id), props.type)
    const record = res.data as any
    form.value = {
      started_at: toLocalDatetime(record.data.started_at),
      ended_at: record.data.ended_at ? toLocalDatetime(record.data.ended_at) : '',
      note: record.data.note || '',
    }
    loadFailed.value = false
  } catch (e) {
    // 单条端点拿不到记录（不存在/越权/离线）→ 内联报错并锁定保存
    loadFailed.value = true
    error.value = writeErrorMessage(e, '记录加载失败')
  } finally {
    loaded.value = true
  }
}

async function save() {
  if (loadFailed.value) {
    error.value = '记录未能加载，无法保存，请返回重试'
    return
  }
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
    window.dispatchEvent(new CustomEvent('record-updated', { detail: null }))
    app.showToast('已保存', 'success')
    router.back()
  } catch (e: any) {
    app.showToast(writeErrorMessage(e, '保存失败'), 'error')
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
  } catch (e: any) {
    app.showToast(writeErrorMessage(e, '删除失败'), 'error')
    showDelete.value = false
  } finally {
    deleting.value = false
  }
}

onMounted(load)
</script>
