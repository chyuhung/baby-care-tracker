<template>
  <div class="min-h-dvh bg-bg-main flex flex-col" :data-theme="app.theme">
    <Toast />
    <!-- 刻意不加 <transition>：深页 push/pop 与登录切换同样瞬时替换。
         name="page" mode="out-in" + opacity 0.2s 的交接空档（旧页淡出结束 → 新页插入之间必然空一帧）
         会被浅灰页面底色放大成可见闪动；去掉后切页零延迟。下方注释记了将来若要上
         iOS 标准横向滑动的约束：transform 层在 iOS 上会吞掉新页首点，须真机验证再用。 -->
    <router-view v-slot="{ Component }">
      <!-- 缓存 MainLayout：进入深页（成长/记录表单等）不销毁标签页，返回时滚动与状态原样恢复。
           登录态驱动 include：登出即清空缓存，避免重登录闪旧账号数据 -->
      <keep-alive :include="cacheInclude">
        <component :is="Component" />
      </keep-alive>
    </router-view>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Toast from '@/components/Toast.vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { onMounted } from 'vue'

const auth = useAuthStore()
const app = useAppStore()

const cacheInclude = computed(() => (auth.isLoggedIn ? 'MainLayout' : ''))

// 状态栏/chrome 颜色由 index.html 的 media-scoped theme-color 处理：
// 浅色 #FFFFFF / 深色 #1C1C1E = --surface（= NavBar 顶栏色）。
// `html` 画布同色（style.css），故安装态与浏览器顶部三处（状态栏/chrome、画布、标题栏）恒为一体。
onMounted(() => {
  auth.restoreSession()
})
</script>
