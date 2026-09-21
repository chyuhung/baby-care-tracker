<template>
  <!-- 统一空状态：图标 + 主文案 + 副文案 + 可选行动按钮（slot） -->
  <div class="text-center px-4" :class="size === 'sm' ? 'py-8' : 'py-16'">
    <img src="/icon-192.png" alt="" class="mx-auto block mb-4" :class="size === 'sm' ? 'w-12 h-12' : 'w-16 h-16'" />
    <p class="text-text-primary font-medium">{{ title }}</p>
    <p v-if="subtitle" class="text-text-secondary text-sm mt-1.5 leading-relaxed">{{ subtitle }}</p>
    <div v-if="$slots.default" class="mt-5">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  title: string
  subtitle?: string
  size?: 'sm' | 'md'
}>(), {
  subtitle: '',
  size: 'md',
})
</script>
