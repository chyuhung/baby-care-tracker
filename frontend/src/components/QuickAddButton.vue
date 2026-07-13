<template>
  <!-- 底部固定快速操作栏 -->
  <div class="fixed bottom-16 left-0 right-0 bg-white border-t border-gray-200 px-4 py-3 z-40">
    <div class="flex gap-3 max-w-lg mx-auto">
      <button @click="quickAdd('feeding')"
        class="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-primary text-white rounded-xl text-sm font-medium shadow-sm btn-press">
        <span class="text-lg">🍼</span>
        <span>喂奶</span>
      </button>
      <button @click="quickAdd('diaper')"
        class="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-white text-gray-700 border-2 border-gray-200 rounded-xl text-sm font-medium btn-press"
        style="border-color: #FF6B6B; color: #FF6B6B">
        <span class="text-lg">🧷</span>
        <span>尿布</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const app = useAppStore()

function quickAdd(type: 'feeding' | 'diaper') {
  if (!app.currentBaby()) {
    app.showToast('请先添加宝宝', 'info')
    router.push('/baby/new')
    return
  }
  router.push(`/record/${type}`)
}
</script>
