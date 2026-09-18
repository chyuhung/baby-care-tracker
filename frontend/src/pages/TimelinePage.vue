<template>
  <div class="flex flex-col h-dvh">
    <PullRefresh class="flex-1 min-h-0" content-class="px-4 py-4 pb-[calc(5rem+env(safe-area-inset-bottom))]"
      :refresh="() => loadRecords(true, true)">
    <template #header>
    <header class="sticky top-0 z-30 glass-surface hairline-bottom pt-safe px-4 py-3">
      <h1 class="text-lg font-bold text-text-primary">时间线</h1>
      <!-- 筛选 -->
      <div class="flex flex-wrap gap-2 mt-2">
        <button v-for="f in filters" :key="f.value"
          @click="activeFilter = f.value"
          :class="['px-3 py-2 min-h-[44px] flex items-center justify-center rounded-full text-xs font-medium transition-colors btn-press whitespace-nowrap',
            activeFilter === f.value ? 'bg-primary-deep text-white' : 'bg-muted text-text-secondary']">
          {{ f.label }}
        </button>
      </div>
    </header>
    </template>

      <div v-if="loading" class="flex justify-center py-20">
        <ActivityIndicator :size="28" class="text-text-secondary" />
      </div>
      <div v-else-if="groupedRecords.length === 0" class="text-center py-16">
        <img src="/icon-192.png" alt="" class="w-14 h-14 mx-auto block mb-4" />
        <p class="text-text-secondary">暂无记录</p>
      </div>
      <div v-else class="space-y-6">
        <div v-for="group in groupedRecords" :key="group.label">
          <h3 class="text-xs font-semibold text-text-secondary mb-3 sticky top-0 bg-bg-main py-1">
            {{ group.label }}
          </h3>
          <div class="space-y-2">
            <RecordCard v-for="(r, i) in group.records" :key="r.record_type + '-' + r.id"
              :record="r" :show-date="false" :style="{ animationDelay: `${i * 40}ms` }" class="card-in"
              @edit="editRecord(r)" @delete="deleteRecord(r)" />
          </div>
        </div>

        <!-- 滚动到底自动加载 -->
        <div ref="sentinelEl" class="h-16 flex items-center justify-center">
          <ActivityIndicator v-if="loadingMore" :size="24" class="text-text-secondary" />
          <span v-else-if="!hasMore" class="text-xs text-text-secondary">没有更多了</span>
        </div>
      </div>
    </PullRefresh>

    <!-- 删除确认（iOS 底部操作表） -->
    <ConfirmSheet :open="showDeleteConfirm" message="确定要删除这条记录吗？删除后无法恢复。"
      @confirm="confirmDelete" @cancel="showDeleteConfirm = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { recordAPI } from '@/api'
import RecordCard from '@/components/RecordCard.vue'
import PullRefresh from '@/components/PullRefresh.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'
import ActivityIndicator from '@/components/ActivityIndicator.vue'
import { WEEKDAY_LONG } from '@/utils'

const app = useAppStore()
const router = useRouter()
const route = useRoute()
const records = ref<any[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const activeFilter = ref('')
const showDeleteConfirm = ref(false)
const recordToDelete = ref<any>(null)
const days = ref(7)
const totalCount = ref(0)
const loadedCount = ref(0)
const sentinelEl = ref<HTMLElement | null>(null)
let io: IntersectionObserver | null = null

const filters = [
  { label: '全部', value: '' },
  { label: '🍼 喂奶', value: 'feeding' },
  { label: '🩲 尿布', value: 'diaper' },
  { label: '😴 睡眠', value: 'sleep' },
  { label: '🌡️ 体温', value: 'temperature' },
  { label: '🌳 户外', value: 'outdoor' },
  { label: '💊 补剂', value: 'supplement' },
]

// 监听路由参数变化，自动切换筛选
watch(() => route.query.filter, (newFilter) => {
  if (newFilter && ['feeding', 'diaper', 'sleep', 'temperature', 'outdoor', 'supplement'].includes(newFilter as string)) {
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
    else {
      const d = new Date(date)
      label = `${d.getMonth() + 1}月${d.getDate()}日 ${WEEKDAY_LONG[d.getDay()]}`
    }
    groups.push({ label, records: recs })
  }

  return groups
})

const hasMore = computed(() => loadedCount.value < totalCount.value)

async function loadRecords(reset: boolean = true, silent: boolean = false) {
  const baby = app.currentBaby
  if (!baby) return
  if (reset) { if (!silent) loading.value = true }
  else loadingMore.value = true
  try {
    const [res, countRes] = await Promise.all([
      recordAPI.list(baby.id, undefined, days.value),
      recordAPI.count(baby.id),
    ])
    records.value = res.data
    totalCount.value = countRes.data.total
    loadedCount.value = res.data.length
  } catch {
    app.showToast('数据加载失败', 'error')
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function loadMore() {
  if (loadingMore.value || !hasMore.value || loading.value) return
  days.value += 7
  loadRecords(false)
}

function editRecord(r: any) {
  if (r.record_type === 'sleep') {
    router.push(`/sleep/${r.id}/edit`)
  } else if (r.record_type === 'temperature') {
    router.push(`/temperature/${r.id}/edit`)
  } else if (r.record_type === 'outdoor') {
    router.push(`/outdoor/${r.id}/edit`)
  } else if (r.record_type === 'supplement') {
    router.push(`/supplement/${r.id}/edit`)
  } else {
    router.push(`/record/${r.record_type}/${r.id}/edit`)
  }
}

function deleteRecord(r: any) {
  recordToDelete.value = r
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (!recordToDelete.value) return
  try {
    const { id, record_type: typ } = recordToDelete.value
    await recordAPI.delete(id, typ)
    window.dispatchEvent(new CustomEvent('record-deleted', { detail: { id, type: typ } }))
    app.showToast('已删除', 'success')
    showDeleteConfirm.value = false
  } catch {
    app.showToast('删除失败', 'error')
  }
}

function onRecordCreated(e: Event) {
  const record = (e as CustomEvent).detail
  if (!record) { loadRecords(); return }
  if (record.baby_id === app.currentBaby?.id) records.value.unshift(record)
}

function onRecordDeleted(e: Event) {
  const { id, type } = (e as CustomEvent).detail || {}
  records.value = records.value.filter(r => !(r.id === id && r.record_type === (type || r.record_type)))
}

onMounted(() => {
  loadRecords()
  window.addEventListener('record-created', onRecordCreated)
  window.addEventListener('record-deleted', onRecordDeleted)
  // 哨兵进入视口（提前 300px）即自动加载更多
  io = new IntersectionObserver((entries) => {
    if (entries[0]?.isIntersecting) loadMore()
  }, { rootMargin: '300px' })
  if (sentinelEl.value) io.observe(sentinelEl.value)
})
onUnmounted(() => {
  window.removeEventListener('record-created', onRecordCreated)
  window.removeEventListener('record-deleted', onRecordDeleted)
  io?.disconnect()
})
</script>
