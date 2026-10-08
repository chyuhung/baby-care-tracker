<template>
  <div class="flex flex-col min-h-dvh bg-bg-main">
    <NavBar :title="pageTitle">
      <template #left>
        <button aria-label="返回" @click="router.back()" class="-ml-2 flex min-h-[44px] min-w-[44px] items-center justify-center btn-press">
          <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
        </button>
      </template>
    </NavBar>

    <main class="flex-1 px-4 py-6 space-y-5 mx-auto w-full max-w-[480px] pb-[calc(6.5rem+env(safe-area-inset-bottom))]">
      <!-- 时间 -->
      <div>
        <label class="text-[13px] text-text-secondary block mb-2">记录时间</label>
        <DateTimeField v-model="form.occurred_at" title="记录时间" aria-label="选择记录时间" />
      </div>

      <!-- 喂奶 -->
      <template v-if="recordType === 'feeding'">
        <div>
          <label class="text-[13px] text-text-secondary block mb-2">喂奶方式</label>
          <Segmented :model-value="feedingForm.type" :options="feedingOptions"
            @update:model-value="(v: string) => feedingForm.type = v" />
        </div>

        <!-- 母乳亲喂：时长 + 侧 -->
        <template v-if="feedingForm.type === 'breast'">
          <div>
            <label class="text-[13px] text-text-secondary block mb-2">时长（分钟）</label>
            <input v-model.number="feedingForm.duration_minutes" type="number" min="0" inputmode="numeric" placeholder="如 15"
              class="w-full px-4 py-3 bg-surface border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
          </div>
          <div>
            <label class="text-[13px] text-text-secondary block mb-2">喂养侧</label>
            <Segmented :model-value="feedingForm.side" :options="sideOptions"
              @update:model-value="(v: string) => feedingForm.side = v" />
          </div>
        </template>

        <!-- 瓶喂：奶量 -->
        <template v-else>
          <div>
            <label class="text-[13px] text-text-secondary block mb-2">奶量（ml）</label>
            <input v-model.number="feedingForm.amount_ml" type="number" min="0" inputmode="numeric" placeholder="如 120"
              class="w-full px-4 py-3 bg-surface border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
          </div>
          <div v-if="feedingForm.type === 'formula'">
            <label class="text-[13px] text-text-secondary block mb-2">品牌（可选）</label>
            <input v-model="feedingForm.brand" type="text" placeholder="奶粉品牌" enterkeyhint="done" autocomplete="off"
              class="w-full px-4 py-3 bg-surface border border-border-color rounded-xl text-text-primary focus:border-primary focus:outline-none transition-colors" />
          </div>
        </template>
      </template>

      <!-- 尿布：大卡片选择（保留彩色 emoji） -->
      <template v-else>
        <div>
          <label class="text-[13px] text-text-secondary block mb-2">尿布类型</label>
          <div class="grid grid-cols-3 gap-3">
            <button type="button" @click="diaperForm.type = 'pee'"
              :class="['flex flex-col items-center justify-center gap-1 py-4 rounded-xl border-2 transition-all btn-press min-h-[88px]',
                diaperForm.type === 'pee' ? 'border-primary bg-primary/5' : 'border-border-color bg-surface']">
              <span class="text-2xl">💧</span>
              <span class="text-sm font-medium" :class="diaperForm.type === 'pee' ? 'text-primary-deep' : 'text-text-secondary'">小便</span>
            </button>
            <button type="button" @click="diaperForm.type = 'poop'"
              :class="['flex flex-col items-center justify-center gap-1 py-4 rounded-xl border-2 transition-all btn-press min-h-[88px]',
                diaperForm.type === 'poop' ? 'border-primary bg-primary/5' : 'border-border-color bg-surface']">
              <span class="text-2xl">💩</span>
              <span class="text-sm font-medium" :class="diaperForm.type === 'poop' ? 'text-primary-deep' : 'text-text-secondary'">大便</span>
            </button>
            <button type="button" @click="diaperForm.type = 'mixed'"
              :class="['flex flex-col items-center justify-center gap-1 py-4 rounded-xl border-2 transition-all btn-press min-h-[88px]',
                diaperForm.type === 'mixed' ? 'border-primary bg-primary/5' : 'border-border-color bg-surface']">
              <span class="text-2xl">🌪️</span>
              <span class="text-sm font-medium" :class="diaperForm.type === 'mixed' ? 'text-primary-deep' : 'text-text-secondary'">混合</span>
            </button>
          </div>
        </div>
      </template>

      <!-- 备注 -->
      <div>
        <label class="text-[13px] text-text-secondary block mb-2">备注</label>
        <textarea v-model="form.note" rows="3" placeholder="可选"
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
    </main>

    <!-- 固定底部操作栏 -->
    <FormBar>
      <button type="button" @click="save" :disabled="saving || app.offline"
        class="btn-press w-full py-3.5 bg-primary/10 text-primary-deep font-semibold rounded-xl disabled:opacity-50 flex items-center justify-center gap-2">
        <ActivityIndicator v-if="saving" :size="20" class="text-primary-deep" />
        <span>{{ saving ? '保存中...' : (isEdit ? '更新记录' : '记录') }}</span>
      </button>
    </FormBar>

    <ConfirmSheet :open="showDelete" :loading="deleting" message="确定要删除这条记录吗？删除后无法恢复。"
      @confirm="doDelete" @cancel="showDelete = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { recordAPI, babyAPI, writeErrorMessage } from '@/api'
