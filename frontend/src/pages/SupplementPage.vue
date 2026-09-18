<template>
  <div class="flex flex-col h-dvh bg-bg-main">
    <PullRefresh class="flex-1 min-h-0"
      content-class="px-4 py-4 space-y-5 pb-[calc(6.5rem+env(safe-area-inset-bottom))]"
      :refresh="refreshAll">
    <template #header>
    <header class="sticky top-0 z-30 glass-surface hairline-bottom pt-safe px-4 py-3 flex items-center gap-3">
      <button aria-label="返回" @click="router.back()" class="p-2 -ml-2 flex items-center justify-center min-w-[44px] min-h-[44px] btn-press">
        <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
      </button>
      <h1 class="text-lg font-bold text-text-primary">{{ isEdit ? '编辑补剂' : '记录补剂' }}</h1>
    </header>
    </template>

      <!-- 补剂名称 -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">补剂名称</label>
        <div class="flex flex-wrap gap-2">
          <button v-for="opt in nameOptions" :key="opt"
            type="button"
            :class="['px-4 py-2.5 min-h-[44px] rounded-full text-sm font-medium transition-colors btn-press',
              form.name === opt ? 'bg-supplement text-white font-semibold' : 'bg-white border border-border-color text-text-secondary']"
            @click="selectName(opt)">
            {{ opt }}
          </button>
        </div>
        <input v-if="form.name === '其他'" v-model="customName" type="text" maxlength="20" placeholder="输入补剂名称"
          class="mt-2 w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
      </div>

      <!-- 剂量 -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">剂量</label>
        <input v-model.number="form.dosage_value" type="number" step="0.1" min="0" inputmode="decimal"
          placeholder="如 1 或 2.5"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-xl text-center font-num font-bold focus:border-primary focus:outline-none transition-colors" />
      </div>

      <!-- 剂量单位（iOS 分段控件） -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">剂量单位</label>
        <Segmented :model-value="form.dosage_unit" :options="unitOptions" compact
          @update:model-value="(v: string) => form.dosage_unit = v" />
      </div>

      <!-- 时间 -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">记录时间</label>
        <input v-model="form.occurred_at" type="datetime-local"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
      </div>

      <!-- 备注 -->
      <div>
        <label class="text-sm text-text-secondary block mb-2">备注</label>
        <textarea v-model="form.note" rows="3" placeholder="如：晚上睡前、随奶服用等"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary resize-none focus:border-primary focus:outline-none transition-colors"></textarea>
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

    <ConfirmSheet :open="showDelete" :loading="deleting" message="确定要删除这条补剂记录吗？删除后无法恢复。"
      @confirm="doDelete" @cancel="showDelete = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { babyAPI, recordAPI } from '@/api'
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

const nameOptions = ['维生素D3', '维生素AD', '钙', '铁', '锌', '益生菌', '鱼油', '其他']

const unitOptions = [
  { value: '滴', label: '滴' },
  { value: '片', label: '片' },
  { value: 'ml', label: 'ml' },
  { value: 'g', label: 'g' },
  { value: '包', label: '包' },
]

const form = reactive({
  occurred_at: nowLocalDatetime(),
  name: '维生素D3',
  dosage_value: null as number | null,
  dosage_unit: '滴',
  note: '',
})
const customName = ref('')

function selectName(opt: string) {
  form.name = opt
}

function effectiveName() {
  if (form.name === '其他') {
    const trimmed = customName.value.trim()
    return trimmed || '其他'
  }
  return form.name
}

async function loadLastSupplement() {
  const baby = app.currentBaby
  if (!baby) return
  try {
    const res = await babyAPI.latestSupplement(baby.id)
    const latest = res.data
    if (!latest.name) return
    form.name = nameOptions.includes(latest.name) ? latest.name : '其他'
    if (latest.name) customName.value = latest.name
    if (latest.dosage_value > 0) form.dosage_value = latest.dosage_value
    if (latest.dosage_unit) form.dosage_unit = latest.dosage_unit
    if (latest.note) form.note = latest.note
  } catch {
    // ignore
  }
}

async function loadRecord() {
  if (!isEdit.value) return
  const baby = app.currentBaby
  if (!baby) return
  try {
    const res = await recordAPI.list(baby.id, 'supplement', 90)
    const record = (res.data as any[]).find(r => r.id === Number(route.params.id))
    if (record) {
      form.occurred_at = toLocalDatetime(record.occurred_at)
      form.name = nameOptions.includes(record.data.name) ? record.data.name : '其他'
      customName.value = form.name === '其他' ? record.data.name : ''
      form.dosage_value = record.data.dosage_value > 0 ? record.data.dosage_value : null
      form.dosage_unit = record.data.dosage_unit || '滴'
      form.note = record.data.note || ''
    }
  } catch {
    app.showToast('加载失败', 'error')
    router.back()
  }
}

async function refreshAll() {
  if (isEdit.value) await loadRecord()
  else await loadLastSupplement()
}

async function save() {
  error.value = ''
  if (!form.occurred_at) { error.value = '请选择时间'; return }
  const baby = app.currentBaby
  if (!baby) { error.value = '请先添加宝宝'; return }

  saving.value = true
  try {
    const occurredAt = new Date(form.occurred_at).toISOString()
    const name = effectiveName()
    const payload = {
      name,
      dosage_value: form.dosage_value || 0,
      dosage_unit: form.dosage_unit || '',
      note: form.note,
      occurred_at: occurredAt,
    }
    if (isEdit.value) {
      await recordAPI.update(Number(route.params.id), 'supplement', payload)
    } else {
      await recordAPI.createSupplement(baby.id, payload)
    }
    window.dispatchEvent(new CustomEvent('record-created', { detail: null }))
    app.showToast(isEdit.value ? '已保存' : '补剂已记录', 'success')
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
    await recordAPI.delete(Number(route.params.id), 'supplement')
    window.dispatchEvent(new CustomEvent('record-deleted', { detail: { id: Number(route.params.id), type: 'supplement' } }))
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
  else loadLastSupplement()
})
</script>