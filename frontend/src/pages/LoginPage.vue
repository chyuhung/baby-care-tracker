<template>
  <div class="min-h-dvh bg-bg-main flex flex-col items-center justify-center px-6">
    <!-- Logo -->
    <div class="mb-10 text-center">
      <img src="/icon-192.png" alt="宝宝护理记录" class="w-20 h-20 mb-4 mx-auto block drop-shadow-sm" />
      <h1 class="text-2xl font-bold text-text-primary">宝宝护理记录</h1>
      <p class="text-text-secondary text-sm mt-1">记录宝宝成长，每一刻都珍贵</p>
    </div>

    <!-- 表单（回车直接提交） -->
    <form class="w-full max-w-xs space-y-4" @submit.prevent="submit">
      <div class="space-y-1">
        <label class="text-sm text-text-secondary">用户名</label>
        <input v-model="form.username" type="text" :placeholder="isRegister ? '2-20位字符' : '输入用户名'"
          autocomplete="username" autocapitalize="none" autocorrect="off" spellcheck="false" enterkeyhint="next"
            class="w-full px-4 py-3 bg-surface border border-border-color rounded-xl text-text-primary placeholder-text-secondary/50 focus:border-primary focus:outline-none transition-colors" />
      </div>

      <div class="space-y-1">
        <label class="text-sm text-text-secondary">{{ isRegister ? '设置密码' : '密码' }}</label>
        <div class="relative">
          <input v-model="form.password" :type="showPassword ? 'text' : 'password'" :placeholder="isRegister ? '至少6位' : '输入密码'"
            :autocomplete="isRegister ? 'new-password' : 'current-password'" enterkeyhint="done"
            class="w-full px-4 py-3 pr-12 bg-surface border border-border-color rounded-xl text-text-primary placeholder-text-secondary/50 focus:border-primary focus:outline-none transition-colors" />
          <button type="button" :aria-label="showPassword ? '隐藏密码' : '显示密码'" @click="showPassword = !showPassword"
            class="absolute right-1 top-1/2 -translate-y-1/2 w-11 h-11 flex items-center justify-center text-text-secondary btn-press">
            <svg v-if="showPassword" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/></svg>
            <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
          </button>
        </div>
      </div>

      <div v-if="error" class="bg-danger-light text-danger text-sm px-4 py-2 rounded-xl text-center">
        {{ error }}
      </div>

      <button type="submit" :disabled="loading"
        class="btn-press w-full py-3.5 bg-primary-fill text-white font-semibold rounded-xl shadow-card hover:shadow-card-hover transition-all disabled:opacity-50 flex items-center justify-center gap-2">
        <ActivityIndicator v-if="loading" :size="18" class="text-white" />
        <span>{{ loading ? '处理中...' : (isRegister ? '注册' : '登录') }}</span>
      </button>

      <div class="text-center">
        <button type="button" @click="isRegister = !isRegister; error = ''" class="text-primary-deep text-sm hover:underline min-h-[44px]">
          {{ isRegister ? '已有账号？登录' : '没有账号？注册' }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import ActivityIndicator from '@/components/ActivityIndicator.vue'

const router = useRouter()
const auth = useAuthStore()
const app = useAppStore()

const isRegister = ref(false)
const loading = ref(false)
const showPassword = ref(false)
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
