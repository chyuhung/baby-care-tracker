<template>
  <div class="flex flex-col h-dvh">
    <PullRefresh class="flex-1 min-h-0" content-class="px-4 py-4 pb-[calc(6.5rem+env(safe-area-inset-bottom))]"
      :refresh="() => loadRecords(true, true)" @scroll="navScroll = $event">
    <template #header>
    <LargeTitleNav title="时间线" :scroll-top="navScroll">
      <template #filters>
        <div class="flex items-center gap-2 mt-2">
          <MenuSelect :model-value="activeFilter" :options="filterOptions" title="筛选记录"
            aria-label="筛选记录类型" @update:model-value="(v: string | number) => activeFilter = String(v)" />
        </div>
      </template>
    </LargeTitleNav>
    </template>

      <SkeletonCard v-if="loading" :count="6" />
      <EmptyState v-else-if="groupedRecords.length === 0" title="暂无记录" icon="clock"
        subtitle="记录宝宝的每一次喂奶、睡眠与成长瞬间">
        <button @click="router.push('/')"
          class="inline-flex items-center gap-2 px-5 py-2.5 bg-primary-fill text-white rounded-xl font-medium text-sm btn-press shadow-card">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
          立即记录
        </button>
      </EmptyState>
      <div v-else class="space-y-6">
        <div v-for="group in groupedRecords" :key="group.label">
          <h3 class="text-xs font-semibold text-text-secondary mb-3 sticky top-0 bg-bg-main py-1">
            {{ group.label }}
          </h3>
          <div class="space-y-2">
            <SwipeToDelete v-for="(r, i) in group.records" :key="r.record_type + '-' + r.id"
              @delete="softDelete(r)">
              <RecordCard :record="r" :show-date="false" :style="{ animationDelay: `${i * 40}ms` }" class="card-in"
                @edit="editRecord(r)" @delete="deleteRecord(r)" @context="openContext" />
            </SwipeToDelete>
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
    <ConfirmSheet :open="showDeleteConfirm" message="确定要删除这条记录吗？删除后可在提示条上撤销。"
      @confirm="confirmDelete" @cancel="showDeleteConfirm = false" />

    <!-- 长按上下文菜单 -->
    <ContextMenu :open="contextOpen" :title="contextRecord?.title" :subtitle="contextRecord?.subtitle"
      :emoji="contextRecord?.emoji" :actions="contextActions"
      @update:open="contextOpen = $event" @select="onContextSelect" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { recordAPI } from '@/api'
import { recordDisplay, CONTEXT_ICONS } from '@/utils/recordDisplay'
import RecordCard from '@/components/RecordCard.vue'
import PullRefresh from '@/components/PullRefresh.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'
import ContextMenu from '@/components/ContextMenu.vue'
import ActivityIndicator from '@/components/ActivityIndicator.vue'
import EmptyState from '@/components/EmptyState.vue'
import LargeTitleNav from '@/components/LargeTitleNav.vue'
import SkeletonCard from '@/components/SkeletonCard.vue'
import MenuSelect from '@/components/MenuSelect.vue'
import SwipeToDelete from '@/components/SwipeToDelete.vue'
import { useUndoDelete } from '@/composables/useUndoDelete'
import { WEEKDAY_LONG } from '@/utils'

const app = useAppStore()
const router = useRouter()
const route = useRoute()
const navScroll = ref(0)
const records = ref<any[]>([])
const { softDelete } = useUndoDelete(records)

// ── 长按上下文菜单 ─────────────────────────────────────────
const contextOpen = ref(false)
const contextRecord = ref<any>(null)
const contextActions = computed(() => [
  { key: 'edit', label: '编辑', icon: CONTEXT_ICONS.edit },
  { key: 'delete', label: '删除', icon: CONTEXT_ICONS.delete, danger: true },
])
function openContext(rec: any) {
  const d = recordDisplay(rec)
  contextRecord.value = { record: rec, ...d }
  contextOpen.value = true
}
function onContextSelect(key: string) {
  const rec = contextRecord.value?.record
  if (!rec) return
  if (key === 'edit') editRecord(rec)
  else if (key === 'delete') deleteRecord(rec)
}
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
  { label: '全部', emoji: '📋', value: '' },
  { label: '喂奶', emoji: '🍼', value: 'feeding' },
  { label: '尿布', emoji: '🩲', value: 'diaper' },
  { label: '睡眠', emoji: '😴', value: 'sleep' },
  { label: '体温', emoji: '🌡️', value: 'temperature' },
  { label: '户外', emoji: '🌳', value: 'outdoor' },
  { label: '补剂', emoji: '💊', value: 'supplement' },
]
const filterOptions = filters

// 监听路由参数变化，自动切换筛选
watch(() => route.query.filter, (newFilter) => {
  if (newFilter && ['feeding', 'diaper', 'sleep', 'temperature', 'outdoor', 'supplement'].includes(newFilter as string)) {
    activeFilter.value = newFilter as string
  }
}, { immediate: true })

// 当前宝宝就绪后重新加载：冷启动/刷新直达本页时 store 可能尚未恢复，
// 此前只在 onMounted 调用一次导致数据存在却显示「暂无记录」
watch(() => app.currentBaby?.id, (id) => {
  if (id) { days.value = 7; loadRecords() }
})

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
  showDeleteConfirm.value = false
  softDelete(recordToDelete.value)
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
