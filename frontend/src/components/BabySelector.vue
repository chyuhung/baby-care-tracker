<template>
  <div v-if="app.babies.length > 0" class="flex gap-2 mt-3 overflow-x-auto pb-1">
    <button v-for="baby in app.babies" :key="baby.id"
      @click="app.setCurrentBaby(baby.id)"
      :class="['flex-shrink-0 flex items-center gap-2 px-3 py-2 rounded-xl text-sm font-medium transition-all btn-press',
        app.currentBabyId === baby.id
          ? 'bg-primary text-white shadow-card'
          : 'bg-white border border-border-color text-text-secondary']">
      <span class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold text-white" :style="{ background: baby.avatar_color }">
        {{ baby.name[0] }}
      </span>
      {{ baby.name }}
    </button>
    <router-link to="/baby/new"
      class="flex-shrink-0 flex items-center gap-1 px-3 py-2 rounded-xl text-sm text-primary bg-primary/10 btn-press">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
      宝宝
    </router-link>
  </div>
</template>

<script setup lang="ts">
import { useAppStore } from '@/stores/app'
const app = useAppStore()
</script>
