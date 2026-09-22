<template>
  <!-- iOS 风左滑删除：仅水平位移超过阈值才激活，纵向滚动不受影响 -->
  <div class="relative rounded-2xl overflow-hidden select-none"
    @touchstart.passive="onStart" @touchmove="onMove" @touchend="onEnd" @touchcancel="onEnd">
    <!-- 底层删除动作 -->
    <div class="absolute inset-y-0 right-0 w-20 flex items-center justify-center bg-danger text-white"
      :style="{ opacity: revealed ? 1 : 0 }" aria-hidden="true">
      <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none">
        <path d="M4 7h16M9 7V5a1 1 0 011-1h4a1 1 0 011 1v2m-8 0v12a1 1 0 001 1h6a1 1 0 001-1V7" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </div>
    <!-- 前景内容 -->
    <div class="relative bg-transparent will-change-transform"
      :style="{ transform: `translateX(${offset}px)`, transition: dragging ? 'none' : 'transform 0.24s cubic-bezier(0.32,0.72,0,1)' }">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { hapticHeavy } from '@/utils/haptic'

const emit = defineEmits<{ (e: 'delete'): void }>()

const MAX = 80          // 最大滑出宽度
const TRIGGER = 56      // 触发删除的阈值
const SLOP = 8          // 方向判定死区

const offset = ref(0)
const dragging = ref(false)
const revealed = ref(false)
let startX = 0, startY = 0, base = 0, horiz: boolean | null = null

function onStart(e: TouchEvent) {
  if (e.touches.length !== 1) return
  startX = e.touches[0].clientX
  startY = e.touches[0].clientY
  base = offset.value
  horiz = null
  dragging.value = true
}

function onMove(e: TouchEvent) {
  if (!dragging.value || e.touches.length !== 1) return
  const dx = e.touches[0].clientX - startX
  const dy = e.touches[0].clientY - startY
  if (horiz === null) {
    if (Math.abs(dx) < SLOP && Math.abs(dy) < SLOP) return
    horiz = Math.abs(dx) > Math.abs(dy) * 1.5
    if (!horiz) { dragging.value = false; return }
  }
  if (!horiz) return
  // 只允许向左滑出 [-MAX, 0]
  let next = base + dx
  if (next > 0) next = 0
  if (next < -MAX) next = -MAX
  offset.value = next
  if (e.cancelable) e.preventDefault()
}

function onEnd() {
  if (!dragging.value && horiz === null) return
  dragging.value = false
  if (offset.value <= -TRIGGER) {
    offset.value = 0
    revealed.value = false
    hapticHeavy()
    emit('delete')
  } else {
    offset.value = 0
  }
  revealed.value = false
  horiz = null
}
</script>
