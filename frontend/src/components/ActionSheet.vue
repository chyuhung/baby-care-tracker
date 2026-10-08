<template>
  <Teleport to="body">
    <transition name="sheet-mask">
      <div v-if="open" class="fixed inset-0 z-[80] flex items-end justify-center bg-black/35" @click.self="cancel">
        <transition name="sheet-panel" appear>
          <div v-if="open" ref="panelRef" class="w-full max-w-[480px] px-3 pb-safe space-y-2" role="menu" aria-modal="true"
            :aria-label="title">
            <!-- 上组：选项列表 -->
            <div class="bg-surface rounded-2xl overflow-hidden shadow-sheet">
              <div class="pt-2.5 flex justify-center">
                <span class="w-9 h-1 rounded-full bg-border-color"></span>
              </div>
              <div v-if="title || description" class="px-4 pt-2.5 pb-1">
                <p v-if="title" class="text-center text-[15px] font-semibold text-text-primary">{{ title }}</p>
                <p v-if="description" class="text-center text-[13px] text-text-secondary mt-1 leading-relaxed">{{ description }}</p>
              </div>
              <div class="divide-y divide-border-color/60">
                <button v-for="opt in options" :key="opt.value" type="button" role="menuitem"
                  class="w-full px-4 py-3.5 flex items-center gap-3 min-h-[44px] text-left btn-press disabled:opacity-60"
                  :disabled="loading !== ''" @click="onSelect(opt.value)">
                  <span class="w-9 h-9 rounded-xl flex items-center justify-center shrink-0 bg-primary/10">
                    <svg class="w-5 h-5 text-primary-deep" fill="none" stroke="currentColor" viewBox="0 0 24 24"
                      stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                      <path :d="opt.icon" />
                    </svg>
                  </span>
                  <span class="flex-1 min-w-0">
                    <span class="block text-[17px] font-medium text-text-primary">{{ opt.label }}</span>
                    <span v-if="opt.sub" class="block text-[13px] text-text-secondary mt-0.5">{{ opt.sub }}</span>
                  </span>
                  <ActivityIndicator v-if="loading === opt.value" :size="18" class="text-text-secondary" />
                </button>
              </div>
            </div>
            <!-- 取消（独立成组） -->
            <button ref="cancelRef" type="button" role="menuitem" @click="cancel"
              class="w-full py-4 bg-surface rounded-2xl text-[17px] font-semibold text-text-primary shadow-sheet btn-press">
              {{ cancelText }}
            </button>
          </div>
        </transition>
      </div>
    </transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onUnmounted } from 'vue'
import ActivityIndicator from './ActivityIndicator.vue'

export interface ActionSheetOption {
  value: string
  label: string
  sub?: string
  icon: string
}

const props = withDefaults(defineProps<{
  open: boolean
  title?: string
  description?: string
  options: ActionSheetOption[]
  cancelText?: string
  /** 正在处理中的选项 value（非空时禁用全部选项并在此行显示 spinner） */
  loading?: string
}>(), {
  title: '',
  description: '',
  cancelText: '取消',
  loading: '',
})

const emit = defineEmits<{
  (e: 'select', value: string): void
  (e: 'cancel'): void
  (e: 'update:open', v: boolean): void
}>()

const panelRef = ref<HTMLElement | null>(null)
const cancelRef = ref<HTMLElement | null>(null)
let prevOverflow = ''

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') { e.preventDefault(); cancel(); return }
  if (e.key !== 'Tab' || !panelRef.value) return
  const focusables = Array.from(
    panelRef.value.querySelectorAll<HTMLElement>('button:not([disabled])'),
  )
  if (!focusables.length) return
  const first = focusables[0]
  const last = focusables[focusables.length - 1]
  const active = document.activeElement as HTMLElement | null
  if (e.shiftKey && (active === first || !panelRef.value.contains(active))) {
    e.preventDefault(); last.focus()
  } else if (!e.shiftKey && active === last) {
    e.preventDefault(); first.focus()
  }
}

watch(() => props.open, async (v) => {
  if (v) {
    prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    window.addEventListener('keydown', onKeydown)
    await nextTick()
    ;(cancelRef.value || panelRef.value)?.focus()
  } else {
    document.body.style.overflow = prevOverflow
    window.removeEventListener('keydown', onKeydown)
  }
})

onUnmounted(() => {
  document.body.style.overflow = prevOverflow
  window.removeEventListener('keydown', onKeydown)
})

function onSelect(value: string) {
  emit('select', value)
}

function cancel() {
  emit('cancel')
  emit('update:open', false)
}
</script>