<template>
  <Teleport to="body">
    <transition name="sheet-mask">
      <div v-if="open" class="fixed inset-0 z-[90] flex items-end justify-center bg-black/35" @click.self="close(false)">
        <transition name="sheet-panel" appear>
          <div v-if="open" ref="panelRef"
            class="w-full max-w-[480px] bg-surface rounded-t-2xl px-3 pb-safe pt-2 space-y-2 shadow-sheet"
            role="dialog" aria-modal="true" :aria-label="title">
            <!-- 顶部工具条 -->
            <div class="flex items-center justify-between px-1 pb-1">
              <button type="button" @click="close(false)"
                class="min-w-[44px] min-h-[44px] px-3 flex items-center justify-center text-[16px] text-primary-deep btn-press">取消</button>
              <span class="text-[14px] font-semibold text-text-primary">{{ title }}</span>
              <button type="button" @click="close(true)"
                class="min-w-[44px] min-h-[44px] px-3 flex items-center justify-center text-[16px] font-semibold text-primary-deep btn-press">完成</button>
            </div>

            <!-- 滚轮区 -->
            <div class="relative select-none">
              <!-- 选中高亮带 -->
              <div class="pointer-events-none absolute left-0 right-0 top-1/2 -translate-y-1/2 h-9 rounded-[10px] bg-muted"
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
  /** 'YYYY-MM-DDTHH:mm'（本地时间，无时区，与 datetime-local 一致）；dateOnly 时为 'YYYY-MM-DD' */
  modelValue: string
  title?: string
  min?: string
  max?: string
  /** 纯日历日模式：只显示日期列，输出 'YYYY-MM-DD'（用于出生日期/测量日期等不关心时刻的字段） */
  dateOnly?: boolean
}>(), {
  title: '选择时间',
  min: '',
  max: '',
  dateOnly: false,
})

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'confirm', v: string): void
}>()

const panelRef = ref<HTMLElement | null>(null)
const colEls = ref<HTMLElement[]>([])

function pad(n: number) { return String(n).padStart(2, '0') }
function parseLocal(s: string) {
  const m = /^(\d{4})-(\d{2})-(\d{2})(?:T(\d{2}):(\d{2}))?/.exec(s || '')
  if (m) return { y: +m[1], mo: +m[2], d: +m[3], h: m[4] ? +m[4] : 0, mi: m[5] ? +m[5] : 0 }
  const n = new Date()
  return { y: n.getFullYear(), mo: n.getMonth() + 1, d: n.getDate(), h: n.getHours(), mi: n.getMinutes() }
}

const draft = ref(clampDraft(parseLocal(props.modelValue)))

/** 日期上下界：min/max 缺省为 今天-10年 ~ 今天+1年（日历日，取整天零点） */
function bound(s: string, which: 'min' | 'max'): Date {
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(s || '')
  if (m) return new Date(+m[1], +m[2] - 1, +m[3])
  const d = new Date()
  d.setHours(0, 0, 0, 0)
  d.setFullYear(d.getFullYear() + (which === 'min' ? -10 : 1))
  return d
}
const minDate = computed(() => bound(props.min, 'min'))
const maxDate = computed(() => {
  const d = bound(props.max, 'max')
  // 属性保证 min <= max，否则日期列表会倒着生成
  return d < minDate.value ? new Date(minDate.value) : d
})

/** 越界的日期取最近的合法日（按 min/max 裁剪），并同步 draft——
 *  否则 draft 保留越界值、而高亮索引被 findIndex 兜成 0，二者互相矛盾。 */
function clampDraft(d: { y: number; mo: number; d: number; h: number; mi: number }) {
  const dt = new Date(d.y, d.mo - 1, d.d)
  const lo = bound(props.min, 'min')
  const hi = bound(props.max, 'max')
  const hiAdj = hi < lo ? lo : hi
  if (dt < lo) dt.setTime(lo.getTime())
  else if (dt > hiAdj) dt.setTime(hiAdj.getTime())
  return { y: dt.getFullYear(), mo: dt.getMonth() + 1, d: dt.getDate(), h: d.h, mi: d.mi }
}

/** 日期列表：[min, max] 内每个日历日（升序）。不再固定「今天前后各 90 天」——
 *  记录补录 90 天前、或宝宝出生日期这类远期日期根本滚不到。 */
const dateItems = computed(() => {
  const out: { label: string; value: string }[] = []
  const cur = new Date(maxDate.value)
  const lo = minDate.value
  // 安全上限：默认区间约 4000 项；props 异常时不至于死循环
  let guard = 0
  while (cur >= lo && guard++ < 8000) {
    const v = `${cur.getFullYear()}-${pad(cur.getMonth() + 1)}-${pad(cur.getDate())}`
    const wd = ['日', '一', '二', '三', '四', '五', '六'][cur.getDay()]
    out.push({ label: `${cur.getMonth() + 1}月${cur.getDate()}日 周${wd}`, value: v })
    cur.setDate(cur.getDate() - 1)
  }
  return out.reverse()
})
const hourItems = Array.from({ length: 24 }, (_, i) => ({ label: pad(i), value: String(i) }))
const minItems = Array.from({ length: 60 }, (_, i) => ({ label: pad(i), value: String(i) }))

const columns = computed(() => {
  const dv = `${draft.value.y}-${pad(draft.value.mo)}-${pad(draft.value.d)}`
  const di = Math.max(0, dateItems.value.findIndex(x => x.value === dv))
  if (props.dateOnly) {
    return [{ key: 'date', items: dateItems.value, index: di, width: '150px' }]
  }
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
  draft.value = clampDraft(parseLocal(props.modelValue))
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
    if (props.dateOnly) {
      emit('confirm', `${draft.value.y}-${pad(draft.value.mo)}-${pad(draft.value.d)}`)
    } else {
      const v = `${draft.value.y}-${pad(draft.value.mo)}-${pad(draft.value.d)}T${pad(draft.value.h)}:${pad(draft.value.mi)}`
      emit('confirm', v)
    }
  }
  emit('update:open', false)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') { e.preventDefault(); close(false) }
  if (e.key === 'Enter') { e.preventDefault(); close(true) }
}

let prevOverflow = ''
watch(() => props.open, async (v) => {
  if (v) {
    syncFromValue()
    prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    window.addEventListener('keydown', onKeydown)
    await nextTick()
    panelRef.value?.focus()
  } else {
    document.body.style.overflow = prevOverflow
    window.removeEventListener('keydown', onKeydown)
  }
})

onUnmounted(() => {
  if (props.open) document.body.style.overflow = prevOverflow
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
