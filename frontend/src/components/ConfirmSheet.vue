<template>
  <Teleport to="body">
    <transition name="sheet-mask">
      <div v-if="open" class="fixed inset-0 z-[80] flex items-end justify-center bg-black/35" @click.self="cancel">
        <transition name="sheet-panel" appear>
          <div v-if="open" ref="panelRef" class="w-full max-w-[480px] px-3 pb-safe space-y-2" role="alertdialog" aria-modal="true"
            :aria-label="title || confirmText">
            <!-- 上组：说明 + 确认按钮 -->
            <div class="bg-surface rounded-2xl overflow-hidden text-center">
              <div class="pt-2.5 flex justify-center">
                <span class="w-9 h-1 rounded-full bg-border-color"></span>
              </div>
              <div class="px-4 pt-2.5 pb-4">
                <p v-if="title" class="text-[15px] font-semibold text-text-primary">{{ title }}</p>
                <p v-if="message" class="text-[13px] text-text-secondary mt-1 leading-relaxed whitespace-pre-line">{{ message }}</p>
              </div>
              <button ref="confirmRef" type="button" @click="onConfirm" :disabled="loading"
                class="w-full py-4 px-4 text-[17px] font-medium border-t border-border-color/70 btn-press disabled:opacity-60 flex items-center justify-center gap-2"
                :class="danger ? 'text-danger' : 'text-primary-deep'">
                <ActivityIndicator v-if="loading" :size="18" />
                <span>{{ confirmText }}</span>
              </button>
            </div>
            <!-- 取消（独立成组） -->
            <button ref="cancelRef" type="button" @click="cancel"
              class="w-full py-4 bg-surface rounded-2xl text-[17px] font-semibold text-text-primary btn-press">
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
import { hapticWarning, hapticTap } from '@/utils/haptic'

const props = withDefaults(defineProps<{
  open: boolean
  title?: string
  message?: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
  loading?: boolean
}>(), {
  title: '',
  message: '',
  confirmText: '确认',
  cancelText: '取消',
  danger: true,
  loading: false,
})

const emit = defineEmits<{
  (e: 'confirm'): void
  (e: 'cancel'): void
  (e: 'update:open', v: boolean): void
}>()

const panelRef = ref<HTMLElement | null>(null)
const confirmRef = ref<HTMLElement | null>(null)
const cancelRef = ref<HTMLElement | null>(null)
let prevOverflow = ''

/* 焦点陷阱（a11y）：弹层打开时把焦点锁在面板内，Esc 关闭 */
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
    ;(cancelRef.value || confirmRef.value || panelRef.value)?.focus()
  } else {
    document.body.style.overflow = prevOverflow
    window.removeEventListener('keydown', onKeydown)
  }
})

onUnmounted(() => {
  document.body.style.overflow = prevOverflow
  window.removeEventListener('keydown', onKeydown)
})

function onConfirm() {
  hapticWarning()
  emit('confirm')
}

function cancel() {
  hapticTap()
  emit('cancel')
  emit('update:open', false)
}
</script>
