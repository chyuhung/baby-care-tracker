<template>
  <div class="flex flex-col min-h-dvh bg-bg-main">
    <router-view v-slot="{ Component }">
      <transition name="page" mode="out-in">
        <!-- 标签页 keep-alive：切换标签/进出深页均保留各标签的 DOM、滚动与数据（iOS 标签栏惯例） -->
        <keep-alive include="HomePage,TimelinePage,TrendPage,ProfilePage">
          <component :is="Component" />
        </keep-alive>
      </transition>
    </router-view>
    <BottomNav />
  </div>
</template>

<script setup lang="ts">
import BottomNav from '@/components/BottomNav.vue'
import { useAppStore } from '@/stores/app'
import { onMounted } from 'vue'

// 显式命名：保证 App.vue 的 <keep-alive include="MainLayout"> 命中缓存
defineOptions({ name: 'MainLayout' })

const app = useAppStore()

onMounted(async () => {
  await app.loadBabies()
  app.connectWebSocket()
})
</script>