<template>
  <div class="flex flex-col min-h-screen">
    <header class="app-header pt-safe px-4 py-3 border-b border-border-color">
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

      <!-- 家庭信息 -->
      <div class="bg-white rounded-2xl p-4 shadow-card space-y-3">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-semibold text-text-secondary uppercase tracking-wide">我的家庭</h2>
          <button v-if="family && family.members.length > 1" @click="leaveFamily" class="text-xs text-red-400 font-medium">退出家庭</button>
        </div>

        <!-- 当前家庭信息 -->
        <div v-if="family" class="space-y-3">
          <div class="bg-bg-secondary rounded-xl p-3">
            <div class="text-xs text-text-secondary mb-1">邀请码</div>
            <div class="flex items-center justify-between">
              <span class="text-lg font-bold tracking-widest text-primary select-all">{{ family.invite_code }}</span>
              <button @click="copyCode" class="text-xs text-primary font-medium">复制</button>
            </div>
          </div>

          <div>
            <div class="text-xs text-text-secondary mb-2">家庭成员 ({{ family.members.length }}人)</div>
            <div class="flex flex-wrap gap-2">
              <div v-for="m in family.members" :key="m.id" class="flex items-center gap-1.5 bg-bg-secondary rounded-full px-3 py-1.5 text-sm">
                <span>👤</span>
                <span>{{ m.username }}</span>
                <span v-if="m.id === auth.user?.id" class="text-xs text-text-secondary">(我)</span>
              </div>
            </div>
          </div>

          <button @click="regenerateCode" class="w-full py-2 text-sm text-primary font-medium rounded-xl border border-primary/30 btn-press">
            重新生成邀请码
          </button>
        </div>

        <!-- 加入其他家庭（始终显示） -->
        <div class="border-t border-border-color pt-3">
          <p class="text-xs text-text-secondary mb-2">加入其他家庭后，你和你的宝宝数据将切换到新家庭</p>
          <div class="flex gap-2">
            <input v-model="joinCode" placeholder="输入对方的邀请码" maxlength="6" class="flex-1 px-3 py-2 border border-border-color rounded-xl text-sm focus:outline-none focus:border-primary uppercase" />
            <button @click="joinFamily" class="px-4 py-2 bg-primary text-white text-sm font-medium rounded-xl btn-press">加入</button>
          </div>
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
          <div class="bg-white rounded-2xl p-4 shadow-card">
            <div class="flex items-center gap-3" @click="router.push(`/baby/${baby.id}/edit`)">
              <div class="w-12 h-12 rounded-full flex items-center justify-center text-xl font-bold text-white cursor-pointer" :style="{ background: baby.avatar_color }">
                {{ baby.name[0] }}
              </div>
              <div class="flex-1 cursor-pointer">
                <div class="font-semibold text-text-primary">{{ baby.name }}</div>
                 <div class="text-xs text-text-secondary mt-0.5">{{ formatBirthDate(baby.birth_date) }}</div>
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
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { familyAPI } from '@/api'

interface FamilyMember {
  id: number
  username: string
}

interface Family {
  id: number
  invite_code: string
  members: FamilyMember[]
}

const router = useRouter()
const auth = useAuthStore()
const app = useAppStore()

const family = ref<Family | null>(null)
const joinCode = ref('')

async function loadFamily() {
  try {
    const res = await familyAPI.getMyFamily()
    family.value = { ...res.data.family, members: res.data.members }
  } catch {}
}

async function joinFamily() {
  if (!joinCode.value.trim()) return
  const code = joinCode.value.trim().toUpperCase()

  // 如果用户有家庭成员，先确认
  if (family.value && family.value.members.length > 1) {
    if (!confirm(`加入新家庭后，将退出当前家庭。\n\n你的宝宝数据会跟随你到新家庭，原家庭成员将无法看到你的数据。\n确定继续？`)) return
  }

  try {
    await familyAPI.join(code)
    joinCode.value = ''
    await loadFamily()
    await app.loadBabies()
    app.showToast('已加入家庭', 'success')
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '加入失败', 'error')
  }
}

async function leaveFamily() {
  if (!confirm('确定退出当前家庭？')) return
  try {
    await familyAPI.leave()
    family.value = null
    await app.loadBabies()
    app.showToast('已退出家庭', 'success')
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '退出失败', 'error')
  }
}

async function regenerateCode() {
  try {
    const res = await familyAPI.regenerateCode()
    family.value!.invite_code = res.data.invite_code
    app.showToast('邀请码已更新', 'success')
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '操作失败', 'error')
  }
}

function copyCode() {
  if (!family.value) return
  navigator.clipboard.writeText(family.value.invite_code).then(() => {
    app.showToast('已复制邀请码', 'success')
  }).catch(() => {
    app.showToast('复制失败', 'error')
  })
}

function formatBirthDate(bd: string) {
  if (!bd) return ''
  // 手动解析，避免 new Date() 在缺少时区/秒数时的跨浏览器歧义
  const m = bd.match(/^(\d{4})-(\d{2})-(\d{2})(?:T(\d{2}):(\d{2}))?/)
  if (m) {
    const [, y, mo, d, h, mi] = m
    if (h) return `${y}-${mo}-${d} ${h}:${mi}`
    return `${y}-${mo}-${d}`
  }
  return bd
}

onMounted(() => {
  loadFamily()
})

function logout() {
  app.disconnectWebSocket()
  auth.logout()
  router.push('/login')
}
</script>
