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

// 状态栏/chrome 颜色：与三段式顶栏保持一致，统一为白色
const THEME_COLOR: Record<string, string> = {
  male: '#FFFFFF',
  female: '#FFFFFF',
  neutral: '#FFFFFF',
}
watch(() => app.theme, (t) => {
  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta) meta.setAttribute('content', THEME_COLOR[t] || THEME_COLOR.neutral)
}, { immediate: true })

onMounted(() => {
  auth.restoreSession()
})
</script>
