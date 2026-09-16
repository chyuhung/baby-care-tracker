<template>
  <!-- iOS 风深色毛玻璃 HUD：顶部居中胶囊，自带状态符号，文案中的 emoji 自动剔除 -->
  <div class="fixed top-[max(env(safe-area-inset-top),0.85rem)] left-1/2 -translate-x-1/2 z-[100] max-w-[92%] space-y-2 pointer-events-none flex flex-col items-center">
    <transition-group name="toast">
      <div v-for="toast in app.toasts" :key="toast.id" role="status" aria-live="polite"
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
</template>

<script setup lang="ts">
import { useAppStore } from '@/stores/app'
const app = useAppStore()

// HUD 已用状态符号表达结果，剔除文案里冗余的 emoji（✅😴🌳 等）
function clean(msg: string) {
  return msg
    .replace(/[\u{1F000}-\u{1FAFF}\u{2600}-\u{27BF}\u{2B00}-\u{2BFF}\u{2190}-\u{21FF}\u{FE0F}\u{200D}]/gu, '')
    .replace(/\s{2,}/g, ' ')
    .trim()
}
</script>
