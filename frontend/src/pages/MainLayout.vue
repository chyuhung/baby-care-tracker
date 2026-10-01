<template>
  <div class="flex flex-col min-h-dvh bg-bg-main">
    <!-- 刻意不加 <transition>：标签页切换按 iOS HIG 瞬时替换（UITabBarController 不做动画）。
         原先的 name="page" mode="out-in" + opacity 0.2s 会「淡出 200ms → 空档 → 淡入 200ms」，
         而页面底色 --bg-main 是浅灰，空档整帧只剩底色，读起来就是一次可见闪动（还白搭 400ms 延迟）。
         keep-alive 已在，切 tab 不会重跑 onMounted/watch，故瞬时切换不会出现骨架屏。 -->
    <router-view v-slot="{ Component }">
      <!-- 标签页 keep-alive：切换标签/进出深页均保留各标签的 DOM、滚动与数据（iOS 标签栏惯例） -->
      <keep-alive include="HomePage,TimelinePage,TrendPage,ProfilePage">
        <component :is="Component" />
      </keep-alive>
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