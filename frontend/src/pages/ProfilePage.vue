<template>
  <div class="flex flex-col min-h-dvh">
    <header class="pt-safe px-4 py-3">
      <h1 class="text-lg font-bold text-text-primary">我的</h1>
    </header>

    <PullRefresh class="flex-1 min-h-0" content-class="px-4 py-4 space-y-4 pb-[calc(5rem+env(safe-area-inset-bottom))]"
      :refresh="refreshAll">
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
          <h2 class="text-sm font-semibold text-text-secondary">我的家庭</h2>
          <button v-if="family && family.members.length > 1" @click="leaveFamily" class="text-xs text-danger/80 font-medium py-2 px-3 -mr-2 flex items-center min-h-[44px]">退出家庭</button>
        </div>

        <!-- 当前家庭信息 -->
        <div v-if="family" class="space-y-3">
          <div class="bg-bg-secondary rounded-xl p-3">
            <div class="text-xs text-text-secondary mb-1">邀请码</div>
            <div class="flex items-center justify-between">
              <span class="text-lg font-bold tracking-widest text-primary-deep select-all">{{ family.invite_code }}</span>
              <button @click="copyCode" class="text-xs text-primary-deep font-medium py-2 px-3 min-h-[44px] flex items-center">复制</button>
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

          <button @click="regenerateCode" class="w-full py-3 text-sm text-primary-deep font-medium rounded-xl border border-primary/30 btn-press min-h-[44px]">
            重新生成邀请码
          </button>
        </div>

        <!-- 加入其他家庭（始终显示） -->
        <div class="border-t border-border-color pt-3">
          <p class="text-xs text-text-secondary mb-2">加入其他家庭后，你和你的宝宝数据将切换到新家庭</p>
          <div class="flex gap-2">
            <input v-model="joinCode" placeholder="输入对方的邀请码" maxlength="6"
              aria-label="邀请码"
              class="flex-1 px-3 py-2 border border-border-color rounded-xl text-sm focus:border-primary focus:outline-none transition-colors uppercase" />
            <button @click="joinFamily" class="px-4 py-3 bg-primary-deep text-white text-sm font-medium rounded-xl btn-press min-h-[44px]">加入</button>
          </div>
        </div>
      </div>

      <!-- 宝宝列表 -->
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-semibold text-text-secondary">宝宝档案</h2>
          <router-link to="/baby/new" class="text-primary-deep text-sm font-medium flex items-center gap-1 min-h-[44px]">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
            添加
          </router-link>
        </div>

        <div v-if="app.babies.length === 0" class="bg-white rounded-2xl p-6 text-center shadow-card">
          <img src="/icon-192.png" alt="" class="w-12 h-12 mx-auto block mb-2" />
          <p class="text-text-secondary text-sm">还没有宝宝档案</p>
        </div>

        <div v-for="baby in app.babies" :key="baby.id" class="bg-white rounded-2xl shadow-card">
          <div class="p-4 flex items-center gap-3 cursor-pointer btn-press rounded-2xl" role="button" tabindex="0"
            @keydown.enter.prevent="router.push(`/baby/${baby.id}/edit`)"
            @click="router.push(`/baby/${baby.id}/edit`)">
            <div class="w-12 h-12 rounded-full flex items-center justify-center text-xl font-bold text-white shrink-0" :style="{ background: baby.avatar_color }">
              {{ baby.name[0] }}
            </div>
            <div class="flex-1 min-w-0">
              <div class="font-semibold text-text-primary">{{ baby.name }}</div>
              <div class="text-xs text-text-secondary mt-0.5">{{ formatBirthDate(baby.birth_date) }}</div>
            </div>
            <svg class="w-5 h-5 text-text-secondary/50 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
          </div>
        </div>
      </div>

      <!-- 登出 -->
      <button @click="logout" class="w-full py-3 bg-white text-danger font-medium rounded-xl shadow-card btn-press mt-8 min-h-[44px]">
        退出登录
      </button>
    </PullRefresh>

    <!-- 加入 / 退出家庭确认（iOS 底部操作表） -->
    <ConfirmSheet :open="sheet.open" :title="sheet.title" :message="sheet.message"
      :confirm-text="sheet.confirmText" @confirm="onSheetConfirm" @cancel="sheetMode = ''" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { familyAPI } from '@/api'
import PullRefresh from '@/components/PullRefresh.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'

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
const sheetMode = ref<'' | 'join' | 'leave'>('')

const sheet = computed(() => ({
  open: sheetMode.value !== '',
  title: sheetMode.value === 'join' ? '加入新家庭' : '退出当前家庭',
  message: sheetMode.value === 'join'
    ? '加入新家庭后，你将退出当前家庭。\n\n你的宝宝数据会跟随你到新家庭，原家庭成员将无法看到你的数据。确定继续？'
    : '退出后，你将无法查看当前家庭的宝宝数据。',
  confirmText: sheetMode.value === 'join' ? '加入' : '退出家庭',
}))

async function loadFamily() {
  try {
    const res = await familyAPI.getMyFamily()
    family.value = { ...res.data.family, members: res.data.members }
  } catch {
    // family not available
  }
}

async function refreshAll() {
  await Promise.all([loadFamily(), app.loadBabies()])
}

function joinFamily() {
  if (!joinCode.value.trim()) return
  // 如果用户已有家庭，先二次确认
  if (family.value && family.value.members.length > 1) {
    sheetMode.value = 'join'
    return
  }
  void doJoin()
}

async function doJoin() {
  const code = joinCode.value.trim().toUpperCase()
  try {
    await familyAPI.join(code)
    joinCode.value = ''
    await loadFamily()
    await app.loadBabies()
    app.showToast('已加入家庭', 'success')
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '加入失败', 'error')
  } finally {
    sheetMode.value = ''
  }
}

function leaveFamily() {
  sheetMode.value = 'leave'
}

async function doLeave() {
  try {
    await familyAPI.leave()
    family.value = null
    await app.loadBabies()
    app.showToast('已退出家庭', 'success')
  } catch (e: any) {
    app.showToast(e.response?.data?.error || '退出失败', 'error')
  } finally {
    sheetMode.value = ''
  }
}

function onSheetConfirm() {
  if (sheetMode.value === 'join') void doJoin()
  else void doLeave()
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
  const text = family.value.invite_code
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(text).then(() => {
      app.showToast('已复制邀请码', 'success')
    }).catch(fallbackCopy)
  } else {
    fallbackCopy()
  }
  function fallbackCopy() {
    const el = document.createElement('textarea')
    el.value = text
    el.style.position = 'fixed'
    el.style.opacity = '0'
    document.body.appendChild(el)
    el.select()
    try {
      document.execCommand('copy')
      app.showToast('已复制邀请码', 'success')
    } catch {
      app.showToast('复制失败，请手动复制', 'error')
    }
    document.body.removeChild(el)
  }
}

function formatBirthDate(bd: string) {
  if (!bd) return ''
  const d = new Date(bd)
  if (isNaN(d.getTime())) return bd
  const p2 = (n: number) => String(n).padStart(2, '0')
  const base = `${d.getFullYear()}-${p2(d.getMonth() + 1)}-${p2(d.getDate())}`
  const h = d.getHours()
  const mi = d.getMinutes()
  if (h || mi) return `${base} ${p2(h)}:${p2(mi)}`
  return base
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
