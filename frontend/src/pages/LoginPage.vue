<template>
  <div class="min-h-screen bg-bg-main flex flex-col items-center justify-center px-6">
    <!-- Logo -->
    <div class="mb-10 text-center">
      <div class="text-6xl mb-4">🍼</div>
      <h1 class="text-2xl font-bold text-text-primary">宝宝护理记录</h1>
      <p class="text-text-secondary text-sm mt-1">记录宝宝成长，每一刻都珍贵</p>
    </div>

    <!-- 表单 -->
    <div class="w-full max-w-xs space-y-4">
      <div class="space-y-1">
        <label class="text-sm text-text-secondary">用户名</label>
        <input v-model="form.username" type="text" :placeholder="isRegister ? '2-20位字符' : '输入用户名'" autocomplete="off"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary placeholder-text-secondary/50 focus:border-primary focus:outline-none transition-colors" />
      </div>

      <div class="space-y-1">
        <label class="text-sm text-text-secondary">{{ isRegister ? '设置密码' : '密码' }}</label>
        <input v-model="form.password" type="password" :placeholder="isRegister ? '至少6位' : '输入密码'" :autocomplete="isRegister ? 'new-password' : 'current-password'"
          class="w-full px-4 py-3 bg-white border border-border-color rounded-xl text-text-primary placeholder-text-secondary/50 focus:border-primary focus:outline-none transition-colors" />
      </div>

      <div v-if="error" class="bg-red-50 text-red-500 text-sm px-4 py-2 rounded-xl text-center">
        {{ error }}
      </div>

      <button @click="submit" :disabled="loading"
        class="btn-press w-full py-3 bg-primary text-white font-semibold rounded-xl shadow-card hover:shadow-card-hover transition-all disabled:opacity-50">
        {{ loading ? '处理中...' : (isRegister ? '注册' : '登录') }}
      </button>

      <div class="text-center">
        <button @click="isRegister = !isRegister; error = ''" class="text-primary text-sm hover:underline">
          {{ isRegister ? '已有账号？登录' : '没有账号？注册' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const auth = useAuthStore()
const app = useAppStore()

const isRegister = ref(false)
const loading = ref(false)
const error = ref('')
const form = reactive({ username: '', password: '' })

async function submit() {
  error.value = ''
  if (!form.username.trim()) { error.value = '请输入用户名'; return }
  if (form.password.length < 6) { error.value = '密码至少6位'; return }
  loading.value = true
  try {
    if (isRegister.value) {
      await auth.register(form.username.trim(), form.password)
    } else {
      await auth.login(form.username.trim(), form.password)
    }
    await app.loadBabies()
    app.connectWebSocket()
    router.push('/')
  } catch (e: any) {
    error.value = e.response?.data?.error || '操作失败，请重试'
  } finally {
    loading.value = false
  }
}
</script>
