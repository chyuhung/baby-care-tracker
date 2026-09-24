<template>
  <Teleport to="body">
    <transition name="sheet-mask">
      <div v-if="open" class="fixed inset-0 z-[90] flex items-end justify-center bg-black/35" @click.self="close(false)">
        <transition name="sheet-panel" appear>
          <div v-if="open" ref="panelRef"
            class="w-full max-w-[480px] bg-surface rounded-t-2xl px-3 pb-safe pt-2 space-y-2"
            role="dialog" aria-modal="true" :aria-label="title">
            <!-- 顶部工具条 -->
            <div class="flex items-center justify-between px-1 pb-1">
              <button type="button" @click="close(false)"
                class="px-2 py-1.5 text-[16px] text-primary-deep btn-press">取消</button>
              <span class="text-[14px] font-semibold text-text-primary">{{ title }}</span>
              <button type="button" @click="close(true)"
                class="px-2 py-1.5 text-[16px] font-semibold text-primary-deep btn-press">完成</button>
            </div>

            <!-- 滚轮区 -->
            <div class="relative select-none">
              <!-- 选中高亮带 -->
              <div class="pointer-events-none absolute left-0 right-0 top-1/2 -translate-y-1/2 h-9 rounded-lg bg-muted"
                aria-hidden="true"></div>
              <!-- 上下渐隐 -->
              <div class="pointer-events-none absolute inset-0 rounded-xl" aria-hidden="true"
                style="background:linear-gradient(to bottom, rgb(var(--surface)) 0%, rgba(255,255,255,0) 34%, rgba(255,255,255,0) 66%, rgb(var(--surface)) 100%)"></div>

              <div class="relative flex justify-center gap-0.5">
                <div v-for="(col, ci) in columns" :key="col.key" ref="colEls"
                  class="wheel-col overflow-y-scroll no-scrollbar snap-y snap-mandatory text-center"
                  :style="{ width: col.width || '72px' }" @scroll="onScroll(ci, $event)">
                  <div class="wheel-spacer" aria-hidden="true"></div>
                  <div v-for="(item, ii) in col.items" :key="item.value"
                    class="h-9 leading-9 snap-center text-[20px] tabular-nums transition-colors duration-150"
                    :class="ii === col.index ? 'text-text-primary font-semibold' : 'text-text-secondary/70'">
                    {{ item.label }}
                  </div>
                  <div class="wheel-spacer" aria-hidden="true"></div>
                </div>
              </div>
            </div>
          </div>
        </transition>
      </div>
    </transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted } from 'vue'

const ITEM_H = 36

const props = withDefaults(defineProps<{
  open: boolean
  /** 'YYYY-MM-DDTHH:mm'（本地时间，无时区，与 datetime-local 一致） */
  modelValue: string
  title?: string
  min?: string
  max?: string
}>(), {
  title: '选择时间',
  min: '',
  max: '',
})

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'confirm', v: string): void
}>()

const panelRef = ref<HTMLElement | null>(null)
const colEls = ref<HTMLElement[]>([])

function pad(n: number) { return String(n).padStart(2, '0') }
function parseLocal(s: string) {
  const m = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})/.exec(s || '')
  if (m) return { y: +m[1], mo: +m[2], d: +m[3], h: +m[4], mi: +m[5] }
  const n = new Date()
  return { y: n.getFullYear(), mo: n.getMonth() + 1, d: n.getDate(), h: n.getHours(), mi: n.getMinutes() }
}

const draft = ref(parseLocal(props.modelValue))

/** 日期列表：今天前后各 90 天 */
const dateItems = computed(() => {
  const out: { label: string; value: string }[] = []
  const base = new Date(); base.setHours(0, 0, 0, 0)
  for (let off = 90; off >= -90; off--) {
    const dt = new Date(base); dt.setDate(dt.getDate() - off)
    const v = `${dt.getFullYear()}-${pad(dt.getMonth() + 1)}-${pad(dt.getDate())}`
    const wd = ['日', '一', '二', '三', '四', '五', '六'][dt.getDay()]
    out.push({ label: `${dt.getMonth() + 1}月${dt.getDate()}日 周${wd}`, value: v })
  }
  return out
})
const hourItems = Array.from({ length: 24 }, (_, i) => ({ label: pad(i), value: String(i) }))
const minItems = Array.from({ length: 60 }, (_, i) => ({ label: pad(i), value: String(i) }))

const columns = computed(() => {
  const dv = `${draft.value.y}-${pad(draft.value.mo)}-${pad(draft.value.d)}`
  const di = Math.max(0, dateItems.value.findIndex(x => x.value === dv))
  return [
    { key: 'date', items: dateItems.value, index: di, width: '150px' },
    { key: 'hour', items: hourItems, index: draft.value.h, width: '56px' },
    { key: 'min', items: minItems, index: draft.value.mi, width: '56px' },
  ]
})

function scrollColTo(ci: number, index: number, smooth = false) {
  const el = colEls.value[ci]
  if (!el) return
  el.scrollTo({ top: index * ITEM_H, behavior: smooth ? 'smooth' : 'auto' })
}

function syncFromValue() {
  draft.value = parseLocal(props.modelValue)
  nextTick(() => columns.value.forEach((c, i) => scrollColTo(i, c.index)))
}

let scrollTimer: number | null = null
function onScroll(ci: number, e: Event) {
  const el = e.target as HTMLElement
  if (scrollTimer) clearTimeout(scrollTimer)
  scrollTimer = window.setTimeout(() => {
    const idx = Math.round(el.scrollTop / ITEM_H)
    const col = columns.value[ci]
    const item = col.items[Math.max(0, Math.min(col.items.length - 1, idx))]
    if (!item) return
    if (col.key === 'date') {
      const [y, mo, d] = item.value.split('-').map(Number)
      if (y !== draft.value.y || mo !== draft.value.mo || d !== draft.value.d) {
        draft.value = { ...draft.value, y, mo, d }
      }
    } else if (col.key === 'hour') {
      const h = +item.value
      if (h !== draft.value.h) draft.value = { ...draft.value, h }
    } else {
      const mi = +item.value
      if (mi !== draft.value.mi) draft.value = { ...draft.value, mi }
    }
  }, 90)
}

function close(commit: boolean) {
  if (commit) {
    const v = `${draft.value.y}-${pad(draft.value.mo)}-${pad(draft.value.d)}T${pad(draft.value.h)}:${pad(draft.value.mi)}`
    emit('confirm', v)
  }
  emit('update:open', false)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') { e.preventDefault(); close(false) }
  if (e.key === 'Enter') { e.preventDefault(); close(true) }
}

watch(() => props.open, async (v) => {
  if (v) {
    syncFromValue()
    document.body.style.overflow = 'hidden'
    window.addEventListener('keydown', onKeydown)
    await nextTick()
    panelRef.value?.focus()
  } else {
    document.body.style.overflow = ''
    window.removeEventListener('keydown', onKeydown)
  }
})

onUnmounted(() => {
  document.body.style.overflow = ''
  window.removeEventListener('keydown', onKeydown)
  if (scrollTimer) clearTimeout(scrollTimer)
})
</script>

<style scoped>
.no-scrollbar::-webkit-scrollbar { display: none; }
.no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
.wheel-col {
  height: 180px;
  -webkit-overflow-scrolling: touch;
}
.wheel-spacer { height: 72px; }
</style>
