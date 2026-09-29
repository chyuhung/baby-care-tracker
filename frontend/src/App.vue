<template>
  <div class="min-h-dvh bg-bg-main flex flex-col" :data-theme="app.theme">
    <Toast />
    <router-view v-slot="{ Component }">
      <transition name="page" mode="out-in">
        <!-- 缓存 MainLayout：进入深页（成长/记录表单等）不销毁标签页，返回时滚动与状态原样恢复。
             登录态驱动 include：登出即清空缓存，避免重登录闪旧账号数据 -->
        <keep-alive :include="cacheInclude">
          <component :is="Component" />
        </keep-alive>
      </transition>
    </router-view>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Toast from '@/components/Toast.vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { onMounted } from 'vue'

const auth = useAuthStore()
const app = useAppStore()

const cacheInclude = computed(() => (auth.isLoggedIn ? 'MainLayout' : ''))

// 状态栏/chrome 颜色由 index.html 的 media-scoped theme-color 处理：
// 浅色 #FFFFFF / 深色 #1C1C1E，与 CSS 的三段式 chrome（--surface）一致。
onMounted(() => {
  auth.restoreSession()
})
</script>
