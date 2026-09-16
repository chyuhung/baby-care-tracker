<template>
  <!-- iOS 风圆弧活动指示器：颜色由父级 currentColor（text-*）控制 -->
  <svg :width="size" :height="size" viewBox="0 0 24 24" fill="none" aria-hidden="true" class="icon-spin shrink-0">
    <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2.6" stroke-linecap="round"
      stroke-dasharray="40 56.55" transform="rotate(-90 12 12)" />
  </svg>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ size?: number }>(), { size: 22 })
</script>
