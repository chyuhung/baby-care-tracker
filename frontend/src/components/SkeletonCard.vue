<template>
  <!-- iOS 风骨架屏：内容型列表首屏占位，避免布局跳动 -->
  <div class="space-y-2" aria-hidden="true">
    <div v-for="i in count" :key="i"
      class="bg-surface rounded-2xl p-4 shadow-card flex items-start gap-3 animate-pulse">
      <div class="w-9 h-9 rounded-xl bg-muted shrink-0"></div>
      <div class="flex-1 min-w-0 space-y-2 pt-0.5">
        <div class="h-3.5 rounded-full bg-muted" :style="{ width: barWidth(i, 1) }"></div>
        <div class="h-3 rounded-full bg-muted/70" :style="{ width: barWidth(i, 2) }"></div>
      </div>
      <div class="w-10 h-3 rounded-full bg-muted shrink-0 mt-1"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ count?: number }>(), { count: 6 })

// 稳定的伪随机宽度，避免每帧跳动
function barWidth(i: number, row: number) {
  const seeds = [[72, 44], [58, 36], [81, 52], [64, 40], [76, 48], [68, 34]]
  const s = seeds[(i - 1) % seeds.length]
  return `${row === 1 ? s[0] : s[1]}%`
}
</script>
