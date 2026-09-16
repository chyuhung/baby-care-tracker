<template>
  <!-- 表单页固定在键盘上方的底部主操作栏（毛玻璃 + 安全区） -->
  <div class="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[480px] z-30 glass-surface border-t border-border-color/60">
    <div class="px-4 pt-3 pb-safe">
      <slot />
    </div>
  </div>
</template>
