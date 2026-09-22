<template>
  <Teleport to="body">
    <transition name="sheet-mask">
      <div v-if="open" class="fixed inset-0 z-[95] bg-black/35" @click.self="onBackdrop">
        <transition name="sheet-panel" appear>
          <div v-if="open"
            class="absolute bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[480px] px-3 pb-safe pt-2 space-y-2"
            role="dialog" aria-modal="true" :aria-label="title">
            <!-- 预览头（iOS 上下文菜单顶部信息行） -->
            <div v-if="title || subtitle" class="flex items-center gap-3 px-4 py-3 bg-surface rounded-2xl">
              <span v-if="emoji" class="text-2xl leading-none">{{ emoji }}</span>
              <div class="min-w-0">
                <p class="text-[15px] font-semibold text-text-primary truncate">{{ title }}</p>
                <p v-if="subtitle" class="text-[13px] text-text-secondary truncate">{{ subtitle }}</p>
              </div>
            </div>
            <!-- 操作组 -->
            <div class="bg-surface rounded-2xl overflow-hidden">
              <button v-for="(a, i) in actions" :key="a.key" type="button" @click="run(a)"
                class="w-full flex items-center gap-3 px-4 min-h-[52px] text-[17px] btn-press active:bg-muted/60"
                :class="[i > 0 ? 'border-t border-border-color/70' : '', a.danger ? 'text-danger' : 'text-text-primary']">
                <svg v-if="a.icon" class="w-5 h-5 shrink-0" viewBox="0 0 24 24" fill="none"
                  :class="a.danger ? 'text-danger' : 'text-text-secondary'">
                  <path :d="a.icon" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
                <span class="flex-1 text-left">{{ a.label }}</span>
              </button>
            </div>
            <!-- 取消 -->
            <button type="button" @click="close"
              class="w-full py-4 bg-surface rounded-2xl text-[17px] font-semibold text-text-primary btn-press">
              取消
            </button>
          </div>
        </transition>
      </div>
    </transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { hapticTap } from '@/utils/haptic'

export interface ContextAction {
  key: string
  label: string
  icon?: string
  danger?: boolean
}

const props = defineProps<{
  open: boolean
  title?: string
  subtitle?: string
  emoji?: string
  actions: ContextAction[]
}>()

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'select', key: string): void
}>()

const close = () => { hapticTap(); emit('update:open', false) }

// 长按触发时，touchend 后紧跟的合成 click 会落在刚出现的遮罩上；
// 用极短的开启宽限期忽略这一击，避免菜单瞬开瞬关。
let openedAt = 0
function onBackdrop() {
  if (Date.now() - openedAt < 380) return
  close()
}

function run(a: ContextAction) {
  emit('select', a.key)
  emit('update:open', false)
}

function onKey(e: KeyboardEvent) { if (e.key === 'Escape') close() }

watch(() => props.open, (v) => {
  if (typeof document === 'undefined') return
  if (v) openedAt = Date.now()
  document.body.style.overflow = v ? 'hidden' : ''
  if (v) window.addEventListener('keydown', onKey)
  else window.removeEventListener('keydown', onKey)
})
onUnmounted(() => {
  if (typeof document !== 'undefined') {
    document.body.style.overflow = ''
    window.removeEventListener('keydown', onKey)
  }
})
</script>