import { nowLocalDatetime, toLocalDatetime } from '@/utils'
import Segmented from '@/components/Segmented.vue'
import DateTimeField from '@/components/DateTimeField.vue'
import FormBar from '@/components/FormBar.vue'
import NavBar from '@/components/NavBar.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'
import ActivityIndicator from '@/components/ActivityIndicator.vue'

const router = useRouter()
const route = useRoute()
const app = useAppStore()

const recordType = computed(() => route.params.type as string)
const isEdit = computed(() => !!route.params.id)
const saving = ref(false)
const deleting = ref(false)
const showDelete = ref(false)
const error = ref('')
// 编辑态取数失败标记：加载失败后禁止保存，避免空表单覆盖真实记录
const loadFailed = ref(false)

const feedingOptions = [
  { value: 'breast', label: '母乳亲喂', emoji: '🤱' },
  { value: 'bottle', label: '母乳瓶喂', emoji: '🍼' },
  { value: 'formula', label: '配方奶', emoji: '🥛' },
]
const sideOptions = [
  { value: 'left', label: '左侧' },
  { value: 'right', label: '右侧' },
  { value: 'both', label: '双边' },
]

const pageTitle = computed(() => {
  const typeLabel = recordType.value === 'feeding' ? '喂奶' : '尿布'
  return isEdit.value ? `编辑${typeLabel}` : `记录${typeLabel}`
})

const form = reactive({
  occurred_at: nowLocalDatetime(),
  note: '',
})

const feedingForm = reactive({
  type: 'breast',
  amount_ml: null as number | null,
  duration_minutes: null as number | null,
  side: 'both',
  brand: '',
})

const diaperForm = reactive({
  type: 'pee',
})

async function loadRecord() {
  if (!isEdit.value) return
  try {
    const res = await recordAPI.get(Number(route.params.id), recordType.value)
    const record = res.data as any
    form.occurred_at = toLocalDatetime(record.occurred_at)
    form.note = record.data.note || ''
    if (record.record_type === 'feeding') {
      feedingForm.type = record.data.type
      feedingForm.amount_ml = record.data.amount_ml
      feedingForm.duration_minutes = record.data.duration_minutes
      feedingForm.side = record.data.side || 'both'
      feedingForm.brand = record.data.brand || ''
    } else {
      diaperForm.type = record.data.type
    }
    loadFailed.value = false
  } catch (e) {
    // 取不到记录时不再静默弹回：内联报错并锁定保存，防止用空白默认表单覆盖真实记录。
    // 覆盖面靠单条端点保证——列表窗口反查拿不到窗口外的记录，此前会走「有记录」分支之外的空当。
    loadFailed.value = true
    error.value = writeErrorMessage(e, '记录加载失败')
  }
}

async function loadLatest() {
  if (isEdit.value || recordType.value !== 'feeding') return
  const baby = app.currentBaby
  if (!baby) return
  try {
    const res = await babyAPI.latestFeeding(baby.id)
    const last = res.data as any
    if (last?.type) {
      feedingForm.type = last.type
      feedingForm.duration_minutes = last.duration_minutes || 0
      feedingForm.amount_ml = last.amount_ml || 0
      feedingForm.side = last.side || 'both'
      feedingForm.brand = last.brand || ''
      if (last.note) form.note = last.note
    }
  } catch {
    // ignore
  }
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
    const note = form.note

    if (recordType.value === 'feeding') {
      const payload: any = { type: feedingForm.type, note, occurred_at: occurredAt }
      if (feedingForm.type === 'breast') {
        payload.duration_minutes = feedingForm.duration_minutes || 0
        payload.side = feedingForm.side
      } else {
        payload.amount_ml = feedingForm.amount_ml || 0
        if (feedingForm.type === 'formula') payload.brand = feedingForm.brand
      }
      if (isEdit.value) {
        await recordAPI.update(Number(route.params.id), 'feeding', payload)
      } else {
        await recordAPI.createFeeding(baby.id, payload)
      }
    } else {
      const payload = { type: diaperForm.type, note, occurred_at: occurredAt }
      if (isEdit.value) {
        await recordAPI.update(Number(route.params.id), 'diaper', payload)
      } else {
        await recordAPI.createDiaper(baby.id, payload)
      }
    }

    window.dispatchEvent(new CustomEvent(isEdit.value ? 'record-updated' : 'record-created', { detail: null }))
    app.showToast(isEdit.value ? '已保存' : '记录成功', 'success')
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
    await recordAPI.delete(Number(route.params.id), recordType.value)
    window.dispatchEvent(new CustomEvent('record-deleted', { detail: { id: Number(route.params.id), type: recordType.value } }))
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
  loadRecord()
  loadLatest()
})
</script>
