<template>
  <div class="min-h-screen bg-bg-main">
    <header class="pt-safe bg-white px-4 py-3 border-b border-border-color flex items-center gap-3">
      <button @click="router.back()" class="p-1 -ml-1 btn-press">
        <svg class="w-6 h-6 text-text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
      </button>
      <h1 class="text-lg font-bold text-text-primary">{{ pageTitle }}</h1>
    </header>

    <main class="px-4 py-6 space-y-5">
      <!-- 喂奶表单 -->
      <template v-if="recordType === 'feeding'">
        <!-- 类型选择 -->
        <div>
          <label class="text-sm text-text-secondary block mb-2">喂奶方式</label>
          <div class="grid grid-cols-3 gap-2">
            <button v-for="t in feedingTypes" :key="t.value"
              @click="form.type = t.value"
              :class="['py-3 rounded-xl text-sm font-medium transition-colors btn-press flex flex-col items-center gap-1',
                form.type === t.value ? 'bg-primary text-white' : 'bg-white border border-border-color text-text-secondary']">
              <span>{{ t.emoji }}</span>
              {{ t.label }}
            </button>
          </div>
        </div>

        <!-- 时间 -->
        <div>
          <label class="text-sm text-text-secondary block mb-2">发生时间</label>
          <input v-model="form.occurred_at" type="datetime-local"
            class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary transition-colors" />
        </div>

        <!-- 亲喂：时长 + 方向 -->
        <template v-if="form.type === 'breast'">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-sm text-text-secondary block mb-2">时长（分钟）</label>
              <input v-model.number="form.duration_minutes" type="number" min="1" placeholder="15"
                class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary transition-colors" />
            </div>
            <div>
              <label class="text-sm text-text-secondary block mb-2">喂养侧</label>
              <div class="flex gap-1">
                <button v-for="s in sides" :key="s.value"
                  @click="form.side = s.value"
                  :class="['flex-1 py-3 rounded-xl text-xs font-medium transition-colors btn-press',
                    form.side === s.value ? 'bg-primary text-white' : 'bg-white border border-border-color text-text-secondary']">
                  {{ s.label }}
                </button>
              </div>
            </div>
          </div>
        </template>

        <!-- 瓶喂/配方奶：奶量 -->
        <template v-else>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-sm text-text-secondary block mb-2">奶量（ml）</label>
              <input v-model.number="form.amount_ml" type="number" min="1" placeholder="120"
                class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary transition-colors" />
            </div>
            <div>
              <label class="text-sm text-text-secondary block mb-2">品牌（可选）</label>
              <input v-model="form.brand" type="text" placeholder="如：爱他美"
                class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary transition-colors" />
            </div>
          </div>
          <div>
            <label class="text-sm text-text-secondary block mb-2">时长（分钟）</label>
            <input v-model.number="form.duration_minutes" type="number" min="1" placeholder="15"
              class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary transition-colors" />
          </div>
        </template>

        <!-- 备注 -->
        <div>
          <label class="text-sm text-text-secondary block mb-2">备注</label>
          <textarea v-model="form.note" rows="2" placeholder="可选"
            class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary transition-colors resize-none"></textarea>
        </div>
      </template>

      <!-- 尿布表单 -->
      <template v-else-if="recordType === 'diaper'">
        <div>
          <label class="text-sm text-text-secondary block mb-2">发生时间</label>
          <input v-model="form.occurred_at" type="datetime-local"
            class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary transition-colors" />
        </div>

        <div>
          <label class="text-sm text-text-secondary block mb-3">尿布类型</label>
          <div class="grid grid-cols-3 gap-3">
            <button v-for="t in diaperTypes" :key="t.value"
              @click="form.type = t.value"
              :class="['py-4 rounded-xl text-sm font-medium transition-colors btn-press flex flex-col items-center gap-2',
                form.type === t.value ? 'bg-primary text-white' : 'bg-white border border-border-color text-text-secondary']">
              <span class="text-2xl">{{ t.emoji }}</span>
              {{ t.label }}
            </button>
          </div>
        </div>

        <div>
          <label class="text-sm text-text-secondary block mb-2">备注</label>
          <textarea v-model="form.note" rows="2" placeholder="可选，如颜色、形状等"
            class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary focus:border-primary transition-colors resize-none"></textarea>
        </div>
      </template>

      <div v-if="error" class="bg-red-50 text-red-500 text-sm px-4 py-2 rounded-xl text-center">{{ error }}</div>

      <button @click="submit" :disabled="loading"
        class="btn-press w-full py-3 bg-primary text-white font-semibold rounded-xl shadow-card disabled:opacity-50">
        {{ loading ? '保存中...' : (isEdit ? '更新记录' : '记录') }}
      </button>

      <button v-if="isEdit" @click="confirmDelete"
        class="btn-press w-full py-3 bg-white text-red-500 font-medium rounded-xl border border-red-200">
        删除此记录
      </button>
    </main>

    <!-- 删除确认 -->
    <div v-if="showDelete" class="fixed inset-0 bg-black/30 flex items-end z-50" @click.self="showDelete = false">
      <div class="bg-white w-full rounded-t-2xl p-6 space-y-4 pb-safe">
        <h3 class="text-lg font-bold text-text-primary text-center">确认删除</h3>
        <div class="flex gap-3">
          <button @click="showDelete = false" class="flex-1 py-3 bg-gray-100 text-text-primary rounded-xl font-medium btn-press">取消</button>
          <button @click="doDelete" class="flex-1 py-3 bg-red-500 text-white rounded-xl font-medium btn-press">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { babyAPI, recordAPI } from '@/api'

