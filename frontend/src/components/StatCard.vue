<template>
  <div class="bg-white rounded-2xl p-4 shadow-card">
    <div class="flex items-center justify-between">
      <div>
        <div class="text-xs text-text-secondary mb-1">{{ title }}</div>
        <div class="text-3xl font-bold font-num" :class="colorClass">{{ value }}</div>
        <div class="text-xs text-text-secondary mt-0.5">{{ unit }}</div>
      </div>
      <div class="text-3xl">{{ icon }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title: string
  value: number
  unit: string
  icon: string
  color: string
}>()

const colorClass = computed(() => ({
  primary: 'text-primary',
  secondary: 'text-[#FF6B6B]',
}[props.color] || 'text-primary'))
</script>
