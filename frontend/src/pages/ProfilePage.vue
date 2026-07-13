<template>
  <div class="flex flex-col min-h-screen">
    <header class="pt-safe bg-white px-4 py-3 border-b border-border-color">
      <h1 class="text-lg font-bold text-text-primary">我的</h1>
    </header>

    <main class="flex-1 px-4 py-4 space-y-4 overflow-y-auto pb-20 min-h-0">
      <!-- 用户信息 -->
      <div class="bg-white rounded-2xl p-4 shadow-card flex items-center gap-4">
        <div class="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center text-2xl">👤</div>
        <div>
          <div class="font-semibold text-text-primary">{{ auth.user?.username }}</div>
          <div class="text-sm text-text-secondary mt-0.5">家庭成员</div>
        </div>
      </div>

      <!-- 宝宝列表 -->
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-semibold text-text-secondary uppercase tracking-wide">宝宝档案</h2>
          <router-link to="/baby/new" class="text-primary text-sm font-medium flex items-center gap-1">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
            添加
          </router-link>
        </div>

        <div v-if="app.babies.length === 0" class="bg-white rounded-2xl p-6 text-center shadow-card">
          <div class="text-4xl mb-2">👶</div>
          <p class="text-text-secondary text-sm">还没有宝宝档案</p>
        </div>

        <div v-for="baby in app.babies" :key="baby.id">
          <!-- 宝宝卡片 -->
          <div class="bg-white rounded-2xl p-4 shadow-card">
            <div class="flex items-center gap-3" @click="router.push(`/baby/${baby.id}/edit`)">
              <div class="w-12 h-12 rounded-full flex items-center justify-center text-xl font-bold text-white cursor-pointer" :style="{ background: baby.avatar_color }">
                {{ baby.name[0] }}
              </div>
              <div class="flex-1 cursor-pointer">
                <div class="font-semibold text-text-primary">{{ baby.name }}</div>
                <div class="text-xs text-text-secondary mt-0.5">{{ baby.birth_date }}</div>
              </div>
              <svg class="w-5 h-5 text-text-secondary/50" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
            </div>
          </div>
        </div>
      </div>

      <!-- 登出 -->
      <button @click="logout" class="w-full py-3 bg-white text-red-500 font-medium rounded-xl shadow-card btn-press mt-8">
        退出登录
      </button>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const auth = useAuthStore()
const app = useAppStore()

function logout() {
  app.disconnectWebSocket()
  auth.logout()
  router.push('/login')
}
</script>
