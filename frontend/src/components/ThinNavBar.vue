<template>
  <!-- 轻薄导航栏：固定在滚动容器顶部，点击标题回顶；actions 区不触发回顶 -->
  <header class="sticky top-0 z-30 nav-surface hairline-bottom pt-safe" role="button" aria-label="返回顶部"
    @click="scrollToTop">
    <div class="flex h-11 items-center justify-between gap-3 px-4">
      <h1 class="min-w-0 truncate text-[17px] font-semibold leading-tight text-text-primary">{{ title }}</h1>
      <div v-if="$slots.actions" class="flex flex-shrink-0 items-center gap-2" @click.stop>
        <slot name="actions" />
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
defineProps<{ title: string }>()

// 单击标题栏 → 各页 PullRefresh 容器回顶
function scrollToTop() {
  window.dispatchEvent(new CustomEvent('app:scroll-to-top'))
}
</script>