<template>
  <div class="flex flex-col min-h-dvh bg-bg-main">
    <router-view v-slot="{ Component }">
      <transition name="page" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
    <BottomNav />
  </div>
</template>

<script setup lang="ts">
import BottomNav from '@/components/BottomNav.vue'
import { useAppStore } from '@/stores/app'
import { onMounted, onUnmounted } from 'vue'
import { loadReminders, startReminderScheduler } from '@/utils/reminders'

const app = useAppStore()
let stopScheduler: (() => void) | null = null

onMounted(async () => {
  await app.loadBabies()
  app.connectWebSocket()
  stopScheduler = startReminderScheduler(() => loadReminders())
})

onUnmounted(() => {
  stopScheduler?.()
})
</script>
