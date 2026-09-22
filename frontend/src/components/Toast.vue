<template>
  <!-- iOS 风 HUD：
       · 纯状态提示（无操作）→ 顶部居中胶囊
       · 带「撤销」等操作 → 底部居中（iOS 可操作 snackbar 惯例） -->
  <Teleport to="body">
    <!-- 顶部：状态提示 -->
    <div class="fixed top-[max(env(safe-area-inset-top),0.85rem)] left-1/2 -translate-x-1/2 z-[100] max-w-[92%] space-y-2 pointer-events-none flex flex-col items-center">
      <transition-group name="toast">
        <div v-for="toast in plainToasts" :key="toast.id" role="status" aria-live="polite"
          class="pointer-events-auto inline-flex items-center gap-2 max-w-full pl-3 pr-4 py-2.5 rounded-full bg-[rgba(28,28,30,0.84)] backdrop-blur-xl shadow-lg text-white text-[13px] font-medium">
          <svg v-if="toast.type === 'success'" width="17" height="17" viewBox="0 0 24 24" fill="none" class="shrink-0">
            <circle cx="12" cy="12" r="9.2" stroke="#30D158" stroke-width="1.8" />
            <path d="M8.2 12.4l2.5 2.5L16 9.4" stroke="#FFFFFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <svg v-else-if="toast.type === 'error'" width="17" height="17" viewBox="0 0 24 24" fill="none" class="shrink-0">
            <circle cx="12" cy="12" r="9.2" stroke="#FF453A" stroke-width="1.8" />
            <path d="M12 7.6v5.2M12 16.5h.01" stroke="#FFFFFF" stroke-width="2" stroke-linecap="round" />
          </svg>
          <span class="whitespace-nowrap overflow-hidden text-ellipsis">{{ clean(toast.message) }}</span>
        </div>
      </transition-group>
    </div>

    <!-- 底部：可操作提示（撤销） -->
    <div class="fixed bottom-[calc(env(safe-area-inset-bottom)+5.5rem)] left-1/2 -translate-x-1/2 z-[100] w-[92%] max-w-[440px] space-y-2 pointer-events-none flex flex-col items-stretch">
      <transition-group name="toast">
        <div v-for="toast in actionToasts" :key="toast.id" role="status" aria-live="polite"
          class="pointer-events-auto flex items-center gap-2 pl-4 pr-2 py-2.5 rounded-2xl bg-[rgba(28,28,30,0.9)] backdrop-blur-xl shadow-lg text-white text-[14px] font-medium">
          <span class="flex-1 min-w-0 truncate">{{ clean(toast.message) }}</span>
          <button type="button" @click="runAction(toast)"
            class="shrink-0 px-3 py-1.5 rounded-full text-[14px] font-semibold text-[#5AC8FA] active:opacity-60">
            {{ toast.action?.label }}
          </button>
        </div>
      </transition-group>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/stores/app'
import type { ToastMessage } from '@/stores/app'
import { hapticSuccess } from '@/utils/haptic'
const app = useAppStore()

const plainToasts = computed(() => app.toasts.filter(t => !t.action))
const actionToasts = computed(() => app.toasts.filter(t => t.action))

function runAction(toast: ToastMessage) {
  hapticSuccess()
  const fn = toast.action?.handler
  app.dismissToast(toast.id)
  fn?.()
}

// HUD 已用状态符号表达结果，剔除文案里冗余的 emoji（✅😴🌳 等）
function clean(msg: string) {
  return msg
    .replace(/[\u{1F000}-\u{1FAFF}\u{2600}-\u{27BF}\u{2B00}-\u{2BFF}\u{2190}-\u{21FF}\u{FE0F}\u{200D}]/gu, '')
    .replace(/\s{2,}/g, ' ')
    .trim()
}
</script>
