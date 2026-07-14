<template>
  <nav class="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[480px] bg-white border-t border-border-color pb-safe z-40">
    <div class="flex items-center justify-around h-16">
      <router-link v-for="tab in tabs" :key="tab.to" :to="tab.to"
        :class="['flex flex-col items-center justify-center w-16 h-full transition-colors relative',
          isActive(tab.to) ? 'text-primary' : 'text-text-secondary']">
        <svg v-if="isActive(tab.to)" class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
          <path :d="tab.activeIcon" />
        </svg>
        <svg v-else class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" :d="tab.icon" />
        </svg>
        <span class="text-[10px] mt-1 font-medium">{{ tab.label }}</span>
        <span v-if="isActive(tab.to)" class="absolute -top-0.5 left-1/2 -translate-x-1/2 w-1.5 h-1.5 bg-primary rounded-full"></span>
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
    icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
    activeIcon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
  },
  {
    to: '/timeline',
    label: '时间线',
    icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
    activeIcon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
  },
  {
    to: '/profile',
    label: '我的',
    icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z',
    activeIcon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z',
  },
]

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>