const router = useRouter()
const route = useRoute()
const app = useAppStore()

const recordType = computed(() => route.params.type as string)
const isEdit = computed(() => !!route.params.id)
const loading = ref(false)
const error = ref('')
const showDelete = ref(false)

const pageTitle = computed(() => {
  if (isEdit.value) return '编辑记录'
  if (recordType.value === 'feeding') return '🍼 记录喂奶'
  if (recordType.value === 'diaper') return '🧷 记录尿布'
  return '记录'
})

const nowDatetime = () => {
  const d = new Date()
  const y = d.getFullYear()
  const M = String(d.getMonth() + 1).padStart(2, '0')
  const D = String(d.getDate()).padStart(2, '0')
  const h = String(d.getHours()).padStart(2, '0')
  const m = String(d.getMinutes()).padStart(2, '0')
  return `${y}-${M}-${D}T${h}:${m}`
}

// 把 UTC ISO 字符串转为本地 datetime-local 格式
function utcToLocalDatetime(utcIso: string) {
  const d = new Date(utcIso)
  const y = d.getFullYear()
  const M = String(d.getMonth() + 1).padStart(2, '0')
  const D = String(d.getDate()).padStart(2, '0')
  const h = String(d.getHours()).padStart(2, '0')
  const m = String(d.getMinutes()).padStart(2, '0')
  return `${y}-${M}-${D}T${h}:${m}`
}

const feedingTypes = [
  { value: 'breast', label: '亲喂', emoji: '🤱' },
  { value: 'bottle', label: '母乳瓶喂', emoji: '🍼' },
  { value: 'formula', label: '配方奶', emoji: '🍼' },
]
const sides = [
  { value: 'left', label: '左' },
  { value: 'right', label: '右' },
  { value: 'both', label: '双边' },
]
const diaperTypes = [
  { value: 'pee', label: '小便', emoji: '💧' },
  { value: 'poop', label: '大便', emoji: '💩' },
  { value: 'mixed', label: '混合', emoji: '💥' },
]

const form = reactive({
  type: 'breast',
  duration_minutes: 15,
  amount_ml: 0,
  side: 'left',
  brand: '',
  note: '',
  occurred_at: nowDatetime(),
})

function resetFormDefaults() {
  if (recordType.value === 'diaper') {
    form.type = 'pee'
    form.duration_minutes = 0
    form.amount_ml = 0
    form.side = 'left'
    form.brand = ''
  } else {
    form.type = 'breast'
    form.duration_minutes = 15
    form.amount_ml = 0
    form.side = 'left'
    form.brand = ''
  }
}

async function loadLatest() {
  const baby = app.currentBaby
  if (!baby) return
  try {
    const res = await babyAPI.latestFeeding(baby.id)
    if (res.data) {
      form.type = res.data.type || 'breast'
      form.duration_minutes = res.data.duration_minutes || 15
      form.amount_ml = res.data.amount_ml || 0
      form.side = res.data.side || 'left'
      form.brand = res.data.brand || ''
      form.note = ''
    }
  } catch {
    // latest data not available, use defaults
  }
}

async function loadRecord() {
  if (!isEdit.value) return
  const id = Number(route.params.id)
  const baby = app.currentBaby
  if (!baby) return
  try {
    const res = await recordAPI.list(baby.id, recordType.value)
    const records = res.data as any[]
    const r = records.find((r: any) => r.id === id)
    if (r) {
      if (recordType.value === 'feeding') {
        const d = r.data
        form.type = d.type
        form.duration_minutes = d.duration_minutes
        form.amount_ml = d.amount_ml
        form.side = d.side || 'left'
        form.brand = d.brand || ''
        form.note = d.note || ''
        form.occurred_at = d.occurred_at ? utcToLocalDatetime(d.occurred_at) : nowDatetime()
      } else {
        form.type = r.data.type
        form.note = r.data.note || ''
        form.occurred_at = r.occurred_at ? utcToLocalDatetime(r.occurred_at) : nowDatetime()
      }
    }
  } catch {
    app.showToast('记录加载失败', 'error')
  }
}

async function submit() {
  error.value = ''
  const baby = app.currentBaby
  if (!baby) { error.value = '请先添加宝宝'; return }
  loading.value = true
  try {
    const payload = {
      ...form,
      occurred_at: new Date(form.occurred_at).toISOString(),
    }
    if (isEdit.value) {
      await recordAPI.update(Number(route.params.id), recordType.value, payload)
    } else if (recordType.value === 'feeding') {
      await recordAPI.createFeeding(baby.id, payload)
    } else {
      await recordAPI.createDiaper(baby.id, payload)
    }
    app.showToast('记录成功 ✅', 'success')
    router.back()
  } catch (e: any) {
    error.value = e.response?.data?.error || '保存失败'
  } finally {
    loading.value = false
  }
}

function confirmDelete() { showDelete.value = true }

async function doDelete() {
  try {
    const id = Number(route.params.id)
    const typ = recordType.value
    await recordAPI.delete(id, typ)
    window.dispatchEvent(new CustomEvent('record-deleted', { detail: { id, type: typ } }))
    app.showToast('已删除', 'success')
    router.back()
  } catch {
    app.showToast('删除失败', 'error')
  }
}

onMounted(async () => {
  resetFormDefaults()
  if (!isEdit.value) {
    form.occurred_at = nowDatetime()
    if (recordType.value === 'feeding') {
      await loadLatest()
    }
  } else {
    await loadRecord()
  }
})
</script>
