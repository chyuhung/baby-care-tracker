<template>
  <div class="flex flex-col min-h-screen">
    <header class="pt-safe bg-white px-4 py-3 border-b border-border-color">
      <h1 class="text-lg font-bold text-text-primary">时间线</h1>
      <!-- 筛选 -->
      <div class="flex gap-2 mt-2 overflow-x-auto">
        <button v-for="f in filters" :key="f.value"
          @click="activeFilter = f.value"
          :class="['px-3 py-1 rounded-full text-xs font-medium transition-colors btn-press whitespace-nowrap',
            activeFilter === f.value ? 'bg-primary text-white' : 'bg-gray-100 text-text-secondary']">
          {{ f.label }}
        </button>
      </div>
    </header>

    <main class="flex-1 min-h-0 px-4 py-4 overflow-y-auto pb-20">
      <div v-if="loading" class="text-center py-16 text-text-secondary">加载中...</div>
      <div v-else-if="groupedRecords.length === 0" class="text-center py-16">
        <div class="text-5xl mb-4">📋</div>
        <p class="text-text-secondary">暂无记录</p>
      </div>
      <div v-else class="space-y-6">
        <div v-for="group in groupedRecords" :key="group.label">
          <h3 class="text-xs font-semibold text-text-secondary uppercase tracking-wide mb-3 sticky top-0 bg-bg-main py-1">
            {{ group.label }}
          </h3>
          <div class="space-y-2">
            <RecordCard v-for="(r, i) in group.records" :key="r.id"
              :record="r" :style="{ animationDelay: `${i * 40}ms` }" class="card-in"
              @edit="editRecord(r)" @delete="deleteRecord(r)" />
          </div>
        </div>
      </div>
    </main>

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/30 flex items-end z-50" @click.self="showDeleteConfirm = false">
      <div class="bg-white w-full rounded-t-2xl p-6 space-y-4 pb-safe animate-[slideUp_0.3s_ease]">
        <h3 class="text-lg font-bold text-text-primary text-center">确认删除</h3>
        <p class="text-text-secondary text-sm text-center">确定要删除这条记录吗？</p>
        <div class="flex gap-3">
          <button @click="showDeleteConfirm = false" class="flex-1 py-3 bg-gray-100 text-text-primary rounded-xl font-medium btn-press">取消</button>
          <button @click="confirmDelete" class="flex-1 py-3 bg-red-500 text-white rounded-xl font-medium btn-press">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { recordAPI } from '@/api'
import RecordCard from '@/components/RecordCard.vue'

const app = useAppStore()
const router = useRouter()
const route = useRoute()
const records = ref<any[]>([])
const loading = ref(false)
const activeFilter = ref('')
const showDeleteConfirm = ref(false)
const recordToDelete = ref<any>(null)

const filters = [
  { label: '全部', value: '' },
  { label: '🍼 喂奶', value: 'feeding' },
  { label: '🧷 尿布', value: 'diaper' },
]

// 监听路由参数变化，自动切换筛选
watch(() => route.query.filter, (newFilter) => {
  if (newFilter && ['feeding', 'diaper'].includes(newFilter as string)) {
    activeFilter.value = newFilter as string
  }
}, { immediate: true })

const groupedRecords = computed(() => {
  const filtered = activeFilter.value ? records.value.filter(r => r.record_type === activeFilter.value) : records.value
  const groups: { label: string; records: any[] }[] = []
  const now = new Date()
  const today = now.toDateString()
  const yesterday = new Date(now.getTime() - 86400000).toDateString()

  const byDate = new Map<string, any[]>()
  for (const r of filtered) {
    const d = new Date(r.occurred_at).toDateString()
    if (!byDate.has(d)) byDate.set(d, [])
    byDate.get(d)!.push(r)
  }

  for (const [date, recs] of byDate) {
    let label = date
    if (date === today) label = '今天'
    else if (date === yesterday) label = '昨天'
    groups.push({ label, records: recs })
  }

  return groups
})

async function loadRecords() {
  const baby = app.currentBaby()
  if (!baby) return
  loading.value = true
  try {
    const res = await recordAPI.list(baby.id)
    records.value = res.data
  } catch {} finally {
    loading.value = false
  }
}

function editRecord(r: any) {
  router.push(`/record/${r.record_type}/${r.id}/edit`)
}

function deleteRecord(r: any) {
  recordToDelete.value = r
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (!recordToDelete.value) return
  try {
    await recordAPI.delete(recordToDelete.value.id, recordToDelete.value.record_type)
    app.showToast('已删除', 'success')
    showDeleteConfirm.value = false
    loadRecords()
  } catch {
    app.showToast('删除失败', 'error')
  }
}

function onRecordChange() { loadRecords() }
onMounted(() => { loadRecords(); window.addEventListener('app:record-changed', onRecordChange) })
onUnmounted(() => { window.removeEventListener('app:record-changed', onRecordChange) })
</script>
