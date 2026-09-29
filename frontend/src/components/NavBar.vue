<template>
  <!-- 微信式居中导航栏：标题固定 17pt 居中（全站 10 处标题栏的唯一来源）。
       Grid 1fr auto 1fr —— 左右两格等宽，标题恒居中且不会被按钮压住（标题 truncate 兜底）；
       不用绝对居中是因为右侧有 actions 时（成长记录「记录」胶囊）会与标题重叠。
       顶栏不透明（.nav-surface），与 iOS 状态栏同色、无接缝；点击标题回顶，actions 区不触发回顶 -->
  <header class="sticky top-0 z-30 nav-surface hairline-bottom pt-safe" role="button" aria-label="返回顶部"
    @click="scrollToTop">
    <div class="grid h-11 grid-cols-[1fr_auto_1fr] items-center gap-2 px-4">
      <div class="flex min-w-0 items-center">
        <slot name="left" />
      </div>
      <h1 class="truncate text-center text-[17px] font-semibold leading-tight text-text-primary">{{ title }}</h1>
      <div class="flex min-w-0 items-center justify-end">
        <div v-if="$slots.actions" class="flex flex-shrink-0 items-center gap-2" @click.stop>
          <slot name="actions" />
        </div>
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
