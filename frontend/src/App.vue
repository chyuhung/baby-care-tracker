<template>
  <div class="min-h-dvh bg-bg-main flex flex-col" :data-theme="app.theme">
    <Toast />
    <router-view v-slot="{ Component }">
      <transition name="page" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </div>
</template>

<script setup lang="ts">
import Toast from '@/components/Toast.vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { onMounted, watch } from 'vue'

const auth = useAuthStore()
const app = useAppStore()

// 主题色随粉/蓝主题动态更新（iOS 状态栏 / Android chrome）
const THEME_COLOR: Record<string, string> = {
  male: '#F4F9FF',
  female: '#FFF6FA',
  neutral: '#FFF6FA',
}
watch(() => app.theme, (t) => {
  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta) meta.setAttribute('content', THEME_COLOR[t] || THEME_COLOR.neutral)
}, { immediate: true })

onMounted(() => {
  auth.restoreSession()
})
</script>
