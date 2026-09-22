<template>
  <!-- 统一空状态：图标 + 主文案 + 副文案 + 可选行动按钮（slot）
       图标采用 SF Symbols 风线性字形（tint 圆底），替代大尺寸位图 logo -->
  <div class="text-center px-4" :class="size === 'sm' ? 'py-8' : 'py-16'">
    <div class="mx-auto mb-4 flex items-center justify-center rounded-full bg-muted"
      :class="size === 'sm' ? 'w-12 h-12' : 'w-16 h-16'" aria-hidden="true">
      <svg class="text-text-secondary/70" :class="size === 'sm' ? 'w-6 h-6' : 'w-8 h-8'"
        viewBox="0 0 24 24" fill="none">
        <path :d="glyph" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </div>
    <p class="text-text-primary font-medium">{{ title }}</p>
    <p v-if="subtitle" class="text-text-secondary text-sm mt-1.5 leading-relaxed">{{ subtitle }}</p>
    <div v-if="$slots.default" class="mt-5">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  title: string
  subtitle?: string
  size?: 'sm' | 'md'
  /** SF Symbols 风字形：clock | chart | folder | moon | drop | sun | pill | bell | box */
  icon?: string
}>(), {
  subtitle: '',
  size: 'md',
  icon: 'box',
})

const GLYPHS: Record<string, string> = {
  // clock.arrow.circlepath
  clock: 'M12 8v4l3 2M4 12a8 8 0 108-8 8 8 0 00-6.9 4M3.5 4v4h4',
  // chart.bar.xaxis
  chart: 'M4 20V10M10 20V4M16 20v-8M4 20h16',
  // folder
  folder: 'M3.5 7.5A1.5 1.5 0 015 6h3.6a1.5 1.5 0 011.2.6l.9 1.2a1.5 1.5 0 001.2.6H19a1.5 1.5 0 011.5 1.5v7A1.5 1.5 0 0119 18.5H5A1.5 1.5 0 013.5 17z',
  // moon.zzz
  moon: 'M20 14.5A8 8 0 019.5 4a8 8 0 1010.5 10.5zM17 5h3l-3 3h3',
  // drop
  drop: 'M12 3.5s5.5 5.6 5.5 9.5a5.5 5.5 0 11-11 0C6.5 9.1 12 3.5 12 3.5z',
  // sun.max
  sun: 'M12 7.5a4.5 4.5 0 100 9 4.5 4.5 0 000-9zM12 2v2.5M12 19.5V22M2 12h2.5M19.5 12H22M4.9 4.9l1.8 1.8M17.3 17.3l1.8 1.8M19.1 4.9l-1.8 1.8M6.7 17.3l-1.8 1.8',
  // pills
  pill: 'M10.5 6.5l7 7a4.95 4.95 0 01-7 7l-7-7a4.95 4.95 0 017-7zM7 10l7 7',
  // bell
  bell: 'M18 8.5a6 6 0 10-12 0c0 5-2 6-2 6h16s-2-1-2-6M10.3 19a2 2 0 003.4 0',
  // tray / box
  box: 'M4 8.5L12 4l8 4.5v7L12 20l-8-4.5zM4 8.5l8 4.5 8-4.5M12 13v7',
}

const glyph = computed(() => GLYPHS[props.icon] || GLYPHS.box)
</script>
