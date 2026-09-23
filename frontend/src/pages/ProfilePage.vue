<template>
  <div class="flex flex-col h-dvh">
    <PullRefresh class="flex-1 min-h-0" content-class="px-4 py-4 space-y-4 pb-[calc(6.5rem+env(safe-area-inset-bottom))]"
      :refresh="refreshAll" @scroll="navScroll = $event">
    <template #header>
    <LargeTitleNav title="我的" :scroll-top="navScroll" />
    </template>

      <!-- 用户信息 -->
      <div class="bg-surface rounded-2xl p-4 shadow-card flex items-center gap-4">
        <div class="w-14 h-14 rounded-full bg-primary/10 flex items-center justify-center text-2xl">👤</div>
        <div>
          <div class="font-semibold text-text-primary">{{ auth.user?.username }}</div>
          <div class="text-sm text-text-secondary mt-0.5">家庭成员</div>
        </div>
      </div>

      <!-- 家庭信息 -->
      <div class="bg-surface rounded-2xl p-4 shadow-card space-y-3">
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
      </div>

      <!-- 加入其他家庭（独立卡片） -->
      <div class="bg-surface rounded-2xl p-4 shadow-card space-y-2">
        <h2 class="text-sm font-semibold text-text-secondary">加入其他家庭</h2>
        <p class="text-xs text-text-secondary">加入其他家庭后，你和你的宝宝数据将切换到新家庭</p>
        <div class="flex gap-2 pt-1">
          <input v-model="joinCode" placeholder="输入对方的邀请码" maxlength="6"
            aria-label="邀请码" inputmode="text" autocapitalize="characters" autocomplete="off" enterkeyhint="done"
            class="flex-1 min-h-[44px] px-3 py-2.5 bg-bg-secondary border border-border-color rounded-xl text-base focus:border-primary focus:outline-none transition-colors uppercase" />
          <button @click="joinFamily" :disabled="!joinCode.trim()"
            class="px-4 py-3 bg-primary-fill text-white text-sm font-medium rounded-xl btn-press min-h-[44px] disabled:opacity-40 disabled:cursor-not-allowed">加入</button>
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

        <div v-if="app.babies.length === 0" class="bg-surface rounded-2xl p-6 text-center shadow-card">
          <EmptyState title="还没有宝宝档案" size="sm" icon="folder" />
        </div>

        <div v-for="baby in app.babies" :key="baby.id" class="bg-surface rounded-2xl shadow-card">
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

      <!-- 成长记录 -->
      <router-link to="/growth" class="bg-surface rounded-2xl shadow-card overflow-hidden block">
        <div class="px-4 py-3.5 flex items-center gap-3 min-h-[44px] btn-press">
          <span class="w-9 h-9 rounded-full bg-sleep/15 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-sleep-deep" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M4 20V10M10 20V4M16 20v-8M4 20h16" />
            </svg>
          </span>
          <div class="flex-1 min-w-0">
            <div class="font-medium text-text-primary">成长记录</div>
            <div class="text-xs text-text-secondary mt-0.5">身高 · 体重 · 头围与生长标准百分位</div>
          </div>
          <svg class="w-5 h-5 text-text-secondary/50 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
        </div>
      </router-link>

      <!-- 提醒 -->
      <RemindersCard />

      <!-- 数据导出 -->
      <div class="bg-surface rounded-2xl shadow-card overflow-hidden">
        <div class="px-4 pt-3 pb-1">
          <h2 class="text-sm font-semibold text-text-secondary">数据</h2>
        </div>
        <button type="button" @click="exportData" :disabled="exporting || app.babies.length === 0"
          class="w-full px-4 py-3.5 flex items-center gap-3 border-t border-border-color/60 min-h-[44px] text-left btn-press disabled:opacity-40">
          <span class="w-9 h-9 rounded-full bg-primary/10 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-primary-deep" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M12 3v12m0 0l-4-4m4 4l4-4M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2" />
            </svg>
          </span>
          <div class="flex-1 min-w-0">
            <div class="font-medium text-text-primary">导出全部记录（CSV）</div>
            <div class="text-xs text-text-secondary mt-0.5">Excel / 医生可直接打开，用于就诊或备份</div>
          </div>
          <ActivityIndicator v-if="exporting" :size="18" class="text-text-secondary" />
        </button>
      </div>

      <!-- 偏好设置 -->
      <div class="bg-surface rounded-2xl shadow-card overflow-hidden">
        <div class="px-4 pt-3 pb-1">
          <h2 class="text-sm font-semibold text-text-secondary">偏好设置</h2>
        </div>
        <button type="button" @click="toggleHaptics"
          class="w-full px-4 py-3.5 flex items-center justify-between border-t border-border-color/60 min-h-[44px] text-left btn-press">
          <span class="flex items-center gap-2 text-text-primary">
            <svg class="w-5 h-5 text-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h6m-6 4h10M5 3h14a2 2 0 012 2v14a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2z" />
            </svg>
            触感反馈
          </span>
          <span :class="['relative w-[51px] h-[31px] rounded-full transition-colors duration-200 shrink-0', hapticsOn ? 'bg-success' : 'bg-border-color']">
            <span :class="['absolute top-[2px] w-[27px] h-[27px] rounded-full bg-white shadow transition-transform duration-200', hapticsOn ? 'translate-x-[22px]' : 'translate-x-[2px]']"></span>
          </span>
        </button>
      </div>

      <!-- 关于 -->
      <div class="bg-surface rounded-2xl shadow-card overflow-hidden">
        <div class="px-4 pt-3 pb-1">
          <h2 class="text-sm font-semibold text-text-secondary">关于</h2>
        </div>
        <div class="px-4 py-3.5 border-t border-border-color/60 flex items-center justify-between min-h-[44px]">
          <span class="text-text-primary">版本</span>
          <span class="text-sm text-text-secondary">v{{ appVersion }}</span>
        </div>
        <div class="px-4 py-3.5 border-t border-border-color/60 text-xs text-text-secondary leading-relaxed">
          宝宝护理记录 · 本地优先的多照护者同步记录工具。所有数据仅存储在自建服务端。
        </div>
      </div>

      <!-- 登出 -->
      <button @click="logout" class="w-full py-3 bg-surface text-danger font-medium rounded-xl shadow-card btn-press mt-2 min-h-[44px]">
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
import { familyAPI, babyAPI } from '@/api'
import PullRefresh from '@/components/PullRefresh.vue'
import ConfirmSheet from '@/components/ConfirmSheet.vue'
import EmptyState from '@/components/EmptyState.vue'
import LargeTitleNav from '@/components/LargeTitleNav.vue'
import RemindersCard from '@/components/RemindersCard.vue'
import ActivityIndicator from '@/components/ActivityIndicator.vue'
import { isHapticsEnabled, setHapticsEnabled, hapticSelection } from '@/utils/haptic'
import { parseLocalDate } from '@/utils'

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
const navScroll = ref(0)
const hapticsOn = ref(isHapticsEnabled())
const appVersion = '1.0.0'
const exporting = ref(false)

async function exportData() {
  const baby = app.currentBaby
  if (!baby) { app.showToast('请先添加宝宝', 'error'); return }
  exporting.value = true
  try {
    const res = await babyAPI.exportRecords(baby.id)
    const blob = new Blob([res.data], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    const d = new Date()
    const p2 = (n: number) => String(n).padStart(2, '0')
    a.href = url
    a.download = `${baby.name}-${d.getFullYear()}${p2(d.getMonth() + 1)}${p2(d.getDate())}.csv`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    setTimeout(() => URL.revokeObjectURL(url), 1000)
    app.showToast('已导出', 'success')
  } catch {
    app.showToast('导出失败', 'error')
  } finally {
    exporting.value = false
  }
}

function toggleHaptics() {
  hapticsOn.value = !hapticsOn.value
  setHapticsEnabled(hapticsOn.value)
  if (hapticsOn.value) hapticSelection()
}

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
  const d = parseLocalDate(bd)
  if (!d) return bd
  const p2 = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p2(d.getMonth() + 1)}-${p2(d.getDate())}`
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
