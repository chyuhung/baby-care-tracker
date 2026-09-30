<template>
  <nav class="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[480px] glass-surface hairline-top pb-safe z-40">
    <div class="flex items-center justify-around h-16">
      <router-link v-for="tab in tabs" :key="tab.to" :to="tab.to"
        :aria-current="isActive(tab.to) ? 'page' : undefined" :aria-label="tab.label"
        :class="['flex flex-col items-center justify-center w-16 h-full transition-colors relative',
          isActive(tab.to) ? 'text-primary-deep' : 'text-text-secondary']">
        <svg class="w-6 h-6"
             :fill="isActive(tab.to) ? 'currentColor' : 'none'"
             :stroke="isActive(tab.to) ? 'none' : 'currentColor'"
             :stroke-width="isActive(tab.to) ? 0 : 2"
             stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
          <path :d="isActive(tab.to) ? tab.activeIcon : tab.icon"
                :fill-rule="isActive(tab.to) ? 'evenodd' : 'inherit'" />
        </svg>
        <span class="text-xs mt-1 font-medium">{{ tab.label }}</span>
      </router-link>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()

const tabs = [
  {
    to: '/',
    label: '记录',
    icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a2 2 0 01.707.293l5.414 5.414a2 2 0 01.293.707V19a2 2 0 01-2 2z',
    // 与描边同轮廓的实心文档（含右下折角）+ evenodd 挖 2 条横线白槽（与原描边线同位）
    activeIcon: 'M17 21H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a2 2 0 01.707.293l5.414 5.414a2 2 0 01.293.707V19a2 2 0 01-2 2zM9 11h6v2H9zM9 15h6v2H9z',
  },
  {
    to: '/timeline',
    label: '时间线',
    icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
    // 同轮廓实心圆盘 + evenodd 挖贯通圆心的 L 形指针白槽（原描边指针的坐标，宽 2）
    activeIcon: 'M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0zM11 7h2v6h-2zM11.293 12.707l3 3l1.414-1.414l-3-3z',
  },
  {
    to: '/trend',
    label: '趋势',
    icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z',
    // 3 根实心圆角柱，外轮廓与描边版逐点一致（顶 11/7/5、底 21、r 2）
    activeIcon: 'M9 13a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h2a2 2 0 0 0 2-2zM9 9a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2zM15 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2z',
  },
  {
    to: '/profile',
    label: '我的',
    icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z',
    // 同轮廓实心填充（头 + 肩，双闭合子路径）
    activeIcon: 'M16 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0zM12 14a7 7 0 0 0-7 7h14a7 7 0 0 0-7-7z',
  },
]

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>
