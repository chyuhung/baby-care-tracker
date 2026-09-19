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
import { onMounted } from 'vue'

const auth = useAuthStore()
const app = useAppStore()

// 状态栏/chrome 颜色由 index.html 的 media-scoped theme-color 处理：
// 浅色 #FFFFFF / 深色 #1C1C1E，与 CSS 的三段式 chrome（--surface）一致。
onMounted(() => {
  auth.restoreSession()
})
</script>
