<template>
  <div class="flex flex-col h-dvh bg-bg-main">
    <PullRefresh class="flex-1 min-h-0"
      content-class="px-4 py-4 space-y-5 pb-[calc(6.5rem+env(safe-area-inset-bottom))]"
      :refresh="refreshAll">
    <template #header>
    <NavBar :title="isEdit ? '编辑补剂' : '记录补剂'">
      <template #left>
        <button aria-label="返回" @click="router.back()" class="-ml-2 flex min-h-[44px] min-w-[44px] items-center justify-center btn-press">
          <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
        </button>
      </template>
    </NavBar>
    </template>

      <!-- 补剂名称 -->
      <div>
        <label class="text-[13px] text-text-secondary block mb-2">补剂名称</label>
        <div class="flex flex-wrap gap-2">
          <button v-for="opt in nameOptions" :key="opt"
            type="button"
            :class="['px-4 py-2.5 min-h-[44px] rounded-full text-sm font-medium transition-colors btn-press',
              form.name === opt ? 'bg-supplement-fill text-white font-semibold' : 'bg-surface border border-border-color text-text-secondary']"
            @click="selectName(opt)">
            {{ opt }}
          </button>
        </div>
        <input v-if="form.name === '其他'" v-model="customName" type="text" maxlength="20" placeholder="输入补剂名称" enterkeyhint="done" autocomplete="off"
          class="mt-2 w-full px-4 py-3 bg-surface border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
      </div>

      <!-- 剂量 -->
      <div>
        <label class="text-[13px] text-text-secondary block mb-2">剂量</label>
        <input v-model.number="form.dosage_value" type="number" step="0.1" min="0" inputmode="decimal"
          placeholder="如 1 或 2.5"
          class="w-full px-4 py-3 bg-surface border border-border-color rounded-xl text-xl text-center font-num font-semibold focus:border-primary focus:outline-none transition-colors" />
      </div>

      <!-- 剂量单位（iOS 分段控件） -->
      <div>
        <label class="text-[13px] text-text-secondary block mb-2">剂量单位</label>
        <Segmented :model-value="form.dosage_unit" :options="unitOptions" compact
          @update:model-value="(v: string) => form.dosage_unit = v" />
      </div>

      <!-- 时间 -->
      <div>
        <label class="text-[13px] text-text-secondary block mb-2">记录时间</label>
        <DateTimeField v-model="form.occurred_at" title="记录时间" aria-label="选择记录时间" />
      </div>

      <!-- 备注 -->
      <div>
        <label class="text-[13px] text-text-secondary block mb-2">备注</label>
        <textarea v-model="form.note" rows="3" placeholder="如：晚上睡前、随奶服用等"
          class="w-full px-4 py-3 bg-surface border border-border-color rounded-xl text-text-primary resize-none focus:border-primary focus:outline-none transition-colors"></textarea>
      </div>

      <div v-if="error" class="bg-danger-light text-danger text-sm px-4 py-2 rounded-xl text-center">
        {{ error }}
      </div>

      <!-- 删除（编辑态，留在内容区） -->
      <button v-if="isEdit" type="button" @click="showDelete = true"
        class="btn-press w-full py-3 bg-surface text-danger font-medium rounded-xl border border-danger/25 min-h-[44px]">
        删除此记录
      </button>
    </PullRefresh>

    <!-- 固定底部操作栏 -->
    <FormBar>
      <button type="button" @click="save" :disabled="saving || app.offline"
        class="btn-press w-full py-3.5 bg-primary/10 text-primary-deep font-semibold rounded-xl disabled:opacity-50 flex items-center justify-center gap-2">
        <ActivityIndicator v-if="saving" :size="20" class="text-primary-deep" />
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
import { babyAPI, recordAPI, writeErrorMessage } from '@/api'
import { nowLocalDatetime, toLocalDatetime } from '@/utils'
import Segmented from '@/components/Segmented.vue'
import DateTimeField from '@/components/DateTimeField.vue'
import FormBar from '@/components/FormBar.vue'
import NavBar from '@/components/NavBar.vue'
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
// 编辑态取数失败标记：加载失败后禁止保存，避免空表单覆盖真实记录
const loadFailed = ref(false)

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
  try {
    const res = await recordAPI.get(Number(route.params.id), 'supplement')
    const record = res.data as any
    form.occurred_at = toLocalDatetime(record.occurred_at)
    form.name = nameOptions.includes(record.data.name) ? record.data.name : '其他'
    customName.value = form.name === '其他' ? record.data.name : ''
    form.dosage_value = record.data.dosage_value > 0 ? record.data.dosage_value : null
    form.dosage_unit = record.data.dosage_unit || '滴'
    form.note = record.data.note || ''
    loadFailed.value = false
  } catch (e) {
    // 取不到记录时不再静默弹回：内联报错并锁定保存，防止用空白默认表单覆盖真实记录
    loadFailed.value = true
    error.value = writeErrorMessage(e, '记录加载失败')
  }
}

async function refreshAll() {
  if (isEdit.value) await loadRecord()
  else await loadLastSupplement()
}

async function save() {
  if (isEdit.value && loadFailed.value) {
    error.value = '记录未能加载，无法保存，请返回重试'
    return
  }
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
    window.dispatchEvent(new CustomEvent(isEdit.value ? 'record-updated' : 'record-created', { detail: null }))
    app.showToast(isEdit.value ? '已保存' : '补剂已记录', 'success')
    router.back()
  } catch (e: any) {
    app.showToast(writeErrorMessage(e, '保存失败'), 'error')
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
    app.showToast(writeErrorMessage(e, '删除失败'), 'error')
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