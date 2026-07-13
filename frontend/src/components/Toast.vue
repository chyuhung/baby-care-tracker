<template>
  <div class="fixed top-4 left-1/2 -translate-x-1/2 z-[100] w-[90%] max-w-[380px] space-y-2 pointer-events-none">
    <transition-group name="toast">
      <div v-for="toast in app.toasts" :key="toast.id"
        :class="['px-4 py-3 rounded-xl shadow-lg text-white text-sm font-medium pointer-events-auto flex items-center gap-2',
          toast.type === 'success' ? 'bg-success' : toast.type === 'error' ? 'bg-red-500' : 'bg-text-secondary']">
        <span>{{ toast.type === 'success' ? '✅' : toast.type === 'error' ? '❌' : 'ℹ️' }}</span>
        {{ toast.message }}
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { useAppStore } from '@/stores/app'
const app = useAppStore()
</script>
